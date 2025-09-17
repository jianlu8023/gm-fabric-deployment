package main

import (
	"_examples/pkg/qa"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/go-ego/gse"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

func FindBestMatchCosine(question string, qaPairs []qa.QAPair) qa.QAPair {
	questionVec := qa.TextToVector(question)

	bestMatch := qa.QAPair{}
	bestScore := 0.0

	for _, pair := range qaPairs {
		qaVec := qa.TextToVector(pair.Question)
		// fmt.Println("Question Vector:", questionVec, " QA Vector:", qaVec)
		score := qa.CosineSimilarityVectorI(qa.ConvertI(questionVec), qa.ConvertI(qaVec))
		if score > bestScore {
			bestScore = score
			bestMatch = pair
		}
	}

	if bestScore == 0.0 && len(qaPairs) > 0 {
		// 如果没有找到匹配项，返回知识库中的第一个 QA 对
		fmt.Println("没有找到匹配项，返回知识库中的第一个 QA 对")
		return qaPairs[0]
	}

	return bestMatch
}

func FindBestMatchTFIDF(question string, qaPairs []qa.QAPair) qa.QAPair {
	questionTfidf := qa.CalcTFIDF(question, qa.CalcIDF(qaPairs))
	bestMatch := qa.QAPair{}
	bestScore := 0.0
	idf := qa.CalcIDF(qaPairs)
	for _, pair := range qaPairs {
		qaTfidf := qa.CalcTFIDF(pair.Question, idf)

		score := qa.CosineSimilarityVectorF(questionTfidf, qaTfidf)
		if score > bestScore {
			bestScore = score
			bestMatch = pair
		}

	}
	if bestScore == 0.0 && len(qaPairs) > 0 {
		// 如果没有找到匹配项，返回知识库中的第一个 QA 对
		fmt.Println("没有找到匹配项，返回知识库中的第一个 QA 对")
		return qaPairs[0]
	}
	return bestMatch
}

func FindBestMatchBM25(question string, qaPairs []qa.QAPair) qa.QAPair {
	bestMatch := qa.QAPair{}
	bestScore := 0.0
	idf := qa.CalcBM25IDF(qaPairs)

	totalDocLength := 0.0
	for _, pair := range qaPairs {
		totalDocLength += float64(len(strings.Fields(pair.Question)))
	}
	avgDocLength := totalDocLength / float64(len(qaPairs))
	for _, pair := range qaPairs {
		score := qa.CalcBM25Score(question, pair, idf, avgDocLength, 1.2, 0.75)
		if score > bestScore {
			bestScore = score
			bestMatch = pair
		}
	}
	if bestScore == 0.0 && len(qaPairs) > 0 {
		// 如果没有找到匹配项，返回知识库中的第一个 QA 对
		fmt.Println("没有找到匹配项，返回知识库中的第一个 QA 对")
		return qaPairs[0]
	}
	return bestMatch
}

func FindBestMatchLevenshtein(question string, qaPairs []qa.QAPair) qa.QAPair {
	bestMatch := qa.QAPair{}
	bestScore := math.MaxInt32
	for _, pair := range qaPairs {
		distance := qa.ComputeDistance(question, pair.Question)
		if distance < bestScore {
			bestScore = distance
			bestMatch = pair
		}
	}
	if bestScore == 0 && len(qaPairs) > 0 {
		// 如果没有找到匹配项，返回知识库中的第一个 QA 对
		fmt.Println("没有找到匹配项，返回知识库中的第一个 QA 对")
		return qaPairs[0]
	}
	return bestMatch
}

func FindBestMatchLevenshteinKeyword(question string, qaPairs []qa.QAPair) qa.QAPair {
	bestMatch := qa.QAPair{}
	bestScore := -1.0  // 初始化为最小值
	threshold := 0.002 // 匹配度阈值 低于则返回最高分答案

	distanceWeight := 0.6
	keywordWeight := 0.4

	for _, pair := range qaPairs {
		// 计算 Levenshtein 距离得分
		distance := qa.ComputeDistance(question, pair.Question)
		distanceScore := 1.0 - float64(distance)/math.Max(float64(len(question)), float64(len(pair.Question))) // 归一化

		// 计算关键词匹配得分
		keywordScore := qa.KeywordMatch(question, pair.Question)

		// 计算综合得分
		score := distanceWeight*distanceScore + keywordWeight*keywordScore

		// fmt.Printf("question %s\tpair %s\t levenshteinScore %.5f\t keywordScore %.5f\t score %.5f\n", question, pair.Question, distanceScore, keywordScore, score)

		if score > bestScore {
			bestScore = score
			bestMatch = pair
		}
	}
	// 如果最佳得分低于阈值，则返回最高得分的答案
	if bestScore < threshold {
		if len(qaPairs) > 0 {
			fmt.Printf("未找到足够好的匹配项，返回最高得分答案，得分: %.5f\n", bestScore)
			return bestMatch // 返回当前最高得分的答案
		} else {
			fmt.Println("知识库为空，无法返回答案")
			return qa.QAPair{} // 返回空答案
		}
	}

	// 如果没有找到匹配项，返回知识库中的第一个 QA 对
	// if bestScore == -1.0 && len(qaPairs) > 0 {
	//
	// 	return qaPairs[0]
	// }

	return bestMatch
}

type QAPair struct {
	Question string
	Answer   string
}

var (
	seg gse.Segmenter

	// 知识库和锁
	qaPairs     []QAPair
	qaPairsLock sync.RWMutex

	knowledgeBaseFile = "knowledge_base.txt" // 知识库文件路径
	lastHash          string                 // 上次文件哈希值
)

func init() {
	seg.SkipLog = true
}

// CalculateFileHash 计算文件哈希值
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func HotReloadKnowledgeBase() {
	for {
		time.Sleep(5 * time.Second) // 每隔 5 秒检查一次

		newHash, err := CalculateFileHash(knowledgeBaseFile)
		if err != nil {
			fmt.Println("Error calculating file hash:", err)
			continue
		}

		if newHash == lastHash {
			// 文件未更改
			continue
		}

		newQAPairs, err := LoadKnowledgeBase(knowledgeBaseFile)
		if err != nil {
			fmt.Println("Error reloading knowledge base:", err)
			continue
		}
		qaPairsLock.Lock() // 获取写锁
		qaPairs = newQAPairs
		qaPairsLock.Unlock() // 释放写锁

		fmt.Println("知识库已更新")
	}
}

// LoadKnowledgeBase 从文件加载知识库
func LoadKnowledgeBase(filePath string) ([]QAPair, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开知识库文件: %w", err)
	}
	defer file.Close()

	var tempQAPairs []QAPair
	scanner := bufio.NewScanner(file)
	var currentQAPair QAPair
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line) // 移除行首尾空格

		if strings.HasPrefix(line, "Q:") {
			if currentQAPair.Question != "" {
				tempQAPairs = append(tempQAPairs, currentQAPair)
			}
			currentQAPair = QAPair{} // 重置 currentQAPair
			currentQAPair.Question = strings.TrimSpace(strings.TrimPrefix(line, "Q:"))
		} else if strings.HasPrefix(line, "A:") {
			currentQAPair.Answer = strings.TrimSpace(strings.TrimPrefix(line, "A:"))
		}
	}

	// 处理最后一个 QA 对
	if currentQAPair.Question != "" {
		tempQAPairs = append(tempQAPairs, currentQAPair)
	}

	if err := scanner.Err(); err != nil {
		return tempQAPairs, fmt.Errorf("读取知识库文件出错: %w", err)
	}

	return tempQAPairs, nil
}

