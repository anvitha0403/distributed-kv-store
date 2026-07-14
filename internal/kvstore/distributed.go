package kvstore

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	api "github.com/anvitha0403/golog/internal/server/api/v1"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/protobuf/proto"
)

type DistributedStore struct {
	config *Config
	store  *FileStore
	raft   *raft.Raft
}

func NewDistributedStore(dataDir string, config *Config) (*DistributedStore, error) {

	db, err := ConnectFileStore(dataDir)
	if err != nil {
		return nil, err
	}
	l := &DistributedStore{
		store:  db,
		config: config,
	}

	if err := l.setupRaft(dataDir); err != nil {
		return nil, err
	}
	return l, nil
}

func (db *DistributedStore) setupRaft(dataDir string) error {
	fsm := &fsm{db: db.store}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Join(dataDir, "raft"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "raft", "snapshots"), 0755); err != nil {
		return err
	}

	// Log store
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft", "log.db"))
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}

	// Stable store
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft", "stable.db"))
	if err != nil {
		return fmt.Errorf("failed to create stable store: %w", err)
	}

	// Snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(filepath.Join(dataDir, "raft", "snapshots"), 1, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create snapshot store: %w", err)
	}

	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
		db.config.Raft.StreamLayer,
		maxPool,
		timeout,
		os.Stderr,
	)

	config := raft.DefaultConfig()

	config.LocalID = db.config.Raft.LocalID
	if db.config.Raft.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = db.config.Raft.HeartbeatTimeout
	}
	if db.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = db.config.Raft.ElectionTimeout
	}
	if db.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = db.config.Raft.LeaderLeaseTimeout
	}
	if db.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = db.config.Raft.CommitTimeout
	}
	config.HeartbeatTimeout = 5 * time.Second
	config.ElectionTimeout = 10 * time.Second
	config.LeaderLeaseTimeout = 5 * time.Second
	config.CommitTimeout = 1 * time.Second

	db.raft, err = raft.NewRaft(
		config,
		fsm,
		logStore,
		stableStore,
		snapshotStore,
		transport,
	)
	if err != nil {
		return err
	}

	if db.config.Raft.Bootstrap {
		config := raft.Configuration{
			Servers: []raft.Server{{
				ID:      config.LocalID,
				Address: raft.ServerAddress(db.config.Raft.BindAddr),
			}},
		}
		db.raft.BootstrapCluster(config).Error()

	}

	return nil
}

func (f *DistributedStore) Del(K string) error {
	_, err := f.apply(&api.Record{Operation: OPERATION_DEL, Data: &api.KVPair{Key: K}})
	return err

}
func (f *DistributedStore) Get(K string) (string, error) {
	return f.store.Get(K)
}
func (f *DistributedStore) Put(K string, V string) (string, error) {
	res, err := f.apply(&api.Record{Operation: OPERATION_PUT, Data: &api.KVPair{Key: K}})
	return res.(string), err
}

func (db *DistributedStore) apply(req *api.Record) (
	interface{},
	error,
) {
	var buf bytes.Buffer

	// Serialize
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}
	timeout := 10 * time.Second
	future := db.raft.Apply(buf.Bytes(), timeout)
	if future.Error() != nil {
		return nil, future.Error()
	}
	res := future.Response()
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res, nil
}

// func (db *DistributedStore) Read(offset uint64) (*api.KVPair, error) {
// 	return db.store.dbFile.ReadRecordAt(int64(offset))
// }

func (db *DistributedStore) Join(id, addr string) error {
	configFuture := db.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}
	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == serverID || srv.Address == serverAddr {
			if srv.ID == serverID && srv.Address == serverAddr {
				// server has already joined
				return nil
			}
			// remove the existing server
			removeFuture := db.raft.RemoveServer(serverID, 0, 0)
			if err := removeFuture.Error(); err != nil {
				return err
			}
		}
	}
	addFuture := db.raft.AddVoter(serverID, serverAddr, 0, 0)
	if err := addFuture.Error(); err != nil {
		return err
	}
	return nil
}

func (db *DistributedStore) Leave(id string) error {
	removeFuture := db.raft.RemoveServer(raft.ServerID(id), 0, 0)
	return removeFuture.Error()
}

func (db *DistributedStore) WaitForLeader(timeout time.Duration) error {
	timeoutc := time.After(timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeoutc:
			return fmt.Errorf("timed out")
		case <-ticker.C:
			if l := db.raft.Leader(); l != "" {
				return nil
			}
		}
	}
}

func (db *DistributedStore) Close() error {
	f := db.raft.Shutdown()
	if err := f.Error(); err != nil {
		return err
	}
	return db.store.Close()
}

// START: get_servers
func (db *DistributedStore) GetServers() ([]*api.Server, error) {
	future := db.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}
	var servers []*api.Server
	for _, server := range future.Configuration().Servers {
		servers = append(servers, &api.Server{
			Id:       string(server.ID),
			RpcAddr:  string(server.Address),
			IsLeader: db.raft.Leader() == server.Address,
		})
	}
	return servers, nil
}

// // END: get_servers

var _ raft.FSM = (*fsm)(nil)

type fsm struct {
	db *FileStore
}

func (l *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	var req api.Record
	err := proto.Unmarshal(buf, &req)
	if err != nil {
		return err
	}
	switch req.Operation {
	case OPERATION_PUT:
		op, err := l.db.Put(req.Data.Key, req.Data.Value)
		if err != nil {
			return err
		}
		return op
	case OPERATION_DEL:
		return l.db.Del(req.Data.Key)

	}
	return nil
}

var _ raft.FSMSnapshot = (*snapshot)(nil)

