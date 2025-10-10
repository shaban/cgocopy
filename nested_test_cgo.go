package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include <stddef.h>

// Nested struct definitions for nested_test.go and deep_nesting_test.go

// Inner nested struct
typedef struct {
    int32_t x;
    int32_t y;
} InnerStruct;

// Middle level struct with nested inner
typedef struct {
    int32_t id;
    InnerStruct inner;
    float value;
} MiddleStruct;

// Outer struct with multiple levels of nesting
typedef struct {
    int32_t outerID;
    MiddleStruct middle;
    int64_t timestamp;
} OuterStruct;

// 5-level deep nesting test: Level1 → Level2 → Level3 → Level4 → Level5
typedef struct {
    int32_t value;
} Level5;

typedef struct {
    int32_t id;
    Level5 level5;
} Level4;

typedef struct {
    int32_t id;
    Level4 level4;
} Level3;

typedef struct {
    int32_t id;
    Level3 level3;
} Level2;

typedef struct {
    int32_t id;
    Level2 level2;
} Level1;

// Create functions
InnerStruct* createInnerStruct() {
    InnerStruct* s = (InnerStruct*)malloc(sizeof(InnerStruct));
    s->x = 10;
    s->y = 20;
    return s;
}

MiddleStruct* createMiddleStruct() {
    MiddleStruct* s = (MiddleStruct*)malloc(sizeof(MiddleStruct));
    s->id = 100;
    s->inner.x = 10;
    s->inner.y = 20;
    s->value = 3.14f;
    return s;
}

OuterStruct* createOuterStruct() {
    OuterStruct* s = (OuterStruct*)malloc(sizeof(OuterStruct));
    s->outerID = 1000;
    s->middle.id = 100;
    s->middle.inner.x = 10;
    s->middle.inner.y = 20;
    s->middle.value = 3.14f;
    s->timestamp = 9999999;
    return s;
}

Level1* createDeepNested() {
    Level1* s = (Level1*)malloc(sizeof(Level1));
    s->id = 1;
    s->level2.id = 2;
    s->level2.level3.id = 3;
    s->level2.level3.level4.id = 4;
    s->level2.level3.level4.level5.value = 999;
    return s;
}

void freeDeepNested(Level1* s) {
    free(s);
}

// Offset functions
size_t innerStructXOffset() { return offsetof(InnerStruct, x); }
size_t innerStructYOffset() { return offsetof(InnerStruct, y); }

size_t middleStructIdOffset() { return offsetof(MiddleStruct, id); }
size_t middleStructInnerOffset() { return offsetof(MiddleStruct, inner); }
size_t middleStructValueOffset() { return offsetof(MiddleStruct, value); }

size_t outerStructOuterIDOffset() { return offsetof(OuterStruct, outerID); }
size_t outerStructMiddleOffset() { return offsetof(OuterStruct, middle); }
size_t outerStructTimestampOffset() { return offsetof(OuterStruct, timestamp); }

// Deep nesting offsets and sizes
size_t sizeofLevel5() { return sizeof(Level5); }
size_t sizeofLevel4() { return sizeof(Level4); }
size_t sizeofLevel3() { return sizeof(Level3); }
size_t sizeofLevel2() { return sizeof(Level2); }
size_t sizeofLevel1() { return sizeof(Level1); }

size_t offsetofLevel4Level5() { return offsetof(Level4, level5); }
size_t offsetofLevel3Level4() { return offsetof(Level3, level4); }
size_t offsetofLevel2Level3() { return offsetof(Level2, level3); }
size_t offsetofLevel1Level2() { return offsetof(Level1, level2); }
*/
import "C"
import "unsafe"

// C struct layout information - populated at init time
var (
	innerStructSize    uintptr
	innerStructXOffset uintptr
	innerStructYOffset uintptr

	middleStructSize        uintptr
	middleStructIdOffset    uintptr
	middleStructInnerOffset uintptr
	middleStructValueOffset uintptr

	outerStructSize            uintptr
	outerStructOuterIDOffset   uintptr
	outerStructMiddleOffset    uintptr
	outerStructTimestampOffset uintptr
)

func init() {
	innerStructSize = uintptr(unsafe.Sizeof(C.InnerStruct{}))
	innerStructXOffset = uintptr(C.innerStructXOffset())
	innerStructYOffset = uintptr(C.innerStructYOffset())

	middleStructSize = uintptr(unsafe.Sizeof(C.MiddleStruct{}))
	middleStructIdOffset = uintptr(C.middleStructIdOffset())
	middleStructInnerOffset = uintptr(C.middleStructInnerOffset())
	middleStructValueOffset = uintptr(C.middleStructValueOffset())

	outerStructSize = uintptr(unsafe.Sizeof(C.OuterStruct{}))
	outerStructOuterIDOffset = uintptr(C.outerStructOuterIDOffset())
	outerStructMiddleOffset = uintptr(C.outerStructMiddleOffset())
	outerStructTimestampOffset = uintptr(C.outerStructTimestampOffset())
}

func CreateInnerStruct() unsafe.Pointer {
	return unsafe.Pointer(C.createInnerStruct())
}

func CreateMiddleStruct() unsafe.Pointer {
	return unsafe.Pointer(C.createMiddleStruct())
}

func CreateOuterStruct() unsafe.Pointer {
	return unsafe.Pointer(C.createOuterStruct())
}

func FreePtr(ptr unsafe.Pointer) {
	C.free(ptr)
}

func InnerStructSize() uintptr {
	return uintptr(unsafe.Sizeof(C.InnerStruct{}))
}

func MiddleStructSize() uintptr {
	return uintptr(unsafe.Sizeof(C.MiddleStruct{}))
}

func OuterStructSize() uintptr {
	return uintptr(unsafe.Sizeof(C.OuterStruct{}))
}

// Deep nesting helpers (5 levels)
func CreateDeepNested() unsafe.Pointer {
	return unsafe.Pointer(C.createDeepNested())
}

func FreeDeepNested(ptr unsafe.Pointer) {
	C.freeDeepNested((*C.Level1)(ptr))
}

func Level1Size() uintptr {
	return uintptr(C.sizeofLevel1())
}

func Level2Size() uintptr {
	return uintptr(C.sizeofLevel2())
}

func Level3Size() uintptr {
	return uintptr(C.sizeofLevel3())
}

func Level4Size() uintptr {
	return uintptr(C.sizeofLevel4())
}

func Level5Size() uintptr {
	return uintptr(C.sizeofLevel5())
}

func Level1Level2Offset() uintptr {
	return uintptr(C.offsetofLevel1Level2())
}

func Level2Level3Offset() uintptr {
	return uintptr(C.offsetofLevel2Level3())
}

func Level3Level4Offset() uintptr {
	return uintptr(C.offsetofLevel3Level4())
}

func Level4Level5Offset() uintptr {
	return uintptr(C.offsetofLevel4Level5())
}
