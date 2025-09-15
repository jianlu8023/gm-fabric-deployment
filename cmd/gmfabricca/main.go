package main

import (
	"fmt"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/client/msp"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/core/config"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/fabsdk"
	"os"
)

// https://www.cnblogs.com/liuhui5599/p/14195513.html
func main() {
	
	// caclient, err := msp.NewCAClient("ca-ogr1", nil)
	// if err != nil {
	// 	panic(err)
	// }
	// register, err := caclient.Register(&api.RegistrationRequest{
	// 	Name:           "user1",
	// 	Type:           "client",
	// 	MaxEnrollments: -1,
	// 	Secret:         "123456",
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("register: %+v\n", register)
	
	// 加载配置文件
	configProvider := config.FromFile("config.yaml")
	
	// 创建 Fabric SDK 实例
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()
	
	// 设置上下文，指定组织
	ctxProvider := sdk.Context(fabsdk.WithOrg("Org1"))
	mspClient, err := msp.New(ctxProvider)
	if err != nil {
		fmt.Printf("Failed to create MSP client: %s\n", err)
		os.Exit(1)
	}
	// 定义要注册的新用户
	username := "newuser1"
	secret := "newuser1_secret"
	
	// 注册新用户
	enrollmentSecret, err := mspClient.Register(
		&msp.RegistrationRequest{
			Name:        username,
			Secret:      secret,
			Type:        "client",
			Affiliation: "org1.department1",
			Attributes: []msp.Attribute{
				{Name: "hf.Revoker", Value: "false", ECert: true},
			},
		},
	)
	if err != nil {
		fmt.Printf("Failed to register new user: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully registered user '%s' with secret: %s\n", username, enrollmentSecret)
	
	// 签发新用户证书
	err = mspClient.Enroll(
		username,
		msp.WithSecret(secret),
	)
	if err != nil {
		fmt.Printf("Failed to enroll new user: %s\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Successfully enrolled new user and received certificates.")
	// 此时，证书和私钥已存储在 SDK 的 CryptoSuite 中，你可以通过 enrollmentResult 获取
	// fmt.Printf("Certificate: %s\n", string(enrollmentResult.Identity.Certificate()))
	// fmt.Printf("Private Key: %s\n", string(enrollmentResult.Key.Bytes()))
}
