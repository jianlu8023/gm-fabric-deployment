package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/common/i18n"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/language"
)

// IsHttpErrServerClosed 判断错误是否是HTTP服务器关闭错误
//
// @description 检查给定的错误是否是http.ErrServerClosed类型的错误
// @param err error 需要检查的错误
// @return bool 如果是http.ErrServerClosed错误则返回true，否则返回false
func IsHttpErrServerClosed(err error) bool {
	return errors.Is(err, http.ErrServerClosed)
}

// Error HTTP错误结构体
//
// @description 定义HTTP响应中的错误信息结构
// @struct
type Error struct {
	// Code 错误码，标识具体的错误类型
	Code ErrCode `json:"code"`
	// Message 错误描述信息，提供给前端显示
	Message string `json:"message"`
}

// NewError 创建新的错误实例
//
// @description 根据错误码和错误信息创建一个新的Error结构体实例
// @param code ErrCode 错误码
// @param message string 错误描述信息
// @return Error 错误结构体实例
func NewError(code ErrCode, message string) Error {
	return Error{
		Code:    code,
		Message: message,
	}
}

// Error 复写Error
// @return string 错误信息
func (e Error) Error() string {
	return fmt.Sprintf("code:%d,message:%s", e.Code, e.Message)
}

// ErrCode 错误码
type ErrCode int

//nolint:unused
const (
	// 通用错误

	NormalFailed        ErrCode = 1000 // 通用业务处理失败
	InternalServerError ErrCode = 1001 // 服务器内部错误
	InvalidParameter    ErrCode = 1002 // 参数无效
	Unauthorized        ErrCode = 1003 // 未授权
	Forbidden           ErrCode = 1004 // 禁止访问
	TooManyRequests     ErrCode = 1005 // 请求过多，限流

	// 文件相关错误

	FileNotFound    ErrCode = 2000 // 文件未找到
	FileTooLarge    ErrCode = 2001 // 文件过大
	FileTypeError   ErrCode = 2002 // 文件类型错误
	FileWriteError  ErrCode = 2003 // 文件写入错误
	FileReadError   ErrCode = 2004 // 文件读取错误
	FileDeleteError ErrCode = 2005 // 文件删除错误

	// 数据库相关错误

	DatabaseError           ErrCode = 3000 // 数据库操作失败
	RecordNotFound          ErrCode = 3001 // 记录未找到
	DuplicateRecord         ErrCode = 3002 // 记录已存在
	DatabaseTimeout         ErrCode = 3003 // 数据库超时
	DatabaseConnectionError ErrCode = 3004 // 数据库连接错误

	// 用户相关错误

	UserNotFound          ErrCode = 4000 // 用户不存在
	InvalidPassword       ErrCode = 4001 // 密码错误
	UserDisabled          ErrCode = 4002 // 用户被禁用
	UsernameAlreadyExists ErrCode = 4003 // 用户名已存在
	EmailAlreadyExists    ErrCode = 4004 // 邮箱已存在
	InvalidToken          ErrCode = 4005 // 无效的 Token
	TokenExpired          ErrCode = 4006 // Token 过期
	PasswordResetFailed   ErrCode = 4007 // 密码重置失败

	// 认证授权相关错误

	AuthenticationFailed ErrCode = 5000 // 认证失败
	PermissionDenied     ErrCode = 5001 // 权限不足
	InvalidCredentials   ErrCode = 5002 // 无效的凭证
	SessionExpired       ErrCode = 5003 // 会话过期

	// 网络相关错误

	NetworkError       ErrCode = 6000 // 网络错误
	TimeoutError       ErrCode = 6001 // 超时错误
	ServiceUnavailable ErrCode = 6002 // 服务不可用
	ConnectionRefused  ErrCode = 6003 // 连接被拒绝

	// 业务逻辑错误

	BusinessLogicError  ErrCode = 7000 // 业务逻辑错误
	InvalidInput        ErrCode = 7001 // 无效的输入
	InsufficientFunds   ErrCode = 7002 // 余额不足
	ResourceUnavailable ErrCode = 7003 // 资源不可用
	DataConflict        ErrCode = 7004 // 数据冲突

	// 验证码相关错误

	InvalidCaptcha          ErrCode = 7100 // 验证码无效
	CaptchaExpired          ErrCode = 7101 // 验证码过期
	CaptchaGenerationFailed ErrCode = 7102 // 验证码生成失败

	// 第三方服务错误

	ThirdPartyServiceError ErrCode = 8000 // 第三方服务错误
	ThirdPartyAPIFailed    ErrCode = 8001 // 调用第三方 API 失败

)

