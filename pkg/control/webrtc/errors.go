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
	Code    int
	Message string
	Cause   error
}

// NewWebRTCError 创建一个新的WebRTC错误
func NewWebRTCError(code int, message string, cause error) *WebRTCError {
	return &WebRTCError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Error 实现error接口
func (e *WebRTCError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap 实现errors.Unwrap接口
func (e *WebRTCError) Unwrap() error {
	return e.Cause
}

// Is 实现errors.Is接口
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
func GetErrorCode(err error) int {
	if we, ok := err.(*WebRTCError); ok {
		return we.Code
	}
	return 0
}
