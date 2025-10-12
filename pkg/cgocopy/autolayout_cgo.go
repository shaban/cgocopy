package cgocopy

import (
	"unsafe"
)

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "../../native/cgocopy_metadata.h"

// Test struct matching the AutoLayout example
typedef struct {
    uint32_t id;
    char* name;
    float value;
} AutoDevice;

// Create a test device
AutoDevice* createAutoDevice() {
    AutoDevice* dev = (AutoDevice*)malloc(sizeof(AutoDevice));
    dev->id = 42;
    dev->name = strdup("Test Device");
    dev->value = 3.14f;
    return dev;
}

void freeAutoDevice(AutoDevice* dev) {
    if (dev) {
        if (dev->name) free(dev->name);
        free(dev);
    }
}

// More complex struct for testing
typedef struct {
    uint8_t flag;
    uint32_t id;
    char* name;
    double value;
    uint16_t port;
} ComplexAutoDevice;

ComplexAutoDevice* createComplexDevice() {
    ComplexAutoDevice* dev = (ComplexAutoDevice*)malloc(sizeof(ComplexAutoDevice));
    dev->flag = 1;
    dev->id = 99;
    dev->name = strdup("Complex Device");
    dev->value = 2.718;
    dev->port = 8080;
    return dev;
}

void freeComplexDevice(ComplexAutoDevice* dev) {
    if (dev) {
        if (dev->name) free(dev->name);
        free(dev);
    }
}

CGOCOPY_STRUCT_BEGIN(AutoDevice)
    CGOCOPY_FIELD_PRIMITIVE(AutoDevice, id, uint32_t),
    CGOCOPY_FIELD_STRING(AutoDevice, name),
    CGOCOPY_FIELD_PRIMITIVE(AutoDevice, value, float),
CGOCOPY_STRUCT_END(AutoDevice)

CGOCOPY_STRUCT_BEGIN(ComplexAutoDevice)
    CGOCOPY_FIELD_PRIMITIVE(ComplexAutoDevice, flag, uint8_t),
    CGOCOPY_FIELD_PRIMITIVE(ComplexAutoDevice, id, uint32_t),
    CGOCOPY_FIELD_STRING(ComplexAutoDevice, name),
    CGOCOPY_FIELD_PRIMITIVE(ComplexAutoDevice, value, double),
    CGOCOPY_FIELD_PRIMITIVE(ComplexAutoDevice, port, uint16_t),
CGOCOPY_STRUCT_END(ComplexAutoDevice)
*/
import "C"

// Simple string converter for testing
type TestConverter struct{}

func (c TestConverter) CStringToGo(ptr unsafe.Pointer) string {
	if ptr == nil {
		return ""
	}
	var buf []byte
	for i := uintptr(0); ; i++ {
		b := *(*byte)(unsafe.Add(ptr, i))
		if b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf)
}

func autoDeviceMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_AutoDevice_info())
}

func complexAutoDeviceMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_ComplexAutoDevice_info())
}

func GetAutoDeviceSize() uintptr {
	return autoDeviceMetadata().Size
}

func GetComplexDeviceSize() uintptr {
	return complexAutoDeviceMetadata().Size
}

// Helper to create C test devices
func CreateAutoDevice() unsafe.Pointer     { return unsafe.Pointer(C.createAutoDevice()) }
func FreeAutoDevice(dev unsafe.Pointer)    { C.freeAutoDevice((*C.AutoDevice)(dev)) }
func CreateComplexDevice() unsafe.Pointer  { return unsafe.Pointer(C.createComplexDevice()) }
func FreeComplexDevice(dev unsafe.Pointer) { C.freeComplexDevice((*C.ComplexAutoDevice)(dev)) }