// IsFileNotFoundError 判断是否是文件未找到的错误
// @param err error 错误
// @return bool 是否是文件未找到的错误
func IsFileNotFoundError(err error) bool {
	var fileNotFoundError *Error
	return errors.As(err, &fileNotFoundError) && fileNotFoundError.Code == FileNotFound
}

//nolint:unused
const (
	// 通用错误信息键

	ErrMsgNormalFailed        = "failed"
	ErrMsgInternalServerError = "internal_error"
	ErrMsgInvalidParameter    = "invalid_parameter"
	ErrMsgUnauthorized        = "unauthorized"
	ErrMsgForbidden           = "forbidden"
	ErrMsgTooManyRequests     = "too_many_requests"

	// 文件相关错误信息键

	ErrMsgFileNotFound    = "file_not_found"
	ErrMsgFileTooLarge    = "file_too_large"
	ErrMsgFileTypeError   = "file_type_error"
	ErrMsgFileWriteError  = "file_write_error"
	ErrMsgFileReadError   = "file_read_error"
	ErrMsgFileDeleteError = "file_delete_error"

	// 数据库相关错误信息键

	ErrMsgDatabaseError           = "database_error"
	ErrMsgRecordNotFound          = "record_not_found"
	ErrMsgDuplicateRecord         = "duplicate_record"
	ErrMsgDatabaseTimeout         = "database_timeout"
	ErrMsgDatabaseConnectionError = "database_connection_error"

	// 用户相关错误信息键

	ErrMsgUserNotFound          = "user_not_found"
	ErrMsgInvalidPassword       = "invalid_password"
	ErrMsgUserDisabled          = "user_disabled"
	ErrMsgUsernameAlreadyExists = "username_already_exists"
	ErrMsgEmailAlreadyExists    = "email_already_exists"
	ErrMsgInvalidToken          = "invalid_token"
	ErrMsgTokenExpired          = "token_expired"
	ErrMsgPasswordResetFailed   = "password_reset_failed"

	// 认证授权相关错误信息键

	ErrMsgAuthenticationFailed = "authentication_failed"
	ErrMsgPermissionDenied     = "permission_denied"
	ErrMsgInvalidCredentials   = "invalid_credentials"
	ErrMsgSessionExpired       = "session_expired"

	// 网络相关错误信息键

	ErrMsgNetworkError       = "network_error"
	ErrMsgTimeoutError       = "timeout_error"
	ErrMsgServiceUnavailable = "service_unavailable"
	ErrMsgConnectionRefused  = "connection_refused"

	// 业务逻辑错误信息键

	ErrMsgBusinessLogicError  = "business_logic_error"
	ErrMsgInvalidInput        = "invalid_input"
	ErrMsgInsufficientFunds   = "insufficient_funds"
	ErrMsgResourceUnavailable = "resource_unavailable"
	ErrMsgDataConflict        = "data_conflict"

	// 验证码相关错误信息键

	ErrMsgInvalidCaptcha          = "invalid_captcha"
	ErrMsgCaptchaExpired          = "captcha_expired"
	ErrMsgCaptchaGenerationFailed = "captcha_generation_failed"

	// 第三方服务错误信息键

	ErrMsgThirdPartyServiceError = "third_party_service_error"
	ErrMsgThirdPartyAPIFailed    = "third_party_api_failed"
)

