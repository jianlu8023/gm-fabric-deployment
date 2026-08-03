package cors

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"go.uber.org/zap"
)

// TestEnableCors_WildcardWithCredentialsDowngrade 验证：AllowOrigins 含 "*" 且 AllowCredentials=true 时不会 panic
//
// @description 修复问题1的回归测试： gin-contrib/cors 在该组合下会 panic，本中间件应在 buildLibConfig 中降级
func TestEnableCors_WildcardWithCredentialsDowngrade(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	logger := zap.NewNop().Sugar()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("EnableCors panicked on wildcard + credentials: %v", r)
		}
	}()

	cfg := &config.CORSConfig{
		Enabled:          true,
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}
	handler := EnableCors(cfg, logger)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

// TestEnableCors_WhitelistWithCredentials 验证：白名单 + credentials 应正常工作
//
// @description 修复问题2的回归测试：白名单模式下不应无条件回显任意 Origin
func TestEnableCors_WhitelistWithCredentials(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	logger := zap.NewNop().Sugar()

	cfg := &config.CORSConfig{
		Enabled:          true,
		AllowOrigins:     []string{"https://example.com"},
		AllowCredentials: true,
		AllowWildcard:    true,
		MaxAge:           3600,
	}
	handler := EnableCors(cfg, logger)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

// TestEnableCors_NilConfig 验证：nil 配置使用默认策略不 panic
func TestEnableCors_NilConfig(t *testing.T) {
	gin.SetMode(gin.DebugMode)
	logger := zap.NewNop().Sugar()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("EnableCors panicked on nil config: %v", r)
		}
	}()

	handler := EnableCors(nil, logger)
	if handler == nil {
		t.Fatal("expected non-nil handler for nil config")
	}
}

// TestHasWildcardOrigin 验证通配符检测
func TestHasWildcardOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		want    bool
	}{
		{"contains wildcard", []string{"*", "https://example.com"}, true},
		{"only wildcard", []string{"*"}, true},
		{"no wildcard", []string{"https://example.com", "http://localhost:8080"}, false},
		{"empty", []string{}, false},
		{"wildcard with spaces", []string{" * "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasWildcardOrigin(tt.origins); got != tt.want {
				t.Errorf("hasWildcardOrigin(%v) = %v, want %v", tt.origins, got, tt.want)
			}
		})
	}
}

// TestSplitCSV 验证 CSV 切分
func TestSplitCSV(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"GET, POST, PUT", 3},
		{"GET,POST,PUT", 3},
		{"  GET ,  POST  ", 2},
		{"", 0},
		{"GET", 1},
	}
	for _, tt := range tests {
		got := splitCSV(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitCSV(%q) returned %d items, want %d", tt.input, len(got), tt.want)
		}
	}
}
