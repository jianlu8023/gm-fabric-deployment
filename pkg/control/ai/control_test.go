package ai

import (
	"fmt"
	"testing"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

func TestNewAIControl(t *testing.T) {

	aiControl, err := NewAIControl(&config.AIConfig{
		Enabled:            true,
		APIKey:             "ollama",
		APIEndpoint:        "http://172.25.133.51:8085/v1/chat/completions",
		DefaultModel:       "qwen3-32b",
		Timeout:            9999,
		InsecureSkipVerify: false,
		UseToonFormat:      true, // 默认不使用TOON格式
	}, logger.NewLoggerControl(nil))
	if err != nil {
		t.Fatal(err)
		return
	}

	defer func(aiControl *Control) {
		err := aiControl.Shutdown()
		if err != nil {
			t.Fatal(err)

		}
	}(aiControl)

	aiControl.StartUp(func(err error) {
		if err != nil {
			t.Fatal(err)
			return
		}
	})

	// 创建聊天请求（非流式）
	request := &Request{
		Model: aiControl.aiConfig.DefaultModel,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Temperature: 0.7,
		MaxTokens:   150,
		ExtraParams: map[string]interface{}{
			"enable_search": false,
		},
		Stream: false, // 非流式请求
	}

	// 发送普通请求（callback为nil）
	response, err := aiControl.SendRequest(request, nil)
	if err != nil {
		fmt.Printf("Failed to send AI request: %v\n", err)
		return
	}

	// 处理响应
	if len(response.Choices) > 0 {
		fmt.Printf("AI Response: %s\n", response.Choices[0].Message.Content)
	}

	// 创建流式请求
	streamRequest := &Request{
		Model: aiControl.aiConfig.DefaultModel,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "Tell me a story about a brave knight.",
			},
		},
		Temperature: 0.7,
		Stream:      true, // 启用流式响应
		ExtraParams: map[string]interface{}{},
	}

	// 发送流式请求（提供callback）
	fmt.Println("Streaming AI response:")
	_, streamErr := aiControl.SendRequest(streamRequest, func(delta string, finishReason string, usage *Usage, err error) bool {
		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return false
		}

		if usage != nil {
			fmt.Printf("\n[Usage: prompt=%d, completion=%d, total=%d]\n",
				usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
			return false
		}

		if finishReason != "" {
			fmt.Printf("\n[Finished: %s]\n", finishReason)
			return false
		}

		fmt.Print(delta)
		return true
	})

	if streamErr != nil {
		fmt.Printf("Failed to send AI stream request: %v\n", streamErr)
	}

	// 等待一段时间
	time.Sleep(1 * time.Second)

}
