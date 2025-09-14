package mapper

type GrpcMapper struct {
	*Mapper
}

func NewGrpcMapper(baseMapper *Mapper) *GrpcMapper {
	return &GrpcMapper{
		Mapper: baseMapper,
	}
}
