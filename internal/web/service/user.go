package service

import (
	"strconv"

	humantime "github.com/jianlu8023/go-tools/v2/pkg/helper/time"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/fabricca"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
)

// UserService 用户服务
// @description 提供用户相关的服务功能，如用户注册、登录等
// @struct
// @property *Service 基础服务
// @property mapper *mapper.UserMapper 用户映射器
// @property sessionManager jwt.SessionManager 会话管理器
type UserService struct {
	*Service
	mapper          *mapper.UserMapper
	sessionManager  jwt.SessionManager
	fabriccaControl *fabricca.Control
}

// NewUserService 创建用户服务实例
// @description 创建并返回一个新的用户服务实例
// @param baseService *Service 基础服务
// @param mapper *mapper.UserMapper 用户映射器
// @param sessionManager jwt.SessionManager 会话管理器
// @return *UserService 用户服务实例
func NewUserService(baseService *Service, userMapper *mapper.UserMapper,
	sessionManager jwt.SessionManager,
	fabriccaControl *fabricca.Control,
) *UserService {
	return &UserService{
		Service:         baseService,
		mapper:          userMapper,
		sessionManager:  sessionManager,
		fabriccaControl: fabriccaControl,
	}
}

// RegisterUser 用户注册服务
// @description 处理用户注册请求，验证用户邮箱是否已存在，创建新用户
// @param ctx *gin.Context Gin上下文
// @param req *request.UserRegisterRequest 用户注册请求参数
func (s *UserService) RegisterUser(ctx *gin.Context, req *request.UserRegisterRequest) {
	s.logger.Debugf("received register user request: %v", req)

	if s.fabriccaControl == nil {
		s.logger.Errorf("fabricca is nil, register user disabled...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, "no available fabric ca server")
		return
	}
	// 验证用户是否存在
	if err := s.mapper.QueryExistUser(model.UserInfo{
		Email: req.Email,
	}); err != nil {
		if datasource.IsAlreadyExists(err) {
			// 用户已存在，返回错误
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.EmailAlreadyExists, commonhttp.ErrMsgEmailAlreadyExists)
			return
		} else {
			// 数据库查询错误，返回错误
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
			return
		}
	}
	// 用户不存在，创建用户
	user := model.NewUserInfo()
	user.Username = req.Username
	user.Password = req.Password
	user.Email = req.Email
	user.UserType = "normal"

	if err := s.fabriccaControl.RegisterUserAndGetCert(user); err != nil {
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "register user failed")
		return
	}

	if err := s.mapper.InsertOneUser(user); err != nil {
		s.logger.Errorf("register user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, commonhttp.ErrMsgBusinessLogicError)
		return
	}

	commonhttp.SuccessResponse(ctx, user)
}

// LoginUser 处理用户登录请求
// @description 处理用户登录请求，验证用户凭据，生成JWT令牌和会话
// @param ctx *gin.Context Gin上下文
// @param req *request.UserLoginRequest 用户登录请求参数
func (s *UserService) LoginUser(ctx *gin.Context, req *request.UserLoginRequest) {
	s.logger.Debugf("received login user request: %v", req)

	if s.fabriccaControl == nil {
		s.logger.Errorf("fabricca is nil, login user disabled...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, "no available fabric ca server")
		return
	}

	// 验证请求参数
	if !req.IsLegal() {
		s.logger.Errorf("login request is illegal: %v", req)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "输入参数不合法")
		return
	}

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
		return
	}

	if err = s.fabriccaControl.EnrollUser(user); err != nil {
		s.logger.Errorf("from fabric ca enroll user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "fabric ca verify failed")
		return
	}

	// 生成会话ID
	sessionID := uuid.New().String()
	// 生成JWT令牌，过期时间设置为24小时
	token, claims, err := jwt.GenerateToken(
		strconv.Itoa(int(user.AutoUid)),
		user.Username,
		"user", // 默认角色为普通用户
		sessionID,
		86400, // 24小时有效期
	)
	if err != nil {
		s.logger.Errorf("generate token failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "生成认证令牌失败")
		return
	}

	if err = s.sessionManager.SetSession(sessionID, claims); err != nil {
		s.logger.Errorf("set session failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "设置会话失败")
		return
	}

	if err = s.mapper.UpdateLastLoginTime(user); err != nil {
		s.logger.Errorf("update last login time failed: %v", err)
	}

	// 返回登录成功响应，包含JWT令牌
	commonhttp.SuccessResponse(ctx, map[string]string{
		"token":           token,
		"user_id":         strconv.Itoa(user.AutoUid),
		"username":        user.Username,
		"last_login_time": humantime.HumanTimeLower(user.LastLoginTime, ""),
	})
}
