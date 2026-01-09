package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// ExampleRequest exampleHandler的请求参数
//
// struct
type ExampleRequest struct {
	commonhttp.PaginationRequest        // 分页请求
	Query                        string `json:"query" form:"query" binding:"required"` // 查询参数
}

// String 重写string方法
//
// @return string: 返回req的json字符串
func (req *ExampleRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证请求参数是否合法
//
// @return bool: 返回true表示合法，false表示不合法
func (req *ExampleRequest) IsLegal() bool {
	if stringer.IsBlank(req.Query) || req.PageNo <= 0 || req.PageSize <= 0 {
		return false
	}
	return true
}
