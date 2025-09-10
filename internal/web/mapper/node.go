package mapper

import (
	"errors"

	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model/node"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/dbpage"
	"gorm.io/gorm"
)

type NodeMapper struct {
	*Mapper
}

func NewNodeMapper(mapper *Mapper) *NodeMapper {
	return &NodeMapper{
		Mapper: mapper,
	}
}

// NodeList 查询节点列表
// @param query 查询条件
// @param isPage 是否分页
// @param pageNo 页码
// @param pageSize 每页大小
// @return dbpage.Info[node.Info] 节点列表
// @return error 错误信息
func (m *NodeMapper) NodeList(query node.Info, isPage bool, pageNo int64, pageSize int64) (dbpage.Info[node.Info], error) {
	page := dbpage.Info[node.Info]{}
	if m.db == nil {
		return page, datasource.ErrNoDataSourceConn
	}
	page.PageNo = pageNo
	page.PageSize = pageSize

	// 计算偏移量
	offset := (pageNo - 1) * pageSize

	// 构建查询
	db := m.db.Model(&node.Info{}).Where(&query)

	// 查询总记录数
	if err := db.Count(&page.Count).Error; err != nil {
		return page, err
	}

	// 计算最大页码
	if page.Count > 0 {
		page.MaxPage = (page.Count + pageSize - 1) / pageSize
	}

	// 查询记录
	records := make([]node.Info, 0)
	if isPage {
		// 分页查询
		if err := db.Offset(int(offset)).Limit(int(pageSize)).Find(&records).Error; err != nil {
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

func (m *NodeMapper) InsertOneWithCheck(record *node.Info) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&node.Info{}).Where(&node.Info{
			NodeId: record.NodeId,
		}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			// 节点已存在
			return datasource.ErrAlreadyExists
		}
		// 节点不存在，插入
		if err := tx.Model(&node.Info{}).Create(record).Error; err != nil {
			return err
		}
		return nil
	})
}

func (m *NodeMapper) InsertOrUpdate(record *node.Info) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var existInfo node.Info
		if err := tx.Model(&node.Info{}).Where(&node.Info{
			NodeId: record.NodeId,
		}).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在
				if err := tx.Model(&node.Info{}).Create(record).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新
			record.AutoUid = existInfo.AutoUid
			if err := tx.Model(&node.Info{}).Where(&node.Info{AutoUid: record.AutoUid}).Updates(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
