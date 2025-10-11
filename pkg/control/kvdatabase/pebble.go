package kvdatabase

import (
	"fmt"

	"github.com/cockroachdb/pebble"
)

const (
	pebbleDatabase = "pebble"
)

// Pebble Pebble数据库实现
type Pebble struct {
	db *pebble.DB
}

// NewPebble 创建Pebble实例
// @param path string 数据库路径
// @return *Pebble Pebble实例
// @return error 创建过程中的错误
func NewPebble(path string) (*Pebble, error) {
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to open pebble at %s: %v", path, err)
	}

	return &Pebble{
		db: db,
	}, nil
}

// Get 获取键对应的值
// @description 获取键对应的值
// @param key string 键
// @return []byte 键对应的值
// @return error 获取过程中的错误
func (p *Pebble) Get(key string) ([]byte, error) {
	get, closer, err := p.db.Get([]byte(key))
	defer closer.Close()
	return get, err
}

// Set 设置键值对
// @description 设置键值对
// @param key string 键
// @param value []byte 值
// @return error 设置过程中的错误
func (p *Pebble) Set(key string, value []byte) error {
	return p.db.Set([]byte(key), value, pebble.NoSync)
}

// Delete 删除键
// @description 删除指定的键
// @param key string 要删除的键
// @return error 删除过程中的错误
func (p *Pebble) Delete(key string) error {
	return p.db.Delete([]byte(key), pebble.NoSync)
}

// Close 关闭数据库连接
// @description 关闭数据库连接并释放资源
// @return error 关闭过程中的错误
func (p *Pebble) Close() error {
	return p.db.Close()
}
