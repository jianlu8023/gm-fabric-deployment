package mapper

import (
	"github.com/jianlu8023/go-tools/v2/pkg/sqlnull"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"gorm.io/gorm"
)

// SystemMapper 系统数据访问接口
//
// @description 定义系统相关的数据访问方法，包括系统初始化状态的查询
// @interface
type SystemMapper interface {
	// GetSystemInit 获取系统初始化状态
	//
	// @description 获取系统的初始化状态，判断系统是否已经完成初始化配置
	// @return bool 系统是否已初始化
	// @return error 错误信息
	GetSystemInit() (bool, error)
}

// systemMapperImpl 系统数据访问层结构体
//
// @description 提供系统相关的数据访问操作
// @struct
type systemMapperImpl struct {
	*Mapper
}

// NewSystemMapper 创建一个新的SystemMapper实例
//
// @description 创建并返回一个新的SystemMapper实例，用于系统相关的数据访问操作
// @param mapper *Mapper 基础Mapper
// @return SystemMapper SystemMapper实例
func NewSystemMapper(mapper *Mapper) SystemMapper {

	m := &systemMapperImpl{Mapper: mapper}

	// 调用init方法
	// m.init()

	return m
}

// GetSystemInit 获取系统初始化状态
//
// @description 获取系统的初始化状态，判断系统是否已经完成初始化配置
// @return bool 系统是否已初始化
// @return error 错误信息
func (m *systemMapperImpl) GetSystemInit() (bool, error) {
	if m.db == nil {
		return false, datasource.ErrNoDataSourceConn
	}
	m.init()
	var systemInit model.SystemInit
	if err := m.db.Model(&model.SystemInit{}).Where(&model.SystemInit{
		Id: 1,
	}).First(&systemInit).Error; err != nil {
		return false, err
	}

	return systemInit.IsInit.Bool, nil
}

// init 初始化系统状态数据
//
// @description 初始化系统状态数据表，确保系统状态记录存在
func (m *systemMapperImpl) init() {
	if m.db == nil {
		return
	}
	if err := m.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.SystemInit{}).Where(&model.SystemInit{
			Id: 1,
		}).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := tx.Model(&model.SystemInit{}).Create(&model.SystemInit{
				Id:     1,
				IsInit: sqlnull.FalseToNull(),
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		m.logger.Errorf("init system status failed: %s", err)
		return
	}
}
