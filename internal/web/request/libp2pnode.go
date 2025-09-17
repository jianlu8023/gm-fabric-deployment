package request

import (
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/json"
)

type Libp2pNodeListRequest struct {
	commonhttp.PaginationRequest
}

func (p *Libp2pNodeListRequest) String() string {
	bytes, _ := json.Marshal(p)
	return string(bytes)
}

func (p *Libp2pNodeListRequest) IsLegal() bool {
	if p.PageNo <= 0 || p.PageSize <= 0 {
		return false
	}
	return true
}
