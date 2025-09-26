package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// CaptchaValidateRequest 验证码请求参数
// @description 用于生成和验证验证码的请求参数
// @property captchaId string 验证码ID
// @property code string 用户输入的验证码
// @property captchaType string 验证码类型
// @property width int 验证码图片宽度
// @property height int 验证码图片高度
type CaptchaValidateRequest struct {
	CaptchaId string `form:"captchaId" json:"captchaId"`
	Code      string `form:"code" json:"code" binding:"required,min=4,max=6"` // gin 也使用这个库`validate:"min=200"`参数错误时，返回InvalidValidationError类型；校验错误时返回ValidationErrors
}

func (c CaptchaValidateRequest) String() string {
	str, _ := json.MarshalString(c)
	return str
}

// IsLegal 验证请求参数是否合法
// @return bool 参数是否合法
func (c CaptchaValidateRequest) IsLegal() bool {
	if stringer.IsBlank(c.CaptchaId) || stringer.IsBlank(c.Code) {
		return false
	}
	return true
}

// CaptchaGenerateRequest 生成验证码请求参数
type CaptchaGenerateRequest struct{}

// IsLegal 验证生成验证码请求参数是否合法
// @return bool 参数是否合法
func (g CaptchaGenerateRequest) IsLegal() bool {
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

func (g CaptchaGenerateRequest) String() string {
	str, _ := json.MarshalString(g)
	return str
}
