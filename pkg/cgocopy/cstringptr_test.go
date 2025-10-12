package cgocopy

import (
	"testing"
	"unsafe"
)

func makeCStringBytes(s string) unsafe.Pointer {
	buf := append([]byte(s), 0)
	return unsafe.Pointer(&buf[0])
}

func TestUTF8Converter(t *testing.T) {
	ptr := makeCStringBytes("Hello, World!")

	conv := UTF8Converter{}
	if got := conv.CStringToGo(ptr); got != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %q", got)
	}
}

func TestUTF8ConverterNil(t *testing.T) {
	conv := UTF8Converter{}
	if got := conv.CStringToGo(nil); got != "" {
		t.Fatalf("expected empty string for nil pointer, got %q", got)
	}
}

func BenchmarkUTF8Converter(b *testing.B) {
	ptr := makeCStringBytes("Benchmark String")
	conv := UTF8Converter{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conv.CStringToGo(ptr)
	}
}
