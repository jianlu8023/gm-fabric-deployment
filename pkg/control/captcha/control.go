package captcha

import (
	"errors"
	"sync"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/mojocn/base64Captcha"
	"go.uber.org/zap"
)

// Control 验证码控制器结构体
type Control struct {
	config  *config.CaptchaConfig
	logger  *zap.SugaredLogger
	store   base64Captcha.Store
	captcha *base64Captcha.Captcha
	drivers map[string]base64Captcha.Driver
	once    sync.Once
}

// NewCaptchaControl 创建一个新的验证码控制器
//
// @param captchaConfig *config.CaptchaConfig 验证码配置
// @param loggerControl *logger.Control 日志控制器
//
// @return *Control 验证码控制器
func NewCaptchaControl(captchaConfig *config.CaptchaConfig, loggerControl *logger.Control) *Control {
	if captchaConfig == nil {
		captchaConfig = getDefaultConfig()
	}
	if !captchaConfig.Enabled {
		return nil
	}

	captchaLogger := loggerControl.GenLogger(logger.ModuleCaptcha)
	captchaLogger.Infof("[control] starting new captcha control...")

	// 初始化内存存储
	store := base64Captcha.DefaultMemStore
	if captchaConfig.MaxAge > 0 {
		store = base64Captcha.DefaultMemStore
	}

	// 初始化驱动映射
	drivers := make(map[string]base64Captcha.Driver)

	captcha := &Control{
		config:  captchaConfig,
		logger:  captchaLogger,
		store:   store,
		drivers: drivers,
	}

	return captcha
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Debugf("[control] starting captcha server...")
			// 初始化驱动
			c.initDrivers()

			c.logger.Debugf("[control] generate captcha...")
			c.captcha = base64Captcha.NewCaptcha(c.getDriver(c.config.DefaultType),
				c.store)
		}
	})
}

func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutting down captcha server...")
	return nil
}

// initDrivers 初始化验证码驱动
func (c *Control) initDrivers() {
	c.logger.Debugf("[control] starting init drivers...")
	// 字符串验证码驱动
	// c.logger.Debugf("[control] starting string drivers...")
	// c.drivers["string"] = c.createStringDriver()
	// 数字验证码驱动
	c.logger.Debugf("[control] starting number drivers...")
	c.drivers["number"] = c.createDigitDriver()
	// 数学计算验证码驱动
	// c.logger.Debugf("[control] starting math drivers...")
	// c.drivers["math"] = c.createMathDriver()
	// 中文验证码驱动
	// c.logger.Debugf("[control] starting chinese drivers...")
	// c.drivers["chinese"] = c.createChineseDriver()

}

// getDriver 根据类型获取验证码驱动
func (c *Control) getDriver(captchaType string) base64Captcha.Driver {
	// 如果类型为空，使用默认类型
	if captchaType == "" {
		captchaType = c.config.DefaultType
		if captchaType == "" {
			captchaType = "string"
		}
	}

	driver, exists := c.drivers[captchaType]
	if !exists {
		// 如果请求的类型不存在，使用字符串类型作为默认
		driver = c.drivers["string"]
	}

	return driver
}

// // createStringDriver 创建字符串验证码驱动
// func (c *Control) createStringDriver() base64Captcha.Driver {
// 	driverString := base64Captcha.NewDriverString(10,
// 		60,
// 		int(c.config.MaxSkew),
// 		c.config.DotCount,
// 		c.config.LineCount,
// 		"1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
// 		nil,
// 		nil,
// 		[]string{})
// 	return driverString.ConvertFonts()
// }

// createDigitDriver 创建数字验证码驱动
func (c *Control) createDigitDriver() base64Captcha.Driver {
	driverDigit := base64Captcha.NewDriverDigit(c.config.Height,
		c.config.Width,
		6,
		c.config.MaxSkew,
		c.config.DotCount,
	)
	return driverDigit
}

// // createMathDriver 创建数学计算验证码驱动
// func (c *Control) createMathDriver() base64Captcha.Driver {
// 	// 80, 240, 5,
// 	// OptionShowSineLine | OptionShowSlimeLine | OptionShowHollowLine,
// 	// nil, []string{"3Dumb.ttf"}
// 	driverMath := base64Captcha.NewDriverMath(
// 		c.config.Height,
// 		c.config.Width,
// 		int(c.config.MaxSkew),
// 		base64Captcha.OptionShowSineLine|base64Captcha.OptionShowSlimeLine|base64Captcha.OptionShowHollowLine,
// 		nil,
// 		nil,
// 		[]string{},
// 	)
// 	return driverMath
// }

// // createChineseDriver 创建中文验证码驱动
// func (c *Control) createChineseDriver() base64Captcha.Driver {
// 	fontPath := "resource/font/chinese.ttf"
// 	chineseChars := "你好世界一二三四五六七八九十百千万东南西北中发白"
// 	driverChinese := base64Captcha.NewDriverChinese(
// 		c.config.Height,
// 		c.config.Width,
// 		1,
// 		int(c.config.MaxSkew),
// 		c.config.DotCount,
// 		strconv.Itoa(c.config.LineCount),
// 		chineseChars,
// 		fontPath,
// 		[]string{},
// 		)
// 	return driverChinese
// }

// GenerateCaptcha 生成验证码
func (c *Control) GenerateCaptcha() (string, string, string, error) {
	c.logger.Debugf("[control] starting generate a new captcha...")

	// 生成验证码
	id, b64s, answer, err := c.captcha.Generate()
	if err != nil {
		c.logger.Errorf("[control] generate captcha failed: %v", err)
		return "", "", "", err
	}

	c.logger.Debugf("[control] captcha generated successfully, id: %s", id)
	return id, b64s, answer, nil
}

// ValidateCaptcha 验证验证码
func (c *Control) ValidateCaptcha(captchaID string, userInput string) (bool, error) {
	c.logger.Infof("Validating captcha, id: %s, user input: %s", captchaID, userInput)

	if captchaID == "" || userInput == "" {
		c.logger.Error("Captcha ID or user input is empty")
		return false, errors.New("captcha id or user input cannot be empty")
	}

	// 验证验证码
	isValid := c.store.Verify(captchaID, userInput, true)

	if isValid {
		c.logger.Debugf("Captcha validated successfully, id: %s", captchaID)
	} else {
		c.logger.Warnf("Captcha validation failed, id: %s", captchaID)
	}

	return isValid, nil
}

// RefreshCaptcha 刷新验证码（与生成验证码逻辑相同）
func (c *Control) RefreshCaptcha() (string, string, string, error) {
	return c.GenerateCaptcha()
}
