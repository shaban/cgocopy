package cgocopy

import (
	"reflect"
	"testing"
)

// Go equivalent struct
type TestStruct struct {
	ID        int32
	Value     float32
	Timestamp int64
}

func TestStructCopyBasic(t *testing.T) {
	registry := NewRegistry()

	// Register the struct mapping (using init-time captured layout)
	cLayout := []FieldInfo{
		{Offset: testStructIdOffset, Size: 4, TypeName: "int32_t"},        // id
		{Offset: testStructValueOffset, Size: 4, TypeName: "float"},       // value
		{Offset: testStructTimestampOffset, Size: 8, TypeName: "int64_t"}, // timestamp
	}

	goType := reflect.TypeOf(TestStruct{})
	err := registry.Register(goType, testStructSize, cLayout)
	if err != nil {
		t.Fatalf("Failed to register struct: %v", err)
	}

	// Print mapping info
	if mapping, ok := registry.GetMapping(goType); ok {
		t.Logf("Registered mapping:\n%s", mapping.String())
	}

	// Create a C struct
	cStruct := CreateTestStruct()
	defer FreeTestStruct(cStruct)

	// Copy to Go struct
	var goStruct TestStruct
	err = registry.Copy(&goStruct, cStruct)
	if err != nil {
		t.Fatalf("Failed to copy struct: %v", err)
	}

	// Verify values
	if goStruct.ID != 42 {
		t.Errorf("Expected ID=42, got %d", goStruct.ID)
	}
	if goStruct.Value != 3.14 {
		t.Errorf("Expected Value=3.14, got %f", goStruct.Value)
	}
	if goStruct.Timestamp != 1234567890 {
		t.Errorf("Expected Timestamp=1234567890, got %d", goStruct.Timestamp)
	}
}

func TestStructCopyFieldMismatch(t *testing.T) {
	registry := NewRegistry()

	// Try to register with wrong number of fields
	cLayout := []FieldInfo{
		{Offset: testStructIdOffset, Size: 4, TypeName: "int32_t"},
		{Offset: testStructValueOffset, Size: 4, TypeName: "float"},
		// Missing timestamp field
	}

	goType := reflect.TypeOf(TestStruct{})
	err := registry.Register(goType, testStructSize, cLayout)
	if err == nil {
		t.Fatal("Expected error for field count mismatch, got nil")
	}
	t.Logf("Correctly rejected: %v", err)
}

func TestStructCopySizeMismatch(t *testing.T) {
	registry := NewRegistry()

	// Try to register with wrong field size
	cLayout := []FieldInfo{
		{Offset: testStructIdOffset, Size: 8, TypeName: "int32_t"}, // Wrong size!
		{Offset: testStructValueOffset, Size: 4, TypeName: "float"},
		{Offset: testStructTimestampOffset, Size: 8, TypeName: "int64_t"},
	}

	goType := reflect.TypeOf(TestStruct{})
	err := registry.Register(goType, testStructSize, cLayout)
	if err == nil {
		t.Fatal("Expected error for size mismatch, got nil")
	}
	t.Logf("Correctly rejected: %v", err)
}

func TestStructCopyTypeMismatch(t *testing.T) {
	registry := NewRegistry()

	// Try to register with incompatible type
	cLayout := []FieldInfo{
		{Offset: testStructIdOffset, Size: 4, TypeName: "float"}, // Wrong type!
		{Offset: testStructValueOffset, Size: 4, TypeName: "float"},
		{Offset: testStructTimestampOffset, Size: 8, TypeName: "int64_t"},
	}

	goType := reflect.TypeOf(TestStruct{})
	err := registry.Register(goType, testStructSize, cLayout)
	if err == nil {
		t.Fatal("Expected error for type mismatch, got nil")
	}
	t.Logf("Correctly rejected: %v", err)
}

func BenchmarkStructCopy(b *testing.B) {
	registry := NewRegistry()

	cLayout := []FieldInfo{
		{Offset: testStructIdOffset, Size: 4, TypeName: "int32_t"},
		{Offset: testStructValueOffset, Size: 4, TypeName: "float"},
		{Offset: testStructTimestampOffset, Size: 8, TypeName: "int64_t"},
	}

	goType := reflect.TypeOf(TestStruct{})
	if err := registry.Register(goType, testStructSize, cLayout); err != nil {
		b.Fatalf("Failed to register: %v", err)
	}

	cStruct := CreateTestStruct()
	defer FreeTestStruct(cStruct)

	var goStruct TestStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&goStruct, cStruct); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDirectUnsafeCopy(b *testing.B) {
	cStruct := CreateTestStruct()
	defer FreeTestStruct(cStruct)

	var goStruct TestStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Direct unsafe copy (fastest possible)
		goStruct = *(*TestStruct)(cStruct)
		// Prevent compiler from optimizing away
		_ = goStruct
	}
}
