package service

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http/middleware/jwt"
)

type UserService struct {
	*Service
	userMapper     *mapper.UserMapper
	sessionManager jwt.SessionManager
}

func NewUserService(baseService *Service, userMapper *mapper.UserMapper, sessionManager jwt.SessionManager) *UserService {
	return &UserService{
		Service:        baseService,
		userMapper:     userMapper,
		sessionManager: sessionManager,
	}
}

func (s *UserService) RegisterUser(ctx *gin.Context, req *request.UserRegisterRequest) {
	s.logger.Debugf("received register user request: %v", req)

	// 验证用户是否存在
	if err := s.userMapper.QueryExistUser(model.UserInfo{
		Email: req.Email,
	}); err != nil {
		if datasource.IsAlreadyExists(err) {
			// 用户已存在，返回错误
			webhttp.FailedResponseWithMessage(ctx, webhttp.EmailAlreadyExists, webhttp.ErrMsgEmailAlreadyExists)
			return
		} else {
			// 数据库查询错误，返回错误
			webhttp.FailedResponseWithMessage(ctx, webhttp.DatabaseError, webhttp.ErrMsgDatabaseError)
			return
		}
	}
	// 用户不存在，创建用户
	user := model.NewUserInfo()
	user.Username = req.Username
	user.Password = req.Password
	user.Email = req.Email

	if err := s.userMapper.InsertOneUser(user); err != nil {
		s.logger.Errorf("register user failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.BusinessLogicError, webhttp.ErrMsgBusinessLogicError)
		return
	}

	webhttp.SuccessResponse(ctx, user)
}

// LoginUser 处理用户登录请求
func (s *UserService) LoginUser(ctx *gin.Context, req *request.UserLoginRequest) {
	s.logger.Debugf("received login user request: %v", req)

	// 验证请求参数
	if !req.IsLegal() {
		s.logger.Errorf("login request is illegal: %v", req)
		webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidParameter, "输入参数不合法")
		return
	}

	// 调用mapper层验证用户凭据
	user, err := s.userMapper.QueryUserByUsernameAndPassword(req.Username, req.Password)
	if err != nil {
		// if datasource.IsRecordNotFound(err) {
		//	 s.logger.Errorf("user not found or password incorrect: %v", req.Username)
		//	 webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidPassword, webhttp.ErrMsgInvalidPassword)
		//	 return
		// }
		s.logger.Errorf("query user failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.DatabaseError, webhttp.ErrMsgDatabaseError)
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
		webhttp.FailedResponseWithMessage(ctx, webhttp.BusinessLogicError, "生成认证令牌失败")
		return
	}

	if err = s.sessionManager.SetSession(sessionID, claims); err != nil {
		s.logger.Errorf("set session failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.BusinessLogicError, "设置会话失败")
		return
	}

	// 返回登录成功响应，包含JWT令牌
	webhttp.SuccessResponse(ctx, map[string]string{
		"token":    token,
		"user_id":  strconv.Itoa(int(user.AutoUid)),
		"username": user.Username,
	})
}
