package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// StreamCallback 定义流式响应的回调函数类型
// @param delta: 增量内容
// @param finishReason: 完成原因
// @param usage: token使用情况
// @param err: 错误信息
// @return bool: 是否继续接收数据
type StreamCallback func(delta string, finishReason string, usage *Usage, err error) bool

// processStreamData 处理流式响应数据
// @param data: 原始流数据
// @param callback: 回调函数
func (c *Control) processStreamData(data []byte, callback StreamCallback) {
	// 使用bufio.Scanner处理数据，按行分割
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if stringer.IsBlank(line) || strings.HasPrefix(line, ":") {
			continue
		}

		// 处理data行
		if strings.HasPrefix(line, "data:") {
			// 提取JSON数据部分（去掉"data: "前缀）
			jsonData := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			// 检查是否是结束标记
			if jsonData == "[DONE]" {
				callback("", "stop", nil, nil)
				return
			}

			// 解析JSON数据
			var response Response
			if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
				// 如果JSON解析失败，记录错误但继续处理其他数据
				// c.logger.Debugf("[control] failed to unmarshal JSON: %v, data: %s", err, jsonData)
				// c.logger.Debugf("[control] current line: %s", line)
				continue
			}

			// 记录接收到的响应数据
			// c.logger.Debugf("[control] received stream response chunk, choices: %d", len(response.Choices))

			// 处理响应数据
			// 添加一个标志来跟踪是否已经发送了完成信号
			finishSent := false
			for _, choice := range response.Choices {
				// c.logger.Debugf("[control] processing choice %d, finish_reason: %s", i, choice.FinishReason)

				// 检查是否有内容需要输出
				if choice.Delta != nil && !stringer.IsBlank(choice.Delta.Content) {
					// c.logger.Debugf("[control] choice %d has content: %s", i, choice.Delta.Content)
					if !callback(choice.Delta.Content, "", nil, nil) {
						// c.logger.Debugf("[control] callback returned false, stopping stream processing")
						return
					}
				}

				// 检查是否完成
				if !finishSent && !stringer.IsBlank(choice.FinishReason) && !stringer.CompareIgnoreCase(choice.FinishReason, "null") {
					// 只有当finishReason不为空时才调用回调
					if !stringer.IsBlank(choice.FinishReason) {
						// c.logger.Debugf("[control] choice %d finished with reason: %s", i, choice.FinishReason)
						if !callback("", choice.FinishReason, nil, nil) {
							// c.logger.Debugf("[control] callback returned false for finish reason, stopping stream processing")
							return
						}
						finishSent = true
					}
				}
			}

			// 检查是否有usage信息
			if response.Usage != nil {
				// 添加token使用量的日志输出
				c.logger.Debugf("[ai/control] token usage - prompt: %d, completion: %d, total: %d",
					response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)

				if !callback("", "", response.Usage, nil) {
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		c.logger.Debugf("[ai/control] scanner error: %v", err)
		callback("", "", nil, err)
	}
}
