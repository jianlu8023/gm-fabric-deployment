package service

import (
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
)

type GrpcService struct {
	*Service
	mapper      *mapper.GrpcMapper
	grpcControl *grpc.Control
}

func NewGrpcService(baseService *Service,
	mapper *mapper.GrpcMapper,
	grpcControl *grpc.Control) *GrpcService {
	return &GrpcService{
		Service:     baseService,
		mapper:      mapper,
		grpcControl: grpcControl,
	}
}
