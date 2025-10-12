package cgocopy

import (
	"reflect"
	"testing"
)

// rawLargeSensorBlock matches the C layout while allowing direct copying.
type rawLargeSensorBlock struct {
	Status   uint8
	_        [7]byte
	Readings [32]GoSensorReading
	Checksum uint64
}

func BenchmarkDirectCopyLargeSensorBlock(b *testing.B) {
	cPtr := createLargeSensorBlock()
	defer freeLargeSensorBlock(cPtr)

	var raw rawLargeSensorBlock

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Direct(&raw, cPtr)
	}
}

func BenchmarkRegistryCopyLargeSensorBlockRegistry(b *testing.B) {
	registry := NewRegistry()

	readingMeta := sensorReadingMetadata()
	if err := registry.Register(reflect.TypeOf(GoSensorReading{}), readingMeta.Size, readingMeta.Fields); err != nil {
		b.Fatalf("register SensorReading: %v", err)
	}

	blockMeta := largeSensorBlockMetadata()
	if err := registry.Register(reflect.TypeOf(GoLargeSensorBlock{}), blockMeta.Size, blockMeta.Fields); err != nil {
		b.Fatalf("register LargeSensorBlock: %v", err)
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

func TestDirectAndRegistryCopyAgreement(t *testing.T) {
	cPtr := createLargeSensorBlock()
	defer freeLargeSensorBlock(cPtr)

	var raw rawLargeSensorBlock
	Direct(&raw, cPtr)

	registry := NewRegistry()
	readingMeta := sensorReadingMetadata()
	if err := registry.Register(reflect.TypeOf(GoSensorReading{}), readingMeta.Size, readingMeta.Fields); err != nil {
		t.Fatalf("register SensorReading: %v", err)
	}

	blockMeta := largeSensorBlockMetadata()
	if err := registry.Register(reflect.TypeOf(GoLargeSensorBlock{}), blockMeta.Size, blockMeta.Fields); err != nil {
		t.Fatalf("register LargeSensorBlock: %v", err)
	}

	var fromRegistry GoLargeSensorBlock
	if err := registry.Copy(&fromRegistry, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if raw.Status != fromRegistry.Status {
		t.Fatalf("status mismatch: raw %d registry %d", raw.Status, fromRegistry.Status)
	}

	for i := range raw.Readings {
		if raw.Readings[i].Temperature != fromRegistry.Readings[i].Temperature {
			t.Fatalf("reading %d temperature mismatch: raw %.2f registry %.2f", i, raw.Readings[i].Temperature, fromRegistry.Readings[i].Temperature)
		}
		if raw.Readings[i].Pressure != fromRegistry.Readings[i].Pressure {
			t.Fatalf("reading %d pressure mismatch: raw %.2f registry %.2f", i, raw.Readings[i].Pressure, fromRegistry.Readings[i].Pressure)
		}
	}

	if raw.Checksum != fromRegistry.Checksum {
		t.Fatalf("checksum mismatch: raw %d registry %d", raw.Checksum, fromRegistry.Checksum)
	}
}
