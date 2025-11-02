package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
)

// StreamCallback 定义流式响应的回调函数类型
// @param delta: 增量内容
// @param finishReason: 完成原因
// @param err: 错误信息
// @return bool: 是否继续接收数据
type StreamCallback func(delta string, finishReason string, err error) bool

// processStreamData 处理流式响应数据
// @param data: 原始流数据
// @param callback: 回调函数
func (c *Control) processStreamData(data []byte, callback StreamCallback) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// 处理data行
		if strings.HasPrefix(line, "data: ") {
			jsonData := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if jsonData == "[DONE]" {
				callback("", "stop", nil)
				return
			}

			// 解析JSON数据
			var response Response
			if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
				if !callback("", "", err) {
					return
				}
				continue
			}

			// 处理响应数据
			for _, choice := range response.Choices {
				if choice.Delta != nil && choice.Delta.Content != "" {
					if !callback(choice.Delta.Content, "", nil) {
						return
					}
				}

				if choice.FinishReason != "" && choice.FinishReason != "null" {
					callback("", choice.FinishReason, nil)
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		callback("", "", err)
	}
}
