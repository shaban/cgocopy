package cgocopy2

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// Test types for Precompile tests
type SimpleStruct struct {
	ID   int
	Name string
}

type TaggedStruct struct {
	UserID   int    `cgocopy:"id"`
	FullName string `cgocopy:"name"`
	Internal string `cgocopy:"-"` // Should be skipped
}

type NestedStruct struct {
	ID      int
	Profile SimpleStruct
}

type ArrayStruct struct {
	Values [5]int
	Names  [3]string
}

type SliceStruct struct {
	Items []int
	Tags  []string
}

type PointerStruct struct {
	Next *SimpleStruct
	Prev *PointerStruct
}

type MixedStruct struct {
	ID        int
	Name      string
	Count     uint32
	Score     float64
	Active    bool
	Tags      []string
	Metadata  SimpleStruct
	Neighbors [2]*SimpleStruct
}

type UnsupportedStruct struct {
	ID       int
	Callback func() // Should fail
}

type UnexportedFieldStruct struct {
	ID       int
	Name     string
	internal string // Should be skipped automatically
}

func TestPrecompile_SimpleStruct(t *testing.T) {
	Reset() // Clear registry
	defer Reset()

	err := Precompile[SimpleStruct]()
	if err != nil {
		t.Fatalf("Precompile() error = %v, want nil", err)
	}

	// Verify it's registered
	if !IsRegistered[SimpleStruct]() {
		t.Error("IsRegistered() = false, want true")
	}

	// Get metadata
	metadata := GetMetadata[SimpleStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil, want metadata")
	}

	if metadata.TypeName != "SimpleStruct" {
		t.Errorf("TypeName = %q, want %q", metadata.TypeName, "SimpleStruct")
	}

	if len(metadata.Fields) != 2 {
		t.Errorf("len(Fields) = %d, want 2", len(metadata.Fields))
	}

	// Check field details
	if metadata.Fields[0].Name != "ID" || metadata.Fields[0].Type != FieldTypePrimitive {
		t.Errorf("Field[0] = %+v, want {Name:ID, Type:Primitive}", metadata.Fields[0])
	}

	if metadata.Fields[1].Name != "Name" || metadata.Fields[1].Type != FieldTypeString {
		t.Errorf("Field[1] = %+v, want {Name:Name, Type:String}", metadata.Fields[1])
	}
}

func TestPrecompile_TaggedStruct(t *testing.T) {
	Reset()
	defer Reset()

	err := Precompile[TaggedStruct]()
	if err != nil {
		t.Fatalf("Precompile() error = %v, want nil", err)
	}

	metadata := GetMetadata[TaggedStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil")
	}

	// Should have 2 fields (Internal is skipped)
	if len(metadata.Fields) != 2 {
		t.Errorf("len(Fields) = %d, want 2 (Internal should be skipped)", len(metadata.Fields))
	}

	// Check that CName uses tag values
	if metadata.Fields[0].CName != "id" {
		t.Errorf("Fields[0].CName = %q, want %q", metadata.Fields[0].CName, "id")
	}

	if metadata.Fields[1].CName != "name" {
		t.Errorf("Fields[1].CName = %q, want %q", metadata.Fields[1].CName, "name")
	}
}

func TestPrecompile_NestedStruct(t *testing.T) {
	Reset()
	defer Reset()

	// Precompile both types
	if err := Precompile[SimpleStruct](); err != nil {
		t.Fatalf("Precompile[SimpleStruct]() error = %v", err)
	}

	if err := Precompile[NestedStruct](); err != nil {
		t.Fatalf("Precompile[NestedStruct]() error = %v", err)
	}

	metadata := GetMetadata[NestedStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil")
	}

	if !metadata.HasNestedStructs {
		t.Error("HasNestedStructs = false, want true")
	}

	// Find the Profile field
	var profileField *FieldInfo
	for i := range metadata.Fields {
		if metadata.Fields[i].Name == "Profile" {
			profileField = &metadata.Fields[i]
			break
		}
	}

	if profileField == nil {
		t.Fatal("Profile field not found")
	}

	if profileField.Type != FieldTypeStruct {
		t.Errorf("Profile.Type = %v, want FieldTypeStruct", profileField.Type)
	}
}

