package mapper

import (
	"errors"
	"math"

	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"gorm.io/gorm"
)

// Libp2pNodeMapper 节点数据访问层结构体
//
// @description 提供节点相关的数据访问操作
// @struct
type Libp2pNodeMapper struct {
	*Mapper
}

// NewLibp2pNodeMapper 创建一个新的NodeMapper实例
//
// @param mapper *Mapper 基础Mapper
// @return *Libp2pNodeMapper NodeMapper实例
func NewLibp2pNodeMapper(mapper *Mapper) *Libp2pNodeMapper {
	return &Libp2pNodeMapper{
		Mapper: mapper,
	}
}

// NodeList 查询节点列表
//
// @param query node.Info 查询条件
// @param isPage bool 是否分页
// @param pageNo int64 页码
// @param pageSize int64 每页大小
// @return dbpage.Info[node.Info] 节点列表
// @return error 错误信息
func (m *Libp2pNodeMapper) NodeList(query model.Libp2pNode,
	isPage bool, pageNo int, pageSize int,
) (dbpage.Info[model.Libp2pNode], error) {
	page := dbpage.Info[model.Libp2pNode]{}
	if m.db == nil {
		return page, datasource.ErrNoDataSourceConn
	}
	page.PageNo = int64(pageNo)
	page.PageSize = int64(pageSize)

	// 计算偏移量
	offset := (pageNo - 1) * pageSize

	// 构建查询
	db := m.db.Model(&model.Libp2pNode{}).Where(&query)

	// 查询总记录数
	if err := db.Count(&page.Count).Error; err != nil {
		return page, err
	}

	// 计算最大页码
	if page.Count > 0 {
		page.MaxPage = int64(math.Ceil(float64(page.Count) / float64(page.PageSize)))
	}

	// 查询记录
	records := make([]model.Libp2pNode, 0)
	if isPage {
		// 分页查询
		if err := db.Offset(int(offset)).
			Limit(int(pageSize)).
			Find(&records).Error; err != nil {
			return page, err
		}
	} else {
		// 不分页查询
		if err := db.Find(&records).Error; err != nil {
			return page, err
		}
	}

	page.Records = records
	return page, nil
}

// InsertOneWithCheck 插入节点并检查是否已存在
//
// @param record *node.Info 节点信息
// @return error 错误信息，如果节点已存在返回ErrAlreadyExists
func (m *Libp2pNodeMapper) InsertOneWithCheck(record *model.Libp2pNode) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Libp2pNode{}).Where(&model.Libp2pNode{
			NodeId: record.NodeId,
		}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			// 节点已存在
			return datasource.ErrAlreadyExists
		}
		// 节点不存在，插入
		if err := tx.Model(&model.Libp2pNode{}).Create(record).
			Error; err != nil {
			return err
		}
		return nil
	})
}

// InsertOrUpdate 插入或更新节点信息
//
// @param record *node.Info 节点信息
// @return error 错误信息
func (m *Libp2pNodeMapper) InsertOrUpdate(record *model.Libp2pNode) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var existInfo model.Libp2pNode
		if err := tx.Model(&model.Libp2pNode{}).Where(&model.Libp2pNode{
			NodeId: record.NodeId,
		}).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在
				if err := tx.Model(&model.Libp2pNode{}).
					Create(record).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新
			record.AutoUid = existInfo.AutoUid
			if err := tx.Model(&model.Libp2pNode{}).
				Where(&model.Libp2pNode{
					AutoUid: record.AutoUid,
				}).
				Updates(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
