package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "../../native/cgocopy_metadata.h"

typedef struct {
    uint16_t values[4];
} V2ArrayHolder;

typedef struct {
    uint32_t code;
    char* label;
} V2NestedInner;

typedef struct {
    V2NestedInner inner;
    V2ArrayHolder data;
} V2NestedOuter;

static inline V2NestedOuter* cgocopy_create_nested(uint32_t code, const char* label, const uint16_t* values) {
    V2NestedOuter* nested = (V2NestedOuter*)malloc(sizeof(V2NestedOuter));
    if (!nested) {
        return NULL;
    }
    nested->inner.code = code;
    if (label) {
        nested->inner.label = strdup(label);
        if (!nested->inner.label) {
            free(nested);
            return NULL;
        }
    } else {
        nested->inner.label = NULL;
    }
    if (values) {
        memcpy(nested->data.values, values, sizeof(uint16_t) * 4);
    } else {
        memset(nested->data.values, 0, sizeof(uint16_t) * 4);
    }
    return nested;
}

static inline void cgocopy_free_nested(V2NestedOuter* nested) {
    if (!nested) {
        return;
    }
    if (nested->inner.label) {
        free(nested->inner.label);
    }
    free(nested);
}

CGOCOPY_STRUCT_BEGIN(V2ArrayHolder)
    CGOCOPY_FIELD_ARRAY(V2ArrayHolder, values, uint16_t, 4)
CGOCOPY_STRUCT_END(V2ArrayHolder)

CGOCOPY_STRUCT_BEGIN(V2NestedInner)
    CGOCOPY_FIELD_PRIMITIVE(V2NestedInner, code, uint32_t),
    CGOCOPY_FIELD_STRING(V2NestedInner, label)
CGOCOPY_STRUCT_END(V2NestedInner)

CGOCOPY_STRUCT_BEGIN(V2NestedOuter)
    CGOCOPY_FIELD_STRUCT(V2NestedOuter, inner, V2NestedInner),
    CGOCOPY_FIELD_STRUCT(V2NestedOuter, data, V2ArrayHolder)
CGOCOPY_STRUCT_END(V2NestedOuter)
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func newNestedSamplePointer(code uint32, label string, values [4]uint16) (unsafe.Pointer, func()) {
	var cLabel *C.char
	if label != "" {
		cLabel = C.CString(label)
		if cLabel == nil {
			return nil, func() {}
		}
		defer C.free(unsafe.Pointer(cLabel))
	}

	var cValues *C.uint16_t
	cValues = (*C.uint16_t)(unsafe.Pointer(&values[0]))

	nested := C.cgocopy_create_nested(C.uint32_t(code), cLabel, cValues)
	if nested == nil {
		return nil, func() {}
	}

	runtime.KeepAlive(values)

	cleanup := func() {
		C.cgocopy_free_nested(nested)
	}

	return unsafe.Pointer(nested), cleanup
}
