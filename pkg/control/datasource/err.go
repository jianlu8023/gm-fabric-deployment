package datasource

import (
	"errors"
)

var (
	// ErrUnknownDataSourceType 未知数据库类型
	ErrUnknownDataSourceType = errors.New("unknown data source type")
	ErrAlreadyExists         = errors.New("already exists")
)

func IsUnknownDataSourceType(err error) bool {
	return errors.Is(err, ErrUnknownDataSourceType)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}
