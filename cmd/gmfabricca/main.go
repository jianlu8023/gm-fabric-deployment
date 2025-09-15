package main

import (
	"fmt"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/msp"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/msp/api"
)

// https://www.cnblogs.com/liuhui5599/p/14195513.html
func main() {
	
	caclient, err := msp.NewCAClient("ca-ogr1", nil)
	if err != nil {
		panic(err)
	}
	register, err := caclient.Register(&api.RegistrationRequest{
		Name:           "user1",
		Type:           "client",
		MaxEnrollments: -1,
		Secret:         "123456",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("register: %+v\n", register)
}
