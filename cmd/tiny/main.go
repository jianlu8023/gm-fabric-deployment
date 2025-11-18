package main

import (
	"log"

	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Printf(colour.Red(err.Error()))
		}
	}()
	if err = conf.LoadConfig(); err != nil {
		log.Printf(colour.Red(err.Error()))
	}

}
