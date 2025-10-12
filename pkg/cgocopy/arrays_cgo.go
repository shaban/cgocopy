package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include "../../native/cgocopy_metadata.h"

typedef struct {
    float x;
    float y;
} SamplePoint;

typedef struct {
    uint32_t id;
    float readings[4];
} PrimitiveArrayDevice;

typedef struct {
    uint32_t id;
    SamplePoint points[3];
} NestedArrayDevice;

typedef struct {
    double temperature;
    double pressure;
} SensorReading;

typedef struct {
    uint8_t status;
    SensorReading readings[32];
    uint64_t checksum;
} LargeSensorBlock;

PrimitiveArrayDevice* createPrimitiveArrayDevice() {
    PrimitiveArrayDevice* dev = (PrimitiveArrayDevice*)malloc(sizeof(PrimitiveArrayDevice));
    dev->id = 7;
    dev->readings[0] = 1.1f;
    dev->readings[1] = 2.2f;
    dev->readings[2] = 3.3f;
    dev->readings[3] = 4.4f;
    return dev;
}

NestedArrayDevice* createNestedArrayDevice() {
    NestedArrayDevice* dev = (NestedArrayDevice*)malloc(sizeof(NestedArrayDevice));
    dev->id = 21;
    for (int i = 0; i < 3; ++i) {
        dev->points[i].x = (float)(i + 1);
        dev->points[i].y = (float)(i + 1) * 10.0f;
    }
    return dev;
}

void freePrimitiveArrayDevice(PrimitiveArrayDevice* dev) {
    free(dev);
}

void freeNestedArrayDevice(NestedArrayDevice* dev) {
    free(dev);
}

LargeSensorBlock* createLargeSensorBlock() {
    LargeSensorBlock* block = (LargeSensorBlock*)malloc(sizeof(LargeSensorBlock));
    block->status = 3;
    block->checksum = 0;
    for (int i = 0; i < 32; ++i) {
        block->readings[i].temperature = 20.0 + (double)i * 0.5;
        block->readings[i].pressure = 101.0 + (double)(i % 5);
        block->checksum += (uint64_t)(block->readings[i].temperature * 10) + (uint64_t)(block->readings[i].pressure * 10);
    }
    return block;
}

void freeLargeSensorBlock(LargeSensorBlock* block) {
    free(block);
}

CGOCOPY_STRUCT_BEGIN(SamplePoint)
    CGOCOPY_FIELD_PRIMITIVE(SamplePoint, x, float),
    CGOCOPY_FIELD_PRIMITIVE(SamplePoint, y, float),
CGOCOPY_STRUCT_END(SamplePoint)

CGOCOPY_STRUCT_BEGIN(PrimitiveArrayDevice)
    CGOCOPY_FIELD_PRIMITIVE(PrimitiveArrayDevice, id, uint32_t),
    CGOCOPY_FIELD_ARRAY(PrimitiveArrayDevice, readings, float, 4),
CGOCOPY_STRUCT_END(PrimitiveArrayDevice)

CGOCOPY_STRUCT_BEGIN(NestedArrayDevice)
    CGOCOPY_FIELD_PRIMITIVE(NestedArrayDevice, id, uint32_t),
    CGOCOPY_FIELD_ARRAY_STRUCT(NestedArrayDevice, points, SamplePoint, 3),
CGOCOPY_STRUCT_END(NestedArrayDevice)

CGOCOPY_STRUCT_BEGIN(SensorReading)
    CGOCOPY_FIELD_PRIMITIVE(SensorReading, temperature, double),
    CGOCOPY_FIELD_PRIMITIVE(SensorReading, pressure, double),
CGOCOPY_STRUCT_END(SensorReading)

CGOCOPY_STRUCT_BEGIN(LargeSensorBlock)
    CGOCOPY_FIELD_PRIMITIVE(LargeSensorBlock, status, uint8_t),
    CGOCOPY_FIELD_ARRAY_STRUCT(LargeSensorBlock, readings, SensorReading, 32),
    CGOCOPY_FIELD_PRIMITIVE(LargeSensorBlock, checksum, uint64_t),
CGOCOPY_STRUCT_END(LargeSensorBlock)
*/
import "C"
import "unsafe"

func samplePointMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_SamplePoint_info())
}

func primitiveArrayDeviceMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_PrimitiveArrayDevice_info())
}

func nestedArrayDeviceMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_NestedArrayDevice_info())
}

func sensorReadingMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_SensorReading_info())
}

func largeSensorBlockMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_LargeSensorBlock_info())
}

func createPrimitiveArrayDevice() unsafe.Pointer {
	return unsafe.Pointer(C.createPrimitiveArrayDevice())
}

func freePrimitiveArrayDevice(ptr unsafe.Pointer) {
	C.freePrimitiveArrayDevice((*C.PrimitiveArrayDevice)(ptr))
}

func createNestedArrayDevice() unsafe.Pointer {
	return unsafe.Pointer(C.createNestedArrayDevice())
}

func freeNestedArrayDevice(ptr unsafe.Pointer) {
	C.freeNestedArrayDevice((*C.NestedArrayDevice)(ptr))
}

func createLargeSensorBlock() unsafe.Pointer {
	return unsafe.Pointer(C.createLargeSensorBlock())
}

func freeLargeSensorBlock(ptr unsafe.Pointer) {
	C.freeLargeSensorBlock((*C.LargeSensorBlock)(ptr))
}
