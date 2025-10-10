package mapper

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Mapper 提供数据库访问的基础结构体
//
// @description 数据库访问的基础结构体，封装了数据库连接和日志记录器
// @struct
type Mapper struct {
	logger *zap.SugaredLogger // logger 日志记录器
	db     *gorm.DB           // db 数据库连接
}

// NewMapper 创建一个新的Mapper实例
//
// @description 创建并返回一个新的Mapper实例
// @param logger *zap.SugaredLogger 日志记录器
// @param db *gorm.DB 数据库连接
// @return *Mapper Mapper实例
func NewMapper(logger *zap.SugaredLogger, db *gorm.DB) *Mapper {
	return &Mapper{
		logger: logger,
		db:     db,
	}
}
