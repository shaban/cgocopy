package cgocopy2

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestValidateStruct_Valid(t *testing.T) {
	Reset()
	defer Reset()

	type ValidStruct struct {
		ID   int
		Name string
	}

	if err := Precompile[ValidStruct](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	err := ValidateStruct[ValidStruct]()
	if err != nil {
		t.Errorf("ValidateStruct() error = %v, want nil", err)
	}
}

func TestValidateStruct_NotRegistered(t *testing.T) {
	Reset()
	defer Reset()

	type UnregisteredStruct struct {
		ID int
	}

	err := ValidateStruct[UnregisteredStruct]()
	if err == nil {
		t.Fatal("ValidateStruct() error = nil, want error for unregistered type")
	}

	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("error message should mention 'not registered', got: %v", err)
	}
}

func TestValidateStruct_NotStruct(t *testing.T) {
	err := ValidateStruct[int]()
	if err == nil {
		t.Fatal("ValidateStruct() error = nil, want error for non-struct type")
	}

	if !strings.Contains(err.Error(), "not a struct") {
		t.Errorf("error message should mention 'not a struct', got: %v", err)
	}
}

func TestValidateStruct_WithNestedStruct(t *testing.T) {
	Reset()
	defer Reset()

	type Inner struct {
		Value int
	}

	type Outer struct {
		ID    int
		Inner Inner
	}

	// Register both types
	if err := Precompile[Inner](); err != nil {
		t.Fatalf("Precompile[Inner]() error = %v", err)
	}

	if err := Precompile[Outer](); err != nil {
		t.Fatalf("Precompile[Outer]() error = %v", err)
	}

	// Should validate successfully
	err := ValidateStruct[Outer]()
	if err != nil {
		t.Errorf("ValidateStruct() error = %v, want nil", err)
	}
}

func TestValidateStruct_NestedStructNotRegistered(t *testing.T) {
	Reset()
	defer Reset()

	type Inner struct {
		Value int
	}

	type Outer struct {
		ID    int
		Inner Inner
	}

	// Only register Outer, not Inner
	if err := Precompile[Outer](); err != nil {
		t.Fatalf("Precompile[Outer]() error = %v", err)
	}

	err := ValidateStruct[Outer]()
	if err == nil {
		t.Fatal("ValidateStruct() error = nil, want error for unregistered nested struct")
	}

	if !strings.Contains(err.Error(), "Inner") || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("error should mention unregistered Inner struct, got: %v", err)
	}
}

