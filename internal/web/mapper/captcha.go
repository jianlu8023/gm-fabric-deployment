package mapper

// CaptchaMapper 验证码数据访问接口
//
// @description 定义验证码相关的数据访问方法，当前为空接口，预留后续扩展
// @interface
type CaptchaMapper interface{}

// captchaMapperImpl 验证码数据访问层结构体
//
// @description 提供验证码相关的数据访问操作，当前为空实现，预留后续扩展
// @struct
type captchaMapperImpl struct {
	*Mapper
}

// NewCaptchaMapper 创建一个新的CaptchaMapper实例
//
// @description 创建并返回一个新的CaptchaMapper实例，用于验证码相关的数据访问操作
// @param baseMapper *Mapper 基础Mapper
// @return CaptchaMapper CaptchaMapper实例
//
// nolint: unused
func NewCaptchaMapper(baseMapper *Mapper) CaptchaMapper {
	return &captchaMapperImpl{
		Mapper: baseMapper,
	}
}
