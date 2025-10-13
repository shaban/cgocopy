package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "../../native/cgocopy_metadata.h"

typedef struct {
    uint32_t id;
    char* name;
} V2Sample;

static inline V2Sample* cgocopy_create_sample(uint32_t id, const char* name) {
    V2Sample* sample = (V2Sample*)malloc(sizeof(V2Sample));
    if (!sample) {
        return NULL;
    }
    sample->id = id;
    if (name) {
        sample->name = strdup(name);
    } else {
        sample->name = NULL;
    }
    return sample;
}

static inline void cgocopy_free_sample(V2Sample* sample) {
    if (!sample) {
        return;
    }
    if (sample->name) {
        free(sample->name);
    }
    free(sample);
}

CGOCOPY_STRUCT_BEGIN(V2Sample)
    CGOCOPY_FIELD_PRIMITIVE(V2Sample, id, uint32_t),
    CGOCOPY_FIELD_STRING(V2Sample, name)
CGOCOPY_STRUCT_END(V2Sample)
*/
import "C"

import "unsafe"

func newSamplePointer(id uint32, name string) (unsafe.Pointer, func()) {
	cName := C.CString(name)
	if cName == nil {
		return nil, func() {}
	}
	sample := C.cgocopy_create_sample(C.uint32_t(id), cName)
	C.free(unsafe.Pointer(cName))
	if sample == nil {
		return nil, func() {}
	}
	cleanup := func() {
		C.cgocopy_free_sample(sample)
	}
	return unsafe.Pointer(sample), cleanup
}
