package kvdatabase

// KvDatabase KV数据库接口
type KvDatabase interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	Close() error
}
