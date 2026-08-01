package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// WebRTCOfferRequest SDP Offer请求结构
//
// @description 用于WebRTC信令交换，客户端提交SDP Offer及ICE候选列表
// @struct
type WebRTCOfferRequest struct {
	Offer      string   `json:"offer" form:"offer" binding:"required"`              // Offer SDP Offer描述（必填）
	Candidates []string `json:"candidates,omitempty" form:"candidates" binding:"-"` // Candidates ICE候选列表（可选）
}

// String 返回请求的字符串表示
//
// @description 将WebRTCOfferRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req WebRTCOfferRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

// IsLegal 验证请求是否合法
//
// @description 检查SDP Offer是否为空
// @return bool 参数是否合法
func (req WebRTCOfferRequest) IsLegal() bool {
	if stringer.IsBlank(req.Offer) {
		return false
	}
	return true
}

// WebRTCIceCandidateRequest ICE候选请求结构
//
// @description 用于WebRTC信令交换，客户端提交ICE候选给服务端
// @struct
type WebRTCIceCandidateRequest struct {
	ConnectionID string `json:"connectionId" form:"connectionId" binding:"required"` // ConnectionID 连接ID（必填）
	Candidate    string `json:"candidate" form:"candidate" binding:"required"`       // Candidate ICE候选描述（必填）
}

// String 返回请求的字符串表示
//
// @description 将WebRTCIceCandidateRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req WebRTCIceCandidateRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

// IsLegal 验证请求是否合法
//
// @description 检查连接ID和ICE候选描述是否为空
// @return bool 参数是否合法
func (req WebRTCIceCandidateRequest) IsLegal() bool {
	if stringer.IsBlank(req.ConnectionID) || stringer.IsBlank(req.Candidate) {
		return false
	}
	return true
}

// GetICECandidatesRequest 获取ICE候选列表请求结构
//
// @description 用于获取指定WebRTC连接的ICE候选列表
// @struct
type GetICECandidatesRequest struct {
	ConnectionID string `json:"connectionId" form:"connectionId" uri:"connectionId" binding:"required"` // ConnectionID 连接ID（必填，从URI获取）
}

// IsLegal 验证请求是否合法
//
// @description 检查连接ID是否为空
// @return bool 参数是否合法
func (req GetICECandidatesRequest) IsLegal() bool {
	if stringer.IsBlank(req.ConnectionID) {
		return false
	}
	return true
}

// String 返回请求的字符串表示
//
// @description 将GetICECandidatesRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req GetICECandidatesRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

// CloseConnectionRequest 关闭WebRTC连接请求结构
//
// @description 用于关闭指定的WebRTC连接
// @struct
type CloseConnectionRequest struct {
	ConnectionID string `json:"connectionId" form:"connectionId" uri:"connectionId" binding:"required"` // ConnectionID 连接ID（必填，从URI获取）
}

// String 返回请求的字符串表示
//
// @description 将CloseConnectionRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req CloseConnectionRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

// IsLegal 验证请求是否合法
//
// @description 检查连接ID是否为空
// @return bool 参数是否合法
func (req CloseConnectionRequest) IsLegal() bool {
	if stringer.IsBlank(req.ConnectionID) {
		return false
	}
	return true
}
