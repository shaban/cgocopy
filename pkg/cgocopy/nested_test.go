package cgocopy

import (
	"reflect"
	"testing"
)

// Go equivalent structs
type GoInnerStruct struct {
	X int32
	Y int32
}

type GoMiddleStruct struct {
	ID    int32
	Inner GoInnerStruct
	Value float32
}

type GoOuterStruct struct {
	OuterID   int32
	Middle    GoMiddleStruct
	Timestamp int64
}

func TestNestedStructSingleLevel(t *testing.T) {
	registry := NewRegistry()

	// Register inner struct first (using init-time captured layout)
	innerLayout := []FieldInfo{
		{Offset: innerStructXOffset, Size: 4, TypeName: "int32_t"},
		{Offset: innerStructYOffset, Size: 4, TypeName: "int32_t"},
	}
	err := registry.Register(reflect.TypeOf(GoInnerStruct{}), innerStructSize, innerLayout, nil)
	if err != nil {
		t.Fatalf("Failed to register inner struct: %v", err)
	}

	// Try to copy inner struct
	cInner := CreateInnerStruct()
	defer FreePtr(cInner)

	var goInner GoInnerStruct
	err = registry.Copy(&goInner, cInner)
	if err != nil {
		t.Fatalf("Failed to copy inner struct: %v", err)
	}

	if goInner.X != 10 || goInner.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", goInner.X, goInner.Y)
	}
}

func TestNestedStructTwoLevels(t *testing.T) {
	registry := NewRegistry()

	// Register inner struct first (using init-time captured layout)
	innerLayout := []FieldInfo{
		{Offset: innerStructXOffset, Size: 4, TypeName: "int32_t"},
		{Offset: innerStructYOffset, Size: 4, TypeName: "int32_t"},
	}
	err := registry.Register(reflect.TypeOf(GoInnerStruct{}), innerStructSize, innerLayout, nil)
	if err != nil {
		t.Fatalf("Failed to register inner struct: %v", err)
	}

	// Register middle struct (contains inner)
	middleLayout := []FieldInfo{
		{Offset: middleStructIdOffset, Size: 4, TypeName: "int32_t"},
		{Offset: middleStructInnerOffset, Size: 8, TypeName: "struct"}, // Nested InnerStruct
		{Offset: middleStructValueOffset, Size: 4, TypeName: "float"},
	}
	err = registry.Register(reflect.TypeOf(GoMiddleStruct{}), middleStructSize, middleLayout, nil)
	if err != nil {
		t.Fatalf("Failed to register middle struct: %v", err)
	}

	// Copy middle struct
	cMiddle := CreateMiddleStruct()
	defer FreePtr(cMiddle)

	var goMiddle GoMiddleStruct
	err = registry.Copy(&goMiddle, cMiddle)
	if err != nil {
		t.Fatalf("Failed to copy middle struct: %v", err)
	}

	// Verify all values including nested
	if goMiddle.ID != 100 {
		t.Errorf("Expected ID=100, got %d", goMiddle.ID)
	}
	if goMiddle.Inner.X != 10 {
		t.Errorf("Expected Inner.X=10, got %d", goMiddle.Inner.X)
	}
	if goMiddle.Inner.Y != 20 {
		t.Errorf("Expected Inner.Y=20, got %d", goMiddle.Inner.Y)
	}
	if goMiddle.Value != 3.14 {
		t.Errorf("Expected Value=3.14, got %f", goMiddle.Value)
	}
}

