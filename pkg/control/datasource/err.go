package datasource

import (
	"errors"
)

var (
	// ErrUnknownDataSourceType 未知数据库类型
	// @description 当配置的数据源类型不被支持时返回此错误
	ErrUnknownDataSourceType = errors.New("unknown data source type")

	// ErrAlreadyExists 已经存在
	// @description 当尝试创建已存在的资源时返回此错误
	ErrAlreadyExists = errors.New("already exists")

	// ErrNoDataSourceConn 没有可用的数据库连接
	// @description 当无法获取有效的数据库连接时返回此错误
	ErrNoDataSourceConn = errors.New("no can used datasource connection")

	ErrNotExists = errors.New("not exists") //nolint:unused
)

// IsUnknownDataSourceType 检查是否为未知数据源类型错误
// @description 判断给定的错误是否为未知数据源类型错误
// @param err error 要检查的错误
// @return bool 是否为未知数据源类型错误
//
//nolint:unused
func IsUnknownDataSourceType(err error) bool {
	return errors.Is(err, ErrUnknownDataSourceType)
}

// IsAlreadyExists 检查是否为已存在错误
// @description 判断给定的错误是否为已存在错误
// @param err error 要检查的错误
// @return bool 是否为已存在错误
//
//nolint:unused
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
