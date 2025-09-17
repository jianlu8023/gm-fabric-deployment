package sm3

import "github.com/tjfoc/gmsm/sm3"

func Hash(context string) []byte {
	hash := sm3.New()
	_, _ = hash.Write([]byte(context))
	return hash.Sum(nil)
}
