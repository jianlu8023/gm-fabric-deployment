package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

type SystemOverviewRequest struct {
}

func (req SystemOverviewRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

func (req SystemOverviewRequest) IsLegal() bool {
	return true
}

type SystemInitStatusRequest struct {
}

func (req SystemInitStatusRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

func (req SystemInitStatusRequest) IsLegal() bool {
	return true
}
