package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include "../../native/cgocopy_metadata.h"

typedef struct {
    int32_t value;
    double ratio;
} MacroInner;

CGOCOPY_STRUCT_BEGIN(MacroInner)
    CGOCOPY_FIELD_PRIMITIVE(MacroInner, value, int32_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroInner, ratio, double),
CGOCOPY_STRUCT_END(MacroInner)

typedef struct {
    int8_t i8;
    uint8_t u8;
    int16_t i16;
    uint16_t u16;
    int32_t i32;
    uint32_t u32;
    int64_t i64;
    uint64_t u64;
    float f32;
    double f64;
    bool flag;
    void* data;
    char* name;
    char label[12];
} MacroPrimitives;

CGOCOPY_STRUCT_BEGIN(MacroPrimitives)
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, i8, int8_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, u8, uint8_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, i16, int16_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, u16, uint16_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, i32, int32_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, u32, uint32_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, i64, int64_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, u64, uint64_t),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, f32, float),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, f64, double),
    CGOCOPY_FIELD_PRIMITIVE(MacroPrimitives, flag, bool),
    CGOCOPY_FIELD_POINTER(MacroPrimitives, data, void*),
    CGOCOPY_FIELD_STRING(MacroPrimitives, name),
    CGOCOPY_FIELD_ARRAY(MacroPrimitives, label, char, 12),
CGOCOPY_STRUCT_END(MacroPrimitives)

typedef struct {
    MacroPrimitives base;
    MacroInner points[2];
} MacroComposite;

CGOCOPY_STRUCT_BEGIN(MacroComposite)
    CGOCOPY_FIELD_STRUCT(MacroComposite, base, MacroPrimitives),
    CGOCOPY_FIELD_ARRAY_STRUCT(MacroComposite, points, MacroInner, 2),
CGOCOPY_STRUCT_END(MacroComposite)

typedef struct {
    char* normal;
    char* empty;
    char* null_value;
    char* unicode;
} MacroStrings;

CGOCOPY_STRUCT_BEGIN(MacroStrings)
    CGOCOPY_FIELD_STRING(MacroStrings, normal),
    CGOCOPY_FIELD_STRING(MacroStrings, empty),
    CGOCOPY_FIELD_STRING(MacroStrings, null_value),
    CGOCOPY_FIELD_STRING(MacroStrings, unicode),
CGOCOPY_STRUCT_END(MacroStrings)

static inline void fill_label(char label[12], const char* text) {
    size_t len = strlen(text);
    if (len > 11) {
        len = 11;
    }
    memset(label, 0, 12);
    memcpy(label, text, len);
}

MacroPrimitives* createMacroPrimitives() {
    MacroPrimitives* mp = (MacroPrimitives*)malloc(sizeof(MacroPrimitives));
    mp->i8 = -8;
    mp->u8 = 8;
    mp->i16 = -160;
    mp->u16 = 160;
    mp->i32 = -3200;
    mp->u32 = 3200;
    mp->i64 = -64000;
    mp->u64 = 64000;
    mp->f32 = 3.5f;
    mp->f64 = 6.75;
    mp->flag = true;
    mp->data = (void*)mp;
    mp->name = strdup("Macro Primitives");
    fill_label(mp->label, "macro");
    return mp;
}

void freeMacroPrimitives(MacroPrimitives* mp) {
    if (!mp) {
        return;
    }
    if (mp->name) {
        free(mp->name);
    }
    free(mp);
}

MacroComposite* createMacroComposite() {
    MacroComposite* mc = (MacroComposite*)malloc(sizeof(MacroComposite));
    MacroPrimitives* mp = createMacroPrimitives();
    mc->base = *mp;
    // data pointer should reference inner MacroComposite for testing
    mc->base.data = (void*)mc;
    mp->name = NULL;
    for (int i = 0; i < 2; ++i) {
        mc->points[i].value = (i + 1) * 10;
        mc->points[i].ratio = 1.5 + (double)i;
    }
    freeMacroPrimitives(mp);
    return mc;
}

void freeMacroComposite(MacroComposite* mc) {
    if (!mc) {
        return;
    }
    if (mc->base.name) {
        free(mc->base.name);
    }
    free(mc);
}

MacroStrings* createMacroStrings() {
    MacroStrings* ms = (MacroStrings*)malloc(sizeof(MacroStrings));
    ms->normal = strdup("Macro String");
    ms->empty = strdup("");
    ms->null_value = NULL;
    ms->unicode = strdup("Macro UTF-8 🚀");
    return ms;
}

void freeMacroStrings(MacroStrings* ms) {
    if (!ms) {
        return;
    }
    if (ms->normal) {
        free(ms->normal);
    }
    if (ms->empty) {
        free(ms->empty);
    }
    if (ms->unicode) {
        free(ms->unicode);
    }
    free(ms);
}

*/
import "C"
import "unsafe"

func macroInnerMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_MacroInner_info())
}

func macroPrimitivesMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_MacroPrimitives_info())
}

func macroCompositeMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_MacroComposite_info())
}

func macroStringsMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_MacroStrings_info())
}

func createMacroPrimitives() unsafe.Pointer {
	return unsafe.Pointer(C.createMacroPrimitives())
}

func freeMacroPrimitives(ptr unsafe.Pointer) {
	C.freeMacroPrimitives((*C.MacroPrimitives)(ptr))
}

func createMacroComposite() unsafe.Pointer {
	return unsafe.Pointer(C.createMacroComposite())
}

func freeMacroComposite(ptr unsafe.Pointer) {
	C.freeMacroComposite((*C.MacroComposite)(ptr))
}

func createMacroStrings() unsafe.Pointer {
	return unsafe.Pointer(C.createMacroStrings())
}

func freeMacroStrings(ptr unsafe.Pointer) {
	C.freeMacroStrings((*C.MacroStrings)(ptr))
}
