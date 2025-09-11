package bytehelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToString(t *testing.T) {
	b := []byte("Hello World")
	str := BytesToString(b)
	assert.Equal(t, string(b), str)
}

func TestStringToBytes(t *testing.T) {
	str := "Hello World"
	b := StringToBytes(str)
	assert.Equal(t, []byte(str), b)
}
