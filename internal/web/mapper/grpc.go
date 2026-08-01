package mapper

// GrpcMapper gRPC数据访问接口
//
// @description 定义gRPC相关的数据访问方法，当前为空接口，预留后续扩展
// @interface
type GrpcMapper interface {
}

// grpcMapperImpl gRPC数据访问层结构体
//
// @description 提供gRPC相关的数据访问操作，当前为空实现，预留后续扩展
// @struct
type grpcMapperImpl struct {
	*Mapper
}

// NewGrpcMapper 创建一个新的GrpcMapper实例
//
// @description 创建并返回一个新的GrpcMapper实例，用于gRPC相关的数据访问操作
// @param baseMapper *Mapper 基础Mapper
// @return GrpcMapper GrpcMapper实例
func NewGrpcMapper(baseMapper *Mapper) GrpcMapper {
	return &grpcMapperImpl{
		Mapper: baseMapper,
	}
}
