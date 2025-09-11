package dbpage

import (
	"encoding/json"
	"testing"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestNewPageInfo(t *testing.T) {
	info := NewPageInfo(1, 2, 10, 5, []Person{
		{Name: "张三", Age: 18},
		{Name: "李四", Age: 20},
	})

	bytes, err := json.Marshal(info)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bytes))
}
