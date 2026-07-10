package kvstore

import (
	"errors"
	"os"
	"path/filepath"

	pb "github.com/anvitha0403/golog/internal/server/api/v1"
)

const (
	OPERATION_PUT = "PUT"
	OPERATION_DEL = "DEL"
)
const (
	OPERATION_PUT_UPDATE = "UPDATE"
	OPERATION_PUT_CREATE = "CREATE"
)

var (
	ErrKeyDoesntExist = errors.New("given key doesn't exist")
)

type FileStore struct {
	dbFile *DataFile
	index  *hashIndex
}

// ConnectFileStore initializes and returns a new FileStore instance at the specified file path.
// It ensures that the directory for the file exists, creating it if necessary.
// If the directory cannot be created or the data file cannot be opened, an error is returned.
// On success, it returns a Store backed by the file at the given path.
func ConnectFileStore(path string) (Store, error) {

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	file, err := NewDataFile(path)
	if err != nil {
		return nil, err
	}

	hashIndex := NewHashIndex(1000000)
	err = hashIndex.LoadFromFile(file)
	if err != nil {
		return nil, err
	}

	return &FileStore{
		dbFile: file,
		index:  hashIndex,
	}, nil
}

// Put stores the given key-value pair in the file store.
// It appends a new record with the specified key and value to the underlying database file.
// Returns an error if writing or flushing the record fails.
func (f *FileStore) Put(K, V string) (string, error) {

	dataToAppend := pb.Record{
		Operation: OPERATION_PUT,
		Data:      &pb.KVPair{Key: K, Value: V},
	}

	startingOffset, err := f.dbFile.Append(&dataToAppend)
	if err != nil {
		return "", err
	}
	return f.index.Insert(K, startingOffset), nil

}

// Get returns the value for key K or an error if not found.
func (f *FileStore) Get(K string) (string, error) {
	offset, err := f.index.GetOffset(K)
	if err != nil {
		return "", err
	}
	recordRead, err := f.dbFile.ReadRecordAt(offset)
	if err != nil {
		return "", err
	}
	return recordRead.GetData().GetValue(), nil
}

// Del deletes the key-value pair associated with the given key K from the file store.
// It appends a delete operation record to the underlying database file and flushes the changes.
// Returns an error if writing or flushing the record fails.
func (f *FileStore) Del(K string) error {
	dataToAppend := pb.Record{
		Operation: OPERATION_DEL,
		Data:      &pb.KVPair{Key: K},
	}
	_, err := f.dbFile.Append(&dataToAppend)
	if err != nil {
		return err
	}
	//delete from index as-well
	if f.index.Delete(K) != nil {
		return err

	}

	return nil
}

// Close closes the underlying database file associated with the FileStore.
// It returns an error if the file cannot be closed.
func (f *FileStore) Close() error {
	return f.dbFile.Close()
}
