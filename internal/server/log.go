package server

import (
	"fmt"
	"sync"
)

var ErrOffsetNotFound = fmt.Errorf("offset not found")

type Log struct {
	mu      *sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{mu: &sync.Mutex{}, records: make([]Record, 0)}
}

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

// pointer reciever
func (l *Log) Append(r Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	offset := len(l.records)
	l.records = append(l.records, r)

	return uint64(offset), nil
}

func (c *Log) Read(offset uint64) (Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if offset >= uint64(len(c.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return c.records[offset], nil
}
