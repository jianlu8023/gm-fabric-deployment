package response

// WebRTCSDPOfferResponse SDP Offer响应结构
//
// @description 用于返回WebRTC信令交换的SDP Answer结果
// @struct
type WebRTCSDPOfferResponse struct {
	ConnectionId string `json:"connectionId"` // 连接ID
	Answer       string `json:"answer"`       // SDP Answer描述
}

// WebRTCICECandidateResponse ICE候选响应结构
//
// @description 用于返回ICE候选提交操作的结果
// @struct
type WebRTCICECandidateResponse struct {
	Success bool   `json:"success"`           // 操作是否成功
	Message string `json:"message,omitempty"` // 操作结果消息（可选）
}

// WebRTCICECandidateListResponse ICE候选列表响应结构
//
// @description 用于返回指定连接的ICE候选列表
// @struct
type WebRTCICECandidateListResponse struct {
	Success    bool                     `json:"success"`              // 操作是否成功
	Candidates []WebRTCICECandidateInfo `json:"candidates,omitempty"` // ICE候选列表（可选）
	Message    string                   `json:"message,omitempty"`    // 操作结果消息（可选）
}

// WebRTCICECandidateInfo ICE候选信息结构
//
// @description 描述单个ICE候选的详细信息，对应WebRTC API中的RTCIceCandidate
// @struct
type WebRTCICECandidateInfo struct {
	Candidate        string `json:"candidate"`                  // ICE候选描述
	SdpMid           string `json:"sdpMid,omitempty"`           // SDP媒体流标识（可选）
	SdpMLineIndex    uint16 `json:"sdpMLineIndex,omitempty"`    // SDP媒体行索引（可选）
	UsernameFragment string `json:"usernameFragment,omitempty"` // 用户名片段（可选）
}

// WebRTCSingleICECandidateResponse 单个ICE候选响应结构
//
// @description 用于返回单个ICE候选查询的结果
// @struct
type WebRTCSingleICECandidateResponse struct {
	Success   bool   `json:"success"`             // 操作是否成功
	Candidate string `json:"candidate,omitempty"` // ICE候选描述（可选）
	Message   string `json:"message,omitempty"`   // 操作结果消息（可选）
}
