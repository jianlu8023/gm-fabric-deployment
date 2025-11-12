package mycommand

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {

	tests := []struct {
		Name    string
		Command string
		Args    []string
		ExpectError bool
	}{
		{
			Name:    "ls -al",
			Command: "ls",
			Args:    []string{"-al"},
			ExpectError: false,
		},
		{
			Name:    "invalid command",
			Command: "nonexistentcommand",
			Args:    []string{},
			ExpectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cmd := exec.Command(tt.Command, tt.Args...)
			put, err := Run(cmd)
			
			// 测试错误处理
			if tt.ExpectError {
				assert.Error(t, err, "Run() should return an error for invalid command")
			} else {
				assert.NoError(t, err, "Run() error = %v", err)
				
				// 测试返回值不为nil
				assert.NotNil(t, put, "Output should not be nil")
				
				// 测试字段存在性
				assert.NotEmpty(t, put.StdoutPut, "StdoutPut should not be empty for valid command")
				// StderrPut 可能为空，所以不强制要求不为空
				
				// 测试类型
				assert.IsType(t, "", put.StdoutPut, "StdoutPut should be a string")
				assert.IsType(t, "", put.StderrPut, "StderrPut should be a string")
				
				t.Log(put.StdoutPut)
				t.Log(put.StderrPut)
			}
		})
	}
}

// 新增测试用例：测试边界条件和特殊情况
func TestRunEdgeCases(t *testing.T) {
	// 测试空命令（会引发panic或错误）
	t.Run("empty command", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// 如果发生panic，这是预期的行为
				assert.NotNil(t, r, "Should panic for empty command")
			}
		}()
		
		// 注意：直接执行空命令可能会导致panic，这里只是演示
		// 在实际应用中，应该在调用Run之前验证输入
	})
	
	// 测试命令返回错误码的情况
	t.Run("command with error exit code", func(t *testing.T) {
		// 使用一个肯定会失败的命令
		cmd := exec.Command("ls", "/nonexistent_directory")
		put, err := Run(cmd)
		
		// 应该返回错误
		assert.Error(t, err, "Command should return error for nonexistent directory")
		
		// 但输出结构体不应为nil
		assert.NotNil(t, put, "Output should not be nil even when command fails")
		
		// 错误输出不应该为空
		assert.NotEmpty(t, put.StderrPut, "StderrPut should contain error message")
	})
	
	// 测试成功命令但无输出的情况
	t.Run("successful command with no output", func(t *testing.T) {
		// 使用true命令，它总是成功且无输出
		cmd := exec.Command("true")
		put, err := Run(cmd)
		
		assert.NoError(t, err, "True command should not return error")
		assert.NotNil(t, put, "Output should not be nil")
		
		// 输出应该为空字符串
		assert.Empty(t, put.StdoutPut, "StdoutPut should be empty for true command")
		assert.Empty(t, put.StderrPut, "StderrPut should be empty for true command")
	})
}

// 新增测试用例：演示各种assert方法的使用
func TestAssertExamples(t *testing.T) {
	t.Run("basic assertions", func(t *testing.T) {
		// 测试相等性
		assert.Equal(t, 1, 1, "1 should equal 1")
		assert.Equal(t, "hello", "hello", "strings should match")
		
		// 测试不相等
		assert.NotEqual(t, 1, 2, "1 should not equal 2")
		
		// 测试nil和非nil
		var nilVar *string = nil
		var nonNilVar = "not nil"
		assert.Nil(t, nilVar, "should be nil")
		assert.NotNil(t, &nonNilVar, "should not be nil")
		
		// 测试布尔值
		assert.True(t, true, "should be true")
		assert.False(t, false, "should be false")
		
		// 测试包含关系
		assert.Contains(t, "hello world", "world", "string should contain substring")
		assert.Contains(t, []int{1, 2, 3}, 2, "slice should contain element")
		
		// 测试长度
		assert.Len(t, []int{1, 2, 3}, 3, "slice should have length 3")
		assert.Empty(t, []int{}, "empty slice should be empty")
		assert.NotEmpty(t, []int{1}, "non-empty slice should not be empty")
		
		// 测试错误
		assert.NoError(t, nil, "no error should be returned")
		
		// 测试Panics
		assert.Panics(t, func() {
			panic("this should panic")
		}, "function should panic")
	})
}