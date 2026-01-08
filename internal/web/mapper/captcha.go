package mapper

type CaptchaMapper interface{}

type captchaMapper struct {
	*Mapper
}

// nolint: unused
func NewCaptchaMapper(baseMapper *Mapper) CaptchaMapper {
	return &captchaMapper{
		Mapper: baseMapper,
	}
}
