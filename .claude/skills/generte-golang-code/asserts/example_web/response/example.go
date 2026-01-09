package response

import (
	"fmt"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/model"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"github.com/jinzhu/copier"
)

// ExampleResponse Example响应结构体
//
// struct
type ExampleResponse struct {
	ExampleKey   string `json:"example_key" yaml:"example_key"`     // example key
	ExampleValue string `json:"example_value" yaml:"example_value"` // example value
}

// String 重写stringer接口
//
// @return string: resp的json字符串
func (resp ExampleResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义json序列化方式
//
// @return []byte: json字节数组
// @return error: 序列化错误
func (resp ExampleResponse) MarshalJSON() ([]byte, error) {
	type Alias ExampleResponse
	aux := struct {
		*Alias
		QueryTime time.Time `json:"query_time" yaml:"query_time"`
	}{
		Alias:     (*Alias)(&resp),
		QueryTime: time.Now(),
	}

	return json.Marshal(aux)
}

// NewExampleResponse 构造ExampleResponse
//
// @param queryPage dbpage.Info[model.ExampleModel]: 查询结果
//
// @return *dbpage.Info[ExampleResponse]: ExampleResponse分页信息
// @return error: 错误信息
func NewExampleResponse(queryPage dbpage.Info[model.ExampleModel]) (*dbpage.Info[ExampleResponse], error) {
	var convert []ExampleResponse

	var src1 int64 = 0
	// src2 := []string{""}
	err := copier.CopyWithOption(&convert, queryPage.GetRecords(), copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
		Converters: []copier.TypeConverter{
			{
				SrcType: src1,
				DstType: time.Time{},
				Fn: func(src interface{}) (dst interface{}, err error) {
					if src == nil {
						return nil, nil
					}
					timeInt, ok := src.(int64)
					if !ok {
						return nil, nil
					}
					local, err := humantime.ParseTimeLocal(fmt.Sprintf("%v", timeInt))
					if err != nil {
						return nil, err
					}
					return local, nil
				},
			},
			// {
			// 	SrcType: src2,
			// 	DstType: copier.String,
			// 	Fn: func(src interface{}) (dst interface{}, err error) {
			// 		if src == nil {
			// 			return nil, nil
			// 		}
			// 		imageNameSrc, ok := src.([]string)
			// 		if !ok {
			// 			return nil, nil
			// 		}
			// 		imageNameDst, err := json.MarshalString(imageNameSrc)
			// 		if err != nil {
			// 			return nil, err
			// 		}
			// 		return imageNameDst, nil
			// 	},
			// },
		},
	})
	if err != nil {
		return nil, err
	}
	return dbpage.NewPageInfo[ExampleResponse](
		queryPage.GetPageNo(),
		queryPage.GetPageSize(),
		queryPage.GetCount(),
		queryPage.GetMaxPage(),
		convert,
	), nil
}
