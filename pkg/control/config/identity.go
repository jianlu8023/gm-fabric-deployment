package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	Ed25519 = "ed25519"
	Rsa     = "rsa"
)

// CreateIdentity 生成libp2p身份
// @param algorithm 身份算法
// @param rsaKeyLen rsa密钥长度
// @return Identity libp2p身份
// @return error 错误信息
func CreateIdentity(algorithm string, rsaKeyLen int) (Identity, error) {
	ident := Identity{}

	var sk crypto.PrivKey
	var pk crypto.PubKey

	switch algorithm {
	case Rsa:
		fmt.Printf("%v  [%v]  config/identity.go:33  [config/identity]  generate rsa key pair with key length: %d\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Blue("INFO "), rsaKeyLen)
		privK, pubK, err := crypto.GenerateKeyPair(crypto.RSA, rsaKeyLen)
		if err != nil {
			fmt.Printf("%v  [%v]  config/identity.go:36  [config/identity]  generate rsa key pair failed: %v\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Red("ERROR"), err)
			return ident, err
		}
		sk = privK
		pk = pubK
	case Ed25519:
		fmt.Printf("%v  [%v]  config/identity.go:42  [config/identity]  generate ed25519 key pair\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Blue("INFO "))
		privK, pubK, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			fmt.Printf("%v  [%v]  config/identity.go:45  [config/identity]  generate ed25519 key pair failed: %v\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Red("ERROR"), err)
			return ident, err
		}
		sk = privK
		pk = pubK
	default:
		fmt.Printf("%v  [%v]  config/identity.go:51  [config/identity]  algorithm no support...\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Red("ERROR"))
		return ident, errors.New("algorithm no support")
	}

	skBytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		fmt.Printf("%v  [%v]  config/identity.go:57  [config/identity]  marshal private key failed: %v\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Red("ERROR"), err)
		return ident, err
	}

	ident.PrivKey = base64.StdEncoding.EncodeToString(skBytes)
	peerId, err := peer.IDFromPublicKey(pk)
	if err != nil {
		fmt.Printf("%v  [%v]  config/identity.go:64  [config/identity]  generate peer id failed: %v\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Red("ERROR"), err)
		return ident, err
	}
	ident.PeerID = peerId.String()
	return ident, nil
}

// DecodePrivateKey 解码用户的私钥
// @param passphrase string 私钥密码（当前版本未使用）
// @return crypto.PrivKey 解码后的私钥对象
// @return error 解码过程中可能产生的错误
//
// nolint: unused
func (c *Identity) DecodePrivateKey(passphrase string) (crypto.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(c.PrivKey)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(pkb)
}
