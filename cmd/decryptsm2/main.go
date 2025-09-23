package main

import (
	"encoding/pem"
	"fmt"
	"os"

	// "github.com/emmansun/gmsm/pkcs8"
	// "github.com/emmansun/gmsm/smx509"

	gmx509 "github.com/tjfoc/gmsm/x509"
)

func tjfocgmsm() {
	// 报错 x509: unknown format
	privKeyPebBytes, err := os.ReadFile("./certs/gmhserver.key")
	if err != nil {
		fmt.Printf("read private key err:%v\n", err)
		return
	}

	privKeyPem, _ := pem.Decode(privKeyPebBytes)

	privateKey, err := gmx509.ParsePKCS8EcryptedPrivateKey(privKeyPem.Bytes, []byte("gmhttp"))
	if err != nil {
		fmt.Printf("parse private key err:%v\n", err)
		return
	}

	unEncryptedPrivateKeyBytes, err := gmx509.MarshalSm2UnecryptedPrivateKey(privateKey)
	if err != nil {
		fmt.Printf("unecrypted private key err:%v\n", err)
		return
	}

	unEncryptedPem := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: unEncryptedPrivateKeyBytes,
	}

	unEncryptedContent := string(pem.EncodeToMemory(unEncryptedPem))
	fmt.Printf("unencrypted private key:%s\n", unEncryptedContent)
}

func main() {
	emmansungmsm()

}
func emmansungmsm() {
	// bytesPem, err := os.ReadFile("./certs/gmhserver.key")
	// if err != nil {
	// 	fmt.Printf("read file err:%v\n", err)
	// 	return
	// }
	//
	// password := []byte("gmhttp")
	// block, _ := pem.Decode(bytesPem)
	// if block == nil {
	// 	fmt.Fprintf(os.Stderr, "Failed to parse PEM block\n")
	// 	return
	// }
	// pk, err := pkcs8.ParsePKCS8PrivateKeySM2(block.Bytes, password)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error from ParsePKCS8PrivateKeySM2: %s\n", err)
	// 	return
	// }
	// if pk != nil {
	// 	fmt.Println("ok")
	// } else {
	// 	fmt.Println("fail")
	// }
	// der, err := smx509.MarshalPKCS8PrivateKey(pk)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error from MarshalPrivateKey: %s\n", err)
	// 	return
	// }
	//
	// // encode der bytes to pem
	// block = &pem.Block{Bytes: der, Type: "PRIVATE KEY"}
	// unEncryptedPrivKeyBytes := pem.EncodeToMemory(block)
	// pemContent := string(unEncryptedPrivKeyBytes)
	// fmt.Printf("%v\n", pemContent)
	//
	// if err = os.WriteFile("./certs/gmhserver.unencrypted.key", unEncryptedPrivKeyBytes, os.FileMode(0o644)); err != nil {
	// 	fmt.Printf("write file err:%v\n", err)
	// }
	// fmt.Printf("success write to file...")
}