func TestNestedStructThreeLevels(t *testing.T) {
	registry := NewRegistry()

	// Register from innermost to outermost (using init-time captured layouts)

	// 1. Inner struct
	innerLayout := []FieldInfo{
		{Offset: innerStructXOffset, Size: 4, TypeName: "int32_t"},
		{Offset: innerStructYOffset, Size: 4, TypeName: "int32_t"},
	}
	err := registry.Register(reflect.TypeOf(GoInnerStruct{}), innerStructSize, innerLayout, nil)
	if err != nil {
		t.Fatalf("Failed to register inner struct: %v", err)
	}

	// 2. Middle struct
	middleLayout := []FieldInfo{
		{Offset: middleStructIdOffset, Size: 4, TypeName: "int32_t"},
		{Offset: middleStructInnerOffset, Size: 8, TypeName: "struct"},
		{Offset: middleStructValueOffset, Size: 4, TypeName: "float"},
	}
	err = registry.Register(reflect.TypeOf(GoMiddleStruct{}), middleStructSize, middleLayout, nil)
	if err != nil {
		t.Fatalf("Failed to register middle struct: %v", err)
	}

	// 3. Outer struct (note: actual offsets determined at init time based on platform)
	outerLayout := []FieldInfo{
		{Offset: outerStructOuterIDOffset, Size: 4, TypeName: "int32_t"},
		{Offset: outerStructMiddleOffset, Size: 16, TypeName: "struct"},    // Nested MiddleStruct
		{Offset: outerStructTimestampOffset, Size: 8, TypeName: "int64_t"}, // Padding-aware offset
	}
	err = registry.Register(reflect.TypeOf(GoOuterStruct{}), outerStructSize, outerLayout, nil)
	if err != nil {
		t.Fatalf("Failed to register outer struct: %v", err)
	}

	// Copy outer struct (3 levels deep!)
	cOuter := CreateOuterStruct()
	defer FreePtr(cOuter)

	var goOuter GoOuterStruct
	err = registry.Copy(&goOuter, cOuter)
	if err != nil {
		t.Fatalf("Failed to copy outer struct: %v", err)
	}

	// Verify all values at all levels
	if goOuter.OuterID != 1000 {
		t.Errorf("Expected OuterID=1000, got %d", goOuter.OuterID)
	}
	if goOuter.Middle.ID != 100 {
		t.Errorf("Expected Middle.ID=100, got %d", goOuter.Middle.ID)
	}
	if goOuter.Middle.Inner.X != 10 {
		t.Errorf("Expected Middle.Inner.X=10, got %d", goOuter.Middle.Inner.X)
	}
	if goOuter.Middle.Inner.Y != 20 {
		t.Errorf("Expected Middle.Inner.Y=20, got %d", goOuter.Middle.Inner.Y)
	}
	if goOuter.Middle.Value != 3.14 {
		t.Errorf("Expected Middle.Value=3.14, got %f", goOuter.Middle.Value)
	}
	if goOuter.Timestamp != 9999999 {
		t.Errorf("Expected Timestamp=9999999, got %d", goOuter.Timestamp)
	}

	t.Logf("✅ Successfully copied 3-level nested struct recursively!")
}

func TestNestedStructMustBeRegisteredFirst(t *testing.T) {
	registry := NewRegistry()

	// Try to register middle struct WITHOUT registering inner first
	middleLayout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: 4, Size: 8, TypeName: "struct"},
		{Offset: 12, Size: 4, TypeName: "float"},
	}

	err := registry.Register(reflect.TypeOf(GoMiddleStruct{}), MiddleStructSize(), middleLayout, nil)
	if err == nil {
		t.Fatal("Expected error when nested struct not registered first, got nil")
	}

	if !contains(err.Error(), "must be registered first") {
		t.Errorf("Expected 'must be registered first' error, got: %v", err)
	}

	t.Logf("✅ Correctly rejected: %v", err)
}

func BenchmarkNestedStructCopy(b *testing.B) {
	registry := NewRegistry()

	// Setup registration
	innerLayout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: 4, Size: 4, TypeName: "int32_t"},
	}
	registry.Register(reflect.TypeOf(GoInnerStruct{}), InnerStructSize(), innerLayout, nil)

	middleLayout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: 4, Size: 8, TypeName: "struct"},
		{Offset: 12, Size: 4, TypeName: "float"},
	}
	registry.Register(reflect.TypeOf(GoMiddleStruct{}), MiddleStructSize(), middleLayout, nil)

	outerLayout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: 4, Size: 16, TypeName: "struct"},
		{Offset: 20, Size: 8, TypeName: "int64_t"},
	}
	registry.Register(reflect.TypeOf(GoOuterStruct{}), OuterStructSize(), outerLayout, nil)

	cOuter := CreateOuterStruct()
	defer FreePtr(cOuter)

	var goOuter GoOuterStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&goOuter, cOuter); err != nil {
			b.Fatal(err)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
