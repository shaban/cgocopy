package cgocopy

import (
	"reflect"
	"testing"
)

// TestAutoLayout verifies that AutoLayout calculates offsets correctly
func TestAutoLayout(t *testing.T) {
	// Generate layout automatically
	layout := AutoLayout("uint32_t", "char*", "float")

	// Verify against actual C offsets
	expectedOffsets := []uintptr{
		GetAutoDeviceIdOffset(),
		GetAutoDeviceNameOffset(),
		GetAutoDeviceValueOffset(),
	}

	if len(layout) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(layout))
	}

	for i, field := range layout {
		if field.Offset != expectedOffsets[i] {
			t.Errorf("Field %d offset mismatch: predicted=%d, actual=%d",
				i, field.Offset, expectedOffsets[i])
		}
		t.Logf("✅ Field %d (%s): offset=%d (matches C)", i, field.TypeName, field.Offset)
	}

	// Verify sizes were auto-filled
	archInfo := GetArchInfo()
	if layout[0].Size != archInfo.Uint32Size {
		t.Errorf("uint32_t size mismatch: got %d, expected %d", layout[0].Size, archInfo.Uint32Size)
	}
	if layout[1].Size != archInfo.PointerSize {
		t.Errorf("char* size mismatch: got %d, expected %d", layout[1].Size, archInfo.PointerSize)
	}
	if layout[2].Size != archInfo.Float32Size {
		t.Errorf("float size mismatch: got %d, expected %d", layout[2].Size, archInfo.Float32Size)
	}

	// Verify IsString was auto-deduced
	if !layout[1].IsString {
		t.Errorf("char* field should have IsString=true")
	}
	if layout[0].IsString || layout[2].IsString {
		t.Errorf("Non-string fields should have IsString=false")
	}
}

// TestRegistryWithAutoLayout verifies the complete workflow
func TestRegistryWithAutoLayout(t *testing.T) {
	// Go struct
	type Device struct {
		ID    uint32
		Name  string
		Value float32
	}

	// Create registry
	registry := NewRegistry()

	// Register using AutoLayout (no manual offsetof needed!)
	layout := AutoLayout("uint32_t", "char*", "float")
	cSize := GetAutoDeviceSize()

	err := registry.Register(
		reflect.TypeOf(Device{}),
		cSize,
		layout,
		TestConverter{},
	)

	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Create C device
	cDevice := CreateAutoDevice()
	defer FreeAutoDevice(cDevice)

	// Copy to Go
	var device Device
	err = registry.Copy(&device, cDevice)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify data
	if device.ID != 42 {
		t.Errorf("ID mismatch: got %d, expected 42", device.ID)
	}
	if device.Name != "Test Device" {
		t.Errorf("Name mismatch: got %q, expected %q", device.Name, "Test Device")
	}
	if device.Value < 3.13 || device.Value > 3.15 {
		t.Errorf("Value mismatch: got %f, expected ~3.14", device.Value)
	}

	t.Logf("✅ Device copied successfully: ID=%d, Name=%q, Value=%.2f",
		device.ID, device.Name, device.Value)
}

// TestAutoLayoutComplex tests with more complex struct (padding, different types)
func TestAutoLayoutComplex(t *testing.T) {
	type ComplexDevice struct {
		Flag  uint8
		ID    uint32
		Name  string
		Value float64
		Port  uint16
	}

	registry := NewRegistry()

	// AutoLayout handles padding automatically!
	layout := AutoLayout("uint8_t", "uint32_t", "char*", "double", "uint16_t")
	cSize := GetComplexDeviceSize()

	err := registry.Register(
		reflect.TypeOf(ComplexDevice{}),
		cSize,
		layout,
		TestConverter{},
	)

	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Create and copy
	cDevice := CreateComplexDevice()
	defer FreeComplexDevice(cDevice)

	var device ComplexDevice
	err = registry.Copy(&device, cDevice)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	// Verify
	if device.Flag != 1 {
		t.Errorf("Flag mismatch: got %d, expected 1", device.Flag)
	}
	if device.ID != 99 {
		t.Errorf("ID mismatch: got %d, expected 99", device.ID)
	}
	if device.Name != "Complex Device" {
		t.Errorf("Name mismatch: got %q, expected %q", device.Name, "Complex Device")
	}
	if device.Value < 2.71 || device.Value > 2.72 {
		t.Errorf("Value mismatch: got %f, expected ~2.718", device.Value)
	}
	if device.Port != 8080 {
		t.Errorf("Port mismatch: got %d, expected 8080", device.Port)
	}

	t.Logf("✅ Complex device copied successfully: Flag=%d, ID=%d, Name=%q, Value=%.3f, Port=%d",
		device.Flag, device.ID, device.Name, device.Value, device.Port)
}

