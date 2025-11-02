package ai

import "errors"

var (
	// ErrAIServiceDisabled AI服务未启用
	ErrAIServiceDisabled = errors.New("ai service is disabled")

	// ErrMissingAPIKey 缺少API密钥
	ErrMissingAPIKey = errors.New("missing api key")

	// ErrMissingAPIEndpoint 缺少API端点
	ErrMissingAPIEndpoint = errors.New("missing api endpoint")

	// ErrInvalidRequest 无效的请求
	ErrInvalidRequest = errors.New("invalid request")

	// ErrRequestTimeout 请求超时
	ErrRequestTimeout = errors.New("request timeout")

	// ErrModelNotSupported 模型不支持
	ErrModelNotSupported = errors.New("model not supported")

	// ErrRateLimitExceeded 速率限制超出
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrInvalidResponse 无效的响应
	ErrInvalidResponse = errors.New("invalid response")
)
