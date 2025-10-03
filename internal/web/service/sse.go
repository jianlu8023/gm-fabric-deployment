package service

import (
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
)

// SSEService SSE服务
// @description 提供Server-Sent Events功能的服务，用于实时推送消息
// @struct
// @property *Service 基础服务
// @property sseMapper *mapper.SSEMapper SSE映射器
type SSEService struct {
	*Service
	sseMapper *mapper.SSEMapper
}

// NewSSEService 创建SSE服务实例
// @description 创建并返回一个新的SSE服务实例
// @param baseService *Service 基础服务
// @param sseMapper *mapper.SSEMapper SSE映射器
// @return *SSEService SSE服务实例
func NewSSEService(baseService *Service, sseMapper *mapper.SSEMapper) *SSEService {
	return &SSEService{
		Service:   baseService,
		sseMapper: sseMapper,
	}
}

// SSE 处理SSE连接请求
// @description 处理客户端的SSE连接请求，发送示例消息
// @param ctx *gin.Context Gin上下文
func (s *SSEService) SSE(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "sseService", "sse")
	defer span.End()
	s.logger.Debugf("received sse request...")
	if err := sse.Encode(ctx.Writer, sse.Event{
		Event: "message",
		Data:  "some data\nmore data",
	}); err != nil {
		s.logger.Error("send sse message failed: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	// also a complex type, like a map, a struct or a slice
	if err := sse.Encode(ctx.Writer, sse.Event{
		Id:    "124",
		Event: "message",
		Data: map[string]interface{}{
			"user":    "manu",
			"date":    time.Now().Unix(),
			"content": "hi!",
		},
	}); err != nil {
		s.logger.Error("send sse message failed: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	span.SetStatus(codes.Ok, "success")
}
