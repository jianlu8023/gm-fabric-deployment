package kvdatabase

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	levelDBDatabase = "leveldb"
)

// LevelDB LevelDB数据库实现
type LevelDB struct {
	db *leveldb.DB
}

// NewLevelDB 创建LevelDB实例
// @param path string 数据库路径
// @return *LevelDB LevelDB实例
// @return error 创建过程中的错误
func NewLevelDB(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(path, &opt.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb at %s: %v", path, err)
	}

	return &LevelDB{
		db: db,
	}, nil
}

// Get 获取键对应的值
// @description 获取键对应的值
// @param key string 键
// @return []byte 键对应的值
// @return error 获取过程中的错误
func (l *LevelDB) Get(key string) ([]byte, error) {
	return l.db.Get([]byte(key), nil)
}

// Set 设置键值对
// @description 设置键值对
// @param key string 键
// @param value []byte 值
// @return error 设置过程中的错误
func (l *LevelDB) Set(key string, value []byte) error {
	return l.db.Put([]byte(key), value, nil)
}

// Delete 删除键
// @description 删除指定的键
// @param key string 要删除的键
// @return error 删除过程中的错误
func (l *LevelDB) Delete(key string) error {
	return l.db.Delete([]byte(key), nil)
}

// Close 关闭数据库连接
// @description 关闭数据库连接并释放资源
// @return error 关闭过程中的错误
func (l *LevelDB) Close() error {
	return l.db.Close()
}
