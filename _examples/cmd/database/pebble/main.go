package main

import (
	"fmt"
	"log"

	"github.com/cockroachdb/pebble"
)

// pebble need go >=1.19
// https://github.com/cockroachdb/pebble
func main() {

	db, err := pebble.Open("./testdata/pebble-demo", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	key := []byte("hello")
	if err := db.Set(key, []byte("world"), pebble.Sync); err != nil {
		log.Fatal(err)
	}
	value, closer, err := db.Get(key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s\n", key, value)
	if err := closer.Close(); err != nil {
		log.Fatal(err)
	}
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}
