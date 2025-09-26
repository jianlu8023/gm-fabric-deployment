package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// CaptchaGenerateResponse 验证码响应数据
// @description 验证码生成响应的数据结构
// @property captchaId string 验证码唯一ID
// @property imgBase64 string 验证码图片的Base64编码
// @property captchaType string 验证码类型
// @property status string 状态信息
// @property message string 消息说明
type CaptchaGenerateResponse struct {
	CaptchaId string `json:"captcha_id,omitempty" yaml:"captcha_id,omitempty"`
	ImgBase64 string `json:"img_base64,omitempty" yaml:"img_base64,omitempty"`
}

func (resp CaptchaGenerateResponse) MarshalJSON() ([]byte, error) {
	type Alias CaptchaGenerateResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

func (resp CaptchaGenerateResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewCaptchaGenerateResponse 创建验证码响应对象
// @param captchaId 验证码唯一ID
// @param imgBase64 验证码图片的Base64编码
// @param captchaType 验证码类型
// @param status 状态信息
// @param message 消息说明
// @return *CaptchaGenerateResponse 验证码响应对象
func NewCaptchaGenerateResponse(captchaId, imgBase64 string) *CaptchaGenerateResponse {
	return &CaptchaGenerateResponse{
		CaptchaId: captchaId,
		ImgBase64: imgBase64,
	}
}

// CaptchaValidateResponse 验证码验证响应数据
// @description 验证码验证结果的响应数据结构
// @property valid bool 验证结果
// @property message string 消息说明
// @property captchaId string 验证码唯一ID
type CaptchaValidateResponse struct {
	Valid     bool   `json:"valid"`
	CaptchaId string `json:"captchaId"`
}

func (resp CaptchaValidateResponse) MarshalJSON() ([]byte, error) {
	type Alias CaptchaValidateResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

func (resp CaptchaValidateResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewCaptchaValidateResponse 创建验证码验证响应对象
// @param valid 验证结果
// @param message 消息说明
// @param captchaId 验证码唯一ID
// @return *CaptchaValidateResponse 验证码验证响应对象
func NewCaptchaValidateResponse(valid bool, captchaId string) *CaptchaValidateResponse {
	return &CaptchaValidateResponse{
		Valid:     valid,
		CaptchaId: captchaId,
	}
}
