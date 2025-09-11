package request

// WSConnectRequest 表示WebSocket连接请求的参数
// 用于在建立WebSocket连接前传递必要的参数

type WSConnectRequest struct {
	NodeID string `json:"node_id" form:"nodeId" binding:"required"` // 节点ID
	Token  string `json:"token" form:"token" binding:"required"`    // 认证令牌
}

// IsLegal 验证WSConnectRequest参数是否合法
func (r *WSConnectRequest) IsLegal() bool {
	if r.NodeID == "" || r.Token == "" {
		return false
	}
	return true
}

// WSMessageRequest 表示WebSocket消息请求的参数
// 用于客户端向服务器发送消息

type WSMessageRequest struct {
	Type    string `json:"type" binding:"required"`    // 消息类型
	Content string `json:"content" binding:"required"` // 消息内容
	NodeID  string `json:"node_id"`                    // 目标节点ID，可选
	UserID  string `json:"user_id"`                    // 目标用户ID，可选
}

// IsLegal 验证WSConnectRequest参数是否合法
func (r *WSMessageRequest) IsLegal() bool {
	// if r.NodeID == "" || r.Token == "" {
	// 	return false
	// }
	return true
}
