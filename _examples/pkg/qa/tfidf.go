package qa

import (
	"math"
)

func CalcTF(text string) VectorF {
	tf := make(VectorF)
	// 分词
	wordFreq := TextToVector(text)
	totalWords := 0

	// 计算总词数
	for _, freq := range wordFreq {
		totalWords += freq
	}
	// 计算tf
	for word, freq := range wordFreq {
		tf[word] = float64(freq) / float64(totalWords)
	}
	return tf
}

func CalcIDF(qaPair []QAPair) VectorF {
	idf := make(VectorF)
	docFreq := make(map[string]int)
	numDocs := len(qaPair)
	for _, pair := range qaPair {
		quesVector := TextToVector(pair.Question)
		for word := range quesVector {
			docFreq[word]++
		}
	}
	for word, freq := range docFreq {
		idf[word] = math.Log(float64(numDocs) / float64(freq+1)) // log(知识库中问题数量 / 问题中词频 + 1)
	}
	return idf
}

// CalcTFIDF 计算TF-IDF
// tf = N(q) / N(d) = 问题中词频 / 知识库中词频
// idf = log(N / N(q)+1) = log(知识库中问题数量 / 问题中词频 + 1)
// tf_idf = tf * idf = 问题中词频 / 知识库中词频 * log(知识库中问题数量 / 问题中词频 + 1)
func CalcTFIDF(text string, idf VectorF) VectorF {
	tf := CalcTF(text)
	tfidf := make(VectorF)

	for word, tfV := range tf {
		tfidf[word] = tfV * idf[word]
	}
	return tfidf
}
