package cgocopy

import (
	"fmt"
	"unsafe"
)

var exampleCStringArena [][]byte

func makeExampleCString(s string) (unsafe.Pointer, func()) {
	buf := append([]byte(s), 0)
	exampleCStringArena = append(exampleCStringArena, buf)
	idx := len(exampleCStringArena) - 1
	return unsafe.Pointer(&exampleCStringArena[idx][0]), func() {
		exampleCStringArena[idx] = nil
	}
}

// Example demonstrating Direct copy for primitive-only structs.
func Example_direct() {
	type CSensor struct {
		ID    uint32
		Value float32
	}

	type Sensor struct {
		ID    uint32
		Value float32
	}

	cSensor := CSensor{ID: 7, Value: 3.14}

	var sensor Sensor
	Direct(&sensor, unsafe.Pointer(&cSensor))

	fmt.Printf("ID=%d Value=%.2f\n", sensor.ID, sensor.Value)
	// Output: ID=7 Value=3.14
}

// Example demonstrating Registry.Copy with automatic string conversion via UTF8Converter.
func Example_registryWithStrings() {
	type CDevice struct {
		ID       uint32
		Name     unsafe.Pointer
		Channels uint32
	}

	type Device struct {
		ID       uint32
		Name     string
		Channels uint32
	}

	cName, cleanup := makeExampleCString("Monitor Output")
	defer cleanup()

	cDevice := CDevice{ID: 42, Name: cName, Channels: 8}

	layout := []FieldInfo{
		{Offset: unsafe.Offsetof(CDevice{}.ID), TypeName: "uint32_t"},
		{Offset: unsafe.Offsetof(CDevice{}.Name), TypeName: "char*"},
		{Offset: unsafe.Offsetof(CDevice{}.Channels), TypeName: "uint32_t"},
	}

	registry := NewRegistry()
	registry.MustRegister(Device{}, unsafe.Sizeof(CDevice{}), layout, DefaultCStringConverter)

	var device Device
	if err := registry.Copy(&device, unsafe.Pointer(&cDevice)); err != nil {
		panic(err)
	}

	fmt.Printf("ID=%d Name=%s Channels=%d\n", device.ID, device.Name, device.Channels)
	// Output: ID=42 Name=Monitor Output Channels=8
}

// Example showcasing nested array copying with the registry metadata helpers.
func Example_registryNestedArrays() {
	registry := NewRegistry()

	readingMeta := sensorReadingMetadata()
	registry.MustRegister(GoSensorReading{}, readingMeta.Size, readingMeta.Fields)

	blockMeta := largeSensorBlockMetadata()
	registry.MustRegister(GoLargeSensorBlock{}, blockMeta.Size, blockMeta.Fields)

	cPtr := createLargeSensorBlock()
	defer freeLargeSensorBlock(cPtr)

	var block GoLargeSensorBlock
	if err := registry.Copy(&block, cPtr); err != nil {
		panic(err)
	}

	first := block.Readings[0]
	last := block.Readings[len(block.Readings)-1]
	fmt.Printf("status=%d first=%.1f/%.1f last=%.1f/%.1f count=%d\n",
		block.Status,
		first.Temperature,
		first.Pressure,
		last.Temperature,
		last.Pressure,
		len(block.Readings))
	// Output: status=3 first=20.0/101.0 last=35.5/102.0 count=32
}
