package captcha

// Interface 定义验证码控制器的接口
//
// @description 该接口定义了验证码控制模块需要实现的核心功能
// @interface
type Interface interface {
	// GenerateCaptcha 生成验证码
	//
	// @return captchaID string 验证码ID
	// @return captchaImage string 验证码图片的Base64编码
	// @return answer string 验证码的答案
	// @return err error 生成过程中的错误
	GenerateCaptcha() (captchaID string, captchaImage string, answer string, err error)

	// ValidateCaptcha 验证验证码
	//
	// @param captchaID string 验证码ID
	// @param userInput string 用户输入的验证码
	// @return isValid bool 验证结果
	// @return err error 验证过程中的错误
	ValidateCaptcha(captchaID string, userInput string) (isValid bool, err error)

	// RefreshCaptcha 刷新验证码
	//
	// @return captchaID string 验证码ID
	// @return captchaImage string 验证码图片的Base64编码
	// @return answer string 验证码的答案
	// @return err error 刷新过程中的错误
	RefreshCaptcha() (captchaID string, captchaImage string, answer string, err error)
}
