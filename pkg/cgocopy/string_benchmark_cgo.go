package cgocopy

/*
#include <stdlib.h>
#include <string.h>

// String helper functions for string_benchmark_test.go

char* createTestString() {
    return strdup("This is a test string for benchmarking C string conversion methods!");
}

char* createCString(const char* str) {
    return strdup(str);
}
*/
import "C"
import "unsafe"

// Helper functions for string benchmarking (wraps CGO calls)
func cStrlen(ptr unsafe.Pointer) int {
	return int(C.strlen((*C.char)(ptr)))
}

func cGoString(ptr unsafe.Pointer) string {
	return C.GoString((*C.char)(ptr))
}

func cGoStringN(ptr unsafe.Pointer, length int) string {
	return C.GoStringN((*C.char)(ptr), C.int(length))
}

func createTestCString() unsafe.Pointer {
	return unsafe.Pointer(C.createTestString())
}

func createCString(s string) unsafe.Pointer {
	cs := C.CString(s)
	return unsafe.Pointer(cs)
}

func freeCString(ptr unsafe.Pointer) {
	C.free(ptr)
}
