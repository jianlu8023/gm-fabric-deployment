package ai

// Request AI请求结构体
type Request struct {
	Model       string                 `json:"model,omitempty"`        // 模型名称
	Prompt      string                 `json:"prompt,omitempty"`       // 提示词
	Messages    []ChatMessage          `json:"messages,omitempty"`     // 聊天消息列表
	Temperature float64                `json:"temperature,omitempty"`  // 温度参数
	MaxTokens   int                    `json:"max_tokens,omitempty"`   // 最大token数
	TopP        float64                `json:"top_p,omitempty"`        // top_p参数
	Stream      bool                   `json:"stream,omitempty"`       // 是否流式输出
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"` // 额外参数
}

type RequestTOON struct {
	Model       string                 `json:"model,omitempty"`        // 模型名称
	Prompt      string                 `json:"prompt,omitempty"`       // 提示词
	Messages    string                 `json:"messages,omitempty"`     // 聊天消息列表
	Temperature float64                `json:"temperature,omitempty"`  // 温度参数
	MaxTokens   int                    `json:"max_tokens,omitempty"`   // 最大token数
	TopP        float64                `json:"top_p,omitempty"`        // top_p参数
	Stream      bool                   `json:"stream,omitempty"`       // 是否流式输出
	ExtraParams map[string]interface{} `json:"extra_params,omitempty"` // 额外参数
}

// ChatMessage AI聊天消息结构体
type ChatMessage struct {
	Role    string `json:"role,omitempty"`    // 角色 (system, user, assistant)
	Content string `json:"content,omitempty"` // 内容
}

// Response AI响应结构体
type Response struct {
	ID      string   `json:"id,omitempty"`      // 响应ID
	Object  string   `json:"object,omitempty"`  // 对象类型
	Created int64    `json:"created,omitempty"` // 创建时间
	Model   string   `json:"model,omitempty"`   // 模型名称
	Choices []Choice `json:"choices,omitempty"` // 选择列表
	Usage   *Usage   `json:"usage,omitempty"`   // 使用情况
	Error   *Error   `json:"error,omitempty"`   // 错误信息
}

// Choice AI选择结构体
type Choice struct {
	Index        int          `json:"index,omitempty"`         // 索引
	Message      *ChatMessage `json:"message,omitempty"`       // 消息
	Delta        *ChatMessage `json:"delta,omitempty"`         // 流式输出增量
	FinishReason string       `json:"finish_reason,omitempty"` // 完成原因
}

// Usage AI使用情况结构体
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`     // 提示token数
	CompletionTokens int `json:"completion_tokens,omitempty"` // 完成token数
	TotalTokens      int `json:"total_tokens,omitempty"`      // 总token数
}

// Error AI错误结构体
type Error struct {
	Message string      `json:"message,omitempty"` // 错误消息
	Type    string      `json:"type,omitempty"`    // 错误类型
	Param   interface{} `json:"param,omitempty"`   // 参数
	Code    interface{} `json:"code,omitempty"`    // 错误码
}
