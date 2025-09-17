package qa

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadKnowledgeBase 从文件加载知识库
func LoadKnowledgeBase(filePath string) ([]QAPair, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开知识库文件: %w", err)
	}
	defer file.Close()

	var qaPairs []QAPair
	scanner := bufio.NewScanner(file)
	var currentQAPair QAPair
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line) // 移除行首尾空格

		if strings.HasPrefix(line, "Q:") {
			if currentQAPair.Question != "" {
				qaPairs = append(qaPairs, currentQAPair)
			}
			currentQAPair = QAPair{} // 重置 currentQAPair
			currentQAPair.Question = strings.TrimSpace(strings.TrimPrefix(line, "Q:"))
		} else if strings.HasPrefix(line, "A:") {
			currentQAPair.Answer = strings.TrimSpace(strings.TrimPrefix(line, "A:"))
		}
	}

	// 处理最后一个 QA 对
	if currentQAPair.Question != "" {
		qaPairs = append(qaPairs, currentQAPair)
	}

	if err := scanner.Err(); err != nil {
		return qaPairs, fmt.Errorf("读取知识库文件出错: %w", err)
	}

	return qaPairs, nil
}

func init() {
	if err := seg.LoadDict(); err != nil {
		panic(err)
	}
	seg.SkipLog = true
}

// preprocessText 对文本进行预处理
func preprocessText(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, "[^a-z0-9\\s]", "") // 移除标点符号

	// 使用 gse 进行分词和词性标注
	segments := seg.Segment([]byte(text))

	// 移除停用词，只保留名词和动词
	var words []string
	for _, segment := range segments {
		if len(segment.Token().Pos()) > 0 &&
			(strings.Contains(segment.Token().Pos(), "n") ||
				strings.Contains(segment.Token().Pos(), "v")) {
			words = append(words, segment.Token().Text())
		}
	}
	return strings.Join(words, " ")
}

// TextToVector 将文本转换为向量
// @param text 文本
// @return 向量
func TextToVector(text string) Vector {
	text = preprocessText(text) // 预处理文本
	words := strings.Fields(text)
	vector := make(Vector)
	for _, word := range words {
		vector[word]++
	}
	return vector
}

func ConvertI(vec Vector) VectorI {
	vector := make(VectorI)
	for k, v := range vec {
		vector[k] = int64(v)
	}
	return vector
}

func ConvertF(vec Vector) VectorF {
	vector := make(VectorF)
	for k, v := range vec {
		vector[k] = float64(v)
	}
	return vector
}