type snapshot struct {
	reader io.Reader
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {}

func (fsm *fsm) Snapshot() (raft.FSMSnapshot, error) {

	buf := new(bytes.Buffer)
	fw := DataFileWriter{writer: bufio.NewWriter(buf)}

	for _, bucket := range fsm.db.index.index {
		for _, key_offset := range bucket {
			recordRead, err := fsm.db.dbFile.ReadRecordAt(key_offset.Offset) // read from file
			if err != nil {
				return nil, err
			}
			fw.Append(recordRead)

		}
	}
	fw.writer.Flush()

	return &snapshot{reader: buf}, nil
}

func (f *fsm) Restore(r io.ReadCloser) error {
	if err := f.db.Reset(); err != nil {
		return err
	}

	// Open file for writing snapshot contents
	file, err := os.Create(f.db.dbFile.fullpath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy snapshot stream directly into file
	if _, err := io.Copy(file, r); err != nil {
		return err
	}
	f.db.dbFile.file = file
	f.db.index.LoadFromFile(f.db.dbFile)

	return nil
}

var _ raft.StreamLayer = (*StreamLayer)(nil)

type StreamLayer struct {
	ln              net.Listener
	serverTLSConfig *tls.Config
	peerTLSConfig   *tls.Config
}

func NewStreamLayer(
	ln net.Listener,
	serverTLSConfig,
	peerTLSConfig *tls.Config,
) *StreamLayer {
	return &StreamLayer{
		ln:              ln,
		serverTLSConfig: serverTLSConfig,
		peerTLSConfig:   peerTLSConfig,
	}
}

const RaftRPC = 1

func (s *StreamLayer) Dial(
	addr raft.ServerAddress,
	timeout time.Duration,
) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	var conn, err = dialer.Dial("tcp", string(addr))
	if err != nil {
		return nil, err
	}
	// identify to mux this is a raft rpc
	_, err = conn.Write([]byte{byte(RaftRPC)})
	if err != nil {
		return nil, err
	}
	if s.peerTLSConfig != nil {
		conn = tls.Client(conn, s.peerTLSConfig)
	}
	return conn, err
}

func (s *StreamLayer) Accept() (net.Conn, error) {
	conn, err := s.ln.Accept()
	if err != nil {
		return nil, err
	}
	b := make([]byte, 1)
	_, err = conn.Read(b)
	if err != nil {
		return nil, err
	}
	if bytes.Compare([]byte{byte(RaftRPC)}, b) != 0 {
		return nil, fmt.Errorf("not a raft rpc")
	}
	if s.serverTLSConfig != nil {
		return tls.Server(conn, s.serverTLSConfig), nil
	}
	return conn, nil
}

func (s *StreamLayer) Close() error {
	return s.ln.Close()
}

func (s *StreamLayer) Addr() net.Addr {
	return s.ln.Addr()
}

// type (
//  KVState map[string]string // THE 'STATE'!

//  ConsensusValidatorFunc func(k, v string) error
// )

// type fsm struct {
//  mux *sync.Mutex // mutex hat
//  state *KVState // link to a curren state
//  prevState KVState // previous state store for proper disaster recovery

//  validators []ConsensusValidatorFunc // custom validators
// }

// func newFSM(validators ...ConsensusValidatorFunc) *fsm {
//  // initializing state to avoid nil state errors
//  // 'genesis' value here is only for observability and not required
//  state := KVState{"genesis": "genesis"}
//  return &fsm{
//    state: &state,
//    prevState: KVState{},
//    mux: new(sync.Mutex),
//    validators: validators,
//  }
// }

// // Apply is invoked by Raft once a log entry is commited. Do not use directly.
// func (fsm *fsm) Apply(rlog *raft.Log) (result interface{}) {
//  fsm.mux.Lock()
//  defer fsm.mux.Unlock()
//  defer func() {
//  // attempt to save state after panic
//    if r := recover(); r != nil {
//      *fsm.state = fsm.prevState
//      result = errors.New("fsm apply panic: rollback")
//    }
//  }()
//  var newState = make(KVState, 1)
//  if err := json.Unmarshal(rlog.Data, &newState); err != nil {
//    return fmt.Errorf("failed to decode log: %w", err)
//  }

//  for _, validator := range fsm.validators {
//    // apply validator on each key/value
//    for k, v := range newState {
//      if err := validator(k, v); err != nil {
//        return err
//      }
//    }
//  // applied validators doesn't stop method Apply from storing data
//  }

//  // save previous state
//  fsm.prevState = make(KVState, len(*fsm.state))
//  for k, v := range *fsm.state {
//    fsm.prevState[k] = v
//  }
//  // save new state
//  for k, v := range newState {
//    (*fsm.state)[k] = v
//  }
//  return fsm.state
// }

// // Snapshot encodes the current state so that we can save a snapshot.
// func (fsm *fsm) Snapshot() (raft.FSMSnapshot, error) {
//  fsm.mux.Lock()
//  defer fsm.mux.Unlock()

//  buf := new(bytes.Buffer)
//  err := json.NewEncoder(buf).Encode(fsm.state)
//  if err != nil {
//    return nil, err
//  }

//  return &fsmSnapshot{state: buf}, nil
// }

// // Restore takes a snapshot and sets the current state from it.
// func (fsm *fsm) Restore(reader io.ReadCloser) (err error) {
//  defer func() {
//    err = reader.Close()
//  }()

//  fsm.mux.Lock()
//  defer fsm.mux.Unlock()

//  err = json.NewDecoder(reader).Decode(fsm.state)
//  if err != nil {
//    return err
//  }

//  // here previous state might be handy
//  fsm.prevState = make(map[string]string, len(*fsm.state))
//  return nil
// }

// // just a buffer for snapshot data
// type fsmSnapshot struct {
//  state *bytes.Buffer
// }
