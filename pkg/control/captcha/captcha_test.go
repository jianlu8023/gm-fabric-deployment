package captcha

import (
	"log"
	"testing"

	"github.com/mojocn/base64Captcha"
)

// configJsonBody json request body.
type configJsonBody struct {
	Id            string
	CaptchaType   string
	VerifyValue   string
	DriverAudio   *base64Captcha.DriverAudio
	DriverString  *base64Captcha.DriverString
	DriverChinese *base64Captcha.DriverChinese
	DriverMath    *base64Captcha.DriverMath
	DriverDigit   *base64Captcha.DriverDigit
}

var store = base64Captcha.DefaultMemStore

// base64Captcha create return id, b64s, err
func GetCaptcha() (string, string, string, error) {

	// 	{
	// 		ShowLineOptions: [],
	// 		CaptchaType: "string",
	// 		Id: '',
	// 		VerifyValue: '',
	// 		DriverAudio: {
	// 			Length: 6,
	// 			Language: 'zh'
	// 		},
	// 		DriverString: {
	// 			Height: 60,
	// 			Width: 240,
	// 			ShowLineOptions: 0,
	// 			NoiseCount: 0,
	// 			Source: "1234567890qwertyuioplkjhgfdsazxcvbnm",
	// 			Length: 6,
	// 			Fonts: ["wqy-microhei.ttc"],
	// 			BgColor: {R: 0, G: 0, B: 0, A: 0},
	// 		},
	// 		DriverMath: {
	// 			Height: 60,
	// 			Width: 240,
	// 			ShowLineOptions: 0,
	// 			NoiseCount: 0,
	// 			Length: 6,
	// 			Fonts: ["wqy-microhei.ttc"],
	// 			BgColor: {R: 0, G: 0, B: 0, A: 0},
	// 		},
	// 		DriverChinese: {
	// 			Height: 60,
	// 			Width: 320,
	// 			ShowLineOptions: 0,
	// 			NoiseCount: 0,
	// 			Source: "设想,你在,处理,消费者,的音,频输,出音,频可,能无,论什,么都,没有,任何,输出,或者,它可,能是,单声道,立体声,或是,环绕立,体声的,,不想要,的值",
	// 			Length: 2,
	// 			Fonts: ["wqy-microhei.ttc"],
	// 			BgColor: {R: 125, G: 125, B: 0, A: 118},
	// 		},
	// 		DriverDigit: {
	// 			Height: 80,
	// 			Width: 240,
	// 			Length: 5,
	// 			MaxSkew: 0.7,
	// 			DotCount: 80
	// 		}
	// 	},
	// 	blob: "",
	// 	loading: false
	// }

	// https://captcha.mojotv.cn/ 调试配置
	var param = configJsonBody{
		Id:          "",
		CaptchaType: "string",
		VerifyValue: "",
		DriverAudio: &base64Captcha.DriverAudio{},
		DriverString: &base64Captcha.DriverString{
			Length:          4,
			Height:          60,
			Width:           240,
			ShowLineOptions: 2,
			NoiseCount:      0,
			Source:          "1234567890qwertyuioplkjhgfdsazxcvbnm",
		},
		DriverChinese: &base64Captcha.DriverChinese{},
		DriverMath:    &base64Captcha.DriverMath{},
		DriverDigit:   &base64Captcha.DriverDigit{},
	}
	var driver base64Captcha.Driver

	// create base64 encoding captcha
	switch param.CaptchaType {
	case "audio":
		driver = param.DriverAudio
	case "string":
		driver = param.DriverString.ConvertFonts()
	case "math":
		driver = param.DriverMath.ConvertFonts()
	case "chinese":
		driver = param.DriverChinese.ConvertFonts()
	default:
		driver = param.DriverDigit
	}
	c := base64Captcha.NewCaptcha(driver, store)
	return c.Generate()
	// id, b64s, err := c.Generate()

	// body := map[string]interface{}{"code": 1, "data": b64s, "captchaId": id, "msg": "success"}
	// if err != nil {
	// 	body = map[string]interface{}{"code": 0, "msg": err.Error()}
	// }
	// var _ = body

	// // log.Println(body)
	// log.Println(1)
	// log.Println(id)

	// log.Printf("store =%+v\n", store)
}

// base64Captcha verify
func VerifyCaptcha(id, VerifyValue string) bool {
	return store.Verify(id, VerifyValue, true)
}

func TestMyCaptcha(t *testing.T) {
	id, b64s, ans, err := GetCaptcha()
	if err != nil {
		return
	}
	var _ = b64s
	log.Println("id =", id)
	log.Println("VerifyValue =", store.Get(id, true))
	result := VerifyCaptcha(id, ans)
	log.Println("result =", result)
}
