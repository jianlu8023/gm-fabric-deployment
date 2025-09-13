package email

// EmailControl 邮件控制器接口
// @description 定义邮件控制器的标准接口
// @interface EmailControl
// @method SendEmail(message *EmailMessage) error 发送邮件
// @method StartUp(failedFunc func(err error)) 启动邮件服务
// @method Shutdown() error 关闭邮件服务
type EmailControl interface {
	// SendEmail 发送邮件
	// @param message *EmailMessage 邮件消息对象
	// @return error 发送过程中的错误
	SendEmail(message *EmailMessage) error

	// StartUp 启动邮件服务
	// @param failedFunc func(err error) 启动失败回调函数
	StartUp(failedFunc func(err error))

	// Shutdown 关闭邮件服务
	// @return error 关闭过程中的错误
	Shutdown() error
}
