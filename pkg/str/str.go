package str

import (
	"strings"
)

const (
	// True 字符串true
	True = "true"
	// False 字符串false
	False = "false"
	// Blank 空字符串
	Blank = ""
)

// IsBlank 判断字符串是否为空
// @param str 字符串
// @return 是否为空
func IsBlank(str string) bool {
	return CompareIgnoreCase(Blank, str)
}

// IsTrue 判断字符串是否为true
// @param str 字符串
// @return 是否为true
func IsTrue(str string) bool {
	return CompareIgnoreCase(True, str)
}

// IsFalse 判断字符串是否为false
// @param str 字符串
// @return 是否为false
func IsFalse(str string) bool {
	return CompareIgnoreCase(False, str)
}

// CompareIgnoreCase 比较两个字符串是否一样
// @param str1: 字符串1
// @param str2: 字符串2
// @return bool: 两个字符串是否一致
func CompareIgnoreCase(str1, str2 string) bool {

	// 去除空格
	str1 = strings.TrimSpace(str1)
	str2 = strings.TrimSpace(str2)

	// 去除str中间的空格
	str1 = strings.Join(strings.Fields(str1), "")
	str2 = strings.Join(strings.Fields(str2), "")

	// 全部转小写
	str1 = strings.ToLower(str1)
	str2 = strings.ToLower(str2)

	// 比较字符串
	return str1 == str2
}

// DeduplicateString 对指定分隔符的字符串进行去重，保留元素首次出现的顺序
// @param str 要去重的字符串
// @param separator 用于分割和拼接字符串的分隔符
// @param ignoreEmpty 是否忽略因连续分隔符或开头/结尾分隔符产生的空字符串元素
// @return string 去重后的字符串
func DeduplicateString(str string, separator string, ignoreEmpty bool) string {
	// 如果输入是空字符串，直接返回空字符串
	if str == "" {
		return ""
	}

	// 如果分隔符是空字符串， Split 会将每个字符作为元素，Join 会直接拼接
	// 这可能是预期行为，所以不做特殊处理
	// if separator == "" { ... }

	// 1. 分割字符串
	elements := strings.Split(str, separator)

	// 2. 去重并保持顺序
	// 使用 map 来记录已经遇到的元素，key 是元素值，value 是空结构体（节省内存）
	seen := make(map[string]struct{})
	// 使用一个切片来存储去重后的元素，按照它们首次出现的顺序
	var uniqueElements []string

	for _, element := range elements {
		// 根据 ignoreEmpty 参数，判断是否跳过空字符串元素
		if ignoreEmpty && IsBlank(element) {
			continue
		}

		// 检查当前元素是否已经在 seen map 中
		if _, ok := seen[element]; !ok {
			// 如果不在，说明是第一次遇到这个元素
			// 将元素添加到 seen map 中
			seen[element] = struct{}{}
			// 将元素添加到 uniqueElements 切片中，保持顺序
			uniqueElements = append(uniqueElements, element)
		}
	}

	// 3. 拼接字符串
	// 使用指定的分隔符将去重后的元素切片拼接回字符串
	return strings.Join(uniqueElements, separator)
}
