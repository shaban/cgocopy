package cgocopy

/*
#include <stdlib.h>
#include <string.h>

// Test helper: create a C string
static char* testCreateString(const char* str) {
    return strdup(str);
}

// Test helper: free C string
static void testFreeString(char* str) {
    free(str);
}

// Test struct with char* pointer
typedef struct {
    uint32_t id;
    char* name;
    uint32_t channels;
} TestCDevicePtr;
*/
import "C"
import "unsafe"

// Test helpers (CGO must be in .go files, not _test.go)

func testCreateCString(s string) unsafe.Pointer {
	cstr := C.CString(s)
	result := C.testCreateString(cstr)
	C.free(unsafe.Pointer(cstr))
	return unsafe.Pointer(result)
}

func testFreeCString(ptr unsafe.Pointer) {
	C.testFreeString((*C.char)(ptr))
}

type testCDevicePtr struct {
	id       uint32
	name     unsafe.Pointer
	channels uint32
}

func testCreateCDevicePtr(id uint32, name string, channels uint32) (testCDevicePtr, func()) {
	namePtr := testCreateCString(name)
	device := testCDevicePtr{
		id:       id,
		name:     namePtr,
		channels: channels,
	}
	cleanup := func() {
		testFreeCString(namePtr)
	}
	return device, cleanup
}
