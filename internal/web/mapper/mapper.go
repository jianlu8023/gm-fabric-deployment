// Package mapper 提供Web模块的数据访问层
//
// @Description 基于GORM实现各业务模块（用户、验证码、系统、节点、Docker镜像、Docker网络、文件、gRPC、WebSocket等）
// 的数据库访问操作，所有写操作均在事务中执行，并统一通过基础Mapper封装数据库连接和日志记录器。
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
