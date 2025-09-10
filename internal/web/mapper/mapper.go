package mapper

import (
	"go.uber.org/zap"
)

type Mapper struct {
	logger *zap.SugaredLogger
}

func NewMapper(logger *zap.SugaredLogger) *Mapper {
	return &Mapper{
		logger: logger,
	}
}
