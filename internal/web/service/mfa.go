package service

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/mfa"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// MFAService MFA服务
// @description MFA相关的业务服务
// @struct
// @property Service 基础服务
// @property mfaControl *mfa.Control MFA控制器

type MFAService struct {
	*Service
	mfaControl *mfa.Control
	userMapper *mapper.UserMapper
}

// NewMFAService 创建MFA服务实例
// @description 创建并返回一个新的MFA服务实例
// @param logger *zap.SugaredLogger 日志记录器
// @param mfaControl *mfa.Control MFA控制器
// @return *MFAService MFA服务实例
func NewMFAService(baseService *Service, userMapper *mapper.UserMapper, mfaControl *mfa.Control) *MFAService {
	return &MFAService{
		Service:    baseService,
		userMapper: userMapper,
		mfaControl: mfaControl,
	}
}

// GenerateRecoverySecret 生成MFA密钥
// @description 为用户生成MFA密钥、二维码URL和二维码图片
// @param req *request.MFARecoverySecretRequest 生成密钥请求
// @return *response.MFASecretResponse MFA密钥响应
// @return error 错误信息
func (s *MFAService) GenerateRecoverySecret(ctx *gin.Context, req *request.MFARecoverySecretRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "mfaService", "generateRecoverySecret",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received mfa generate secret request with params: %v", req)

	exist, err := s.userMapper.QueryExistUser(model.UserInfo{
		UserId: req.UserID,
	})
	if err != nil {
		s.logger.Errorf("query exist user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	if !exist {
		// 用户不存在
		s.logger.Errorf("user %s not exist", req.UserID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.UserNotFound, commonhttp.ErrMsgUserNotFound)
		span.SetStatus(codes.Error, commonhttp.ErrMsgUserNotFound)
		return
	}
	// 查询用户详细信息，检查是否已启用MFA
	user, err := s.userMapper.QueryUserByQuery(model.UserInfo{
		UserId: req.UserID,
	})
	if err != nil {
		s.logger.Errorf("query user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	if user.MFAEnabled.Bool {
		// 用户已启用过MFA，返回错误
		s.logger.Errorf("user %s already enabled MFA", req.UserID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "MFA already enabled for this user")
		span.SetStatus(codes.Error, "MFA already enabled for this user")
		return
	}

	// 使用MFA控制器生成密钥
	secret, qrCodeURL, err := s.mfaControl.GenerateSecret(user.Username)
	if err != nil {
		s.logger.Errorf("generate secret for user %s failed: %v", req.UserID, err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	user.MFAEnabled = sql.NullBool{Bool: true, Valid: true}
	user.MFASecret = secret
	user.MFAOTPAuthURL = qrCodeURL

	recoverySecret, err := s.mfaControl.GenerateRecoverySecret(user.Username, 8)
	if err != nil {
		s.logger.Errorf("generate recovery secret for user %s failed: %v", req.UserID, err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	user.MFARecoverySecret = stringer.Join(recoverySecret, ",")

	if err = s.userMapper.UpdateUser(&user); err != nil {
		s.logger.Errorf("update user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	commonhttp.SuccessResponse(ctx, response.NewMFASecretResponse(recoverySecret))
	span.SetStatus(codes.Ok, "success")
}

// VerifyCode 验证MFA代码
// @description 验证用户提供的MFA代码是否有效
// @param req *request.MFAVerifyCodeRequest 验证代码请求
// @return *response.MFAVerifyResponse MFA验证响应
// @return error 错误信息
func (s *MFAService) VerifyCode(ctx *gin.Context, req *request.MFAVerifyCodeRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "mfaService", "verifyCode",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received mfa verify code request with params: %v", req)
	exist, err := s.userMapper.QueryExistUser(model.UserInfo{
		UserId: req.UserID,
	})
	if err != nil {
		s.logger.Errorf("query exist user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !exist {
		// 用户不存在
		s.logger.Errorf("user %s not exist", req.UserID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.UserNotFound, commonhttp.ErrMsgUserNotFound)
		span.SetStatus(codes.Error, commonhttp.ErrMsgUserNotFound)
		return
	}
	// 查询用户详细信息，检查是否已启用MFA
	user, err := s.userMapper.QueryUserByQuery(model.UserInfo{
		UserId: req.UserID,
	})
	if err != nil {
		s.logger.Errorf("query user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !user.MFAEnabled.Bool {
		s.logger.Errorf("user %v not enabled MFA", req.UserID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "user not enabled MFA")
		span.SetStatus(codes.Error, "user not enabled MFA")
		return
	}
	if stringer.IsBlank(user.MFASecret) {
		s.logger.Errorf("user %v MFA secret is blank...", req.UserID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "user MFA secret is blank")
		span.SetStatus(codes.Error, "user MFA secret is blank")
		return
	}
	valid := s.mfaControl.VerifyCode(user.MFASecret, req.Code)
	if !valid {
		// 验证不通过
		s.logger.Errorf("user %v MFA verify code failed: %v", req.UserID, req.Code)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "verify code failed")
		span.SetStatus(codes.Error, "verify code failed")
		return
	}
	commonhttp.SuccessResponse(ctx, response.NewMFAVerifyResponse(valid))
	span.SetStatus(codes.Ok, "success")
}

func (s *MFAService) GenerateQrCodeImage(ctx *gin.Context, req *request.MFAQrcodeRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "mfaService", "generateQrCodeImage",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received mfa generate qrcode request with params: %v", req)
	exist, err := s.userMapper.QueryExistUser(model.UserInfo{
		UserId: req.UserID,
	})
	if err != nil {
		s.logger.Errorf("query exist user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !exist {
		// 用户不存在
		s.logger.Errorf("user %s not exist", req.UserID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.UserNotFound, commonhttp.ErrMsgUserNotFound)
		span.SetStatus(codes.Error, commonhttp.ErrMsgUserNotFound)
		return
	}
	// 查询用户详细信息，检查是否已启用MFA
	user, err := s.userMapper.QueryUserByQuery(model.UserInfo{
		UserId: req.UserID,
	})
	if err != nil {
		s.logger.Errorf("query user failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, commonhttp.ErrMsgDatabaseError)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	// 如果用户没有启用,则返回错误
	if !user.MFAEnabled.Bool {
		s.logger.Errorf("user %s not enabled MFA", req.UserID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "user not enabled MFA")
		span.SetStatus(codes.Error, "user not enabled MFA")
		return
	}

	qrCodeImage := s.mfaControl.GenerateQrCodeImage(user.MFAOTPAuthURL)

	commonhttp.SuccessResponse(ctx, response.NewMFAQrCodeResponse(qrCodeImage))
	span.SetStatus(codes.Ok, "success")
}
