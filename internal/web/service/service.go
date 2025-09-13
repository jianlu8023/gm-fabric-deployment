package service

import (
	"go.uber.org/zap"
)

// Service 基础服务结构体
// @description 所有业务服务的基础结构体，提供日志功能
// @struct
// @property logger *zap.SugaredLogger 日志记录器
type Service struct {
	logger *zap.SugaredLogger
}

// NewService 创建新的基础服务实例
// @description 创建并返回一个新的基础服务实例
// @param logger *zap.SugaredLogger 日志记录器
// @return *Service 基础服务实例
func NewService(logger *zap.SugaredLogger) *Service {
	return &Service{
		logger: logger,
	}
}
