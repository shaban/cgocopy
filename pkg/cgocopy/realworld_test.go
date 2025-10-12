package cgocopy

import (
	"reflect"
	"testing"
)

type GoCoordinate struct {
	Latitude  float64
	Longitude float64
}

type GoDeviceRecord struct {
	ID                  uint64
	Location            GoCoordinate
	Name                string
	FirmwareVersion     [8]byte
	SensorCount         uint16
	TemperatureReadings [5]float64
}

func TestDeviceRecordCopy(t *testing.T) {
	registry := NewRegistry()

	coordMeta := coordinateMetadata()
	if err := registry.Register(reflect.TypeOf(GoCoordinate{}), coordMeta.Size, coordMeta.Fields); err != nil {
		t.Fatalf("failed to register Coordinate: %v", err)
	}

	recordMeta := deviceRecordMetadata()
	if len(recordMeta.Fields) != 6 {
		t.Fatalf("expected 6 fields, got %d", len(recordMeta.Fields))
	}
	if recordMeta.Fields[1].Kind != FieldStruct {
		t.Fatalf("expected location field to be struct, got %v", recordMeta.Fields[1].Kind)
	}
	if recordMeta.Fields[2].Kind != FieldString {
		t.Fatalf("expected name field to be string, got %v", recordMeta.Fields[2].Kind)
	}
	if recordMeta.Fields[3].Kind != FieldArray || recordMeta.Fields[3].ElemCount != 8 {
		t.Fatalf("expected firmware array length 8, got %+v", recordMeta.Fields[3])
	}
	if recordMeta.Fields[5].Kind != FieldArray || recordMeta.Fields[5].ElemCount != 5 {
		t.Fatalf("expected temperature readings to have 5 entries, got %+v", recordMeta.Fields[5])
	}

	if err := registry.Register(reflect.TypeOf(GoDeviceRecord{}), recordMeta.Size, recordMeta.Fields, TestConverter{}); err != nil {
		t.Fatalf("failed to register DeviceRecord: %v", err)
	}

	cPtr := createDeviceRecord()
	defer freeDeviceRecord(cPtr)

	var out GoDeviceRecord
	if err := registry.Copy(&out, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if out.ID != 9001 {
		t.Fatalf("expected ID 9001, got %d", out.ID)
	}
	if out.Location.Latitude != 37.7749 || out.Location.Longitude != -122.4194 {
		t.Fatalf("unexpected location: %+v", out.Location)
	}
	if out.Name != "Gateway Alpha" {
		t.Fatalf("unexpected name %q", out.Name)
	}

	expectedFirmware := [8]byte{1, 2, 0, 5, 9, 8, 1, 3}
	if out.FirmwareVersion != expectedFirmware {
		t.Fatalf("unexpected firmware bytes: %v", out.FirmwareVersion)
	}
	if out.SensorCount != 5 {
		t.Fatalf("unexpected sensor count: %d", out.SensorCount)
	}

	for i, reading := range out.TemperatureReadings {
		expected := 65.0 + float64(i)*1.5
		if reading != expected {
			t.Fatalf("reading %d got %.2f want %.2f", i, reading, expected)
		}
	}
}
