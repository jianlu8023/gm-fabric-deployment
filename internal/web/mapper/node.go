package mapper

import (
	"errors"
	"github.com/jianlu8023/gm-fabric-deployment/internal/datasource/node"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"gorm.io/gorm"
)

type NodeMapper struct {
	*Mapper
	db *gorm.DB
}

func NewNodeMapper(mapper *Mapper, db *gorm.DB) *NodeMapper {
	return &NodeMapper{
		Mapper: mapper,
		db:     db,
	}
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
