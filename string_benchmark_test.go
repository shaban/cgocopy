package cgocopy

import (
	"testing"
	"unsafe"
)

// CStringToGoString converts a C string (char*) to a Go string
// This is the optimal approach - uses C.GoString which does strlen + copy in one CGO call
func CStringToGoString(cPtr unsafe.Pointer) string {
	if cPtr == nil {
		return ""
	}
	return cGoString(cPtr)
}

// Benchmark C.GoString conversion
func BenchmarkCStringConversion(b *testing.B) {
	ptr := createTestCString()
	defer freeCString(ptr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CStringToGoString(ptr)
	}
}

// Test different string lengths
func BenchmarkCStringByLength(b *testing.B) {
	testCases := []struct {
		name string
		str  string
	}{
		{"short_5_chars", "Short"},
		{"medium_30_chars", "This is a medium length test!"},
		{"long_200_chars", "This is a much longer string that we want to test to see how the performance scales with string length. We're making it quite long to test the overhead of C.GoString which does both strlen and memory copy in a single CGO call. This should be around 200 characters."},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			ptr := createCString(tc.str)
			defer freeCString(ptr)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = CStringToGoString(ptr)
			}
		})
	}
}

// Test correctness
func TestCStringConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single_char", "a", "a"},
		{"simple", "Hello", "Hello"},
		{"with_spaces", "This is a test string with some length to it!", "This is a test string with some length to it!"},
		{"unicode", "Unicode: 你好世界 🚀", "Unicode: 你好世界 🚀"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == "" {
				// Test nil pointer
				result := CStringToGoString(nil)
				if result != "" {
					t.Error("Nil pointer should return empty string")
				}
				return
			}

			ptr := createCString(tt.input)
			result := CStringToGoString(ptr)
			freeCString(ptr)

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
