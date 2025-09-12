package handler

import (
	"go.uber.org/zap"
)

// Handler 基础Handler
type Handler struct {
	logger *zap.SugaredLogger
}

// NewHandler 创建Handler
// @param logger *zap.SugaredLogger 日志
// @return *Handler Handler 实例
func NewHandler(logger *zap.SugaredLogger) *Handler {
	return &Handler{
		logger: logger,
	}
}
