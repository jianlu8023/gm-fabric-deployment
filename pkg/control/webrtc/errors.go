package webrtc

import "errors"

// WebRTC相关错误定义
var (
	// 连接错误
	ErrConnectionNotFound      = errors.New("webrtc: connection not found")
	ErrConnectionClosed        = errors.New("webrtc: connection closed")
	ErrConnectionFailed        = errors.New("webrtc: connection failed")
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
//
// @description 自定义WebRTC错误类型，包含错误码、错误消息和错误原因，支持errors.Is/errors.As链式解包
// @struct
type WebRTCError struct {
	// Code 错误码
	Code int
	// Message 错误消息
	Message string
	// Cause 错误原因
	Cause error
}

// NewWebRTCError 创建一个新的WebRTC错误
//
// @description 创建一个新的WebRTC错误对象，包含错误码、错误消息和错误原因
// @param code int 错误码
// @param message string 错误消息
// @param cause error 错误原因
// @return *WebRTCError 创建的WebRTC错误对象
func NewWebRTCError(code int, message string, cause error) *WebRTCError {
	return &WebRTCError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Error 实现error接口
//
// @description 返回错误消息字符串，如果存在错误原因则附加原因信息；对nil接收者做防护避免panic
// @return string 错误消息字符串
func (e *WebRTCError) Error() string {
	if e == nil {
		return "webrtc: nil error"
	}
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap 实现errors.Unwrap接口
//
// @description 返回错误原因，支持errors.Unwrap链式调用；对nil接收者做防护
// @return error 错误原因
func (e *WebRTCError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Is 实现errors.Is接口
//
// @description 判断当前错误是否匹配目标错误，仅按错误码进行匹配；对nil接收者和nil目标做防护
// @param target error 目标错误
// @return bool 如果错误码相同则返回true，否则返回false
func (e *WebRTCError) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	if we, ok := target.(*WebRTCError); ok {
		return e.Code == we.Code
	}
	return false
}

// GetErrorCode 获取错误码
//
// @description 从错误对象中提取WebRTC错误码，支持通过errors.As解包被包装的错误链
// @param err error 错误对象
// @return int 错误码，如果err为nil或不是WebRTCError类型则返回0
func GetErrorCode(err error) int {
	if err == nil {
		return 0
	}
	var we *WebRTCError
	if errors.As(err, &we) {
		if we != nil {
			return we.Code
		}
	}
	return 0
}
