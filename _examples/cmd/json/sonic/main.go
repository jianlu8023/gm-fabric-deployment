package main

import (
	"github.com/bytedance/sonic"

	"_examples/internal/logger"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {

	p := &Person{
		Name: "Tom",
		Age:  18,
	}

	marshalString, err := sonic.MarshalString(p)
	if err != nil {
		logger.GetAppLogger().Errorf("sonic marshal error %v", err)
		return
	}

	logger.GetAppLogger().Infof("sonic marshal string %v", marshalString)

}
