package datasource

// CustomTableName 自定义表名接口
// @description 定义实体模型自定义表名的接口
type CustomTableName interface {
	// TableName 获取表名
	// @return string 表名
	TableName() string
}
