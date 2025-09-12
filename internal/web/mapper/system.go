package mapper

// SystemMapper 系统管理
type SystemMapper struct {
	*Mapper
}

// NewSystemMapper 创建系统mapper
// @param *mapper *Mapper 基础mapper
// @return *SystemMapper 系统mapper
func NewSystemMapper(mapper *Mapper) *SystemMapper {
	return &SystemMapper{Mapper: mapper}
}
