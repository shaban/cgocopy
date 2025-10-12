package cgocopy

/*
#include <stdlib.h>
#include <string.h>
#include <stddef.h>

void cgocopy_memcpy(void* dst, const void* src, size_t n) {
    memcpy(dst, src, n);
}
*/
import "C"
import "unsafe"

func cgoMemCopy(dst, src unsafe.Pointer, size uintptr) {
	if size == 0 {
		return
	}
	C.cgocopy_memcpy(dst, src, C.size_t(size))
}
