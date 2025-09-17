package sm2

import (
	"crypto/rand"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

func Encrypt(pubKey string, plaintext string, c1c2c3 bool) ([]byte, error) {
	return EncryptBytes([]byte(pubKey), []byte(plaintext), c1c2c3)
}

func EncryptBytes(pubKey []byte, plaintext []byte, c1c2c3 bool) ([]byte, error) {
	var (
		encrypt []byte
		err     error
	)
	publicKey, err := x509.ReadPublicKeyFromPem(pubKey)
	if err != nil {
		return nil, err
	}
	if c1c2c3 {
		encrypt, err = sm2.Encrypt(publicKey, plaintext, rand.Reader, sm2.C1C2C3)

	} else {
		encrypt, err = sm2.Encrypt(publicKey, plaintext, rand.Reader, sm2.C1C3C2)
	}
	if err != nil {
		return nil, err
	}
	return encrypt, nil
}

func Decrypt(privKey, pwd, ciphertext string, c1c2c3 bool) ([]byte, error) {
	return DecryptBytes([]byte(privKey), []byte(pwd), []byte(ciphertext), c1c2c3)
}

func DecryptBytes(privKey, pwd, ciphertext []byte, c1c2c3 bool) ([]byte, error) {
	var (
		decrypt []byte
		err     error
	)
	priKey, err := x509.ReadPrivateKeyFromPem(privKey, pwd)
	if err != nil {
		return nil, err
	}
	if c1c2c3 {
		decrypt, err = sm2.Decrypt(priKey, ciphertext, sm2.C1C2C3)

	} else {
		decrypt, err = sm2.Decrypt(priKey, ciphertext, sm2.C1C2C3)
	}
	if err != nil {
		return nil, err
	}
	return decrypt, nil
}
