package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"sync"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	emailConfig *config.EmailConfig
	logger      *zap.SugaredLogger
	once        sync.Once
}

// NewEmailControl 创建邮件控制器
func NewEmailControl(emailConfig *config.EmailConfig, loggerControl *logger.Control) (*Control, error) {
	emailConfig = check.IF[*config.EmailConfig](emailConfig == nil,
		getDefaultConfig(),
		emailConfig,
	)
	if !emailConfig.Enabled {
		return nil, errors.New("email is not enabled")
	}
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)

	// emailLogger := loggerControl.GenLogger("email")
	emailLogger := loggerControl.GenLogger("")
	emailLogger.Infof("[email/control] starting new email control...")

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

	c.logger.Debugf("[email/control] preparing to send email...")

	// 验证必填字段
	if len(message.To) == 0 {
		return ErrNoRecipients
	}

	if stringer.IsBlank(message.Subject) {
		return ErrEmptySubject
	}

	// 构建收件人列表
	recipients := append(append(message.To, message.Cc...), message.Bcc...)

	// 构建邮件内容
	from := c.emailConfig.SenderAddress
	if !stringer.IsBlank(c.emailConfig.SenderName) {
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
	c.logger.Debugf("[email/control] sending email to: %v, subject: %s", message.To, message.Subject)

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
		c.logger.Errorf("[email/control] failed to send email: %v", err)
		return err
	}

	c.logger.Infof("[email/control] email sent successfully to: %v", message.To)
	return nil
}

// StartUp 启动邮件服务
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		// 验证配置
		if c.emailConfig.Enabled {
			c.logger.Debugf("[email/control] starting up email service...")
			if c.emailConfig.SmtpHost == "" {
				c.logger.Errorf("[email/control] email service enabled but smtp host is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingSmtpHost)
				}
				return
			}

			if c.emailConfig.SmtpPort <= 0 {
				c.logger.Errorf("[email/control] email service enabled but smtp port is invalid")
				if failedFunc != nil {
					failedFunc(ErrInvalidSmtpPort)
				}
				return
			}

			if c.emailConfig.Username == "" || c.emailConfig.Password == "" {
				c.logger.Errorf("[email/control] email service enabled but username or password is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingCredentials)
				}
				return
			}

			if c.emailConfig.SenderAddress == "" {
				c.logger.Errorf("[email/control] email service enabled but sender address is empty")
				if failedFunc != nil {
					failedFunc(ErrMissingSenderAddress)
				}
				return
			}

			c.logger.Infof("[email/control] email service started successfully")
		} else {
			c.logger.Info("[email/control] email service is disabled")
		}
	})
}

// Shutdown 关闭邮件服务
func (c *Control) Shutdown() error {
	c.logger.Debugf("[email/control] shutting down email service...")
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
