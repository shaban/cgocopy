package cgocopy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "native/cgocopy_metadata.h"

typedef struct {
    double latitude;
    double longitude;
} Coordinate;

typedef struct {
    uint64_t id;
    Coordinate location;
    char* name;
    uint8_t firmwareVersion[8];
    uint16_t sensorCount;
    double temperatureReadings[5];
} DeviceRecord;

DeviceRecord* createDeviceRecord() {
    DeviceRecord* record = (DeviceRecord*)malloc(sizeof(DeviceRecord));
    record->id = 9001;
    record->location.latitude = 37.7749;
    record->location.longitude = -122.4194;
    record->name = strdup("Gateway Alpha");
    const uint8_t firmware[8] = {1, 2, 0, 5, 9, 8, 1, 3};
    memcpy(record->firmwareVersion, firmware, sizeof(firmware));
    record->sensorCount = 5;
    for (int i = 0; i < 5; ++i) {
        record->temperatureReadings[i] = 65.0 + (double)i * 1.5;
    }
    return record;
}

void freeDeviceRecord(DeviceRecord* record) {
    if (!record) {
        return;
    }
    if (record->name) {
        free(record->name);
    }
    free(record);
}

CGOCOPY_STRUCT_BEGIN(Coordinate)
    CGOCOPY_FIELD_PRIMITIVE(Coordinate, latitude, double),
    CGOCOPY_FIELD_PRIMITIVE(Coordinate, longitude, double),
CGOCOPY_STRUCT_END(Coordinate)

CGOCOPY_STRUCT_BEGIN(DeviceRecord)
    CGOCOPY_FIELD_PRIMITIVE(DeviceRecord, id, uint64_t),
    CGOCOPY_FIELD_STRUCT(DeviceRecord, location, Coordinate),
    CGOCOPY_FIELD_STRING(DeviceRecord, name),
    CGOCOPY_FIELD_ARRAY(DeviceRecord, firmwareVersion, uint8_t, 8),
    CGOCOPY_FIELD_PRIMITIVE(DeviceRecord, sensorCount, uint16_t),
    CGOCOPY_FIELD_ARRAY(DeviceRecord, temperatureReadings, double, 5),
CGOCOPY_STRUCT_END(DeviceRecord)
*/
import "C"
import "unsafe"

func coordinateMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_Coordinate_info())
}

func deviceRecordMetadata() StructMetadata {
	return loadStructMetadata(C.cgocopy_get_DeviceRecord_info())
}

func createDeviceRecord() unsafe.Pointer {
	return unsafe.Pointer(C.createDeviceRecord())
}

func freeDeviceRecord(ptr unsafe.Pointer) {
	C.freeDeviceRecord((*C.DeviceRecord)(ptr))
}
