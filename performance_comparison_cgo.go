package cgocopy

/*
#include <stdlib.h>
#include <string.h>
#include <stddef.h>

typedef struct {
    uint32_t id;
    char* name;
    char* manufacturer;
    char* category;
    float version;
} PluginInfo;

PluginInfo* createPluginInfo() {
    PluginInfo* p = (PluginInfo*)malloc(sizeof(PluginInfo));
    p->id = 42;
    p->name = strdup("SuperDelay");
    p->manufacturer = strdup("Waves");
    p->category = strdup("Delay");
    p->version = 1.5f;
    return p;
}

void freePluginInfo(PluginInfo* p) {
    free(p->name);
    free(p->manufacturer);
    free(p->category);
    free(p);
}

// Offsets
size_t pluginInfoIdOffset() { return offsetof(PluginInfo, id); }
size_t pluginInfoNameOffset() { return offsetof(PluginInfo, name); }
size_t pluginInfoMfgOffset() { return offsetof(PluginInfo, manufacturer); }
size_t pluginInfoCategoryOffset() { return offsetof(PluginInfo, category); }
size_t pluginInfoVersionOffset() { return offsetof(PluginInfo, version); }
*/
import "C"
import "unsafe"

func createPluginInfo() unsafe.Pointer {
	return unsafe.Pointer(C.createPluginInfo())
}

func freePluginInfo(ptr unsafe.Pointer) {
	C.freePluginInfo((*C.PluginInfo)(ptr))
}

func pluginInfoSize() uintptr {
	return uintptr(unsafe.Sizeof(C.PluginInfo{}))
}

func pluginInfoOffsets() (id, name, mfg, category, version uintptr) {
	return uintptr(C.pluginInfoIdOffset()),
		uintptr(C.pluginInfoNameOffset()),
		uintptr(C.pluginInfoMfgOffset()),
		uintptr(C.pluginInfoCategoryOffset()),
		uintptr(C.pluginInfoVersionOffset())
}

// String converter for Registry
type PluginStringConverter struct{}

func (c PluginStringConverter) CStringToGo(ptr unsafe.Pointer) string {
	if ptr == nil {
		return ""
	}
	return C.GoString((*C.char)(ptr))
}
