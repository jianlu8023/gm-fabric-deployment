package main

import (
	"fmt"
	configcontrol "github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"

	// "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	// "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	// "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/client/msp"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/core/config"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/fabsdk"
	"os"
)

// https://www.cnblogs.com/liuhui5599/p/14195513.html
func main() {

	loggerConfig := &configcontrol.LoggerConfig{
		DefaultLogLevel: "debug",
		StackLogLevel:   "error",
		PrintFormat:     "console",
		FilePath:        "./logs/sdk.log",
		MaxAge:          7,
		RotationTime:    1,
		LoggerLevel: map[string]string{
			"main": "debug",
			"grpc": "debug",
		},
	}

	loggerControl := logger.NewLoggerControl(loggerConfig)
	// 加载配置文件
	configProvider := config.FromFile("test/baasca/config.json")
	// 创建 Fabric SDK 实例
	sdk, err := fabsdk.New(configProvider,
		fabsdk.WithLoggerPkg(newFabricLogger(loggerControl.GenLogger("basic"))),
	)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	// 设置上下文，指定组织
	ctxProvider := sdk.Context(fabsdk.WithOrg("baasCA"))
	mspClient, err := msp.New(ctxProvider)
	if err != nil {
		fmt.Printf("Failed to create MSP client: %s\n", err)
		os.Exit(1)
	}

	info, err := mspClient.GetCAInfo()
	if err != nil {
		fmt.Printf("failed to get ca info: %s\n", err)
		return
	}
	fmt.Printf("CA info: %+v\n", info)

	identity, err := mspClient.GetIdentity("gmadmin")
	if err != nil {
		fmt.Printf("failed to get identity: %s\n", err)
		return
	}
	fmt.Printf("Identity: %+v\n", identity)

	signingIdentity, err := mspClient.GetSigningIdentity("gmadmin")
	if err != nil {
		fmt.Printf("Failed to get signing identity: %s\n", err)
		return
	}
	serialize, err := signingIdentity.Serialize()
	if err != nil {
		fmt.Printf("Failed to serialize signing identity: %s\n", err)
		return
	}

	fmt.Printf("Signing identity: %+v\n", string(serialize))

	certificate := signingIdentity.EnrollmentCertificate()

	fmt.Printf("Enrollment certificate: %+v\n", string(certificate))

	// 定义要注册的新用户
	// username := "newuser1"
	// secret := "newuser1_secret"

	// 注册新用户
	// enrollmentSecret, err := mspClient.Register(
	// 	&msp.RegistrationRequest{
	// 		Name:        username,
	// 		Secret:      secret,
	// 		Type:        "client",
	// 		Affiliation: "org1.department1",
	// 		Attributes: []msp.Attribute{
	// 			{Name: "hf.Revoker", Value: "false", ECert: true},
	// 		},
	// 	},
	// )

	// if err != nil {
	// 	fmt.Printf("Failed to register new user: %s\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Printf("Successfully registered user '%s' with secret: %s\n", username, enrollmentSecret)

	// 签发新用户证书
	// err = mspClient.Enroll(
	// 	username,
	// 	msp.WithSecret(secret),
	// )
	// if err != nil {
	// 	fmt.Printf("Failed to enroll new user: %s\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Println("Successfully enrolled new user and received certificates.")
	// 此时，证书和私钥已存储在 SDK 的 CryptoSuite 中，你可以通过 enrollmentResult 获取
	// fmt.Printf("Certificate: %s\n", string(enrollmentResult.Identity.Certificate()))
	// fmt.Printf("Private Key: %s\n", string(enrollmentResult.Key.Bytes()))
}
