package model

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/golang-example/internal/web/model"
)

const (
	// exampleModelTableName 表名定义
	exampleModelTableName = "t_example_model"
)

// ExampleModel example模型
//
// struct
type ExampleModel struct {
	model.Model
	AutoUid      int    `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`       // 自增id
	ExampleKey   string `json:"example_key,omitempty" yaml:"example_key,omitempty" gorm:"column:example_key;type:varchar(255);"`       // example key
	ExampleValue string `json:"example_value,omitempty" yaml:"example_value,omitempty" gorm:"column:example_value;type:varchar(255);"` // example value
}

// TableName 表名
//
// @return string: 返回表名
func (model *ExampleModel) TableName() string {
	return exampleModelTableName
}

// String 重写stringer接口
//
// @return string: 返回model的json字符串
func (model *ExampleModel) String() string {
	str, _ := json.MarshalString(model)
	return str
}

// NewExampleModel example模型实例化
//
// @return *ExampleModel: 返回ExampleModel实例
func NewExampleModel() *ExampleModel {
	return &ExampleModel{}
}
