package service

import (
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
)

type SSEService struct {
	*Service
	sseMapper *mapper.SSEMapper
}

func NewSSEService(baseService *Service, sseMapper *mapper.SSEMapper) *SSEService {
	return &SSEService{
		Service:   baseService,
		sseMapper: sseMapper,
	}
}

func (s *SSEService) SSE(ctx *gin.Context) {
	s.logger.Debugf("received sse request...")
	if err := sse.Encode(ctx.Writer, sse.Event{
		Event: "message",
		Data:  "some data\nmore data",
	}); err != nil {
		s.logger.Error("send sse message failed: %v", err)
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
		return
	}
}
