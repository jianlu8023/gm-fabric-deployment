package webrtc

import "errors"

// WebRTC相关错误定义
var (
	// 连接错误
	ErrConnectionNotFound      = errors.New("webrtc: connection not found")
	ErrConnectionClosed        = errors.New("webrtc: connection closed")
	ErrConnectionBufferFull    = errors.New("webrtc: connection buffer full")
	ErrConnectionAlreadyExists = errors.New("webrtc: connection already exists")

	// SDP错误
	ErrSDPInvalid     = errors.New("webrtc: invalid SDP")
	ErrSDPParseFailed = errors.New("webrtc: failed to parse SDP")
	ErrSDPMissing     = errors.New("webrtc: SDP missing")

	// ICE错误
	ErrICECandidateInvalid  = errors.New("webrtc: invalid ICE candidate")
	ErrICENegotiationFailed = errors.New("webrtc: ICE negotiation failed")
	ErrICEConnectionTimeout = errors.New("webrtc: ICE connection timeout")

	// 媒体错误
	ErrMediaTrackNotFound       = errors.New("webrtc: media track not found")
	ErrMediaTrackCreationFailed = errors.New("webrtc: media track creation failed")
	ErrMediaEncodingFailed      = errors.New("webrtc: media encoding failed")
	ErrMediaDecodingFailed      = errors.New("webrtc: media decoding failed")

	// 其他错误
	ErrControllerNotInitialized = errors.New("webrtc: controller not initialized")
	ErrInvalidConfiguration     = errors.New("webrtc: invalid configuration")
	ErrNotSupported             = errors.New("webrtc: operation not supported")
)

// WebRTCError 自定义WebRTC错误类型
type WebRTCError struct {
	// Code 错误码
	Code int
	// Message 错误消息
	Message string
	// Cause 错误原因
	Cause error
}

// NewWebRTCError 创建一个新的WebRTC错误
// 参数:
//   - code: 错误码
//   - message: 错误消息
//   - cause: 错误原因
//
// 返回值:
//   - *WebRTCError: 创建的WebRTC错误对象
func NewWebRTCError(code int, message string, cause error) *WebRTCError {
	return &WebRTCError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Error 实现error接口
// 返回值:
//   - string: 错误消息字符串
func (e *WebRTCError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap 实现errors.Unwrap接口
// 返回值:
//   - error: 错误原因
func (e *WebRTCError) Unwrap() error {
	return e.Cause
}

// Is 实现errors.Is接口
// 参数:
//   - target: 目标错误
//
// 返回值:
//   - bool: 如果错误码相同则返回true，否则返回false
func (e *WebRTCError) Is(target error) bool {
	if target == nil {
		return false
	}
	if we, ok := target.(*WebRTCError); ok {
		return e.Code == we.Code
	}
	return false
}

// GetErrorCode 获取错误码
// 参数:
//   - err: 错误对象
//
// 返回值:
//   - int: 错误码，如果err不是WebRTCError类型则返回0
func GetErrorCode(err error) int {
	if we, ok := err.(*WebRTCError); ok {
		return we.Code
	}
	return 0
}
