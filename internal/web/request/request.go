package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

type Request struct {
}

func (req Request) IsLegal() bool {
	return true
}

func (req Request) GoString() string {
	return req.String()
}

func (req Request) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}
