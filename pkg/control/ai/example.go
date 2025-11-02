package ai

import (
	"fmt"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

// ExampleUsage AI控制器使用示例
func ExampleUsage(aiConfig *config.AIConfig, loggerControl *logger.Control) {
	// 创建AI控制器
	aiControl, err := NewAIControl(aiConfig, loggerControl)
	if err != nil {
		fmt.Printf("Failed to create AI control: %v\n", err)
		return
	}

	// 启动AI服务
	aiControl.StartUp(func(err error) {
		fmt.Printf("Failed to start AI service: %v\n", err)
	})

	// 创建聊天请求
	request := &Request{
		Model: "gemini-2.0-flash-exp",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Temperature: 0.7,
		MaxTokens:   150,
	}

	// 发送普通请求
	response, err := aiControl.SendRequest(request)
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
		Model: "gemini-2.0-flash-exp",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "Tell me a story about a brave knight.",
			},
		},
		Temperature: 0.7,
		MaxTokens:   300,
		Stream:      true, // 启用流式响应
	}

	// 发送流式请求
	fmt.Println("Streaming AI response:")
	streamErr := aiControl.StreamRequest(streamRequest, func(delta string, finishReason string, err error) bool {
		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
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

	// 关闭AI服务
	if err := aiControl.Shutdown(); err != nil {
		fmt.Printf("Failed to shutdown AI service: %v\n", err)
	}
}
