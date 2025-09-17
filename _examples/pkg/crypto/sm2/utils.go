package sm2

import (
	"encoding/pem"
	"os"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

func PrivateKeySave(key *sm2.PrivateKey, pwd, filename string) error {
	privBytes, err := x509.MarshalSm2PrivateKey(key, []byte(pwd))
	if err != nil {
		return err
	}
	privBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}
	create, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = pem.Encode(create, privBlock)
	if err != nil {
		return err
	}
	return nil
}

func PublicKeySave(key *sm2.PublicKey, filename string) error {
	publicKeyBytes, err := x509.MarshalSm2PublicKey(key)
	if err != nil {
		return err
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	create, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = pem.Encode(create, pubBlock)
	if err != nil {
		return err
	}
	return nil
}
