package main

import (
	"github.com/bytedance/sonic"

	"_examples/internal/logger"
)

// User user结构体
type User struct {
	Name string `json:"name"`
	Age  int64  `json:"age"`
}

// String string方法
// @return string: json格式的string
func (u User) String() string {
	marshalString, _ := sonic.MarshalString(u)
	return marshalString
}

// Animal 结构体
type Animal struct {
	Name string `json:"name"`
	Age  int64  `json:"age"`
}

// String string方法
// @return string: json格式的string
func (a *Animal) String() string {
	marshalString, _ := sonic.MarshalString(a)
	return marshalString
}

var (
	log = logger.GetAppLogger()
)

// main main方法
func main() {
	u1 := User{Name: "u1 User", Age: 18}
	log.Infof("%v", u1)
	u2 := &User{Name: "u2 *User", Age: 18}
	log.Infof("%v", u2)

	a1 := Animal{
		Name: "a1 Animal",
		Age:  18,
	}
	log.Infof("%v", a1)
	a2 := &Animal{
		Name: "a2 *Animal",
		Age:  18,
	}
	log.Infof("%v", a2)

}
