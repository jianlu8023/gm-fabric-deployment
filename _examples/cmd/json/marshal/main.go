package main

import (
	"encoding/json"
	"fmt"

	"_examples/pkg/marshaler"
)

type Person struct {
	Public  string `json:"public"`
	Uid     int    `json:"uid"`
	Age     int    `json:"age"`
	Omit    string `json:"omit,omitempty"`
	Cid     string `json:"cid" mask:"cid"`
	NodeUid string `json:"nodeUid" mask:"uid"`
	private string `json:"private"`
	Email   string `json:"email" mask:"email"`
	Address string `json:"address" mask:"all"`
}

func (p Person) MarshalJSON() ([]byte, error) {
	return marshaler.Marshal(p)
}

func main() {
	p := Person{
		Public:  "public",
		private: "private",
		Uid:     100,
		Age:     18,
		Omit:    "",
		Cid:     "bafybeiagkdyf5emc7honmyls2e2y6ogrcjp5ey5q7yytndszyy5fbga7zq",
		NodeUid: "12D3KooWJvJKwUJdNnYhEaKc7NX9rnBXkSd2D6JnktCawn7t9m1h",
		Address: "这是测试地址",
		Email:   "t@e.com",
	}

	bytes, err := json.MarshalIndent(p, "  ", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(bytes))
}
