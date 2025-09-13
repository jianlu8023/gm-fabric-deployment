package response

// CaptchaResponse 验证码响应数据
// @description 验证码生成响应的数据结构
// @property captchaId string 验证码唯一ID
// @property imgBase64 string 验证码图片的Base64编码
// @property captchaType string 验证码类型
// @property status string 状态信息
// @property message string 消息说明
type CaptchaResponse struct {
	CaptchaId   string `json:"captchaId"`
	ImgBase64   string `json:"imgBase64"`
	CaptchaType string `json:"captchaType"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// NewCaptchaResponse 创建验证码响应对象
// @param captchaId 验证码唯一ID
// @param imgBase64 验证码图片的Base64编码
// @param captchaType 验证码类型
// @param status 状态信息
// @param message 消息说明
// @return *CaptchaResponse 验证码响应对象
func NewCaptchaResponse(captchaId, imgBase64, captchaType, status, message string) *CaptchaResponse {
	return &CaptchaResponse{
		CaptchaId:   captchaId,
		ImgBase64:   imgBase64,
		CaptchaType: captchaType,
		Status:      status,
		Message:     message,
	}
}

// ValidateCaptchaResponse 验证码验证响应数据
// @description 验证码验证结果的响应数据结构
// @property valid bool 验证结果
// @property message string 消息说明
// @property captchaId string 验证码唯一ID
type ValidateCaptchaResponse struct {
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`
	CaptchaId string `json:"captchaId"`
}

// NewValidateCaptchaResponse 创建验证码验证响应对象
// @param valid 验证结果
// @param message 消息说明
// @param captchaId 验证码唯一ID
// @return *ValidateCaptchaResponse 验证码验证响应对象
func NewValidateCaptchaResponse(valid bool, message, captchaId string) *ValidateCaptchaResponse {
	return &ValidateCaptchaResponse{
		Valid:     valid,
		Message:   message,
		CaptchaId: captchaId,
	}
}
