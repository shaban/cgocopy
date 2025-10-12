package cgocopy

import (
	"reflect"
	"testing"
)

type GoSensorReading struct {
	Temperature float64
	Pressure    float64
}

type GoLargeSensorBlock struct {
	Status   uint8
	Readings [32]GoSensorReading
	Checksum uint64
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

func BenchmarkRegistryCopyLargeSensorBlock(b *testing.B) {
	registry := NewRegistry()

	readingMeta := sensorReadingMetadata()
	if err := registry.Register(reflect.TypeOf(GoSensorReading{}), readingMeta.Size, readingMeta.Fields); err != nil {
		b.Fatalf("failed to register SensorReading: %v", err)
	}

	blockMeta := largeSensorBlockMetadata()
	if err := registry.Register(reflect.TypeOf(GoLargeSensorBlock{}), blockMeta.Size, blockMeta.Fields); err != nil {
		b.Fatalf("failed to register LargeSensorBlock: %v", err)
	}

	cPtr := createLargeSensorBlock()
	defer freeLargeSensorBlock(cPtr)

	var block GoLargeSensorBlock
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&block, cPtr); err != nil {
			b.Fatalf("copy failed: %v", err)
		}
	}
}
