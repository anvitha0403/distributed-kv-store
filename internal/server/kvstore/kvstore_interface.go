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
	Put(K, V string) (string, error)

	// Del removes the given key and its associated value from the storage engine.
	// Returns error if operation fails.
	Del(K string) error

	Close() error
}
