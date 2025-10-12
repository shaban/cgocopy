package cgocopy

import (
	"testing"
	"unsafe"
)

var copySink byte

func BenchmarkMemCopyPureGo(b *testing.B) {
	size := 32 * 1024
	src := make([]byte, size)
	for i := range src {
		src[i] = byte(i)
	}
	dst := make([]byte, size)
	ptrSrc := unsafe.Pointer(&src[0])
	ptrDst := unsafe.Pointer(&dst[0])

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memCopy(ptrDst, ptrSrc, uintptr(size))
	}

	copySink ^= dst[13]
}

func BenchmarkMemCopyCgo(b *testing.B) {
	size := 32 * 1024
	src := make([]byte, size)
	for i := range src {
		src[i] = byte(i * 3)
	}
	dst := make([]byte, size)
	ptrSrc := unsafe.Pointer(&src[0])
	ptrDst := unsafe.Pointer(&dst[0])

	b.SetBytes(int64(size))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cgoMemCopy(ptrDst, ptrSrc, uintptr(size))
	}

	copySink ^= dst[17]
}
