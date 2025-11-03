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

	// 构建请求体
	requestBody, err := json.Marshal(request)
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

		return nil, c.sendStreamRequest(url, headers, request, callback)
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
func (c *Control) sendStreamRequest(url string, headers map[string]string, request *Request, callback StreamCallback) error {
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
	err := c.client.SetHeaders(headers).StreamPOST(url, request, streamCallback)
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
			// 	c.logger.Errorf("[control] ai service enabled but api key is empty")
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
