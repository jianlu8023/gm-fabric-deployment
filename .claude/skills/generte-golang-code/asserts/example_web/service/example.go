package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/mapper"
	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/model"
	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/request"
	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ExampleService ExampleService接口
//
// interface
type ExampleService interface {
	// ExampleService example具体实现逻辑
	//
	// @param ctx *gin.Context: gin上下文
	// @param req *request.ExampleRequest: 请求参数
	ExampleService(ctx *gin.Context, req *request.ExampleRequest)
}

// exampleService example服务层具体实现
//
// struct
type exampleService struct {
	*Service
	exampleMapper mapper.ExampleMapper
}

// NewExampleService 创建ExampleService实例
//
// @param baseService *Service: Service实例
// @param exampleMapper mapper.ExampleMapper: ExampleMapper实例
//
// @return ExampleService: ExampleService实例
func NewExampleService(baseService *Service, exampleMapper mapper.ExampleMapper) ExampleService {
	return &exampleService{
		Service:       baseService,
		exampleMapper: exampleMapper,
	}
}

// ExampleService example具体实现逻辑
//
// @param ctx *gin.Context: gin上下文
// @param req *request.ExampleRequest: 请求参数
func (s *exampleService) ExampleService(ctx *gin.Context, req *request.ExampleRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "exampleService", "ExampleService",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()

	s.logger.Debugf("received example service request with params: %v", req)
	// 具体实现逻辑 以分页查询为例

	queryPage, err := s.exampleMapper.Pagination(ctx.Request.Context(), model.ExampleModel{
		ExampleKey: req.Query,
	}, dbpage.Info[model.ExampleModel]{
		PageNo:   int64(req.PageNo),
		PageSize: int64(req.PageSize),
	})
	if err != nil {
		s.logger.Errorf("get example service list failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取example信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	listResponse, err := response.NewExampleResponse(queryPage)
	if err != nil {
		s.logger.Errorf("convert example list list err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	commonhttp.SuccessResponse(ctx, listResponse)
	span.SetStatus(codes.Ok, "success")
}
