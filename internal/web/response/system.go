package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

type SystemInitStatusResponse struct {
	Init bool `json:"init" yaml:"init"`
}

func (resp SystemInitStatusResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func (resp SystemInitStatusResponse) MarshalJSON() ([]byte, error) {
	type Alias SystemInitStatusResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

func NewSystemInitStatusResponse(init bool) SystemInitStatusResponse {
	return SystemInitStatusResponse{
		Init: init,
	}
}
