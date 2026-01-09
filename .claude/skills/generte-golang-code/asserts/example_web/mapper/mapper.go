package mapper

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Mapper 数据访问层
//
// struct
type Mapper struct {
	db     *gorm.DB           // db 数据库连接
	logger *zap.SugaredLogger // logger 日志记录器
}

// NewMapper 创建 Mapper 实例
//
// @param db *gorm.DB: 数据库连接
// @param logger *zap.SugaredLogger: 日志记录器
//
// @return *Mapper: Mapper 实例
func NewMapper(db *gorm.DB, logger *zap.SugaredLogger) *Mapper {
	return &Mapper{
		db:     db,
		logger: logger,
	}
}
