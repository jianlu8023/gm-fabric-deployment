package i18n

import (
	"net/http/httptest"
	"testing"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetLanguageFromContext(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		queryLang    string
		acceptLang   string
		expectedLang string
	}{
		{
			name:         "Query parameter zh",
			queryLang:    "zh",
			acceptLang:   "",
			expectedLang: "zh",
		},
		{
			name:         "Query parameter en",
			queryLang:    "en",
			acceptLang:   "",
			expectedLang: "en",
		},
		{
			name:         "Accept-Language header zh-CN",
			queryLang:    "",
			acceptLang:   "zh-CN,zh;q=0.9,en;q=0.8",
			expectedLang: "zh",
		},
		{
			name:         "Accept-Language header en-US",
			queryLang:    "",
			acceptLang:   "en-US,en;q=0.9,zh;q=0.8",
			expectedLang: "en",
		},
		{
			name:         "Default language",
			queryLang:    "",
			acceptLang:   "",
			expectedLang: "zh",
		},
		{
			name:         "Unsupported language fallback",
			queryLang:    "fr",
			acceptLang:   "",
			expectedLang: "zh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建Gin上下文
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// 创建请求
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.queryLang != "" {
				q := req.URL.Query()
				q.Add("lang", tt.queryLang)
				req.URL.RawQuery = q.Encode()
			}
			if tt.acceptLang != "" {
				req.Header.Set("Accept-Language", tt.acceptLang)
			}
			ctx.Request = req

			// 测试获取语言
			lang := GetLanguageFromContext(ctx)
			assert.Equal(t, tt.expectedLang, lang)
		})
	}
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		key          string
		expectedText string
	}{
		{
			name:         "Chinese success message",
			lang:         "zh",
			key:          "success",
			expectedText: "业务处理成功",
		},
		{
			name:         "English success message",
			lang:         "en",
			key:          "success",
			expectedText: "Operation successful",
		},
		{
			name:         "Unsupported language fallback",
			lang:         "fr",
			key:          "success",
			expectedText: "业务处理成功",
		},
		{
			name:         "Non-existent key",
			lang:         "zh",
			key:          "nonexistent",
			expectedText: "nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := Translate(tt.lang, tt.key)
			assert.Equal(t, tt.expectedText, text)
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		expectedLang string
	}{
		{
			name:         "Lowercase zh",
			lang:         "zh",
			expectedLang: "zh",
		},
		{
			name:         "Uppercase ZH",
			lang:         "ZH",
			expectedLang: "zh",
		},
		{
			name:         "zh-CN with region",
			lang:         "zh-CN",
			expectedLang: "zh",
		},
		{
			name:         "en-US with region",
			lang:         "en-US",
			expectedLang: "en",
		},
		{
			name:         "Unsupported language",
			lang:         "fr",
			expectedLang: "zh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用反射调用私有函数normalizeLanguage
			// 这里我们通过间接方式测试，即直接调用GetLanguageFromContext
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/test?lang="+tt.lang, nil)
			ctx.Request = req

			lang := GetLanguageFromContext(ctx)
			assert.Equal(t, tt.expectedLang, lang)
		})
	}
}
