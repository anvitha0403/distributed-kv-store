package kvstore

type hashIndex struct {
	index   [][]keyOffset
	maxHash int
}

type keyOffset struct {
	Key    string
	Offset int64
}

// NewHashIndex creates a hashIndex with the given number of buckets (maxHash).
func NewHashIndex(maxHash int) *hashIndex {

	return &hashIndex{
		index:   make([][]keyOffset, maxHash),
		maxHash: maxHash,
	}
}

// Insert adds a key and its corresponding offset to the hash index.
// If the key already exists, its offset is updated.
// The key is hashed to determine its position in the index.
// Collisions are handled by storing multiple key-offset pairs in a slice at each position.
func (hi *hashIndex) Insert(key string, offset int64) string {
	hashOfKey := hash(key)
	pos := hashOfKey % int64(hi.maxHash)
	if hi.index[pos] == nil {
		hi.index[pos] = []keyOffset{}
	}
	// Check if the key already exists and update the offset if needed
	for i, ko := range hi.index[pos] {
		if ko.Key == key {
			hi.index[pos][i].Offset = offset
			return OPERATION_PUT_UPDATE
		}
	}
	hi.index[pos] = append(hi.index[pos], keyOffset{Key: key, Offset: offset})

	return OPERATION_PUT_CREATE
}

// GetOffset retrieves the offset associated with the given key from the hash index.
// It computes the hash of the key, determines its position in the index, and searches
// for the key in the corresponding bucket. If the key is found, its offset is returned.
// If the key is not found, it returns -1 and an ErrorKeyNotFound error.
//
// Parameters:
//
//	key string - the key to look up in the hash index.
//
// Returns:
//
//	int64 - the offset associated with the key, or -1 if not found.
//	error - an error if the key is not found.
func (hi *hashIndex) GetOffset(key string) (int64, error) {
	hashOfKey := hash(key)
	pos := hashOfKey % int64(hi.maxHash)
	if hi.index[pos] == nil {
		return -1, ErrKeyDoesntExist
	}
	for _, ko := range hi.index[pos] {
		if ko.Key == key {
			return ko.Offset, nil
		}
	}
	return -1, ErrKeyDoesntExist
}

// Delete removes the entry associated with the given key from the hash index.
// If the key does not exist in the index, then its a no-op.
// This operation is safe to call even if the key is not present.
func (hi *hashIndex) Delete(key string) error {
	hashOfKey := hash(key)
	pos := hashOfKey % int64(hi.maxHash)
	bucket := hi.index[pos]
	if bucket == nil {
		return ErrKeyDoesntExist
	}

	for i, ko := range bucket {
		if ko.Key == key {
			bucket[i] = bucket[len(bucket)-1]
			hi.index[pos] = bucket[:len(bucket)-1]
			
			return nil
		}
	}
	return ErrKeyDoesntExist
}

// LoadFromFile rebuilds the index by replaying records from file (from offset 0).
// PUT -> Insert(key, offset); DEL -> Delete(key). Returns any iterator error.
func (hi *hashIndex) LoadFromFile(file *DataFile) error {
	iterator, err := file.GetIterator(0)
	if err != nil {
		return err
	}

	for iterator.HasNext() {
		record, startingOffset, err := iterator.Get()
		if err != nil {
			return err
		}
		switch record.Operation {
		case OPERATION_PUT:
			hi.Insert(record.GetData().GetKey(), startingOffset)
		case OPERATION_DEL:
			hi.Delete(record.GetData().GetKey())
		}
	}
	return nil
}

// hash computes a simple hash value for the given string key by summing the ASCII values
// of each character in the key. It returns the resulting sum as an int64.
func hash(key string) int64 {
	var hash int64 = 0
	for i := 0; i < len(key); i++ {
		hash += int64(key[i])
	}
	return hash
}
