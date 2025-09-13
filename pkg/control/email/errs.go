package email

import "errors"

// 邮件服务相关错误定义
var (
	// ErrEmailServiceDisabled 邮件服务已禁用
	ErrEmailServiceDisabled = errors.New("email service is disabled")

	// ErrNoRecipients 没有指定收件人
	ErrNoRecipients = errors.New("no recipients specified")

	// ErrEmptySubject 邮件主题为空
	ErrEmptySubject = errors.New("email subject is empty")

	// ErrMissingSmtpHost 缺少SMTP主机配置
	ErrMissingSmtpHost = errors.New("missing smtp host configuration")

	// ErrInvalidSmtpPort SMTP端口无效
	ErrInvalidSmtpPort = errors.New("invalid smtp port")

	// ErrMissingCredentials 缺少认证凭据
	ErrMissingCredentials = errors.New("missing username or password")

	// ErrMissingSenderAddress 缺少发件人地址
	ErrMissingSenderAddress = errors.New("missing sender address")

	// ErrInvalidEmailFormat 邮件格式无效
	ErrInvalidEmailFormat = errors.New("invalid email format")

	// ErrAttachmentTooLarge 附件过大
	ErrAttachmentTooLarge = errors.New("attachment is too large")

	// ErrSendFailed 发送失败
	ErrSendFailed = errors.New("failed to send email")

	// ErrConnectionFailed 连接服务器失败
	ErrConnectionFailed = errors.New("failed to connect to email server")

	// ErrAuthenticationFailed 认证失败
	ErrAuthenticationFailed = errors.New("email authentication failed")
)
