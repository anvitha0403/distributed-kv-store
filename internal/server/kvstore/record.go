package kvstore

import (
	"fmt"
	"strings"
)

type record struct {
	operation string
	data      KVPair
}

type KVPair struct {
	key, val string
}

func (r *record) String() string {
	return fmt.Sprintf("%s|%s|%s", r.operation, r.data.key, r.data.val)
}

func (r *record) FromString(data string) {
	parts := strings.SplitN(data, "|", 3)
	r.operation = parts[0]
	r.data.key = parts[1]
	r.data.val = parts[2]
}

func (r *record) GetKey() string {
	return r.data.key
}

func (r *record) GetValue() string {
	return r.data.val
}
