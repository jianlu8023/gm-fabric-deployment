package mapper

type GrpcMapper interface {
}

type grpcMapper struct {
	*Mapper
}

func NewGrpcMapper(baseMapper *Mapper) GrpcMapper {
	return &grpcMapper{
		Mapper: baseMapper,
	}
}
