package service

import (
	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
)

type UserService struct {
	*Service
	userMapper *mapper.UserMapper
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

func NewUserService(baseService *Service, userMapper *mapper.UserMapper) *UserService {
	return &UserService{
		Service:    baseService,
		userMapper: userMapper,
	}
}
