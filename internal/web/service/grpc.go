package service

import (
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/pkg/control/grpc"
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
