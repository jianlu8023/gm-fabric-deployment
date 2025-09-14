package bytebufferpool

import (
	"fmt"
	"github.com/valyala/bytebufferpool"
)

func ByteBufferPoolExample() {
	b := bytebufferpool.Get()
	b.WriteString("hello")
	b.WriteByte(',')
	b.WriteString(" world!")

	fmt.Println("b: ", b.String())

	bytebufferpool.Put(b)

	b1 := bytebufferpool.Get()
	fmt.Println("b1: ", b1.String())
}
