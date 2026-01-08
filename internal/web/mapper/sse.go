package mapper

type SSEMapper interface {
}

// sSEMapper SSE数据访问层结构体
//
// @description 提供Server-Sent Events相关的数据访问操作
// @struct
type sSEMapper struct {
	*Mapper
}

// NewSSEMapper 创建一个新的SSEMapper实例
//
// @param baseMapper *Mapper 基础Mapper
// @return SSEMapper SSEMapper实例
func NewSSEMapper(baseMapper *Mapper) SSEMapper {
	return &sSEMapper{
		Mapper: baseMapper,
	}
}
