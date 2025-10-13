package cgocopy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../../native/cgocopy_metadata.h"

typedef struct {
    uint32_t id;
    double value;
} V2BenchPrimitive;

typedef struct {
    uint32_t count;
    uint16_t readings[64];
} V2BenchSlice;

typedef struct {
    uint32_t id;
    char* name;
} V2BenchMember;

typedef struct {
    uint32_t count;
    V2BenchMember members[4];
} V2BenchTeam;

static inline V2BenchPrimitive* cgocopy_create_bench_primitive(uint32_t id, double value) {
    V2BenchPrimitive* prim = (V2BenchPrimitive*)malloc(sizeof(V2BenchPrimitive));
    if (!prim) {
        return NULL;
    }
    prim->id = id;
    prim->value = value;
    return prim;
}

static inline void cgocopy_free_bench_primitive(V2BenchPrimitive* prim) {
    free(prim);
}

static inline V2BenchSlice* cgocopy_create_bench_slice(const uint16_t* values, uint32_t count) {
    V2BenchSlice* slice = (V2BenchSlice*)malloc(sizeof(V2BenchSlice));
    if (!slice) {
        return NULL;
    }
    if (count > 64) {
        count = 64;
    }
    slice->count = count;
    if (values && count > 0) {
        memcpy(slice->readings, values, sizeof(uint16_t) * count);
        if (count < 64) {
            memset(slice->readings + count, 0, sizeof(uint16_t) * (64 - count));
        }
    } else {
        memset(slice->readings, 0, sizeof(uint16_t) * 64);
    }
    return slice;
}

static inline void cgocopy_free_bench_slice(V2BenchSlice* slice) {
    free(slice);
}

static inline V2BenchTeam* cgocopy_create_bench_team(const char** names, const uint32_t* ids, uint32_t count) {
    V2BenchTeam* team = (V2BenchTeam*)malloc(sizeof(V2BenchTeam));
    if (!team) {
        return NULL;
    }
    if (count > 4) {
        count = 4;
    }
    team->count = count;
    for (uint32_t i = 0; i < 4; ++i) {
        team->members[i].id = 0;
        team->members[i].name = NULL;
    }
    for (uint32_t i = 0; i < count; ++i) {
        team->members[i].id = ids ? ids[i] : (1000 + i);
        if (names && names[i]) {
            team->members[i].name = strdup(names[i]);
            if (!team->members[i].name) {
                // clean up and return NULL on allocation failure
                for (uint32_t j = 0; j <= i; ++j) {
                    free(team->members[j].name);
                }
                free(team);
                return NULL;
            }
        } else {
            team->members[i].name = strdup("member");
            if (!team->members[i].name) {
                for (uint32_t j = 0; j <= i; ++j) {
                    free(team->members[j].name);
                }
                free(team);
                return NULL;
            }
        }
    }
    return team;
}

static inline void cgocopy_free_bench_team(V2BenchTeam* team) {
    if (!team) {
        return;
    }
    for (uint32_t i = 0; i < team->count && i < 4; ++i) {
        if (team->members[i].name) {
            free(team->members[i].name);
        }
    }
    free(team);
}

static inline char* cgocopy_create_bench_team_json(const char** names, const uint32_t* ids, uint32_t count) {
    if (count > 4) {
        count = 4;
    }

    size_t total = 32; // base for count and brackets
    for (uint32_t i = 0; i < count; ++i) {
        const char* name = (names && names[i]) ? names[i] : "";
        total += 32 + strlen(name); // estimate for each member entry
    }
    total += 1; // null terminator

    char* json = (char*)malloc(total);
    if (!json) {
        return NULL;
    }

    char* ptr = json;
    ptr += sprintf(ptr, "{\"count\":%u,\"members\":[", count);
    for (uint32_t i = 0; i < count; ++i) {
        if (i > 0) {
            ptr += sprintf(ptr, ",");
        }
        const char* name = (names && names[i]) ? names[i] : "";
        uint32_t id = ids ? ids[i] : (1000 + i);
        ptr += sprintf(ptr, "{\"id\":%u,\"name\":\"%s\"}", id, name);
    }
    sprintf(ptr, "]}");
    return json;
}

static inline void cgocopy_free_json(char* json) {
    free(json);
}

CGOCOPY_STRUCT_BEGIN(V2BenchPrimitive)
    CGOCOPY_FIELD_PRIMITIVE(V2BenchPrimitive, id, uint32_t),
    CGOCOPY_FIELD_PRIMITIVE(V2BenchPrimitive, value, double)
