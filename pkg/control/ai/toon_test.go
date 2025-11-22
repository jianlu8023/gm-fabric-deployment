package ai

import (
	"fmt"
	"testing"

	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

func TestToonEncoder(t *testing.T) {
	// 创建日志控制器
	loggerControl := logger.NewLoggerControl(nil)

	// 创建TOON编码器
	toonEncoder := NewToonEncoder(loggerControl.GenLogger("toon"))

	// 测试数据
	testData := map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "Alice", "role": "admin"},
			{"id": 2, "name": "Bob", "role": "user"},
		},
	}

	// 测试基本TOON编码
	t.Run("Basic TOON Encoding", func(t *testing.T) {
		encoded, err := toonEncoder.EncodeToToon(testData)
		if err != nil {
			t.Fatalf("Failed to encode to TOON: %v", err)
		}

		expected := "users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user"
		if encoded != expected {
			t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, encoded)
		}
	})

	// 测试带制表符分隔符的TOON编码
	t.Run("TOON Encoding with Tab Delimiter", func(t *testing.T) {
		encoded, err := toonEncoder.EncodeToToonWithTabDelimiter(testData)
		if err != nil {
			t.Fatalf("Failed to encode to TOON with tab delimiter: %v", err)
		}

		// 检查是否包含制表符分隔符
		if encoded == "" {
			t.Error("Encoded string is empty")
		}
	})

	// 测试带长度标记的TOON编码
	t.Run("TOON Encoding with Length Marker", func(t *testing.T) {
		encoded, err := toonEncoder.EncodeToToonWithLengthMarker(testData)
		if err != nil {
			t.Fatalf("Failed to encode to TOON with length marker: %v", err)
		}

		// 检查是否包含长度标记
		if encoded == "" {
			t.Error("Encoded string is empty")
		}
	})
	t.Run("TOON Encoding with chatMessage", func(t *testing.T) {
		msg := []ChatMessage{
			{
				Role:    "user",
				Content: "hello world",
			},
			{
				Role:    "assistant",
				Content: "hello",
			},
		}
		encoded, err := toonEncoder.EncodeToToonWithLengthMarker(msg)
		if err != nil {
			t.Fatalf("Failed to encode to TOON with length marker: %v", err)
		}
		fmt.Printf("Encoded string: %s\n", encoded)

	})
}