func TestValidateStruct_WithArrays(t *testing.T) {
	Reset()
	defer Reset()

	type WithArrays struct {
		Values [5]int
		Names  [3]string
	}

	if err := Precompile[WithArrays](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	err := ValidateStruct[WithArrays]()
	if err != nil {
		t.Errorf("ValidateStruct() error = %v, want nil", err)
	}
}

func TestValidateStruct_WithSlices(t *testing.T) {
	Reset()
	defer Reset()

	type WithSlices struct {
		Items []int
		Tags  []string
	}

	if err := Precompile[WithSlices](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	err := ValidateStruct[WithSlices]()
	if err != nil {
		t.Errorf("ValidateStruct() error = %v, want nil", err)
	}
}

func TestValidateStruct_WithPointers(t *testing.T) {
	Reset()
	defer Reset()

	type Simple struct {
		X int
	}

	type WithPointers struct {
		Next *Simple
	}

	// Register both types
	if err := Precompile[Simple](); err != nil {
		t.Fatalf("Precompile[Simple]() error = %v", err)
	}

	if err := Precompile[WithPointers](); err != nil {
		t.Fatalf("Precompile[WithPointers]() error = %v", err)
	}

	err := ValidateStruct[WithPointers]()
	if err != nil {
		t.Errorf("ValidateStruct() error = %v, want nil", err)
	}
}

func TestValidateStruct_ComplexNesting(t *testing.T) {
	Reset()
	defer Reset()

	type Level1 struct {
		Value int
	}

	type Level2 struct {
		Data Level1
		Ref  *Level1
	}

	type Level3 struct {
		Items [2]Level2
		Ptr   *Level2
	}

	// Register all types
	if err := Precompile[Level1](); err != nil {
		t.Fatalf("Precompile[Level1]() error = %v", err)
	}

	if err := Precompile[Level2](); err != nil {
		t.Fatalf("Precompile[Level2]() error = %v", err)
	}

	if err := Precompile[Level3](); err != nil {
		t.Fatalf("Precompile[Level3]() error = %v", err)
	}

	// Validate all
	if err := ValidateStruct[Level1](); err != nil {
		t.Errorf("ValidateStruct[Level1]() error = %v", err)
	}

	if err := ValidateStruct[Level2](); err != nil {
		t.Errorf("ValidateStruct[Level2]() error = %v", err)
	}

	if err := ValidateStruct[Level3](); err != nil {
		t.Errorf("ValidateStruct[Level3]() error = %v", err)
	}
}

func TestValidateStruct_ArrayOfStructsNotRegistered(t *testing.T) {
	Reset()
	defer Reset()

	type Element struct {
		Value int
	}

	type Container struct {
		Items [3]Element
	}

	// Only register Container, not Element
	if err := Precompile[Container](); err != nil {
		t.Fatalf("Precompile[Container]() error = %v", err)
	}

	err := ValidateStruct[Container]()
	if err == nil {
		t.Fatal("ValidateStruct() error = nil, want error for unregistered array element type")
	}

	if !strings.Contains(err.Error(), "Element") {
		t.Errorf("error should mention Element type, got: %v", err)
	}
}

func TestValidateAll(t *testing.T) {
	Reset()
	defer Reset()

	type Type1 struct {
		X int
	}

	type Type2 struct {
		Y string
	}

	// Register both
	Precompile[Type1]()
	Precompile[Type2]()

	errors := ValidateAll()
	if len(errors) != 0 {
		t.Errorf("ValidateAll() returned %d errors, want 0: %v", len(errors), errors)
	}
}

func TestValidateAll_WithErrors(t *testing.T) {
	Reset()
	defer Reset()

	type Inner struct {
		Value int
	}

	type Outer struct {
		Data Inner
	}

	// Register Outer but not Inner (this creates an invalid state)
	Precompile[Outer]()

	errors := ValidateAll()
	if len(errors) == 0 {
		t.Error("ValidateAll() returned no errors, want at least 1 for unregistered nested struct")
	}
}

func TestMustValidateStruct_Valid(t *testing.T) {
	Reset()
	defer Reset()

	type ValidStruct struct {
		ID int
	}

	if err := Precompile[ValidStruct](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	// Should not panic
	MustValidateStruct[ValidStruct]()
}

func TestMustValidateStruct_Invalid_Panics(t *testing.T) {
	Reset()
	defer Reset()

	type UnregisteredStruct struct {
		ID int
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustValidateStruct should panic for unregistered type, but didn't")
		}
	}()

	MustValidateStruct[UnregisteredStruct]()
}

func TestGetRegisteredTypes(t *testing.T) {
	Reset()
	defer Reset()

	type Type1 struct {
		X int
	}

	type Type2 struct {
		Y string
	}

	// Initially empty
	types := GetRegisteredTypes()
	if len(types) != 0 {
		t.Errorf("GetRegisteredTypes() = %v, want empty slice", types)
	}

	// Register types
	Precompile[Type1]()
	Precompile[Type2]()

	types = GetRegisteredTypes()
	if len(types) != 2 {
		t.Errorf("GetRegisteredTypes() length = %d, want 2", len(types))
	}

	// Check both types are present
	found1, found2 := false, false
	for _, name := range types {
		if name == "Type1" {
			found1 = true
		}
		if name == "Type2" {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("GetRegisteredTypes() = %v, missing Type1 or Type2", types)
	}
}

func TestValidateField_InvalidPrimitive(t *testing.T) {
	// Create a field with wrong type marking
	field := &FieldInfo{
		Name:        "TestField",
		Type:        FieldTypePrimitive,
		ReflectType: stringType(),
	}

	err := validateField("TestType", field)
	if err == nil {
		t.Error("validateField() error = nil, want error for string marked as primitive")
	}
}

func TestValidateField_InvalidArray(t *testing.T) {
	// Array with no length
	field := &FieldInfo{
		Name:        "TestArray",
		Type:        FieldTypeArray,
		ReflectType: arrayType(),
		ArrayLen:    0, // Invalid
	}

	err := validateField("TestType", field)
	if err == nil {
		t.Error("validateField() error = nil, want error for array with invalid length")
	}
}

func TestValidateField_MissingReflectType(t *testing.T) {
	field := &FieldInfo{
		Name:        "TestField",
		Type:        FieldTypePrimitive,
		ReflectType: nil, // Missing
	}

	err := validateField("TestType", field)
	if err == nil {
		t.Error("validateField() error = nil, want error for missing reflect type")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("error should be ValidationError, got %T", err)
	}
}

// Helper functions for tests
func stringType() reflect.Type {
	return reflect.TypeOf("")
}

func arrayType() reflect.Type {
	return reflect.TypeOf([5]int{})
}