// convertErrorMessage 获取错误信息（支持国际化）
// @param code ErrCode 错误码
// @param lang string 语言代码
// @return string 错误信息
func convertErrorMessage(code ErrCode, lang string) string {
	var key string

	switch code {
	case NormalFailed:
		key = ErrMsgNormalFailed
	case InternalServerError:
		key = ErrMsgInternalServerError
	case InvalidParameter:
		key = ErrMsgInvalidParameter
	case Unauthorized:
		key = ErrMsgUnauthorized
	case Forbidden:
		key = ErrMsgForbidden
	case TooManyRequests:
		key = ErrMsgTooManyRequests
	case FileNotFound:
		key = ErrMsgFileNotFound
	case FileTooLarge:
		key = ErrMsgFileTooLarge
	case FileTypeError:
		key = ErrMsgFileTypeError
	case FileWriteError:
		key = ErrMsgFileWriteError
	case FileReadError:
		key = ErrMsgFileReadError
	case FileDeleteError:
		key = ErrMsgFileDeleteError
	case DatabaseError:
		key = ErrMsgDatabaseError
	case RecordNotFound:
		key = ErrMsgRecordNotFound
	case DuplicateRecord:
		key = ErrMsgDuplicateRecord
	case DatabaseTimeout:
		key = ErrMsgDatabaseTimeout
	case DatabaseConnectionError:
		key = ErrMsgDatabaseConnectionError
	case UserNotFound:
		key = ErrMsgUserNotFound
	case InvalidPassword:
		key = ErrMsgInvalidPassword
	case UserDisabled:
		key = ErrMsgUserDisabled
	case UsernameAlreadyExists:
		key = ErrMsgUsernameAlreadyExists
	case EmailAlreadyExists:
		key = ErrMsgEmailAlreadyExists
	case InvalidToken:
		key = ErrMsgInvalidToken
	case TokenExpired:
		key = ErrMsgTokenExpired
	case PasswordResetFailed:
		key = ErrMsgPasswordResetFailed
	case AuthenticationFailed:
		key = ErrMsgAuthenticationFailed
	case PermissionDenied:
		key = ErrMsgPermissionDenied
	case InvalidCredentials:
		key = ErrMsgInvalidCredentials
	case SessionExpired:
		key = ErrMsgSessionExpired
	case NetworkError:
		key = ErrMsgNetworkError
	case TimeoutError:
		key = ErrMsgTimeoutError
	case ServiceUnavailable:
		key = ErrMsgServiceUnavailable
	case ConnectionRefused:
		key = ErrMsgConnectionRefused
	case BusinessLogicError:
		key = ErrMsgBusinessLogicError
	case InvalidInput:
		key = ErrMsgInvalidInput
	case InsufficientFunds:
		key = ErrMsgInsufficientFunds
	case ResourceUnavailable:
		key = ErrMsgResourceUnavailable
	case DataConflict:
		key = ErrMsgDataConflict
	case InvalidCaptcha:
		key = ErrMsgInvalidCaptcha
	case CaptchaExpired:
		key = ErrMsgCaptchaExpired
	case CaptchaGenerationFailed:
		key = ErrMsgCaptchaGenerationFailed
	case ThirdPartyServiceError:
		key = ErrMsgThirdPartyServiceError
	case ThirdPartyAPIFailed:
		key = ErrMsgThirdPartyAPIFailed
	default:
		key = "unknown_error"
	}

	return i18n.Translate(lang, key)
}

// GetErrorMessageWithContext 从Gin上下文中获取错误信息（支持国际化）
// @param ctx *gin.Context Gin上下文
// @param code ErrCode 错误码
// @return string 错误信息
func GetErrorMessageWithContext(ctx *gin.Context, code ErrCode) string {
	return convertErrorMessage(code, language.GetLanguageFromContext(ctx))
}
