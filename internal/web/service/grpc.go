package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/grpc"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
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

func (s *GrpcService) GrpcPingMessage(ctx *gin.Context, req *request.GrpcSendPingMessageRequest) {
	s.logger.Debugf("received grpc send ping message request with params: %v", req)

	baseResponse, err := s.grpcControl.Call(&pb.BaseRequest{
		MessageType: grpc.BasePing,
		MessageBody: []byte("ping"),
	})
	if err != nil {
		s.logger.Errorf("grpc send ping message failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, commonhttp.ErrMsgInternalServerError)
		return
	}
	s.logger.Debugf("grpc send ping message success: %v", baseResponse)
	messageResponse, err := response.NewGrpcSendPingMessageResponse(baseResponse)
	if err != nil {
		s.logger.Errorf("convert grpc ping message err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		return
	}
	commonhttp.SuccessResponse(ctx, messageResponse)

}
