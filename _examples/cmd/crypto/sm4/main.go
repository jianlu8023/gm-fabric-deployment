package main

import (
	"os"

	"github.com/jianlu8023/go-tools/pkg/path"

	"_examples/internal/logger"
	"_examples/pkg/crypto/sm4"
)

var log = logger.GetAppLogger()

func main() {
	var (
		key      = "0123456789abcdef"
		original = "testdata/sm4/sm4.txt"
		encrypt  = "testdata/sm4/sm4-enc.txt"
		decrypt  = "testdata/sm4/sm4-dec.txt"
		context  = "hello world this is a sm4 crypto test file"
	)

	if err := path.Ensure(original, false); err != nil {
		log.Errorf("ensure original file failed: %v", err)
		return
	}

	if err := os.WriteFile(original, []byte(context), os.FileMode(0644)); err != nil {
		log.Errorf("write original file failed: %v", err)
		return
	}

	err := sm4.EncryptFile(key, original, encrypt)
	if err != nil {
		log.Errorf("sm4 encrypt file failed: %v", err)
		return
	}

	err = sm4.DecryptFile(key, encrypt, decrypt)
	if err != nil {
		log.Errorf("sm4 decrypt file failed: %v", err)
		return
	}
}
