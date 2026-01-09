package service

import (
	"go.uber.org/zap"
)

// Service 服务层
//
// struct
type Service struct {
	logger *zap.SugaredLogger // logger 日志
}

// NewService 创建服务层
//
// @param logger *zap.SugaredLogger: 日志
//
// @return *Service: 服务层
func NewService(logger *zap.SugaredLogger) *Service {
	return &Service{
		logger: logger,
	}
}
