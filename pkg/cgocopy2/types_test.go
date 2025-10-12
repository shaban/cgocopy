package cgocopy2

import (
	"reflect"
	"testing"
)

func TestFieldType_String(t *testing.T) {
	tests := []struct {
		name     string
		ft       FieldType
		expected string
	}{
		{"Primitive", FieldTypePrimitive, "Primitive"},
		{"String", FieldTypeString, "String"},
		{"Struct", FieldTypeStruct, "Struct"},
		{"Array", FieldTypeArray, "Array"},
		{"Slice", FieldTypeSlice, "Slice"},
		{"Pointer", FieldTypePointer, "Pointer"},
		{"Invalid", FieldTypeInvalid, "Invalid"},
		{"Unknown", FieldType(999), "Invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ft.String()
			if result != tt.expected {
				t.Errorf("FieldType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	
	if r.metadata == nil {
		t.Error("registry metadata map is nil")
	}
	
	if r.cTypeMap == nil {
		t.Error("registry cTypeMap is nil")
	}
	
	if r.Count() != 0 {
		t.Errorf("new registry Count() = %d, want 0", r.Count())
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	
	type TestStruct struct {
		ID   int
		Name string
	}
	
	goType := reflect.TypeOf(TestStruct{})
	metadata := &StructMetadata{
		TypeName:  "TestStruct",
		CTypeName: "TestStruct",
		GoType:    goType,
		Size:      16,
		Fields: []FieldInfo{
			{Name: "ID", CName: "ID", Type: FieldTypePrimitive, Index: 0},
			{Name: "Name", CName: "Name", Type: FieldTypeString, Index: 1},
		},
	}
	
	// Register the type
	r.Register(goType, metadata)
	
	// Verify Count
	if count := r.Count(); count != 1 {
		t.Errorf("Count() = %d, want 1", count)
	}
	
	// Verify Get
	retrieved := r.Get(goType)
	if retrieved == nil {
		t.Fatal("Get() returned nil")
	}
	
	if retrieved.TypeName != metadata.TypeName {
		t.Errorf("retrieved TypeName = %q, want %q", retrieved.TypeName, metadata.TypeName)
	}
	
	if len(retrieved.Fields) != len(metadata.Fields) {
		t.Errorf("retrieved Fields length = %d, want %d", len(retrieved.Fields), len(metadata.Fields))
	}
}

func TestRegistry_GetByCName(t *testing.T) {
	r := NewRegistry()
	
	type User struct {
		ID int
	}
	
	goType := reflect.TypeOf(User{})
	metadata := &StructMetadata{
		TypeName:  "User",
		CTypeName: "User",
		GoType:    goType,
		Size:      8,
	}
	
	r.Register(goType, metadata)
	
	// Get by C name
	retrieved := r.GetByCName("User")
	if retrieved == nil {
		t.Fatal("GetByCName() returned nil")
	}
	
	if retrieved.TypeName != "User" {
		t.Errorf("retrieved TypeName = %q, want %q", retrieved.TypeName, "User")
	}
	
	// Get non-existent type
	notFound := r.GetByCName("NonExistent")
	if notFound != nil {
		t.Error("GetByCName() for non-existent type should return nil")
	}
}

func TestRegistry_IsRegistered(t *testing.T) {
	r := NewRegistry()
	
	type Product struct {
		Price float64
	}
	
	goType := reflect.TypeOf(Product{})
	
	// Should not be registered initially
	if r.IsRegistered(goType) {
		t.Error("IsRegistered() = true for unregistered type, want false")
	}
	
	// Register it
	metadata := &StructMetadata{
		TypeName:  "Product",
		CTypeName: "Product",
		GoType:    goType,
		Size:      8,
	}
	r.Register(goType, metadata)
	
	// Should be registered now
	if !r.IsRegistered(goType) {
		t.Error("IsRegistered() = false for registered type, want true")
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()
	
	type Item struct {
		Value int
	}
	
	goType := reflect.TypeOf(Item{})
	metadata := &StructMetadata{
		TypeName:  "Item",
		CTypeName: "Item",
		GoType:    goType,
		Size:      8,
	}
	
	r.Register(goType, metadata)
	
	if r.Count() != 1 {
		t.Fatalf("Count() = %d before Clear(), want 1", r.Count())
	}
	
	// Clear the registry
	r.Clear()
	
	if r.Count() != 0 {
		t.Errorf("Count() = %d after Clear(), want 0", r.Count())
	}
	
	if r.IsRegistered(goType) {
		t.Error("IsRegistered() = true after Clear(), want false")
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	r := NewRegistry()
	
	type Concurrent struct {
		Value int
	}
	
	// Test concurrent Register and Get operations
	done := make(chan bool, 10)
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			goType := reflect.TypeOf(Concurrent{})
			metadata := &StructMetadata{
				TypeName:  "Concurrent",
				CTypeName: "Concurrent",
				GoType:    goType,
				Size:      8,
			}
			r.Register(goType, metadata)
			done <- true
		}(i)
	}
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			goType := reflect.TypeOf(Concurrent{})
			_ = r.Get(goType)
			_ = r.IsRegistered(goType)
			_ = r.Count()
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should have successfully registered
	goType := reflect.TypeOf(Concurrent{})
	if !r.IsRegistered(goType) {
		t.Error("type not registered after concurrent operations")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Test that globalRegistry is initialized
	if globalRegistry == nil {
		t.Fatal("globalRegistry is nil")
	}
	
	// Clear it for a clean test
	globalRegistry.Clear()
	
	type GlobalTest struct {
		X int
	}
	
	goType := reflect.TypeOf(GlobalTest{})
	metadata := &StructMetadata{
		TypeName:  "GlobalTest",
		CTypeName: "GlobalTest",
		GoType:    goType,
		Size:      8,
	}
	
	globalRegistry.Register(goType, metadata)
	
	if !globalRegistry.IsRegistered(goType) {
		t.Error("globalRegistry does not have registered type")
	}
	
	// Clean up
	globalRegistry.Clear()
}
