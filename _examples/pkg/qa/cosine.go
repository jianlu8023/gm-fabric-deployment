package qa

import (
	"math"
)

// CosineSimilarityVectorI 计算两个向量的余弦相似度
// 公式 cos_sim = A•B / (||A||*||B||) = sum(a[i]*b[i]) / (sqrt(sum(a[i]^2))*sqrt(sum(b[i]^2)))
func CosineSimilarityVectorI(vec1, vec2 VectorI) float64 {
	dotProduct := 0.0
	vec1Norm := 0.0
	vec2Norm := 0.0

	for k, v1 := range vec1 {
		vec1Norm += float64(v1) * float64(v1)
		if v2, ok := vec2[k]; ok {
			dotProduct += float64(v1) * float64(v2)
		}
	}

	for _, v2 := range vec2 {
		vec2Norm += float64(v2) * float64(v2)
	}

	// 处理零向量情况
	if vec1Norm == 0 || vec2Norm == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(vec1Norm) * math.Sqrt(vec2Norm))
}

// CosineSimilarityVectorF 计算余弦相似度
func CosineSimilarityVectorF(vec1, vec2 VectorF) float64 {
	dotProduct := 0.0
	vec1Norm := 0.0
	vec2Norm := 0.0

	for k, v1 := range vec1 {
		vec1Norm += v1 * v1
		if v2, ok := vec2[k]; ok {
			dotProduct += v1 * v2
		}
	}

	for _, v2 := range vec2 {
		vec2Norm += v2 * v2
	}

	// 处理零向量情况
	if vec1Norm == 0 || vec2Norm == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(vec1Norm) * math.Sqrt(vec2Norm))
}
