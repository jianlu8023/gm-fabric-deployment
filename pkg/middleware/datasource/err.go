package datasource

import (
	"errors"
)

var (
	ErrUnknownDataSourceType = errors.New("unknown data source type")
)

func IsUnknownDataSourceType(err error) bool {
	return errors.Is(err, ErrUnknownDataSourceType)
}
