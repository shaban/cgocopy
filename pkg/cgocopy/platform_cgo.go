package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include <stddef.h>

// Primitive type size functions for platform_test.go
size_t sizeOfChar() { return sizeof(char); }
size_t sizeOfShort() { return sizeof(short); }
size_t sizeOfInt() { return sizeof(int); }
size_t sizeOfLong() { return sizeof(long); }
size_t sizeOfLongLong() { return sizeof(long long); }
size_t sizeOfInt8() { return sizeof(int8_t); }
size_t sizeOfInt16() { return sizeof(int16_t); }
size_t sizeOfInt32() { return sizeof(int32_t); }
size_t sizeOfInt64() { return sizeof(int64_t); }
size_t sizeOfUInt8() { return sizeof(uint8_t); }
size_t sizeOfUInt16() { return sizeof(uint16_t); }
size_t sizeOfUInt32() { return sizeof(uint32_t); }
size_t sizeOfUInt64() { return sizeof(uint64_t); }
size_t sizeOfFloat() { return sizeof(float); }
size_t sizeOfDouble() { return sizeof(double); }
size_t sizeOfPointer() { return sizeof(void*); }
size_t sizeOfSizeT() { return sizeof(size_t); }
*/
import "C"

// C primitive type sizes - populated at init time
var (
	CCharSize     uintptr
	CShortSize    uintptr
	CIntSize      uintptr
	CLongSize     uintptr
	CLongLongSize uintptr
	CInt8Size     uintptr
	CInt16Size    uintptr
	CInt32Size    uintptr
	CInt64Size    uintptr
	CUInt8Size    uintptr
	CUInt16Size   uintptr
	CUInt32Size   uintptr
	CUInt64Size   uintptr
	CFloatSize    uintptr
	CDoubleSize   uintptr
	CPointerSize  uintptr
	CSizeTSize    uintptr
)

func init() {
	CCharSize = uintptr(C.sizeOfChar())
	CShortSize = uintptr(C.sizeOfShort())
	CIntSize = uintptr(C.sizeOfInt())
	CLongSize = uintptr(C.sizeOfLong())
	CLongLongSize = uintptr(C.sizeOfLongLong())
	CInt8Size = uintptr(C.sizeOfInt8())
	CInt16Size = uintptr(C.sizeOfInt16())
	CInt32Size = uintptr(C.sizeOfInt32())
	CInt64Size = uintptr(C.sizeOfInt64())
	CUInt8Size = uintptr(C.sizeOfUInt8())
	CUInt16Size = uintptr(C.sizeOfUInt16())
	CUInt32Size = uintptr(C.sizeOfUInt32())
	CUInt64Size = uintptr(C.sizeOfUInt64())
	CFloatSize = uintptr(C.sizeOfFloat())
	CDoubleSize = uintptr(C.sizeOfDouble())
	CPointerSize = uintptr(C.sizeOfPointer())
	CSizeTSize = uintptr(C.sizeOfSizeT())
}
