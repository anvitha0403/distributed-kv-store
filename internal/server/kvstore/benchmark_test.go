package kvstore

import (
	"fmt"
	"testing"
)

func BenchmarkGet1MKeys(b *testing.B) {

	inputData := []struct {
		input int
	}{
		{input: 10000},
		{input: 100000},
		{input: 1000000},
	}
	for _, data := range inputData {
		b.Run(fmt.Sprintf("Get-%d", data.input), func(b *testing.B) {
			tmpDir := b.TempDir()
			store, err := ConnectFileStore(tmpDir)
			if err != nil {
				b.Fatalf("ConnectFileStore failed: %v", err)
			}
			defer store.Close()

			if err := addNItemsToKVStore(store, data.input); err != nil {
				b.Fatalf("addNItemsToKVStore failed: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := getNthItemFromKVStore(store, i%data.input); err != nil {
					b.Errorf("Get failed: %v", err)
				}
			}
		})
	}
}

func addNItemsToKVStore(store Store, N int) error {
	for i := 0; i < N; i++ {
		key := fmt.Sprintf("key-%d", i)
		val := fmt.Sprintf("value-%d", i)
		if err := store.Put(key, val); err != nil {
			return fmt.Errorf("failed to put key %s: %w", key, err)
		}
	}
	return nil
}

func getNthItemFromKVStore(store Store, N int) (string, error) {
	key := fmt.Sprintf("key-%d", N)
	val, err := store.Get(key)
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}
