package captcha

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.CaptchaConfig {
	return &config.CaptchaConfig{
		Enabled:     false,
		MaxAge:      300,      // 验证码有效期，单位秒
		Width:       240,      // 验证码图片宽度
		Height:      80,       // 验证码图片高度
		MaxSkew:     0.8,      // 最大倾斜度
		DotCount:    20,       // 干扰点数量
		LineCount:   3,        // 干扰线数量
		DefaultType: "number", // 默认验证码类型：string, number, math, chinese
	}
}
