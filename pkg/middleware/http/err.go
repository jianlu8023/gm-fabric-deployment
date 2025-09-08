package http

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
}

func NewError(code ErrCode, message string) Error {
	return Error{
		Code:    code,
		Message: message,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("code:%d,message:%s", e.Code, e.Message)
}

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

	// 第三方服务错误

	ThirdPartyServiceError ErrCode = 8000 // 第三方服务错误
	ThirdPartyAPIFailed    ErrCode = 8001 // 调用第三方 API 失败

)

// IsFileNotFoundError 判断是否是文件未找到的错误 (示例)
func IsFileNotFoundError(err error) bool {
	var fileNotFoundError *Error
	return errors.As(err, &fileNotFoundError) && fileNotFoundError.Code == FileNotFound
}

//nolint:unused
const (
	// 通用错误信息

	ErrMsgNormalFailed        = "通用业务处理失败"
	ErrMsgInternalServerError = "服务器内部错误，请稍后再试"
	ErrMsgInvalidParameter    = "请求参数无效，请检查参数"
	ErrMsgUnauthorized        = "未授权，请先登录"
	ErrMsgForbidden           = "禁止访问，权限不足"
	ErrMsgTooManyRequests     = "请求过于频繁，请稍后再试"

	// 文件相关错误信息

	ErrMsgFileNotFound    = "文件未找到"
	ErrMsgFileTooLarge    = "文件过大，请上传较小的文件"
	ErrMsgFileTypeError   = "文件类型错误，请上传正确的文件类型"
	ErrMsgFileWriteError  = "文件写入失败，请检查磁盘空间或权限"
	ErrMsgFileReadError   = "文件读取失败，请检查文件是否损坏"
	ErrMsgFileDeleteError = "文件删除失败，请检查权限"

	// 数据库相关错误信息

	ErrMsgDatabaseError           = "数据库操作失败，请稍后再试"
	ErrMsgRecordNotFound          = "记录未找到"
	ErrMsgDuplicateRecord         = "记录已存在"
	ErrMsgDatabaseTimeout         = "数据库操作超时，请稍后再试"
	ErrMsgDatabaseConnectionError = "数据库连接失败，请检查数据库配置"

	// 用户相关错误信息

	ErrMsgUserNotFound          = "用户不存在"
	ErrMsgInvalidPassword       = "密码错误"
	ErrMsgUserDisabled          = "用户已被禁用"
	ErrMsgUsernameAlreadyExists = "用户名已存在"
	ErrMsgEmailAlreadyExists    = "邮箱已存在"
	ErrMsgInvalidToken          = "无效的 Token"
	ErrMsgTokenExpired          = "Token 已过期，请重新登录"
	ErrMsgPasswordResetFailed   = "密码重置失败，请稍后再试"

	// 认证授权相关错误信息

	ErrMsgAuthenticationFailed = "认证失败，请检查用户名和密码"
	ErrMsgPermissionDenied     = "权限不足，无法访问该资源"
	ErrMsgInvalidCredentials   = "无效的凭证"
	ErrMsgSessionExpired       = "会话已过期，请重新登录"

	// 网络相关错误信息

	ErrMsgNetworkError       = "网络错误，请检查网络连接"
	ErrMsgTimeoutError       = "请求超时，请稍后再试"
	ErrMsgServiceUnavailable = "服务不可用，请稍后再试"
	ErrMsgConnectionRefused  = "连接被拒绝，请检查服务是否启动"

	// 业务逻辑错误信息

	ErrMsgBusinessLogicError  = "业务逻辑错误"
	ErrMsgInvalidInput        = "无效的输入"
	ErrMsgInsufficientFunds   = "余额不足"
	ErrMsgResourceUnavailable = "资源不可用"
	ErrMsgDataConflict        = "数据冲突，请稍后再试"

	// 第三方服务错误信息

	ErrMsgThirdPartyServiceError = "第三方服务错误，请稍后再试"
	ErrMsgThirdPartyAPIFailed    = "调用第三方 API 失败，请检查配置"
)
