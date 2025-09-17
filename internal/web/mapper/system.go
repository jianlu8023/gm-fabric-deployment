package mapper

import (
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
)

// SystemMapper 系统数据访问层结构体
//
// @description 提供系统相关的数据访问操作
// @struct
type SystemMapper struct {
	*Mapper
}

func (m *SystemMapper) GetSystemInit() (bool, error) {
	if m.db == nil {
		return false, datasource.ErrNoDataSourceConn
	}
	var systemInit model.SystemInit
	if err := m.db.Model(&model.SystemInit{}).Where(&model.SystemInit{
		Id: 1,
	}).First(&systemInit).Error; err != nil {
		return false, err
	}

	return systemInit.IsInit.Bool, nil
}

// NewSystemMapper 创建一个新的SystemMapper实例
//
// @param mapper *Mapper 基础Mapper
// @return *SystemMapper SystemMapper实例
func NewSystemMapper(mapper *Mapper) *SystemMapper {
	return &SystemMapper{Mapper: mapper}
}
