package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// CaptchaGenerateResponse 验证码生成响应结构体
//
// @description 用于返回验证码生成结果的响应数据
// @struct
type CaptchaGenerateResponse struct {
	CaptchaId string `json:"captcha_id" yaml:"captcha_id"` // 验证码唯一ID
	ImgBase64 string `json:"img_base64" yaml:"img_base64"` // 验证码图片的Base64编码
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 自定义CaptchaGenerateResponse结构体的JSON序列化逻辑
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp CaptchaGenerateResponse) MarshalJSON() ([]byte, error) {
	type Alias CaptchaGenerateResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// String 将验证码生成响应转换为字符串表示
//
// @description 将CaptchaGenerateResponse结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp CaptchaGenerateResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewCaptchaGenerateResponse 创建验证码响应对象
//
// @description 创建一个新的验证码生成响应对象
// @param captchaId string 验证码唯一ID
// @param imgBase64 string 验证码图片的Base64编码
// @return *CaptchaGenerateResponse 验证码响应对象
func NewCaptchaGenerateResponse(captchaId, imgBase64 string) *CaptchaGenerateResponse {
	return &CaptchaGenerateResponse{
		CaptchaId: captchaId,
		ImgBase64: imgBase64,
	}
}

// CaptchaValidateResponse 验证码验证响应结构体
//
// @description 用于返回验证码验证结果的响应数据
// @struct
type CaptchaValidateResponse struct {
	Valid     bool   `json:"valid" yaml:"valid"`           // 验证结果
	CaptchaId string `json:"captcha_id" yaml:"captcha_id"` // 验证码唯一ID
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 自定义CaptchaValidateResponse结构体的JSON序列化逻辑
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp CaptchaValidateResponse) MarshalJSON() ([]byte, error) {
	type Alias CaptchaValidateResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// String 将验证码验证响应转换为字符串表示
//
// @description 将CaptchaValidateResponse结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp CaptchaValidateResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewCaptchaValidateResponse 创建验证码验证响应对象
//
// @description 创建一个新的验证码验证响应对象
// @param valid bool 验证结果
// @param captchaId string 验证码唯一ID
// @return *CaptchaValidateResponse 验证码验证响应对象
func NewCaptchaValidateResponse(valid bool, captchaId string) *CaptchaValidateResponse {
	return &CaptchaValidateResponse{
		Valid:     valid,
		CaptchaId: captchaId,
	}
}

