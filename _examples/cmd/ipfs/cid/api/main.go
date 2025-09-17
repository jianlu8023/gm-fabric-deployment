package main

import (
	"fmt"
	"os"

	"_examples/pkg/randomfile"
	shell "github.com/ipfs/go-ipfs-api"
)

func main() {
	fmt.Println("starting use api get ipfs cid")

	sh := shell.NewShell("/ip4/127.0.0.1/tcp/5001")

	up := sh.IsUp()

	if !up {
		fmt.Println("ipfs is not up")
		return
	}

	filePath := "./testdata/ipfs/cid/test.txt"

	if !randomfile.RandomFile(filePath) {
		fmt.Println("random file error")
		return
	}

	v0, err := getCidV0(sh, filePath)
	if err != nil {
		fmt.Println("get cid v0 error:", err)
		return
	}
	fmt.Println("cid v0:", v0)

	v1, err := getCidV1(sh, filePath)
	if err != nil {
		fmt.Println("get cid v1 error:", err)
		return
	}
	fmt.Println("cid v1:", v1)
}

func getCidV0(sh *shell.Shell, filePath string) (string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("open file error:", err)
		return "", err
	}

	add, err := sh.Add(file,
		shell.OnlyHash(true),
		// shell.RawLeaves(true),
	)

	if err != nil {
		fmt.Println("add file error:", err)
		return "", err
	}

	return add, nil

}

func getCidV1(sh *shell.Shell, filePath string) (string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("open file error:", err)
		return "", err
	}

	add, err := sh.Add(file,
		shell.OnlyHash(true),
		shell.CidVersion(1),
		// shell.RawLeaves(true),
	)

	if err != nil {
		fmt.Println("add file error:", err)
		return "", err
	}

	return add, nil
}
