package mapper

type CaptchaMapper struct {
	*Mapper
}

// nolint: unused
func NewCaptchaMapper(baseMapper *Mapper) *CaptchaMapper {
	return &CaptchaMapper{
		Mapper: baseMapper,
	}
}