func TestPrecompile_ArrayStruct(t *testing.T) {
	Reset()
	defer Reset()

	err := Precompile[ArrayStruct]()
	if err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	metadata := GetMetadata[ArrayStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil")
	}

	// Check Values field
	valuesField := metadata.Fields[0]
	if valuesField.Type != FieldTypeArray {
		t.Errorf("Values.Type = %v, want FieldTypeArray", valuesField.Type)
	}

	if valuesField.ArrayLen != 5 {
		t.Errorf("Values.ArrayLen = %d, want 5", valuesField.ArrayLen)
	}

	if valuesField.ElemType.Kind() != reflect.Int {
		t.Errorf("Values.ElemType = %v, want int", valuesField.ElemType)
	}

	// Check Names field
	namesField := metadata.Fields[1]
	if namesField.ArrayLen != 3 {
		t.Errorf("Names.ArrayLen = %d, want 3", namesField.ArrayLen)
	}
}

func TestPrecompile_SliceStruct(t *testing.T) {
	Reset()
	defer Reset()

	err := Precompile[SliceStruct]()
	if err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	metadata := GetMetadata[SliceStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil")
	}

	// Check Items field
	itemsField := metadata.Fields[0]
	if itemsField.Type != FieldTypeSlice {
		t.Errorf("Items.Type = %v, want FieldTypeSlice", itemsField.Type)
	}

	if itemsField.ElemType.Kind() != reflect.Int {
		t.Errorf("Items.ElemType = %v, want int", itemsField.ElemType)
	}
}

func TestPrecompile_PointerStruct(t *testing.T) {
	Reset()
	defer Reset()

	err := Precompile[PointerStruct]()
	if err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	metadata := GetMetadata[PointerStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil")
	}

	// Check Next field
	nextField := metadata.Fields[0]
	if nextField.Type != FieldTypePointer {
		t.Errorf("Next.Type = %v, want FieldTypePointer", nextField.Type)
	}

	if nextField.ElemType.Name() != "SimpleStruct" {
		t.Errorf("Next.ElemType = %v, want SimpleStruct", nextField.ElemType)
	}
}

func TestPrecompile_MixedStruct(t *testing.T) {
	Reset()
	defer Reset()

	// Precompile dependencies first
	Precompile[SimpleStruct]()

	err := Precompile[MixedStruct]()
	if err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	metadata := GetMetadata[MixedStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil")
	}

	// Verify all field types are correctly categorized
	expectedTypes := map[string]FieldType{
		"ID":        FieldTypePrimitive,
		"Name":      FieldTypeString,
		"Count":     FieldTypePrimitive,
		"Score":     FieldTypePrimitive,
		"Active":    FieldTypePrimitive,
		"Tags":      FieldTypeSlice,
		"Metadata":  FieldTypeStruct,
		"Neighbors": FieldTypeArray,
	}

	for _, field := range metadata.Fields {
		expected, ok := expectedTypes[field.Name]
		if !ok {
			t.Errorf("unexpected field %q", field.Name)
			continue
		}

		if field.Type != expected {
			t.Errorf("field %q: Type = %v, want %v", field.Name, field.Type, expected)
		}
	}
}

func TestPrecompile_UnsupportedType(t *testing.T) {
	Reset()
	defer Reset()

	err := Precompile[UnsupportedStruct]()
	if err == nil {
		t.Fatal("Precompile() error = nil, want error for unsupported type")
	}

	// The error should be a RegistrationError wrapping a ValidationError
	var regErr *RegistrationError
	if !errors.As(err, &regErr) {
		t.Fatalf("error is not RegistrationError: %v", err)
	}

	// The underlying cause should mention unsupported type
	errMsg := err.Error()
	if !strings.Contains(errMsg, "unsupported") {
		t.Errorf("error message %q should contain 'unsupported'", errMsg)
	}
}

func TestPrecompile_UnexportedFields(t *testing.T) {
	Reset()
	defer Reset()

	err := Precompile[UnexportedFieldStruct]()
	if err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	metadata := GetMetadata[UnexportedFieldStruct]()
	if metadata == nil {
		t.Fatal("GetMetadata() = nil")
	}

	// Should only have 2 exported fields
	if len(metadata.Fields) != 2 {
		t.Errorf("len(Fields) = %d, want 2 (unexported field should be skipped)", len(metadata.Fields))
	}
}

func TestPrecompile_AlreadyRegistered(t *testing.T) {
	Reset()
	defer Reset()

	// First registration should succeed
	err := Precompile[SimpleStruct]()
	if err != nil {
		t.Fatalf("first Precompile() error = %v", err)
	}

	// Second registration should fail
	err = Precompile[SimpleStruct]()
	if err == nil {
		t.Fatal("second Precompile() error = nil, want error")
	}

	if !errors.Is(err, ErrAlreadyRegistered) {
		t.Errorf("error does not wrap ErrAlreadyRegistered: %v", err)
	}
}

