package mapper

type SSEMapper struct {
	*Mapper
}

func NewSSEMapper(baseMapper *Mapper) *SSEMapper {
	return &SSEMapper{
		Mapper: baseMapper,
	}
}
