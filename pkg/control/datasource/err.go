package datasource

import (
	"errors"
)

var (
	// ErrUnknownDataSourceType 未知数据库类型
	ErrUnknownDataSourceType = errors.New("unknown data source type")
	// ErrAlreadyExists 已经存在
	ErrAlreadyExists = errors.New("already exists")

	// ErrNoDataSourceConn 没有可用的数据库连接
	ErrNoDataSourceConn = errors.New("no can used datasource connection")

	ErrNotExists = errors.New("not exists") //nolint:unused
)

//nolint:unused
func IsUnknownDataSourceType(err error) bool {
	return errors.Is(err, ErrUnknownDataSourceType)
}

//nolint:unused
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
