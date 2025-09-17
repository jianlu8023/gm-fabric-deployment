package request

import (
	"github.com/jianlu8023/golang-example/pkg/json"
)

// CaptchaRequest 验证码请求参数
// @description 用于生成和验证验证码的请求参数
// @property captchaId string 验证码ID
// @property code string 用户输入的验证码
// @property captchaType string 验证码类型
// @property width int 验证码图片宽度
// @property height int 验证码图片高度
type CaptchaRequest struct {
	CaptchaId string `form:"captchaId" json:"captchaId"`
	Code      string `form:"code" json:"code" binding:"required,min=4,max=6"` // gin 也使用这个库`validate:"min=200"`参数错误时，返回InvalidValidationError类型；校验错误时返回ValidationErrors
	// CaptchaType string `form:"captchaType" json:"captchaType" binding:"omitempty,oneof=string math chinese"`
	// Width       int    `form:"width" json:"width" binding:"omitempty,min=80,max=400"`
	// Height      int    `form:"height" json:"height" binding:"omitempty,min=40,max=200"`
}

func (c *CaptchaRequest) String() string {
	bytes, _ := json.Marshal(c)
	return string(bytes)
}

// IsLegal 验证请求参数是否合法
// @return bool 参数是否合法
func (c *CaptchaRequest) IsLegal() bool {
	// 验证CaptchaType是否合法
	// if c.CaptchaType == "" {
	// 	c.CaptchaType = "string"
	// }
	// 验证宽度和高度
	// if c.Width == 0 {
	// 	c.Width = 240
	// }
	// if c.Height == 0 {
	// 	c.Height = 80
	// }
	// 验证CaptchaId
	// if c.CaptchaId != "" {
	// 验证UUID格式
	// _, err := uuid.Parse(c.CaptchaId)
	// if err != nil {
	// 	return false
	// }
	// }
	return true
}

// GenerateCaptchaRequest 生成验证码请求参数
// @description 仅用于生成验证码的请求参数
// @property captchaType string 验证码类型
// @property width int 验证码图片宽度
// @property height int 验证码图片高度
type GenerateCaptchaRequest struct {
	// CaptchaType string `form:"captchaType" json:"captchaType" binding:"omitempty,oneof=string math chinese"`
	// Width       int    `form:"width" json:"width" binding:"omitempty,min=80,max=400"`
	// Height      int    `form:"height" json:"height" binding:"omitempty,min=40,max=200"`
}

// IsLegal 验证生成验证码请求参数是否合法
// @return bool 参数是否合法
func (g *GenerateCaptchaRequest) IsLegal() bool {
	// 验证CaptchaType是否合法
	// if g.CaptchaType == "" {
	// 	g.CaptchaType = "string"
	// }
	// 验证宽度和高度
	// if g.Width == 0 {
	// 	g.Width = 240
	// }
	// if g.Height == 0 {
	// 	g.Height = 80
	// }
	return true
}

func (g *GenerateCaptchaRequest) String() string {
	bytes, _ := json.Marshal(g)
	return string(bytes)
}