// LevenshteinDistance 计算 Levenshtein 距离
func LevenshteinDistance(s1, s2 string) int {
	len1 := len(s1)
	len2 := len(s2)

	// 创建一个二维数组来存储距离
	dp := make([][]int, len1+1)
	for i := range dp {
		dp[i] = make([]int, len2+1)
	}

	// 初始化第一行和第一列
	for i := 0; i <= len1; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		dp[0][j] = j
	}

	// 填充二维数组
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = int(math.Min(math.Min(float64(dp[i-1][j]+1), float64(dp[i][j-1]+1)), float64(dp[i-1][j-1]+1)))
			}
		}
	}
	// 打印 DP 表格
	// fmt.Println("DP Table:")
	// for i := 0; i <= len1; i++ {
	// 	for j := 0; j <= len2; j++ {
	// 		fmt.Printf("%d\t", dp[i][j])
	// 	}
	// 	fmt.Println()
	// }
	return dp[len1][len2]
}

// KeywordMatch 计算关键词匹配度 (使用 gse 分词)
func KeywordMatch(question, qaQuestion string) float64 {
	questionSegments := seg.Segment([]byte(question))
	qaQuestionSegments := seg.Segment([]byte(qaQuestion))

	// 提取分词结果
	var questionWords []string
	for _, segment := range questionSegments {
		questionWords = append(questionWords, segment.Token().Text())
	}
	var qaQuestionWords []string
	for _, segment := range qaQuestionSegments {
		qaQuestionWords = append(qaQuestionWords, segment.Token().Text())
	}

	// 统计匹配的关键词数量
	matchedCount := 0
	for _, qWord := range questionWords {
		for _, qaWord := range qaQuestionWords {
			if qWord == qaWord {
				matchedCount++
				break
			}
		}
	}

	// 计算 Jaccard 相似度
	unionCount := len(questionWords) + len(qaQuestionWords) - matchedCount
	if unionCount == 0 {
		return 0 // 避免除以零
	}
	return float64(matchedCount) / float64(unionCount)
}