func TestPrecompile_NonStructType(t *testing.T) {
	Reset()
	defer Reset()

	// Test with int (primitive)
	err := Precompile[int]()
	if err == nil {
		t.Error("Precompile[int]() error = nil, want error")
	}

	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("error does not wrap ErrInvalidType: %v", err)
	}

	// Test with string
	err = Precompile[string]()
	if err == nil {
		t.Error("Precompile[string]() error = nil, want error")
	}

	// Test with slice
	err = Precompile[[]int]()
	if err == nil {
		t.Error("Precompile[[]int]() error = nil, want error")
	}
}

func TestPrecompile_PointerType(t *testing.T) {
	Reset()
	defer Reset()

	// Should dereference pointer and register the struct
	err := Precompile[*SimpleStruct]()
	if err != nil {
		t.Fatalf("Precompile[*SimpleStruct]() error = %v", err)
	}

	// Should be registered under the struct type (not pointer type)
	if !IsRegistered[SimpleStruct]() {
		t.Error("IsRegistered[SimpleStruct]() = false, want true")
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		wantName string
		wantSkip bool
	}{
		{"empty tag", "", "", false},
		{"skip marker", "-", "", true},
		{"field name", "c_field_name", "c_field_name", false},
		{"with spaces", "  field  ", "field", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotSkip := parseTag(tt.tag)
			if gotName != tt.wantName {
				t.Errorf("parseTag(%q) name = %q, want %q", tt.tag, gotName, tt.wantName)
			}
			if gotSkip != tt.wantSkip {
				t.Errorf("parseTag(%q) skip = %v, want %v", tt.tag, gotSkip, tt.wantSkip)
			}
		})
	}
}

func TestCategorizeFieldType(t *testing.T) {
	tests := []struct {
		name    string
		typ     reflect.Type
		want    FieldType
		wantErr bool
	}{
		{"int", reflect.TypeOf(0), FieldTypePrimitive, false},
		{"int32", reflect.TypeOf(int32(0)), FieldTypePrimitive, false},
		{"uint64", reflect.TypeOf(uint64(0)), FieldTypePrimitive, false},
		{"float64", reflect.TypeOf(float64(0)), FieldTypePrimitive, false},
		{"bool", reflect.TypeOf(true), FieldTypePrimitive, false},
		{"string", reflect.TypeOf(""), FieldTypeString, false},
		{"struct", reflect.TypeOf(SimpleStruct{}), FieldTypeStruct, false},
		{"array", reflect.TypeOf([5]int{}), FieldTypeArray, false},
		{"slice", reflect.TypeOf([]int{}), FieldTypeSlice, false},
		{"pointer", reflect.TypeOf((*int)(nil)), FieldTypePointer, false},
		{"func", reflect.TypeOf(func() {}), FieldTypeInvalid, true},
		{"map", reflect.TypeOf(map[string]int{}), FieldTypeInvalid, true},
		{"chan", reflect.TypeOf(make(chan int)), FieldTypeInvalid, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := categorizeFieldType(tt.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("categorizeFieldType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("categorizeFieldType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReset(t *testing.T) {
	// Register some types
	Precompile[SimpleStruct]()
	Precompile[TaggedStruct]()

	if globalRegistry.Count() == 0 {
		t.Fatal("registry is empty before Reset()")
	}

	// Reset should clear everything
	Reset()

	if globalRegistry.Count() != 0 {
		t.Errorf("registry Count() = %d after Reset(), want 0", globalRegistry.Count())
	}

	if IsRegistered[SimpleStruct]() {
		t.Error("SimpleStruct is still registered after Reset()")
	}
}

func TestIsPrimitiveKind(t *testing.T) {
	primitiveKinds := []reflect.Kind{
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool,
	}

	for _, kind := range primitiveKinds {
		if !isPrimitiveKind(kind) {
			t.Errorf("isPrimitiveKind(%v) = false, want true", kind)
		}
	}

	nonPrimitiveKinds := []reflect.Kind{
		reflect.String, reflect.Struct, reflect.Array, reflect.Slice,
		reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func,
	}

	for _, kind := range nonPrimitiveKinds {
		if isPrimitiveKind(kind) {
			t.Errorf("isPrimitiveKind(%v) = true, want false", kind)
		}
	}
}
