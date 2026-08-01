package mapper

import (
	"context"
	"errors"
	"math"

	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

// Libp2pNodeMapper 节点数据访问接口
//
// @description 定义libp2p节点相关的数据访问方法，包括节点列表查询、本机节点查询及节点的插入与更新
// @interface
type Libp2pNodeMapper interface {
	// NodeList 查询节点列表
	//
	// @description 根据查询条件获取节点列表，支持分页和不分页查询
	// @param ctx context.Context 上下文
	// @param query model.Libp2pNode 查询条件
	// @param isPage bool 是否分页
	// @param pageNo int 页码
	// @param pageSize int 每页大小
	// @return dbpage.Info[model.Libp2pNode] 节点列表分页结果
	// @return error 错误信息
	NodeList(ctx context.Context, query model.Libp2pNode, isPage bool, pageNo int, pageSize int) (dbpage.Info[model.Libp2pNode], error)
	// NodeMyself 查询本机节点信息
	//
	// @description 根据peerId查询本机节点的信息
	// @param ctx context.Context 上下文
	// @param peerId string 节点PeerID
	// @return model.Libp2pNode 节点信息
	// @return error 错误信息
	NodeMyself(ctx context.Context, peerId string) (model.Libp2pNode, error)
	// InsertOneWithCheck 插入节点并检查是否已存在
	//
	// @description 在事务中插入节点信息，若节点已存在则返回datasource.ErrAlreadyExists
	// @param ctx context.Context 上下文
	// @param record *model.Libp2pNode 节点信息
	// @return error 错误信息，如果节点已存在返回ErrAlreadyExists
	InsertOneWithCheck(ctx context.Context, record *model.Libp2pNode) error
	// InsertOrUpdate 插入或更新节点信息
	//
	// @description 在事务中插入或更新节点信息，根据NodeId判断记录是否存在
	// @param ctx context.Context 上下文
	// @param record *model.Libp2pNode 节点信息
	// @return error 错误信息
	InsertOrUpdate(ctx context.Context, record *model.Libp2pNode) error
}

// libp2pNodeMapperImpl 节点数据访问层结构体
//
// @description 提供节点相关的数据访问操作
// @struct
type libp2pNodeMapperImpl struct {
	*Mapper
}

// NewLibp2pNodeMapper 创建一个新的NodeMapper实例
//
// @description 创建并返回一个新的Libp2pNodeMapper实例，用于节点相关的数据访问操作
// @param mapper *Mapper 基础Mapper
// @return Libp2pNodeMapper NodeMapper实例
func NewLibp2pNodeMapper(mapper *Mapper) Libp2pNodeMapper {
	return &libp2pNodeMapperImpl{
		Mapper: mapper,
	}
}

// NodeList 查询节点列表
//
// @description 根据查询条件获取节点列表，支持分页和不分页查询，并集成链路追踪
// @param ctx context.Context 上下文
// @param query node.Info 查询条件
// @param isPage bool 是否分页
// @param pageNo int64 页码
// @param pageSize int64 每页大小
// @return dbpage.Info[node.Info] 节点列表
// @return error 错误信息
func (m *libp2pNodeMapperImpl) NodeList(ctx context.Context, query model.Libp2pNode, isPage bool, pageNo int, pageSize int) (dbpage.Info[model.Libp2pNode], error) {
	_, span := tracer.StartSpan(ctx, "libp2pNodeMapperImpl", "list")
	defer span.End()
	span.SetAttributes(
		attribute.String("query", query.String()),
	)
	page := dbpage.Info[model.Libp2pNode]{}
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return page, datasource.ErrNoDataSourceConn
	}
	page.PageNo = int64(pageNo)
	page.PageSize = int64(pageSize)

	// 计算偏移量
	offset := (pageNo - 1) * pageSize

	// 构建查询
	db := m.db.WithContext(ctx).Model(&model.Libp2pNode{}).Where(&query)

	// 查询总记录数
	if err := db.WithContext(ctx).Count(&page.Count).Error; err != nil {
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
		if err := db.WithContext(ctx).Offset(offset).
			Limit(pageSize).
			Find(&records).Error; err != nil {
			return page, err
		}
	} else {
		// 不分页查询
		if err := db.WithContext(ctx).Find(&records).Error; err != nil {
			return page, err
		}
	}
	span.SetStatus(codes.Ok, "success")
	page.Records = records
	return page, nil
}

// InsertOneWithCheck 插入节点并检查是否已存在
//
// @description 在事务中插入节点信息，若NodeId已存在则返回datasource.ErrAlreadyExists，并集成链路追踪
// @param ctx context.Context 上下文
// @param record *node.Info 节点信息
// @return error 错误信息，如果节点已存在返回ErrAlreadyExists
func (m *libp2pNodeMapperImpl) InsertOneWithCheck(ctx context.Context, record *model.Libp2pNode) error {
	_, span := tracer.StartSpan(ctx, "libp2pNodeMapperImpl", "insertOrUpdateOneWithCheck")
	defer span.End()
	span.SetAttributes(
		attribute.String("insert", record.String()),
	)
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return datasource.ErrNoDataSourceConn
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.WithContext(ctx).Model(&model.Libp2pNode{}).Where(&model.Libp2pNode{
			NodeId: record.NodeId,
		}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			// 节点已存在
			span.RecordError(datasource.ErrAlreadyExists)
			span.SetStatus(codes.Error, datasource.ErrAlreadyExists.Error())
			return datasource.ErrAlreadyExists
		}
		// 节点不存在，插入
		if err := tx.WithContext(ctx).Model(&model.Libp2pNode{}).Create(record).
			Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		return nil
	})
}

// InsertOrUpdate 插入或更新节点信息
//
// @description 在事务中插入或更新节点信息，根据NodeId判断记录是否存在，存在则更新，不存在则插入
// @param ctx context.Context 上下文
// @param record *node.Info 节点信息
// @return error 错误信息
func (m *libp2pNodeMapperImpl) InsertOrUpdate(ctx context.Context, record *model.Libp2pNode) error {
	_, span := tracer.StartSpan(ctx, "libp2pNodeMapperImpl", "insertOrUpdate")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return datasource.ErrNoDataSourceConn
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existInfo model.Libp2pNode
		if err := tx.WithContext(ctx).Model(&model.Libp2pNode{}).Where(&model.Libp2pNode{
			NodeId: record.NodeId,
		}).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在
				if err := tx.WithContext(ctx).Model(&model.Libp2pNode{}).
					Create(record).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新
			record.AutoUid = existInfo.AutoUid
			if err := tx.WithContext(ctx).Model(&model.Libp2pNode{}).
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

// NodeMyself 查询本机节点信息
//
// @description 根据peerId查询本机节点的信息，并集成链路追踪
// @param ctx context.Context 上下文
// @param peerId string 节点PeerID
// @return model.Libp2pNode 节点信息
// @return error 错误信息
func (m *libp2pNodeMapperImpl) NodeMyself(ctx context.Context, peerId string) (model.Libp2pNode, error) {
	_, span := tracer.StartSpan(ctx, "libp2pNodeMapperImpl", "myself")
	defer span.End()
	span.SetAttributes(
		attribute.String("node_id", peerId),
	)
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return model.Libp2pNode{}, datasource.ErrNoDataSourceConn
	}
	var node model.Libp2pNode
	if err := m.db.WithContext(ctx).Model(&model.Libp2pNode{}).Where(&model.Libp2pNode{
		NodeId: peerId,
	}).First(&node).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return model.Libp2pNode{}, err
	}
	span.SetStatus(codes.Ok, "success")
	return node, nil
}
