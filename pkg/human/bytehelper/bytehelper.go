package bytehelper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

const (
	Byte = 1

	KiloByte = Byte * 1000
	MegaByte = KiloByte * 1000
	GigaByte = MegaByte * 1000
	TeraByte = GigaByte * 1000

	KibiByte = Byte * 1024
	MebiByte = KibiByte * 1024
	GibiByte = MebiByte * 1024
)

func HumanBytes1000(b int64) string {
	var value float64
	var unit string

	switch {
	case b >= TeraByte:
		value = float64(b) / TeraByte
		unit = "TB"
	case b >= GigaByte:
		value = float64(b) / GigaByte
		unit = "GB"
	case b >= MegaByte:
		value = float64(b) / MegaByte
		unit = "MB"
	case b >= KiloByte:
		value = float64(b) / KiloByte
		unit = "KB"
	default:
		return fmt.Sprintf("%d B", b)
	}

	switch {
	case value >= 10:
		return fmt.Sprintf("%d %s", int(value), unit)
	case value != math.Trunc(value):
		return fmt.Sprintf("%.1f %s", value, unit)
	default:
		return fmt.Sprintf("%d %s", int(value), unit)
	}
}

func HumanBytes1024(b uint64) string {
	switch {
	case b >= GibiByte:
		return fmt.Sprintf("%.1f GiB", float64(b)/GibiByte)
	case b >= MebiByte:
		return fmt.Sprintf("%.1f MiB", float64(b)/MebiByte)
	case b >= KibiByte:
		return fmt.Sprintf("%.1f KiB", float64(b)/KibiByte)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// func HumanNumber(b uint64) string {
// 	switch {
// 	case b >= GigaByte:
// 		number := float64(b) / GigaByte
// 		if number == math.Floor(number) {
// 			return fmt.Sprintf("%.0fB", number) // no decimals if whole number
// 		}
// 		return fmt.Sprintf("%.1fB", number) // one decimal if not a whole number
// 	case b >= MegaByte:
// 		number := float64(b) / MegaByte
// 		if number == math.Floor(number) {
// 			return fmt.Sprintf("%.0fM", number) // no decimals if whole number
// 		}
// 		return fmt.Sprintf("%.2fM", number) // two decimals if not a whole number
// 	case b >= KiloByte:
// 		return fmt.Sprintf("%.0fK", float64(b)/KiloByte)
// 	default:
// 		return strconv.FormatUint(b, 10)
// 	}
// }

// BytesToInt le bytehelper to int32, little endian
func BytesToInt(b []byte) (int32, error) {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	err := binary.Read(bytesBuffer, binary.LittleEndian, &x)
	if err != nil {
		return -1, err
	}
	return x, nil
}

// BytesToInt64 le bytehelper to int64, little endian
func BytesToInt64(b []byte) (int64, error) {
	bytesBuffer := bytes.NewBuffer(b)
	var x int64
	err := binary.Read(bytesBuffer, binary.LittleEndian, &x)
	if err != nil {
		return -1, err
	}
	return x, nil
}

// IntToBytes int32 to le bytehelper, little endian
func IntToBytes(x int32) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.LittleEndian, x)
	if err != nil {
		return nil
	}
	return bytesBuffer.Bytes()
}

// Int64ToBytes int64 to le bytehelper, little endian
func Int64ToBytes(x int64) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.LittleEndian, x)
	if err != nil {
		return nil, err
	}
	return bytesBuffer.Bytes(), nil
}

// BytesToUint64 le bytehelper to uint64, little endian
func BytesToUint64(b []byte) (uint64, error) {
	bytesBuffer := bytes.NewBuffer(b)
	var x uint64
	err := binary.Read(bytesBuffer, binary.LittleEndian, &x)
	if err != nil {
		return 0, err
	}
	return x, nil
}

// Uint64ToBytes uint64 to le bytehelper, little endian
func Uint64ToBytes(x uint64) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.LittleEndian, x)
	if err != nil {
		return nil, err
	}
	return bytesBuffer.Bytes(), nil
}

// BytesPrefix returns key range that satisfy the given prefix.
// This only applicable for the standard 'bytehelper comparer'.
func BytesPrefix(prefix []byte) ([]byte, []byte) {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	return prefix, limit
}

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
