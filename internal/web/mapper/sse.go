package mapper

// SSEMapper Server-Sent Events数据访问接口
//
// @description 定义Server-Sent Events相关的数据访问方法，当前为空接口，预留后续扩展
// @interface
type SSEMapper interface {
}

// sSEMapperImpl SSE数据访问层结构体
//
// @description 提供Server-Sent Events相关的数据访问操作
// @struct
type sSEMapperImpl struct {
	*Mapper
}

// NewSSEMapper 创建一个新的SSEMapper实例
//
// @param baseMapper *Mapper 基础Mapper
// @return SSEMapper SSEMapper实例
func NewSSEMapper(baseMapper *Mapper) SSEMapper {
	return &sSEMapperImpl{
		Mapper: baseMapper,
	}
}
