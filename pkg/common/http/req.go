package http

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

type PaginationRequest struct {
	IsPage   bool `json:"is_page,omitempty" yaml:"is_page,omitempty" form:"isPage" binding:"-"`
	PageNo   int  `json:"page_no,omitempty" yaml:"page_no,omitempty" form:"pageNo" binding:"required"`
	PageSize int  `json:"page_size,omitempty" yaml:"page_size,omitempty" form:"pageSize" binding:"required"`
}

func (p PaginationRequest) String() string {
	str, _ := json.MarshalString(p)
	return str
}
