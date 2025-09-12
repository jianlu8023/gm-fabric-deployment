package service

import (
	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
)

type SystemService struct {
	*Service
	mapper *mapper.SystemMapper
}

func (s *SystemService) GetSystemOverview(ctx *gin.Context) {
	s.logger.Debugf("received get system overview request...")
	webhttp.SuccessResponse(ctx, gin.H{
		"version": "1.0.0",
	})
}

func NewSystemService(service *Service, mapper *mapper.SystemMapper) *SystemService {
	return &SystemService{
		Service: service,
		mapper:  mapper,
	}
}
