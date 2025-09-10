package mapper

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Mapper struct {
	logger *zap.SugaredLogger
	db     *gorm.DB
}

func NewMapper(logger *zap.SugaredLogger, db *gorm.DB) *Mapper {
	return &Mapper{
		logger: logger,
		db:     db,
	}
}
