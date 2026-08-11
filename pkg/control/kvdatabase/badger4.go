package kvdatabase

import (
	"fmt"
	"go.uber.org/zap"

	"github.com/dgraph-io/badger/v4"
)

const (
	badger4Database = "badger4"
)

// Badger Badger数据库实现
type Badger struct {
	db     *badger.DB
	logger *zap.SugaredLogger
}

// NewBadger4 创建Badger实例
// @param path string 数据库路径
// @return *Badger Badger实例
// @return error 创建过程中的错误
func NewBadger4(path string, logger *zap.SugaredLogger) (*Badger, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open badger at %s: %v", path, err)
	}

	return &Badger{
		db:     db,
		logger: logger,
	}, nil
}

// Get 获取键对应的值
// @description 获取键对应的值
// @param key string 键
// @return []byte 键对应的值
// @return error 获取过程中的错误
func (b *Badger) Get(key string) ([]byte, error) {
	var value []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			value = append([]byte{}, val...)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Set 设置键值对
// @description 设置键值对
// @param key string 键
// @param value []byte 值
// @return error 设置过程中的错误
func (b *Badger) Set(key string, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// Delete 删除键
// @description 删除指定的键
// @param key string 要删除的键
// @return error 删除过程中的错误
func (b *Badger) Delete(key string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Close 关闭数据库连接
// @description 关闭数据库连接并释放资源
// @return error 关闭过程中的错误
func (b *Badger) Close() error {
	return b.db.Close()
}
