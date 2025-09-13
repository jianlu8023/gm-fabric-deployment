package dataparse

import (
	"fmt"
	"github.com/araddon/dateparse"
	"log"
)

func parse() {
	t1, err := dateparse.ParseAny("3/1/2014")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("t1: %v\n", t1)
}
