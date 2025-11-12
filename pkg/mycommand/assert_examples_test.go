package mycommand

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertExamples 演示各种assert方法的使用场景
func TestAssertUsageExamples(t *testing.T) {

	// 1. 基本相等性断言
	t.Run("equality assertions", func(t *testing.T) {
		// Equal - 测试值相等
		assert.Equal(t, 42, 42, "数值应该相等")
		assert.Equal(t, "hello", "hello", "字符串应该相等")

		// NotEqual - 测试值不相等
		assert.NotEqual(t, 1, 2, "数值应该不相等")
		assert.NotEqual(t, "hello", "world", "字符串应该不相等")

		// Same - 测试相同实例（指针）
		a := "hello"
		b := a
		c := "hello"
		assert.Same(t, &a, &b, "应该是相同实例")
		assert.NotSame(t, &a, &c, "应该是不同实例")
	})

	// 2. Nil和非Nil断言
	t.Run("nil assertions", func(t *testing.T) {
		var nilPtr *string = nil
		var notNilPtr = new(string)

		// Nil - 测试为nil
		assert.Nil(t, nilPtr, "应该为nil")

		// NotNil - 测试不为nil
		assert.NotNil(t, notNilPtr, "不应该为nil")

		// Empty - 测试为空
		assert.Empty(t, "", "字符串应该为空")
		assert.Empty(t, []int{}, "切片应该为空")
		assert.Empty(t, map[string]int{}, "map应该为空")

		// NotEmpty - 测试非空
		assert.NotEmpty(t, "hello", "字符串应该非空")
		assert.NotEmpty(t, []int{1}, "切片应该非空")
	})

	// 3. 布尔值断言
	t.Run("boolean assertions", func(t *testing.T) {
		assert.True(t, 1 == 1, "应该为true")
		assert.False(t, 1 == 2, "应该为false")
	})

	// 4. 错误处理断言
	t.Run("error assertions", func(t *testing.T) {
		// NoError - 测试无错误
		err := func() error { return nil }()
		assert.NoError(t, err, "应该没有错误")

		// Error - 测试有错误
		err = errors.New("some error")
		assert.Error(t, err, "应该有错误")

		// ErrorContains - 测试错误包含特定文本
		assert.ErrorContains(t, err, "error", "错误应该包含'error'")

		// ErrorAs - 测试错误类型转换
		var target *MyError
		err = &MyError{Msg: "test error"}
		assert.ErrorAs(t, err, &target, "错误应该能转换为MyError类型")

		// ErrorIs - 测试错误相等性
		assert.ErrorIs(t, err, err, "错误应该等于自身")
	})

	// 5. 包含关系断言
	t.Run("containment assertions", func(t *testing.T) {
		// Contains - 测试包含
		assert.Contains(t, "hello world", "world", "字符串应该包含子串")
		assert.Contains(t, []int{1, 2, 3}, 2, "切片应该包含元素")
		assert.Contains(t, map[string]int{"a": 1, "b": 2}, "a", "map应该包含键")

		// NotContains - 测试不包含
		assert.NotContains(t, "hello world", "xyz", "字符串不应该包含子串")
		assert.NotContains(t, []int{1, 2, 3}, 4, "切片不应该包含元素")
	})

	// 6. 长度断言
	t.Run("length assertions", func(t *testing.T) {
		assert.Len(t, []int{1, 2, 3}, 3, "切片长度应该为3")
		assert.Len(t, "hello", 5, "字符串长度应该为5")
		assert.Len(t, map[string]int{"a": 1, "b": 2}, 2, "map长度应该为2")
	})

	// 7. Panics断言
	t.Run("panic assertions", func(t *testing.T) {
		// Panics - 测试会panic
		assert.Panics(t, func() {
			panic("something went wrong")
		}, "函数应该panic")

		// NotPanics - 测试不会panic
		assert.NotPanics(t, func() {
			// 正常函数
		}, "函数不应该panic")

		// PanicsWithValue - 测试panic特定值
		assert.PanicsWithValue(t, "expected panic", func() {
			panic("expected panic")
		}, "应该panic特定值")
	})

	// 8. 类型断言
	t.Run("type assertions", func(t *testing.T) {
		var i interface{} = 42
		assert.IsType(t, 0, i, "应该为int类型")

		var s interface{} = "hello"
		assert.IsType(t, "", s, "应该为string类型")
	})

	// 9. 文件系统断言
	t.Run("file assertions", func(t *testing.T) {
		// FileExists - 测试文件存在
		// assert.FileExists(t, "existing_file.txt", "文件应该存在")

		// NoFileExists - 测试文件不存在
		// assert.NoFileExists(t, "nonexistent_file.txt", "文件不应该存在")

		// DirExists - 测试目录存在
		// assert.DirExists(t, "/tmp", "目录应该存在")
	})

	// 10. HTTP断言
	t.Run("http assertions", func(t *testing.T) {
		// HTTPSuccess - 测试HTTP请求成功
		// assert.HTTPSuccess(t, myHandler, "GET", "/path", nil)

		// HTTPBodyContains - 测试HTTP响应体包含
		// assert.HTTPBodyContains(t, myHandler, "GET", "/path", nil, "expected content")
	})

	// 11. YAML断言
	t.Run("yaml assertions", func(t *testing.T) {
		// YAMLContains - 测试YAML包含
		// assert.YAMLEq(t, `a: 1`, `a: 1`, "YAML应该相等")
	})

	// 12. JSON断言
	t.Run("json assertions", func(t *testing.T) {
		// JSONEq - 测试JSON相等
		// assert.JSONEq(t, `{"a": 1}`, `{"a": 1}`, "JSON应该相等")
	})
}

// 辅助类型用于错误测试
type MyError struct {
	Msg string
}

func (e *MyError) Error() string {
	return e.Msg
}

// 示例：何时使用各种assert方法的场景说明
func TestWhenToUseAsserts(t *testing.T) {

	// 何时使用Nil/NotNil:
	// - Nil: 当函数返回指针且期望为nil时（如查找未找到返回nil）
	// - NotNil: 当函数返回指针且期望不为nil时（如创建对象成功）

	// 何时使用Error/NoError:
	// - Error: 当函数应该返回错误时（如无效输入、资源不存在）
	// - NoError: 当函数不应该返回错误时（如正常操作）

	// 何时使用Empty/NotEmpty:
	// - Empty: 当期望集合为空时（如初始化状态、清空操作后）
	// - NotEmpty: 当期望集合不为空时（如查询结果、列表数据）

	// 何时使用Contains/NotContains:
	// - Contains: 当期望在集合中找到元素时（如白名单检查）
	// - NotContains: 当期望在集合中找不到元素时（如黑名单检查）

	// 何时使用Panics/NotPanics:
	// - Panics: 当函数应该panic时（如违反前置条件）
	// - NotPanics: 当函数不应该panic时（如正常流程）
}