CGOCOPY_STRUCT_END(V2BenchPrimitive)

CGOCOPY_STRUCT_BEGIN(V2BenchSlice)
    CGOCOPY_FIELD_PRIMITIVE(V2BenchSlice, count, uint32_t),
    CGOCOPY_FIELD_ARRAY(V2BenchSlice, readings, uint16_t, 64)
CGOCOPY_STRUCT_END(V2BenchSlice)

CGOCOPY_STRUCT_BEGIN(V2BenchMember)
    CGOCOPY_FIELD_PRIMITIVE(V2BenchMember, id, uint32_t),
    CGOCOPY_FIELD_STRING(V2BenchMember, name)
CGOCOPY_STRUCT_END(V2BenchMember)

CGOCOPY_STRUCT_BEGIN(V2BenchTeam)
    CGOCOPY_FIELD_PRIMITIVE(V2BenchTeam, count, uint32_t),
    CGOCOPY_FIELD_ARRAY_STRUCT(V2BenchTeam, members, V2BenchMember, 4)
CGOCOPY_STRUCT_END(V2BenchTeam)
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func newBenchPrimitivePointer(id uint32, value float64) (unsafe.Pointer, func()) {
	prim := C.cgocopy_create_bench_primitive(C.uint32_t(id), C.double(value))
	if prim == nil {
		return nil, func() {}
	}
	cleanup := func() {
		C.cgocopy_free_bench_primitive(prim)
	}
	return unsafe.Pointer(prim), cleanup
}

func newBenchSlicePointer(values []uint16) (unsafe.Pointer, func()) {
	var ptr *C.uint16_t
	if len(values) > 0 {
		ptr = (*C.uint16_t)(unsafe.Pointer(&values[0]))
	}
	slice := C.cgocopy_create_bench_slice(ptr, C.uint32_t(len(values)))
	if slice == nil {
		return nil, func() {}
	}
	runtime.KeepAlive(values)
	cleanup := func() {
		C.cgocopy_free_bench_slice(slice)
	}
	return unsafe.Pointer(slice), cleanup
}

func newBenchTeamPointer(ids []uint32, names []string) (unsafe.Pointer, func()) {
	count := len(names)
	if len(ids) < count {
		count = len(ids)
	}
	if count > 4 {
		count = 4
	}

	var cNames **C.char
	if count > 0 {
		nameArray := make([]*C.char, count)
		for i := 0; i < count; i++ {
			nameArray[i] = C.CString(names[i])
		}
		cNames = (**C.char)(unsafe.Pointer(&nameArray[0]))

		cIDs := make([]C.uint32_t, count)
		for i := 0; i < count; i++ {
			cIDs[i] = C.uint32_t(ids[i])
		}

		team := C.cgocopy_create_bench_team(cNames, &cIDs[0], C.uint32_t(count))
		for i := 0; i < count; i++ {
			C.free(unsafe.Pointer(nameArray[i]))
		}
		if team == nil {
			return nil, func() {}
		}
		cleanup := func() {
			C.cgocopy_free_bench_team(team)
		}
		return unsafe.Pointer(team), cleanup
	}

	team := C.cgocopy_create_bench_team(nil, nil, 0)
	if team == nil {
		return nil, func() {}
	}
	cleanup := func() {
		C.cgocopy_free_bench_team(team)
	}
	return unsafe.Pointer(team), cleanup
}

func newBenchTeamJSON(ids []uint32, names []string) (string, func()) {
	count := len(names)
	if len(ids) < count {
		count = len(ids)
	}
	if count > 4 {
		count = 4
	}

	var jsonPtr *C.char
	if count > 0 {
		nameArray := make([]*C.char, count)
		for i := 0; i < count; i++ {
			nameArray[i] = C.CString(names[i])
		}
		cNames := (**C.char)(unsafe.Pointer(&nameArray[0]))

		cIDs := make([]C.uint32_t, count)
		for i := 0; i < count; i++ {
			cIDs[i] = C.uint32_t(ids[i])
		}

		jsonPtr = C.cgocopy_create_bench_team_json(cNames, &cIDs[0], C.uint32_t(count))
		for i := 0; i < count; i++ {
			C.free(unsafe.Pointer(nameArray[i]))
		}
	} else {
		jsonPtr = C.cgocopy_create_bench_team_json(nil, nil, 0)
	}

	if jsonPtr == nil {
		return "", func() {}
	}

	jsonStr := C.GoString(jsonPtr)
	cleanup := func() {
		C.cgocopy_free_json(jsonPtr)
	}
	return jsonStr, cleanup
}
