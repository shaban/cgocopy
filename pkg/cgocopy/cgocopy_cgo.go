package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include <stddef.h>

// Basic test struct for structcopy_test.go
typedef struct {
    int32_t id;
    float value;
    int64_t timestamp;
} TestStruct;

// Create test struct
TestStruct* createTestStruct() {
    TestStruct* s = (TestStruct*)malloc(sizeof(TestStruct));
    s->id = 42;
    s->value = 3.14f;
    s->timestamp = 1234567890;
    return s;
}

void freeTestStruct(TestStruct* s) {
    free(s);
}

// Field offset functions using offsetof()
size_t testStructIdOffset() { return offsetof(TestStruct, id); }
size_t testStructValueOffset() { return offsetof(TestStruct, value); }
size_t testStructTimestampOffset() { return offsetof(TestStruct, timestamp); }
*/
import "C"
import "unsafe"

// C struct layout information - populated at init time
var (
	testStructSize            uintptr
	testStructIdOffset        uintptr
	testStructValueOffset     uintptr
	testStructTimestampOffset uintptr
)

func init() {
	testStructSize = uintptr(unsafe.Sizeof(C.TestStruct{}))
	testStructIdOffset = uintptr(C.testStructIdOffset())
	testStructValueOffset = uintptr(C.testStructValueOffset())
	testStructTimestampOffset = uintptr(C.testStructTimestampOffset())
}

func CreateTestStruct() unsafe.Pointer {
	return unsafe.Pointer(C.createTestStruct())
}

func FreeTestStruct(ptr unsafe.Pointer) {
	C.freeTestStruct((*C.TestStruct)(ptr))
}

func TestStructSize() uintptr {
	return uintptr(unsafe.Sizeof(C.TestStruct{}))
}
