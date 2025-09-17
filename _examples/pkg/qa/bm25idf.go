package qa

import (
	"math"
	"strings"
)

func CalcBM25IDF(qaPair []QAPair) VectorF {
	idf := make(VectorF)
	docFreq := make(Vector)
	numDocs := len(qaPair)

	for _, qa := range qaPair {
		quesVector := TextToVector(qa.Question)
		for word := range quesVector {
			docFreq[word]++
		}
	}

	for word, freq := range docFreq {
		idf[word] = math.Log((float64(numDocs) - float64(freq) + 0.5) / (float64(freq) + 0.5))
	}
	return idf
}

// CalcBM25Score 计算BM25得分
// @param question 问题
// @param qa QAPair
// @param idf 词频-逆文档频率向量
// @param avgDocLength 平均文档长度 计算公式 所有文档长度之和 / 文档数量
//
// totalDocLength += float64(len(strings.Fields(qa.Question)))
// avgDocLength := totalDocLength / float64(len(qaPairs))
//
// @param k1 常数，用于控制词频对得分的影响，默认值为1.2
// @param b 常数，用于控制文档长度对得分的影响，默认值为0.75
func CalcBM25Score(question string, qa QAPair, idf VectorF, avgDocLength float64, k1, b float64) float64 {
	quesVector := TextToVector(question)
	qaVector := TextToVector(qa.Question)
	docLength := float64(len(strings.Fields(qa.Question)))
	score := 0.0
	for word, qtf := range quesVector {
		if idfValue, ok := idf[word]; ok {
			dtf := 0
			if val, ok := qaVector[word]; ok {
				dtf = val
			}
			numerator := idfValue * float64(dtf) * (k1 + 1)
			denominator := float64(dtf) + k1*(1-b+b*(docLength/avgDocLength))
			score += float64(qtf) * numerator / denominator
		}
	}

	return score
}
