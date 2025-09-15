package gm_fabric

import (
	"fmt"
	gmconfig "github.com/hxx258456/fabric-sdk-go-gm/pkg/core/config"
	gmgw "github.com/hxx258456/fabric-sdk-go-gm/pkg/gateway"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/gm-fabric/wallethelper"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
)

func ChaincodeCallExample() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wd, err := os.Getwd()
	
	if err != nil {
		fmt.Println(fmt.Errorf("get current work dir has error: %s", err))
		return
	}
	wallet, err := gmgw.NewFileSystemWallet(filepath.Clean(filepath.Join(wd, "testdata", "wallet")))
	if err != nil {
		fmt.Println(fmt.Errorf("get file system wallet has error: %s", err))
		return
	}
	if !wallet.Exists("appUser") {
		err = wallethelper.GMCreateWallet(wallet)
		if err != nil {
			fmt.Println(fmt.Sprintf("create user appUser has error:%s", err))
			return
		}
	}
	
	connectionPath := filepath.Join(wd, "testdata", "organizations", "peerOrganizations", "org2.example.com", "connection-org2.json")
	
	gw, err := gmgw.Connect(
		gmgw.WithConfig(gmconfig.FromFile(filepath.Clean(connectionPath))),
		gmgw.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Println(fmt.Sprintf("get gatewayClient has error:%s", err))
		return
	}
	
	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Println(fmt.Sprintf("get channel mychannel has error:%s", err))
		return
	}
	
	contract := network.GetContract("baisc")
	allAssert, err := contract.EvaluateTransaction("GetAllAssets")
	
	if err != nil {
		fmt.Println(fmt.Sprintf("call GetAllAsserts has error:%s", err))
		return
	}
	
	fmt.Println(string(allAssert))
}

func getwd() {
	// method 1
	wd, _ := os.Getwd()
	fmt.Println("os.Getwd(): ", wd)
	
	lookPath, _ := exec.LookPath(os.Args[0])
	abs, _ := filepath.Abs(lookPath)
	dir := path.Dir(abs)
	fmt.Println("os.Args[0]: ", dir)
	
	_, file, _, _ := runtime.Caller(0)
	fmt.Println("runtime.Caller(0): ", file)
}
