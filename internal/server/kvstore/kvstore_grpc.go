package kvstore

import (
	"context"

	pb "github.com/anvitha0403/golog/internal/server/api/v1"
)

// Store defines the interface for a simple Key-Value storage engine.
//
// It provides basic functions to retrieve, insert and delete data by key.
type Store interface {

	// Get retrieves a value associated with the given key.
	// If the key doesn't exists then ErrKeyDoesntExist is thrown.
	Get(K string) (string, error)

	// Put inserts or updates the value of the given key.
	// Returns an error if operation fails.
	Put(K, V string) (string, error)

	// Del removes the given key and its associated value from the storage engine.
	// Returns error if operation fails.
	Del(K string) error

	Close() error
}

type KvStore struct {
	store Store
	idx   *hashIndex
	pb.UnimplementedKVStoreServiceServer
}

func New(dataDir string) (*KvStore, error) {
	store, err := ConnectFileStore(dataDir)
	if err != nil {
		return nil, err
	}
	return &KvStore{
		store: store,
	}, nil

}

func (s *KvStore) Put(ctx context.Context, pair *pb.KVPair) (*pb.Response, error) {

	op, err := s.store.Put(pair.Key, pair.Value)
	if err != nil {

		return nil, err
	}

	return &pb.Response{Value: op}, nil
}

func (s *KvStore) Get(ctx context.Context, pair *pb.KVPair) (*pb.Response, error) {

	value, err := s.store.Get(pair.Key)
	if err != nil {
		return nil, err
	}
	return &pb.Response{Value: value}, nil
}

func (s *KvStore) Delete(ctx context.Context, pair *pb.KVPair) (*pb.Empty, error) {

	err := s.store.Del(pair.Key)
	if err != nil {
		if err == ErrKeyDoesntExist {
			return nil, &pb.ErrKeyDoesntExist{}
		}
		return nil, err
	}

	return &pb.Empty{}, nil
}
