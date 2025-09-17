package main

import (
	"encoding/pem"
	"fmt"
	"log"
	"os"

	ecies "github.com/ecies/go/v2"
	"github.com/jianlu8023/go-tools/pkg/path"
)

func main() {
	var (
		privKeyPath = "testdata/ecc/private.pem"
		pubKeyPath  = "testdata/ecc/public.pem"
	)

	if err := path.Ensure(privKeyPath, false); err != nil {
		log.Fatalln("ensure private key file error", err)
		return
	}

	key, err := ecies.GenerateKey()

	if err != nil {
		log.Fatalln("generate key error", err)
		return
	}

	privKeyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: key.Bytes(),
	}

	privFile, err := os.Create(privKeyPath)
	if err != nil {
		log.Fatalln("create private key file error", err)
		return
	}

	if err := pem.Encode(privFile, privKeyBlock); err != nil {
		log.Fatalln("encode private key error", err)
	}

	publicKey := key.PublicKey

	pubKeyFile, err := os.Create(pubKeyPath)
	if err != nil {
		log.Fatalln("create public key file error", err)
		return
	}

	pubKeyBlock := &pem.Block{
		Type:  "EC PUBLIC KEY",
		Bytes: publicKey.Bytes(false),
	}

	if err := pem.Encode(pubKeyFile, pubKeyBlock); err != nil {
		log.Fatalln("encode public key error", err)
	}

	original := "hello world"
	fmt.Println("original:", original)

	pubFileRead, err := os.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatalln("read public key file error", err)
		return
	}
	pubBlockR, _ := pem.Decode(pubFileRead)

	publicKeyR, err := ecies.NewPublicKeyFromBytes(pubBlockR.Bytes)
	if err != nil {
		log.Fatalln("read public key error", err)
		return
	}
	encrypt, err := ecies.Encrypt(publicKeyR, []byte(original))
	if err != nil {
		log.Fatalln("encrypt error", err)
		return
	}
	fmt.Println("encrypt:", string(encrypt))

	privFileRead, err := os.ReadFile(privKeyPath)
	if err != nil {
		log.Fatalln("read private key file error", err)
		return
	}
	priBlockR, _ := pem.Decode(privFileRead)

	privateKeyR := ecies.NewPrivateKeyFromBytes(priBlockR.Bytes)

	decrypt, err := ecies.Decrypt(privateKeyR, encrypt)
	if err != nil {
		log.Fatalln("decrypt error", err)
		return
	}
	fmt.Println("decrypt:", string(decrypt))
}
