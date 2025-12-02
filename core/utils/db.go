package utils

type NewDb func(name string) Db

type Db interface {
	NewTable(name string) Table
}

type Table interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Del(key string) error
	Map() map[string][]byte
}
