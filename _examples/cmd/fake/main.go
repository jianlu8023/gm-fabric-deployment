package main

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jianlu8023/go-tools/pkg/format/json"

	"_examples/internal/logger"
)

var log = logger.GetAppLogger()

func main() {

	gofakeit.New(0)
	car := gofakeit.Car()
	prettyJSON, _ := json.PrettyJSON(car)
	log.Infof("fake car :\n %v", prettyJSON)
	log.Infof("fake name : %v", gofakeit.Name())
	second := gofakeit.Second()
	log.Infof("second: %d", second)
}
