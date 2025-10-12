package cgocopy

import (
	"reflect"
	"testing"
)

type GoPrimitiveArrayDevice struct {
	ID       uint32
	Readings [4]float32
}

type GoSamplePoint struct {
	X float32
	Y float32
}

type GoNestedArrayDevice struct {
	ID     uint32
	Points [3]GoSamplePoint
}

type GoSensorReading struct {
	Temperature float64
	Pressure    float64
}

type GoLargeSensorBlock struct {
	Status   uint8
	Readings [32]GoSensorReading
	Checksum uint64
}

func TestPrimitiveArrayMetadataCopy(t *testing.T) {
	registry := NewRegistry()

	meta := primitiveArrayDeviceMetadata()
	if len(meta.Fields) != 2 {
		t.Fatalf("expected 2 fields from metadata, got %d", len(meta.Fields))
	}
	if meta.Fields[1].Kind != FieldArray {
		t.Fatalf("expected second field to be array, got kind %v", meta.Fields[1].Kind)
	}
	if meta.Fields[1].ElemCount != 4 {
		t.Fatalf("expected array length 4, got %d", meta.Fields[1].ElemCount)
	}

	if err := registry.Register(reflect.TypeOf(GoPrimitiveArrayDevice{}), meta.Size, meta.Fields); err != nil {
		t.Fatalf("failed to register primitive array struct: %v", err)
	}

	cPtr := createPrimitiveArrayDevice()
	defer freePrimitiveArrayDevice(cPtr)

	var device GoPrimitiveArrayDevice
	if err := registry.Copy(&device, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	expected := [4]float32{1.1, 2.2, 3.3, 4.4}
	if device.ID != 7 {
		t.Errorf("expected ID 7, got %d", device.ID)
	}
	if device.Readings != expected {
		t.Errorf("expected readings %v, got %v", expected, device.Readings)
	}
}

func TestNestedArrayMetadataCopy(t *testing.T) {
	registry := NewRegistry()

	pointMeta := samplePointMetadata()
	if err := registry.Register(reflect.TypeOf(GoSamplePoint{}), pointMeta.Size, pointMeta.Fields); err != nil {
		t.Fatalf("failed to register SamplePoint: %v", err)
	}

	deviceMeta := nestedArrayDeviceMetadata()
	if len(deviceMeta.Fields) != 2 {
		t.Fatalf("expected 2 fields from metadata, got %d", len(deviceMeta.Fields))
	}
	if deviceMeta.Fields[1].Kind != FieldArray {
		t.Fatalf("expected array field, got kind %v", deviceMeta.Fields[1].Kind)
	}
	if err := registry.Register(reflect.TypeOf(GoNestedArrayDevice{}), deviceMeta.Size, deviceMeta.Fields); err != nil {
		t.Fatalf("failed to register nested array struct: %v", err)
	}

	cPtr := createNestedArrayDevice()
	defer freeNestedArrayDevice(cPtr)

	var device GoNestedArrayDevice
	if err := registry.Copy(&device, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if device.ID != 21 {
		t.Errorf("expected ID 21, got %d", device.ID)
	}

	for i := 0; i < len(device.Points); i++ {
		expectedX := float32(i + 1)
		expectedY := float32(i+1) * 10
		if device.Points[i].X != expectedX {
			t.Errorf("point %d expected X %.1f, got %.1f", i, expectedX, device.Points[i].X)
		}
		if device.Points[i].Y != expectedY {
			t.Errorf("point %d expected Y %.1f, got %.1f", i, expectedY, device.Points[i].Y)
		}
	}
}

func TestLargeSensorBlockCopy(t *testing.T) {
	registry := NewRegistry()

	readingMeta := sensorReadingMetadata()
	if err := registry.Register(reflect.TypeOf(GoSensorReading{}), readingMeta.Size, readingMeta.Fields); err != nil {
		t.Fatalf("failed to register SensorReading: %v", err)
	}

	blockMeta := largeSensorBlockMetadata()
	if blockMeta.Fields[0].Kind != FieldPrimitive {
		t.Fatalf("expected status field to be primitive, got %v", blockMeta.Fields[0].Kind)
	}
	if blockMeta.Fields[1].Kind != FieldArray || blockMeta.Fields[1].ElemCount != 32 {
		t.Fatalf("expected readings to be array of 32, got %+v", blockMeta.Fields[1])
	}
	if blockMeta.Size == 0 {
		t.Fatalf("expected non-zero struct size")
	}

	if err := registry.Register(reflect.TypeOf(GoLargeSensorBlock{}), blockMeta.Size, blockMeta.Fields); err != nil {
		t.Fatalf("failed to register LargeSensorBlock: %v", err)
	}

	cPtr := createLargeSensorBlock()
	defer freeLargeSensorBlock(cPtr)

	var block GoLargeSensorBlock
	if err := registry.Copy(&block, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if block.Status != 3 {
		t.Fatalf("expected status 3, got %d", block.Status)
	}

	var expectedChecksum uint64
	for i := range block.Readings {
		expectedTemp := 20.0 + float64(i)*0.5
		expectedPressure := 101.0 + float64(i%5)
		if block.Readings[i].Temperature != expectedTemp {
			t.Fatalf("reading %d temperature: got %.2f want %.2f", i, block.Readings[i].Temperature, expectedTemp)
		}
		if block.Readings[i].Pressure != expectedPressure {
			t.Fatalf("reading %d pressure: got %.2f want %.2f", i, block.Readings[i].Pressure, expectedPressure)
		}
		expectedChecksum += uint64(expectedTemp*10) + uint64(expectedPressure*10)
	}

	if block.Checksum != expectedChecksum {
		t.Fatalf("checksum mismatch: got %d want %d", block.Checksum, expectedChecksum)
	}
}
