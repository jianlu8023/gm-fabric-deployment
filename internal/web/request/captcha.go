package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// CaptchaValidateRequest 验证码验证请求参数
// @description 用于验证用户输入的验证码是否正确
// @struct
// @property captchaId string 验证码ID
// @property code string 用户输入的验证码（必填，4-6位）
type CaptchaValidateRequest struct {
	CaptchaId string `json:"captcha_id,omitempty" yaml:"captcha_id,omitempty" form:"captchaId" binding:"required"`
	Code      string `json:"code,omitempty" yaml:"code,omitempty" form:"code"  binding:"required,min=4,max=6"`
}

// String 将验证码验证请求参数转换为字符串表示
// @description 将CaptchaValidateRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req CaptchaValidateRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证请求参数是否合法
// @description 检查验证码验证请求的参数是否满足基本要求
// @return bool 参数是否合法
func (req CaptchaValidateRequest) IsLegal() bool {
	if stringer.IsBlank(req.CaptchaId) || stringer.IsBlank(req.Code) {
		return false
	}
	return true
}

// CaptchaGenerateRequest 生成验证码请求参数
// @description 用于请求生成新的验证码图片
// @struct
type CaptchaGenerateRequest struct{}

// IsLegal 验证生成验证码请求参数是否合法
// @description 检查生成验证码请求的参数是否满足基本要求
// @return bool 参数是否合法
func (req CaptchaGenerateRequest) IsLegal() bool {
	return true
}

// String 将生成验证码请求参数转换为字符串表示
// @description 将CaptchaGenerateRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req CaptchaGenerateRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
