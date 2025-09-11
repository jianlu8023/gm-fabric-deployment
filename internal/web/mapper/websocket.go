package mapper

import (
	"errors"

	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/dbpage"
	"gorm.io/gorm"
)

// WebSocketMapper 提供WebSocket连接相关的数据库操作

type WebSocketMapper struct {
	*Mapper
}

// NewWebSocketMapper 创建一个新的WebSocketMapper实例
func NewWebSocketMapper(mapper *Mapper) *WebSocketMapper {
	return &WebSocketMapper{
		Mapper: mapper,
	}
}

// SaveConnection 保存WebSocket连接信息
func (m *WebSocketMapper) SaveConnection(conn *model.ConnectionInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Model(&model.ConnectionInfo{}).Create(conn).Error
}

// UpdateConnectionStatus 更新WebSocket连接状态
func (m *WebSocketMapper) UpdateConnectionStatus(id string, status string) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Model(&model.ConnectionInfo{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateConnectionLastActive 更新WebSocket连接的最后活动时间
func (m *WebSocketMapper) UpdateConnectionLastActive(id string) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Model(&model.ConnectionInfo{}).Where("id = ?", id).Update("last_active", gorm.Expr("NOW()")).Error
}

// GetConnectionByID 根据ID获取WebSocket连接信息
func (m *WebSocketMapper) GetConnectionByID(id string) (*model.ConnectionInfo, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}
	conn := &model.ConnectionInfo{}
	err := m.db.Model(&model.ConnectionInfo{}).Where("id = ?", id).First(conn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("连接不存在")
		}
		return nil, err
	}
	return conn, nil
}

// GetConnectionsByUserID 根据用户ID获取WebSocket连接列表
func (m *WebSocketMapper) GetConnectionsByUserID(userID string, isPage bool, pageNo int64, pageSize int64) (dbpage.Info[model.ConnectionInfo], error) {
	page := dbpage.Info[model.ConnectionInfo]{}
	if m.db == nil {
		return page, datasource.ErrNoDataSourceConn
	}
	page.PageNo = pageNo
	page.PageSize = pageSize

	// 计算偏移量
	offset := (pageNo - 1) * pageSize

	// 构建查询
	db := m.db.Model(&model.ConnectionInfo{}).Where("user_id = ?", userID)

	// 查询总记录数
	if err := db.Count(&page.Count).Error; err != nil {
		return page, err
	}

	// 计算最大页码
	if page.Count > 0 {
		page.MaxPage = (page.Count + pageSize - 1) / pageSize
	}

	// 查询记录
	records := make([]model.ConnectionInfo, 0)
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

// DeleteConnection 删除WebSocket连接信息
func (m *WebSocketMapper) DeleteConnection(id string) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Model(&model.ConnectionInfo{}).Where("id = ?", id).Delete(nil).Error
}
