package main

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func main() {
	jsonString := `{"name":"Alice","age":30}`

	// 使用 gjson 获取 name 字段
	name := gjson.Get(jsonString, "name").String()
	age := gjson.Get(jsonString, "age").Int()

	fmt.Printf("Name: %s, Age: %d\n", name, age)
}