// FindBestMatch 在知识库中查找最佳匹配的答案
func FindBestMatch(question string, qaPairs []QAPair, distanceWeight, keywordWeight float64) QAPair {
	bestMatch := QAPair{}
	bestScore := -1.0 // 初始化为最小值

	for _, qa := range qaPairs {
		// 计算 Levenshtein 距离得分
		distance := LevenshteinDistance(question, qa.Question)
		distanceScore := 1.0 - float64(distance)/math.Max(float64(len(question)), float64(len(qa.Question))) // 归一化

		// 计算关键词匹配得分
		keywordScore := KeywordMatch(question, qa.Question)

		// 计算综合得分
		score := distanceWeight*distanceScore + keywordWeight*keywordScore

		fmt.Printf("question %s\tqa %s\t levenshteinScore %.5f\t keywordScore %.5f\t score %.5f\n",
			question, qa.Question, distanceScore, keywordScore, score)

		if score > bestScore {
			bestScore = score
			bestMatch = qa
		}
	}

	// 如果没有找到匹配项，返回知识库中的第一个 QA 对
	if bestScore == -1.0 && len(qaPairs) > 0 {
		return qaPairs[0]
	}

	return bestMatch
}

func main() {

	// s1 := "kitten"
	// s2 := "sitting"
	// distance := LevenshteinDistance(s1, s2)
	// fmt.Printf("Levenshtein Distance between '%s' and '%s' is: %d\n", s1, s2, distance)

	// 初始化 gse 分词器
	err := seg.LoadDict()
	if err != nil {
		fmt.Println("Error loading gse dictionary:", err)
		return
	}

	qaPairs, err = LoadKnowledgeBase("knowledge_base.txt")
	if err != nil {
		fmt.Println("Error loading knowledge base:", err)
		return
	}

	if len(qaPairs) == 0 {
		fmt.Println("知识库为空，请添加 QA 对")
		return
	}

	go HotReloadKnowledgeBase()

	reader := bufio.NewReader(os.Stdin)

	// 设置权重
	distanceWeight := 0.6
	keywordWeight := 0.4

	for {
		fmt.Println("请输入你的问题 (输入 'exit' 退出):")
		userQuestion, _ := reader.ReadString('\n')
		userQuestion = strings.TrimSpace(userQuestion)

		if userQuestion == "exit" {
			fmt.Println("程序退出。")
			break
		}

		qaPairsLock.RLock() // 获取读锁
		bestMatch := FindBestMatch(userQuestion, qaPairs, distanceWeight, keywordWeight)
		qaPairsLock.RUnlock() // 释放读锁

		if bestMatch.Question == "" {
			fmt.Println("无法找到匹配的答案")
		} else {
			fmt.Printf("最佳匹配答案：\nQ: %s\nA: %s\n", bestMatch.Question, bestMatch.Answer)
		}
	}
}
