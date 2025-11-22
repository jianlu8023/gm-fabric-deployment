package email

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
