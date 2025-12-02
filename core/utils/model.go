package utils

import (
	"bytes"
	"encoding/gob"
)

type BoundStruct[T any] struct {
	table Table
	key   string
	V     *T
}

func Bind[T any](table Table, key string, value *T) *BoundStruct[T] {
	return &BoundStruct[T]{table: table, key: key, V: value}
}

func (b *BoundStruct[T]) Load() error {
	data, err := b.table.Get(b.key)
	if err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(b.V)
}

func (b *BoundStruct[T]) Save() error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(b.V); err != nil {
		return err
	}
	return b.table.Put(b.key, buf.Bytes())
}

func (b *BoundStruct[T]) Delete() error {
	return b.table.Del(b.key)
}
