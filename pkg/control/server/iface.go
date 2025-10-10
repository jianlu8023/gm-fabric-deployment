package server

// ControlInterface 服务器控制器接口
// @description 定义服务器控制器需要实现的接口方法
type ControlInterface interface {
	// StartUp 启动服务器控制器
	// @param failedFunc func(err error) 启动失败时的回调函数
	StartUp(failedFunc func(err error))

	// Shutdown 关闭服务器控制器
	// @return error 关闭过程中可能产生的错误
	Shutdown() error
}