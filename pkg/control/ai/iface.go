package ai

// ControlInterface AI控制器接口
// @description 定义AI控制器的标准接口
// @interface ControlInterface
type ControlInterface interface {
	// SendRequest 发送AI请求
	// @param request *AIRequest AI请求对象
	// @return *AIResponse AI响应对象
	// @return error 发送过程中的错误
	SendRequest(request *Request) (*Response, error)

	// StreamRequest 发送AI流式请求
	// @param request *AIRequest AI请求对象
	// @param callback StreamCallback 流式响应回调函数
	// @return error 发送过程中的错误
	StreamRequest(request *Request, callback StreamCallback) error
}
