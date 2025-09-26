package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

type Libp2pNodeListRequest struct {
	commonhttp.PaginationRequest
}

func (req Libp2pNodeListRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

func (req Libp2pNodeListRequest) IsLegal() bool {
	if req.PageNo <= 0 || req.PageSize <= 0 {
		return false
	}
	return true
}

type Libp2pNodeMyselfRequest struct{}

func (req Libp2pNodeMyselfRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

func (req Libp2pNodeMyselfRequest) IsLegal() bool {
	return true
}
