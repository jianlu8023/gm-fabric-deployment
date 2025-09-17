package qa

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
