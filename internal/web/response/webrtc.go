package response

// WebRTCSDPOfferResponse SDP Offer响应结构
type WebRTCSDPOfferResponse struct {
	ConnectionId string `json:"connectionId"`
	Answer       string `json:"answer"`
}

// WebRTCICECandidateResponse ICE候选响应结构
type WebRTCICECandidateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// WebRTCICECandidateListResponse ICE候选列表响应结构
type WebRTCICECandidateListResponse struct {
	Success    bool                    `json:"success"`
	Candidates []WebRTCICECandidateInfo `json:"candidates,omitempty"`
	Message    string                  `json:"message,omitempty"`
}

// WebRTCICECandidateInfo ICE候选信息结构
type WebRTCICECandidateInfo struct {
	Candidate        string `json:"candidate"`
	SdpMid           string `json:"sdpMid,omitempty"`
	SdpMLineIndex    uint16 `json:"sdpMLineIndex,omitempty"`
	UsernameFragment string `json:"usernameFragment,omitempty"`
}