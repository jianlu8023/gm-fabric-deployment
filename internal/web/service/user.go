package service

import (
	"strconv"

	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"

	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/auth"
)

// UserService 用户服务接口
//
// @description 定义用户服务需要实现的方法，包括用户注册和登录
// @interface
type UserService interface {
	// LoginUser 用户登录
	//
	// @description 处理用户登录请求，验证用户凭据，生成JWT令牌和会话
	// @param ctx *gin.Context Gin上下文
	// @param req *request.UserLoginRequest 用户登录请求参数
	LoginUser(ctx *gin.Context, req *request.UserLoginRequest)
	// RegisterUser 用户注册
	//
	// @description 处理用户注册请求，验证用户邮箱是否已存在，创建新用户
	// @param ctx *gin.Context Gin上下文
	// @param req *request.UserRegisterRequest 用户注册请求参数
	RegisterUser(ctx *gin.Context, req *request.UserRegisterRequest)
}

// userServiceImpl 用户服务结构体
//
// @description 提供用户相关的服务功能，如用户注册、登录等
// @struct
type userServiceImpl struct {
	*Service                         // Service 基础服务
	mapper        mapper.UserMapper  // mapper 用户数据访问对象
	authenticator auth.Authenticator // authenticator 认证器，用于登录/登出/续期
}

// NewUserService 创建用户服务实例
//
// @description 创建并返回一个新的用户服务实例
// @param baseService *Service 基础服务
// @param userMapper mapper.UserMapper 用户映射器
// @param authenticator auth.Authenticator 认证器
// @return UserService 用户服务实例
func NewUserService(baseService *Service, userMapper mapper.UserMapper,
	authenticator auth.Authenticator,
) UserService {
	return &userServiceImpl{
		Service:       baseService,
		mapper:        userMapper,
		authenticator: authenticator,
	}
}

// RegisterUser 用户注册服务
//
// @description 处理用户注册请求，验证用户邮箱是否已存在，创建新用户
// @param ctx *gin.Context Gin上下文
// @param req *request.UserRegisterRequest 用户注册请求参数
func (s *userServiceImpl) RegisterUser(ctx *gin.Context, req *request.UserRegisterRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "userServiceImpl", "registerUser",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received register user request: %v", req)

	// 验证用户是否存在
	exist, err := s.mapper.QueryExistUser(model.UserInfo{
		Email:    req.Email,
		Username: req.Username,
	})
	if err != nil {
		// 数据库查询错误，返回错误
		s.logger.Errorf("query user exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.RecordError(err)
		span.SetStatus(codes.Error, "query user exist failed")
		return
	}

	if exist {
		// 用户已存在，返回错误
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.EmailAlreadyExists, commonhttp.ErrMsgEmailAlreadyExists)
		span.SetStatus(codes.Error, "email already exists")
		return
	}
	// 用户不存在，创建用户
	user := model.NewUserInfo()
	user.Username = req.Username
	user.Password = req.Password
	user.Email = req.Email
	user.UserType = "normal"
	user.UserId = uuid.GetUUID()

	if err = s.mapper.InsertOneUser(user); err != nil {
		s.logger.Errorf("register user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, commonhttp.ErrMsgBusinessLogicError)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	userResponse, err := response.NewRegisterUserResponse(user)
	if err != nil {
		s.logger.Errorf("convert response err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	commonhttp.SuccessResponse(ctx, userResponse)
	span.SetStatus(codes.Ok, "success")
}

// LoginUser 处理用户登录请求
//
// @description 处理用户登录请求，验证用户凭据，生成JWT令牌和会话
// @param ctx *gin.Context Gin上下文
// @param req *request.UserLoginRequest 用户登录请求参数
func (s *userServiceImpl) LoginUser(ctx *gin.Context, req *request.UserLoginRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "userServiceImpl", "loginUser",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received login user request: %v", req)

	// 调用mapper层验证用户凭据
	user, err := s.mapper.QueryUserByUsernameAndPassword(req.Username, req.Password)
	if err != nil {
		// if datasource.IsRecordNotFound(err) {
		//	 s.logger.Errorf("user not found or password incorrect: %v", req.Username)
		//	 webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidPassword, webhttp.ErrMsgInvalidPassword)
		//	 return
		// }
		s.logger.Errorf("query user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.RecordError(err)
		span.SetStatus(codes.Error, "query user failed")
		return
	}

	// 通过认证器登录：生成 sessionID、创建会话、签发 token
	token, err := s.authenticator.Login(
		strconv.Itoa(user.AutoUid),
		user.Username,
		"user", // 默认角色为普通用户
	)
	if err != nil {
		s.logger.Errorf("login failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "生成认证令牌失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if err = s.mapper.UpdateLastLoginTime(user); err != nil {
		s.logger.Errorf("update last login time failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}

	// 返回登录成功响应，包含JWT令牌
	commonhttp.SuccessResponse(ctx, response.NewLoginUserResponse(user, token))
	span.SetStatus(codes.Ok, "success")
}
