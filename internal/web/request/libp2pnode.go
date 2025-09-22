package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

type Libp2pNodeListRequest struct {
	commonhttp.PaginationRequest
}

func (p *Libp2pNodeListRequest) String() string {
	str, _ := json.MarshalString(p)
	return str
}

func (p *Libp2pNodeListRequest) IsLegal() bool {
	if p.PageNo <= 0 || p.PageSize <= 0 {
		return false
	}
	return true
}
