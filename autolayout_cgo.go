package cgocopy

import (
	"unsafe"
)

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

// Test struct matching the AutoLayout example
typedef struct {
    uint32_t id;
    char* name;
    float value;
} AutoDevice;

// Helper functions for manual offset verification
size_t autoDeviceIdOffset() { return offsetof(AutoDevice, id); }
size_t autoDeviceNameOffset() { return offsetof(AutoDevice, name); }
size_t autoDeviceValueOffset() { return offsetof(AutoDevice, value); }
size_t autoDeviceSize() { return sizeof(AutoDevice); }

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

size_t complexDeviceSize() { return sizeof(ComplexAutoDevice); }

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
*/
import "C"

// Simple string converter for testing
type TestConverter struct{}

func (c TestConverter) CStringToGo(ptr unsafe.Pointer) string {
	if ptr == nil {
		return ""
	}
	return C.GoString((*C.char)(ptr))
}

// Helper functions to expose C offsets for testing
func GetAutoDeviceIdOffset() uintptr    { return uintptr(C.autoDeviceIdOffset()) }
func GetAutoDeviceNameOffset() uintptr  { return uintptr(C.autoDeviceNameOffset()) }
func GetAutoDeviceValueOffset() uintptr { return uintptr(C.autoDeviceValueOffset()) }
func GetAutoDeviceSize() uintptr        { return uintptr(C.autoDeviceSize()) }
func GetComplexDeviceSize() uintptr     { return uintptr(C.complexDeviceSize()) }

// Helper to create C test devices
func CreateAutoDevice() unsafe.Pointer     { return unsafe.Pointer(C.createAutoDevice()) }
func FreeAutoDevice(dev unsafe.Pointer)    { C.freeAutoDevice((*C.AutoDevice)(dev)) }
func CreateComplexDevice() unsafe.Pointer  { return unsafe.Pointer(C.createComplexDevice()) }
func FreeComplexDevice(dev unsafe.Pointer) { C.freeComplexDevice((*C.ComplexAutoDevice)(dev)) }
