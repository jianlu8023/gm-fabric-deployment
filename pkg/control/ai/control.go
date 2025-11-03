package ai

import (
	"crypto/tls"
	"errors"
	"fmt"
	gohttp "net/http"
	"sync"

	"github.com/jianlu8023/go-tools/v2/pkg/http"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

// Control AI控制器
type Control struct {
	aiConfig   *config.AIConfig
	logger     *zap.SugaredLogger
	once       sync.Once
	client     *http.Client
	toonEncode *ToonEncoder
}

// NewAIControl 创建AI控制器
func NewAIControl(aiConfig *config.AIConfig, loggerControl *logger.Control) (*Control, error) {
	if aiConfig == nil {
		aiConfig = getDefaultConfig()
	}
	if !aiConfig.Enabled {
		return nil, errors.New("ai is disabled")
	}

	aiLogger := loggerControl.GenLogger(logger.ModuleAI)
	aiLogger.Infof("[control] starting new ai control...")

	// 创建HTTP客户端
	client := http.NewClient().
		SetTimeout(aiConfig.Timeout*1000). // 转换为毫秒
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: aiConfig.InsecureSkipVerify}).
		SetRetry(3, 500, 2000)

	// 创建TOON编码器
	toonEncoder := NewToonEncoder(loggerControl.GenLogger("Toon"))

	ctl := &Control{
		aiConfig:   aiConfig,
		logger:     aiLogger,
		client:     client,
		toonEncode: toonEncoder,
	}

	return ctl, nil
}

// SendRequest 发送AI请求
// @param request *Request AI请求对象
// @param callback StreamCallback 流式响应回调函数，如果为nil则为普通请求，否则为流式请求
// @return *Response AI响应对象（仅在非流式请求时返回）
// @return error 发送过程中的错误
func (c *Control) SendRequest(request *Request, callback StreamCallback) (*Response, error) {
	if !c.aiConfig.Enabled {
		return nil, ErrAIServiceDisabled
	}

	c.logger.Debugf("[control] preparing to send ai request...")

	// 验证必填字段
	if stringer.IsBlank(request.Model) {
		return nil, ErrInvalidRequest
	}

	if stringer.IsBlank(request.Prompt) && len(request.Messages) == 0 {
		return nil, ErrInvalidRequest
	}

	// 构建请求URL
	url := c.aiConfig.APIEndpoint
	if stringer.IsBlank(url) {
		return nil, ErrMissingAPIEndpoint
	}

	// 构建请求头
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", c.aiConfig.APIKey),
	}

	// 添加自定义请求头
	for k, v := range c.aiConfig.Headers {
		headers[k] = v
	}

	// 根据是否启用TOON格式决定使用哪种请求结构体
	var requestBody []byte
	var sendRequest interface{}

	if c.aiConfig.UseToonFormat {
		// 启用TOON格式，使用RequestTOON结构体
		toonRequest := &Request{
			Model:       request.Model,
			Prompt:      request.Prompt,
			Temperature: request.Temperature,
			MaxTokens:   request.MaxTokens,
			TopP:        request.TopP,
			Stream:      request.Stream,
			ExtraParams: request.ExtraParams,
		}

		// 处理聊天模式请求 - 只转换Messages中的Content为TOON格式
		if len(request.Messages) > 0 {
			// 创建新的Messages副本，只转换Content字段
			convertedMessages := make([]ChatMessage, len(request.Messages))
			for i, msg := range request.Messages {
				convertedMessages[i] = ChatMessage{
					Role: msg.Role,
				}

				// 尝试将Content转换为TOON格式
				var contentData interface{}
				if err := json.Unmarshal([]byte(msg.Content), &contentData); err == nil {
					// Content是JSON格式，可以转换为TOON
					toonStr, err := c.toonEncode.EncodeToToon(contentData)
					if err != nil {
						c.logger.Warnf("[control] failed to encode message content to TOON format, using original: %v", err)
						convertedMessages[i].Content = msg.Content
					} else {
						// 记录大小差异
						contentJSON, _ := json.Marshal(contentData)
						jsonSize := len(contentJSON)
						toonSize := len(toonStr)
						reduction := float64(jsonSize-toonSize) / float64(jsonSize) * 100

						c.logger.Infof("[control] message content token usage reduction - JSON: %d bytes, TOON: %d bytes, reduction: %.2f%%",
							jsonSize, toonSize, reduction)

						convertedMessages[i].Content = toonStr
					}
				} else {
					// Content是普通文本，不需要转换
					convertedMessages[i].Content = msg.Content
					c.logger.Debugf("[control] message content is plain text, no TOON conversion needed")
				}
			}
			toonRequest.Messages = convertedMessages

		} else if !stringer.IsBlank(request.Prompt) {
			// 处理补全模式请求 - 转换Prompt为TOON格式
			var promptData interface{}
			// 尝试解析Prompt为JSON对象，如果失败则作为普通字符串处理
			if err := json.Unmarshal([]byte(request.Prompt), &promptData); err == nil {
				// Prompt是JSON格式，可以转换为TOON
				toonStr, err := c.toonEncode.EncodeToToon(promptData)
				if err != nil {
					c.logger.Warnf("[control] failed to encode prompt to TOON format, using JSON: %v", err)
				} else {
					// 记录大小差异
					promptJSON, _ := json.Marshal(promptData)
					jsonSize := len(promptJSON)
					toonSize := len(toonStr)
					reduction := float64(jsonSize-toonSize) / float64(jsonSize) * 100

					c.logger.Infof("[control] prompt token usage reduction - JSON: %d bytes, TOON: %d bytes, reduction: %.2f%%",
						jsonSize, toonSize, reduction)

					// 将TOON格式的内容设置到Prompt字段
					toonRequest.Prompt = toonStr
				}
			} else {
				// Prompt是普通字符串，不需要转换
				c.logger.Debugf("[control] prompt is plain text, no TOON conversion needed")
			}
		}

		sendRequest = toonRequest
	} else {
		// 未启用TOON格式，使用原始Request结构体
		sendRequest = request
	}

	// 构建请求体
	var err error
	requestBody, err = json.Marshal(sendRequest)
	if err != nil {
		c.logger.Errorf("[control] failed to marshal request: %v", err)
		return nil, err
	}

	c.logger.Debugf("[control] sending ai request to: %s, model: %s", url, request.Model)

	// 根据request.Stream和callback参数决定处理方式
	if request.Stream {
		// 流式请求
		if callback == nil {
			// 流式请求但未提供callback
			return nil, fmt.Errorf("callback is required for stream request")
		}

		// 设置流式请求头
		headers["Accept"] = "text/event-stream"
		headers["Cache-Control"] = "no-cache"
		headers["Connection"] = "keep-alive"

		return nil, c.sendStreamRequest(url, headers, requestBody, callback)
	} else {
		// 普通请求
		if callback != nil {
			// 普通请求但提供了callback
			c.logger.Warnf("[control] callback provided for non-stream request, ignoring callback")
		}

		return c.sendNormalRequest(url, headers, requestBody)
	}
}

