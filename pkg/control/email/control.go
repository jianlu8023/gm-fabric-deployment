package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"sync"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

// EmailMessage 邮件消息结构
// @description 定义邮件消息的数据结构
// @struct EmailMessage
// @param To []string 收件人列表
// @param Cc []string 抄送人列表
// @param Bcc []string 密送人列表
// @param Subject string 邮件主题
// @param Body string 邮件正文
// @param IsHTML bool 是否为HTML格式
// @param Attachments []Attachment 附件列表
// @param ReplyTo string 回复地址
// @param Headers map[string]string 自定义邮件头

// Attachment 邮件附件结构
// @description 定义邮件附件的数据结构
// @struct Attachment
// @param Filename string 文件名
// @param ContentType string 文件类型
// @param Content []byte 文件内容

// Control 邮件控制器
// @description 负责邮件发送功能的控制器
// @struct Control
// @param emailConfig *config.EmailConfig 邮件配置
// @param logger *zap.SugaredLogger 日志记录器
// @param once sync.Once 单例控制

// NewEmailControl 创建邮件控制器
// @description 创建并初始化邮件控制器
// @param configControl *config.Control 配置控制器
// @param loggerControl *logger.Control 日志控制器
// @return *Control 邮件控制器实例
// @return error 创建过程中的错误

// SendEmail 发送邮件
// @description 发送一封邮件
// @param message *EmailMessage 邮件消息对象
// @return error 发送过程中的错误

// StartUp 启动邮件服务
// @description 启动邮件服务（主要用于初始化和验证配置）
// @param failedFunc func(err error) 启动失败回调函数

// Shutdown 关闭邮件服务
// @description 关闭邮件服务（空操作）
// @return error 关闭过程中的错误

type EmailMessage struct {
	To          []string          // 收件人列表
	Cc          []string          // 抄送人列表
	Bcc         []string          // 密送人列表
	Subject     string            // 邮件主题
	Body        string            // 邮件正文
	IsHTML      bool              // 是否为HTML格式
	Attachments []Attachment      // 附件列表
	ReplyTo     string            // 回复地址
	Headers     map[string]string // 自定义邮件头
}

type Attachment struct {
	Filename    string // 文件名
	ContentType string // 文件类型
	Content     []byte // 文件内容
}

type Control struct {
	emailConfig *config.EmailConfig
	logger      *zap.SugaredLogger
	once        sync.Once
}

// NewEmailControl 创建邮件控制器
func NewEmailControl(configControl *config.Control, loggerControl *logger.Control) (*Control, error) {
	emailLogger := loggerControl.GenLogger("email")
	emailLogger.Infof("[control] starting new email control...")

	emailConfig := configControl.GetEmailConfig()
	if emailConfig == nil {
		emailLogger.Warnf("[control] email config is nil, using default config...")
		emailConfig = &config.EmailConfig{
			Enabled: false,
		}
	}

	ctl := &Control{
		emailConfig: emailConfig,
		logger:      emailLogger,
	}

	return ctl, nil
}

// SendEmail 发送邮件
func (c *Control) SendEmail(message *EmailMessage) error {
	if !c.emailConfig.Enabled {
		return ErrEmailServiceDisabled
	}

	c.logger.Debugf("[control] preparing to send email...")

	// 验证必填字段
	if len(message.To) == 0 {
		return ErrNoRecipients
	}

	if message.Subject == "" {
		return ErrEmptySubject
	}

	// 构建收件人列表
	recipients := append(append(message.To, message.Cc...), message.Bcc...)

	// 构建邮件内容
	from := c.emailConfig.SenderAddress
	if c.emailConfig.SenderName != "" {
		from = fmt.Sprintf("%s <%s>", c.emailConfig.SenderName, c.emailConfig.SenderAddress)
	}

	// 设置认证信息
	auth := smtp.PlainAuth("", c.emailConfig.Username, c.emailConfig.Password, c.emailConfig.SmtpHost)

	// 创建邮件头
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = formatAddresses(message.To)
	headers["Subject"] = message.Subject
	headers["Date"] = time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST")

	// 添加抄送
	if len(message.Cc) > 0 {
		headers["Cc"] = formatAddresses(message.Cc)
	}

	// 添加自定义邮件头
	if message.Headers != nil {
		for k, v := range message.Headers {
			headers[k] = v
		}
	}

	// 设置邮件内容类型
	contentType := "text/plain; charset=utf-8"
	if message.IsHTML {
		contentType = "text/html; charset=utf-8"
	}
	headers["Content-Type"] = contentType

	// 构建完整的邮件内容
	emailBody := buildEmailBody(headers, message.Body)

	// 发送邮件
	c.logger.Debugf("[control] sending email to: %v, subject: %s", message.To, message.Subject)

	var err error
	if c.emailConfig.TlsEnabled {
		// 使用TLS连接
		err = sendEmailWithTLS(c.emailConfig.SmtpHost, c.emailConfig.SmtpPort, auth, from, recipients, []byte(emailBody), c.emailConfig.Debug)
	} else {
		// 使用普通连接
		addr := fmt.Sprintf("%s:%d", c.emailConfig.SmtpHost, c.emailConfig.SmtpPort)
		err = smtp.SendMail(addr, auth, from, recipients, []byte(emailBody))
	}

	if err != nil {
		c.logger.Errorf("[control] failed to send email: %v", err)
		return err
	}

	c.logger.Infof("[control] email sent successfully to: %v", message.To)
	return nil
}

// StartUp 启动邮件服务
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting up email service...")

		// 验证配置
		if c.emailConfig.Enabled {
			if c.emailConfig.SmtpHost == "" {
				c.logger.Errorf("[control] email service enabled but smtp host is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingSmtpHost)
				}
				return
			}

			if c.emailConfig.SmtpPort <= 0 {
				c.logger.Errorf("[control] email service enabled but smtp port is invalid")
				if failedFunc != nil {
					failedFunc(ErrInvalidSmtpPort)
				}
				return
			}

			if c.emailConfig.Username == "" || c.emailConfig.Password == "" {
				c.logger.Errorf("[control] email service enabled but username or password is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingCredentials)
				}
				return
			}

			if c.emailConfig.SenderAddress == "" {
				c.logger.Errorf("[control] email service enabled but sender address is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingSenderAddress)
				}
				return
			}

			c.logger.Infof("[control] email service started successfully")
		} else {
			c.logger.Info("[control] email service is disabled")
		}
	})
}

// Shutdown 关闭邮件服务
func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutting down email service...")
	// no-op
	return nil
}

// 辅助函数：格式化地址列表
func formatAddresses(addresses []string) string {
	result := ""
	for i, addr := range addresses {
		if i > 0 {
			result += ", "
		}
		result += addr
	}
	return result
}

// 辅助函数：构建邮件正文
func buildEmailBody(headers map[string]string, body string) string {
	result := ""
	for k, v := range headers {
		result += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	result += "\r\n"
	result += body
	return result
}

// 辅助函数：使用TLS发送邮件
func sendEmailWithTLS(host string, port int, auth smtp.Auth, from string, to []string, msg []byte, debug bool) error {
	// 创建TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: debug, // 如果是调试模式，可以跳过证书验证
		ServerName:         host,
	}

	// 连接到SMTP服务器
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	// 认证
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return err
	}

	// 添加收件人
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 写入邮件内容
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	// 关闭数据写入器
	if err = w.Close(); err != nil {
		return err
	}

	// 退出SMTP会话
	return client.Quit()
}