// TestSizeDeduction verifies Size field is optional (deduced from TypeName)
func TestSizeDeduction(t *testing.T) {
	type SimpleStruct struct {
		A uint32
		B float64
	}

	registry := NewRegistry()
	archInfo := GetArchInfo()

	// Omit Size field - should be deduced from TypeName
	layout := []FieldInfo{
		{Offset: 0, Size: 0, TypeName: "uint32_t"}, // Size=0, will be deduced
		{Offset: 8, Size: 0, TypeName: "double"},   // Size=0, will be deduced
	}

	err := registry.Register(
		reflect.TypeOf(SimpleStruct{}),
		16, // C struct size
		layout,
	)

	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Verify sizes were deduced correctly
	mapping, ok := registry.GetMapping(reflect.TypeOf(SimpleStruct{}))
	if !ok {
		t.Fatal("Mapping not found")
	}

	if mapping.Fields[0].Size != archInfo.Uint32Size {
		t.Errorf("Field 0 size not deduced: got %d, expected %d",
			mapping.Fields[0].Size, archInfo.Uint32Size)
	}
	if mapping.Fields[1].Size != archInfo.Float64Size {
		t.Errorf("Field 1 size not deduced: got %d, expected %d",
			mapping.Fields[1].Size, archInfo.Float64Size)
	}

	t.Logf("✅ Sizes deduced correctly: uint32_t=%d, double=%d",
		mapping.Fields[0].Size, mapping.Fields[1].Size)
}

// TestIsStringDeduction verifies IsString is deduced from TypeName
func TestIsStringDeduction(t *testing.T) {
	type StringStruct struct {
		ID   uint32
		Name string
	}

	registry := NewRegistry()

	// Don't specify IsString - should be auto-deduced from TypeName="char*"
	layout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "uint32_t"},
		{Offset: 8, Size: 8, TypeName: "char*"}, // No IsString specified
	}

	err := registry.Register(
		reflect.TypeOf(StringStruct{}),
		16,
		layout,
		TestConverter{},
	)

	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Verify IsString was deduced
	mapping, ok := registry.GetMapping(reflect.TypeOf(StringStruct{}))
	if !ok {
		t.Fatal("Mapping not found")
	}

	if !mapping.Fields[1].IsString {
		t.Errorf("IsString not deduced for char* field")
	}
	if mapping.Fields[0].IsString {
		t.Errorf("Non-string field incorrectly marked as string")
	}

	t.Logf("✅ IsString deduced correctly from TypeName")
}

// BenchmarkAutoLayout measures AutoLayout performance
func BenchmarkAutoLayout(b *testing.B) {
	typeNames := []string{"uint8_t", "uint32_t", "char*", "double", "uint16_t"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AutoLayout(typeNames...)
	}
}

// BenchmarkManualLayout compares with manual layout creation
func BenchmarkManualLayout(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = []FieldInfo{
			{Offset: 0, Size: 1, TypeName: "uint8_t"},
			{Offset: 4, Size: 4, TypeName: "uint32_t"},
			{Offset: 8, Size: 8, TypeName: "char*", IsString: true},
			{Offset: 16, Size: 8, TypeName: "double"},
			{Offset: 24, Size: 2, TypeName: "uint16_t"},
		}
	}
}
