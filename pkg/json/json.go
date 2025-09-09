package json

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

func Marshal(v interface{}) ([]byte, error) {
	return jsoniter.Marshal(v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return jsoniter.MarshalIndent(v, prefix, indent)
}

func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return jsoniter.NewEncoder(w)
}

func Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}

func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return jsoniter.NewDecoder(r)
}
