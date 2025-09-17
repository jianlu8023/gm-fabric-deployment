package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"_examples/pkg/randomfile"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func getCidV0(filePath string) (cid.Cid, error) {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("open file failed, err:", err)
		return cid.Undef, err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Println("close file failed, err:", err)
		}
	}(file)

	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		fmt.Println("copy file failed, err:", err)
		return cid.Undef, err
	}
	hashSum := hash.Sum(nil)

	multihash, err := mh.Encode(hashSum, mh.SHA2_256)
	if err != nil {
		fmt.Println("encode failed, err:", err)
	}

	v0 := cid.NewCidV0(multihash)
	return v0, nil
}

func getCidV1(filePath string) (cid.Cid, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("open file failed, err:", err)
		return cid.Undef, err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Println("close file failed, err:", err)
		}
	}(file)

	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		fmt.Println("copy file failed, err:", err)
		return cid.Undef, err
	}
	hashSum := hash.Sum(nil)

	multihash, err := mh.Encode(hashSum, mh.SHA2_256)
	if err != nil {
		fmt.Println("encode failed, err:", err)
	}

	v1 := cid.NewCidV1(cid.Raw, multihash)
	return v1, nil
}

func main() {

	fmt.Println("starting calculate cid")

	filePath := "./testdata/ipfs/cid/test.txt"
	if !randomfile.RandomFile(filePath) {
		fmt.Println("random file error")
		return
	}
	v0, err := getCidV0(filePath)
	if err != nil {
		fmt.Println("build v0 failed, err:", err)
		return
	}
	fmt.Println("v0:", v0.String())

	v1, err := getCidV1(filePath)
	if err != nil {
		fmt.Println("build v1 failed, err:", err)
		return
	}
	fmt.Println("v1:", v1.String())

}