// sendNormalRequest 发送普通请求
func (c *Control) sendNormalRequest(url string, headers map[string]string, body []byte) (*Response, error) {
	respBody, statusCode, err := c.client.SetHeaders(headers).POST(url, body)
	if err != nil {
		return nil, err
	}

	if statusCode != gohttp.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", statusCode)
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	// 检查是否有错误响应
	if response.Error != nil {
		return nil, fmt.Errorf("ai service error: %s", response.Error.Message)
	}

	return &response, nil
}

// sendStreamRequest 发送流式请求
func (c *Control) sendStreamRequest(url string, headers map[string]string, requestBody []byte, callback StreamCallback) error {
	// 定义流式回调函数
	streamCallback := func(data []byte, err error) bool {
		if err != nil {
			// 如果有错误，调用用户回调
			return callback("", "", nil, err)
		}

		if data == nil {
			// 不调用callback("", "stop", nil, nil)，因为processStreamData已经处理了完成信号
			return false
		}

		// 处理接收到的数据块
		// c.logger.Debugf("[control] received stream chunk: %s", string(data))

		// 处理流数据
		c.processStreamData(data, callback)

		return true // 继续接收
	}

	// 发送流式请求
	err := c.client.SetHeaders(headers).StreamPOST(url, requestBody, streamCallback)
	if err != nil {
		return err
	}

	return nil
}

// StartUp 启动AI服务
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting up ai service...")

		// 验证配置
		if c.aiConfig.Enabled {
			// if stringer.IsBlank(c.aiConfig.APIKey) {
			// 	c.logger.Errorf("[control] Ai service enabled but api key is empty")
			// 	if failedFunc != nil {
			// 		failedFunc(ErrMissingAPIKey)
			// 	}
			// 	return
			// }

			if stringer.IsBlank(c.aiConfig.APIEndpoint) {
				c.logger.Errorf("[control] ai service enabled but api endpoint is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingAPIEndpoint)
				}
				return
			}

			c.logger.Infof("[control] ai service started successfully")
		} else {
			c.logger.Warnf("[control] ai service is disabled")
		}
	})
}

// Shutdown 关闭AI服务
func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutting down ai service...")
	// no-op
	return nil
}
