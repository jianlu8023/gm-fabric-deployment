package gotoon

import (
	"testing"
	"time"
)

func TestEncodePrimitives(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "null",
			input:    nil,
			expected: "null",
		},
		{
			name:     "boolean true",
			input:    true,
			expected: "true",
		},
		{
			name:     "boolean false",
			input:    false,
			expected: "false",
		},
		{
			name:     "integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "float",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with spaces",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeObject(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "simple object",
			input: map[string]interface{}{
				"id":   123,
				"name": "Ada",
			},
			expected: "id: 123\nname: Ada",
		},
		{
			name: "nested object",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   123,
					"name": "Ada",
				},
			},
			expected: "user:\n  id: 123\n  name: Ada",
		},
		{
			name: "object with boolean",
			input: map[string]interface{}{
				"active": true,
				"name":   "test",
			},
			expected: "active: true\nname: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestEncodePrimitiveArray(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: "[0]:",
		},
		{
			name:     "array of strings",
			input:    []string{"reading", "gaming"},
			expected: "[2]: reading,gaming",
		},
		{
			name:     "array of numbers",
			input:    []int{1, 2, 3},
			expected: "[3]: 1,2,3",
		},
		{
			name: "object with array",
			input: map[string]interface{}{
				"tags": []string{"admin", "ops", "dev"},
			},
			expected: "tags[3]: admin,ops,dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestEncodeTabularArray(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "array of objects with same keys",
			input: map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": 1, "name": "Alice", "role": "admin"},
					{"id": 2, "name": "Bob", "role": "user"},
				},
			},
			expected: "items[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user",
		},
		{
			name: "array of objects with numbers",
			input: map[string]interface{}{
				"items": []map[string]interface{}{
					{"sku": "A1", "qty": 2, "price": 9.99},
					{"sku": "B2", "qty": 1, "price": 14.5},
				},
			},
			expected: "items[2]{price,qty,sku}:\n  9.99,2,A1\n  14.5,1,B2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestEncodeMixedArray(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "mixed array",
			input: map[string]interface{}{
				"items": []interface{}{
					1,
					"text",
					map[string]interface{}{"a": 1},
				},
			},
			expected: "items[3]:\n  - 1\n  - text\n  - a: 1",
		},
		{
			name: "array of objects with different keys",
			input: map[string]interface{}{
				"items": []map[string]interface{}{
					{"id": 1, "name": "First"},
					{"id": 2, "name": "Second", "extra": true},
				},
			},
			expected: "items[2]:\n  - id: 1\n    name: First\n  - extra: true\n    id: 2\n    name: Second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestEncodeWithOptions(t *testing.T) {
	input := map[string]interface{}{
		"tags": []string{"a", "b", "c"},
	}

	t.Run("custom indent", func(t *testing.T) {
		result, err := Encode(input, WithIndent(4))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "tags[3]: a,b,c"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("tab delimiter", func(t *testing.T) {
		result, err := Encode(input, WithDelimiter("\t"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "tags[3\t]: a\tb\tc"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("length marker", func(t *testing.T) {
		result, err := Encode(input, WithLengthMarker())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "tags[#3]: a,b,c"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

func TestEncodeStruct(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
	}

	input := map[string]interface{}{
		"users": []User{
			{ID: 1, Name: "Alice", Role: "admin"},
			{ID: 2, Name: "Bob", Role: "user"},
		},
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user"
	if result != expected {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, result)
	}
}

func TestEncodeTime(t *testing.T) {
	tm := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	input := map[string]interface{}{
		"timestamp": tm,
	}

	result, err := Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "timestamp: \"2025-01-15T10:30:00Z\""
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeQuoting(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "string with comma",
			input: map[string]interface{}{
				"note": "hello, world",
			},
			expected: "note: \"hello, world\"",
		},
		{
			name: "string looking like boolean",
			input: map[string]interface{}{
				"items": []string{"true", "false"},
			},
			expected: "items[2]: \"true\",\"false\"",
		},
		{
			name: "empty string",
			input: map[string]interface{}{
				"name": "",
			},
			expected: "name: \"\"",
		},
		{
			name: "string with quotes",
			input: map[string]interface{}{
				"text": "say \"hi\"",
			},
			expected: "text: \"say \\\"hi\\\"\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestREADMEExample(t *testing.T) {
	data := map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "Alice", "role": "admin"},
			{"id": 2, "name": "Bob", "role": "user"},
		},
	}

	result, err := Encode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "users[2]{id,name,role}:\n  1,Alice,admin\n  2,Bob,user"
	if result != expected {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, result)
	}
}
