package mapper

// SystemMapper 系统数据访问层结构体
//
// @description 提供系统相关的数据访问操作
// @struct
//
type SystemMapper struct {
	*Mapper
}

// NewSystemMapper 创建一个新的SystemMapper实例
//
// @param mapper *Mapper 基础Mapper
// @return *SystemMapper SystemMapper实例
//
func NewSystemMapper(mapper *Mapper) *SystemMapper {
	return &SystemMapper{Mapper: mapper}
}
