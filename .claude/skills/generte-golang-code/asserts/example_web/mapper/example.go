package mapper

import (
	"context"
	"math"

	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
)

// ExampleMapper ExampleMapper接口
//
// interface
type ExampleMapper interface {
	// Pagination 分页查询
	//
	// @param ctx context.Context: 上下文
	// @param query model.ExampleModel: 查询条件
	// @param queryPage dbpage.Info[model.ExampleModel]: 分页信息
	//
	// @return dbpage.Info[model.ExampleModel]: 分页结果
	// @return error: 错误信息
	Pagination(ctx context.Context, query model.ExampleModel, queryPage dbpage.Info[model.ExampleModel]) (dbpage.Info[model.ExampleModel], error)
}

// exampleMapper
//
// struct
type exampleMapper struct {
	*Mapper // Mapper 基础mapper
}

// NewExampleMapper 创建ExampleMapper实例
//
// @param baseMapper *Mapper: Mapper实例
//
// @return ExampleMapper: ExampleMapper实例
func NewExampleMapper(baseMapper *Mapper) ExampleMapper {
	return &exampleMapper{
		Mapper: baseMapper,
	}
}

// Pagination 分页查询
//
// @param ctx context.Context: 上下文
// @param query model.ExampleModel: 查询条件
// @param queryPage dbpage.Info[model.ExampleModel]: 分页信息
//
// @return dbpage.Info[model.ExampleModel]: 分页结果
// @return error: 错误信息
func (m *exampleMapper) Pagination(ctx context.Context, query model.ExampleModel, queryPage dbpage.Info[model.ExampleModel]) (dbpage.Info[model.ExampleModel], error) {

	if m.db == nil {
		return queryPage, datasource.ErrNoDataSourceConn
	}

	// 计算偏移量
	offset := int((queryPage.PageNo - 1) * queryPage.PageSize)

	// 构建查询
	db := m.db.WithContext(ctx).Model(&model.ExampleModel{}).Where(&query)

	// 查询总记录数
	if err := db.WithContext(ctx).Count(&queryPage.Count).Error; err != nil {
		return queryPage, err
	}

	// 计算最大页码
	if queryPage.Count > 0 {
		queryPage.MaxPage = int64(math.Ceil(float64(queryPage.Count) / float64(queryPage.PageSize)))
	}

	// 查询记录
	records := make([]model.ExampleModel, 0)

	// 分页查询
	if err := db.WithContext(ctx).Offset(offset).
		Limit(int(queryPage.PageSize)).
		Find(&records).Error; err != nil {
		return queryPage, err
	}

	queryPage.Records = records
	return queryPage, nil

}
