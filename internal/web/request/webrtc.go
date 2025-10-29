package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// WebRTCOfferRequest SDP Offer请求结构
type WebRTCOfferRequest struct {
	Offer      string   `json:"offer" form:"offer" binding:"required"`
	Candidates []string `json:"candidates,omitempty" form:"candidates" binding:"-"`
}

// String 返回请求的字符串表示
func (req WebRTCOfferRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

// IsLegal 验证请求是否合法
func (req WebRTCOfferRequest) IsLegal() bool {
	if stringer.IsBlank(req.Offer) {
		return false
	}
	return true
}

// WebRTCIceCandidateRequest ICE候选请求结构
type WebRTCIceCandidateRequest struct {
	ConnectionID string `json:"connectionId" form:"connectionId" binding:"required"`
	Candidate    string `json:"candidate" form:"candidate" binding:"required"`
}

// String 返回请求的字符串表示
func (req WebRTCIceCandidateRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

// IsLegal 验证请求是否合法
func (req WebRTCIceCandidateRequest) IsLegal() bool {
	if stringer.IsBlank(req.ConnectionID) || stringer.IsBlank(req.Candidate) {
		return false
	}
	return true
}

// GetICECandidatesRequest 获取ICE候选列表请求结构
type GetICECandidatesRequest struct {
	ConnectionID string `json:"connectionId" form:"connectionId" uri:"connectionId" binding:"required"`
}

// IsLegal 验证请求是否合法
func (req GetICECandidatesRequest) IsLegal() bool {
	if stringer.IsBlank(req.ConnectionID) {
		return false
	}
	return true
}

// String 返回请求的字符串表示
func (req GetICECandidatesRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

type CloseConnectionRequest struct {
	ConnectionID string `json:"connectionId" form:"connectionId" uri:"connectionId" binding:"required"`
}

func (req CloseConnectionRequest) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}

// IsLegal 验证请求是否合法
func (req CloseConnectionRequest) IsLegal() bool {
	if stringer.IsBlank(req.ConnectionID) {
		return false
	}
	return true
}
