package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

type Response struct{}

func (resp Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

func (resp Response) GoString() string {
	return resp.String()
}

func (resp Response) String() string {
	marshalString, _ := json.MarshalString(resp)
	return marshalString
}

func NewResponse() Response {
	return Response{}
}
