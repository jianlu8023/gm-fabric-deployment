// Package service 提供Web模块的业务逻辑层
//
// @Description 实现各业务模块（用户、验证码、系统、节点、Docker镜像、Docker网络、文件、gRPC、MFA、WebRTC、WebSocket等）
// 的业务逻辑处理，负责调用mapper层进行数据操作，并集成链路追踪和统一响应封装。
package service

import (
	"go.uber.org/zap"
)

// Service 基础服务结构体
//
// @description 所有业务服务的基础结构体，提供日志功能
// @struct
type Service struct {
	logger *zap.SugaredLogger // logger 日志记录器
}

// NewService 创建新的基础服务实例
//
// @description 创建并返回一个新的基础服务实例
// @param logger *zap.SugaredLogger 日志记录器
// @return *Service 基础服务实例
func NewService(logger *zap.SugaredLogger) *Service {
	return &Service{
		logger: logger,
	}
}
