package kvstore

// Store defines the interface for a simple Key-Value storage engine.
//
// It provides basic functions to retrieve, insert and delete data by key.
type Store interface {

	// Get retrieves a value associated with the given key.
	// If the key doesn't exists then ErrKeyDoesntExist is thrown.
	Get(K string) (string, error)

	// Put inserts or updates the value of the given key.
	// Returns an error if operation fails.
	Put(K, V string) error

	// Del removes the given key and its associated value from the storage engine.
	// Returns error if operation fails.
	Del(K string) error

	Close() error
}

type kvStore struct {
	store Store
}

func New() (Store, error) {
	store, err := ConnectFileStore("./data/")
	if err != nil {
		return nil, err
	}
	return kvStore{
		store: store,
	}, nil

}

func (db kvStore) Put(K, V string) error {
	return db.store.Put(K, V)
}

func (db kvStore) Del(K string) error {
	return db.store.Del(K)
}
func (db kvStore) Get(K string) (string, error) {
	return db.store.Get(K)
}
func (db kvStore) Close() error {
	return db.store.Close()
}
