package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/grpc"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// GrpcService gRPC服务结构体
//
// @description 提供gRPC相关的服务功能
// @struct
type GrpcService struct {
	*Service        // Service 基础服务，提供日志功能
	mapper         *mapper.GrpcMapper  // mapper gRPC映射器，用于数据访问
	grpcControl    *grpc.Control       // grpcControl gRPC控制器，用于gRPC操作
}

// NewGrpcService 创建新的gRPC服务实例
//
// @description 创建并返回一个新的gRPC服务实例
// @param baseService *Service 基础服务
// @param mapper *mapper.GrpcMapper gRPC映射器
// @param grpcControl *grpc.Control gRPC控制器
// @return *GrpcService gRPC服务实例
func NewGrpcService(baseService *Service,
	mapper *mapper.GrpcMapper,
	grpcControl *grpc.Control) *GrpcService {
	return &GrpcService{
		Service:     baseService,
		mapper:      mapper,
		grpcControl: grpcControl,
	}
}

// GrpcPingMessage 发送gRPC ping消息
//
// @description 通过gRPC发送ping消息并返回响应
// @param ctx *gin.Context HTTP上下文
// @param req *request.GrpcSendPingMessageRequest gRPC发送ping消息请求
func (s *GrpcService) GrpcPingMessage(ctx *gin.Context, req *request.GrpcSendPingMessageRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "grpcService", "grpcPingMessage",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received grpc send ping message request with params: %v", req)

	baseResponse, err := s.grpcControl.Call(&pb.BaseRequest{
		MessageType: grpc.BasePing,
		MessageBody: []byte("ping"),
	})
	if err != nil {
		s.logger.Errorf("grpc send ping message failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, commonhttp.ErrMsgInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	s.logger.Debugf("grpc send ping message success: %v", baseResponse)
	messageResponse, err := response.NewGrpcSendPingMessageResponse(baseResponse)
	if err != nil {
		s.logger.Errorf("convert grpc ping message err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	commonhttp.SuccessResponse(ctx, messageResponse)
	span.SetStatus(codes.Ok, "success")
}
