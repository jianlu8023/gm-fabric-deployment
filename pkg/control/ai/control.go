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
	aiConfig *config.AIConfig
	logger   *zap.SugaredLogger
	once     sync.Once
	client   *http.Client
}

// NewAIControl 创建AI控制器
func NewAIControl(aiConfig *config.AIConfig, loggerControl *logger.Control) (*Control, error) {
	if aiConfig == nil {
		aiConfig = getDefaultConfig()
	}
	if !aiConfig.Enabled {
		return nil, errors.New("ai is disabled")
	}

	aiLogger := loggerControl.GenLogger("ai")
	aiLogger.Infof("[control] starting new ai control...")

	// 创建HTTP客户端
	client := http.NewClient().
		SetTimeout(aiConfig.Timeout*1000). // 转换为毫秒
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: aiConfig.InsecureSkipVerify}).
		SetRetry(3, 500, 2000)

	ctl := &Control{
		aiConfig: aiConfig,
		logger:   aiLogger,
		client:   client,
	}

	return ctl, nil
}

// SendRequest 发送AI请求
func (c *Control) SendRequest(request *Request) (*Response, error) {
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

	// 构建请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		c.logger.Errorf("[control] failed to marshal request: %v", err)
		return nil, err
	}

	c.logger.Debugf("[control] sending ai request to: %s, model: %s", url, request.Model)

	// 发送请求
	var response *Response
	if request.Stream {
		// 流式请求处理
		response, err = c.sendStreamRequest(url, headers, request)
	} else {
		// 普通请求处理
		response, err = c.sendNormalRequest(url, headers, requestBody)
	}

	if err != nil {
		c.logger.Errorf("[control] failed to send ai request: %v", err)
		return nil, err
	}

	c.logger.Infof("[control] ai request sent successfully")
	return response, nil
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
func (c *Control) sendStreamRequest(url string, headers map[string]string, request *Request) (*Response, error) {
	// 使用go-tools的流式处理功能
	response := &Response{
		Choices: make([]Choice, 0),
	}

	// 定义流式回调函数
	callback := func(data []byte, err error) bool {
		if err != nil {
			c.logger.Errorf("[control] stream error: %v", err)
			return false // 停止接收
		}

		if data == nil {
			// 流结束
			c.logger.Debugf("[control] stream ended")
			return false
		}

		// 处理接收到的数据块
		c.logger.Debugf("[control] received stream chunk: %s", string(data))
		return true // 继续接收
	}

	// 发送流式请求
	err := c.client.SetHeaders(headers).StreamPOST(url, request, callback)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// StartUp 启动AI服务
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting up ai service...")

		// 验证配置
		if c.aiConfig.Enabled {
			if stringer.IsBlank(c.aiConfig.APIKey) {
				c.logger.Errorf("[control] ai service enabled but api key is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingAPIKey)
				}
				return
			}

			if stringer.IsBlank(c.aiConfig.APIEndpoint) {
				c.logger.Errorf("[control] ai service enabled but api endpoint is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingAPIEndpoint)
				}
				return
			}

			c.logger.Infof("[control] ai service started successfully")
		} else {
			c.logger.Info("[control] ai service is disabled")
		}
	})
}

// StreamRequest 发送AI流式请求
func (c *Control) StreamRequest(request *Request, callback StreamCallback) error {
	if !c.aiConfig.Enabled {
		return ErrAIServiceDisabled
	}

	c.logger.Debugf("[control] preparing to send ai stream request...")

	// 验证必填字段
	if request.Model == "" {
		return ErrInvalidRequest
	}

	if request.Prompt == "" && len(request.Messages) == 0 {
		return ErrInvalidRequest
	}

	// 构建请求URL
	url := c.aiConfig.APIEndpoint
	if stringer.IsBlank(url) {
		return ErrMissingAPIEndpoint
	}

	// 构建请求头
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.aiConfig.APIKey,
	}

	// 添加自定义请求头
	for k, v := range c.aiConfig.Headers {
		headers[k] = v
	}

	// 设置流式标志
	headers["Accept"] = "text/event-stream"
	headers["Cache-Control"] = "no-cache"
	headers["Connection"] = "keep-alive"

	// 确保请求设置为流式
	request.Stream = true

	// 定义流式回调函数
	streamCallback := func(data []byte, err error) bool {
		if err != nil {
			return callback("", "", err)
		}

		if data == nil {
			// 流结束
			callback("", "stop", nil)
			return false
		}

		// 处理流数据
		c.processStreamData(data, callback)
		return true
	}

	// 发送流式请求
	return c.client.SetHeaders(headers).StreamPOST(url, request, streamCallback)
}

// Shutdown 关闭AI服务
func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutting down ai service...")
	// no-op
	return nil
}
