package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"

	"github.com/jianlu8023/go-tools/v2/pkg/helper/json"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Identity struct {
	PeerID  string
	PrivKey string `json:",omitempty"`
}

// DecodePrivateKey is a helper to decode the users PrivateKey.
func (i *Identity) DecodePrivateKey(passphrase string) (crypto.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(i.PrivKey)
	if err != nil {
		return nil, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	return crypto.UnmarshalPrivateKey(pkb)
}

func main() {

	var algorithm string
	var rsaKeyLen int
	flag.StringVar(&algorithm, "algorithm", "ed25519", "algorithm rsa or ed25519")
	flag.IntVar(&rsaKeyLen, "rsakeylen", 2048, "rsakeylen 2048")

	var sk crypto.PrivKey
	var pk crypto.PubKey

	switch algorithm {
	case "rsa":
		privK, pubK, err := crypto.GenerateKeyPair(crypto.RSA, rsaKeyLen)
		if err != nil {
			fmt.Printf("generate key pair failed: %v\n", err)
			return
		}
		sk = privK
		pk = pubK
	case "ed25519":
		privK, pubK, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			fmt.Printf("generate key pair failed: %v\n", err)
			return
		}
		sk = privK
		pk = pubK
	default:
		fmt.Println("algorithm no support...")
		return
	}

	skBytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		fmt.Printf("marshal private key failed: %v\n", err)
		return
	}
	ident := Identity{}

	skBase64 := base64.StdEncoding.EncodeToString(skBytes)
	ident.PrivKey = skBase64
	peerId, err := peer.IDFromPublicKey(pk)
	if err != nil {
		fmt.Printf("generate peer id failed: %v\n", err)
		return
	}
	ident.PeerID = peerId.String()
	fmt.Printf("peer identity: %s\n", ident.PeerID)
	bytes, err := json.Marshal(ident)
	if err != nil {
		fmt.Printf("marshal identity failed: %v\n", err)
		return
	}
	fmt.Printf("peer identity: %s\n", string(bytes))

}
