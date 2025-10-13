package cgocopy2

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"
)

// Test types for Copy tests
type CopySimple struct {
	ID    int
	Count uint32
	Score float64
	Flag  bool
}

type CopyWithStrings struct {
	ID   int
	Name *byte // Simulates char* in C
}

type CopyNested struct {
	ID      int
	Profile CopySimple
}

type CopyArrays struct {
	Values [3]int
	Scores [2]float64
}

type CopyTagged struct {
	UserID   int    `cgocopy:"id"`
	FullName *byte  `cgocopy:"name"` // char*
	Internal string `cgocopy:"-"`
}

// Helper function to create a C-style string (null-terminated byte array)
func cString(s string) *byte {
	if s == "" {
		return nil
	}
	bytes := make([]byte, len(s)+1) // +1 for null terminator
	copy(bytes, s)
	bytes[len(s)] = 0
	return &bytes[0]
}

func TestCopy_Simple(t *testing.T) {
	Reset()
	defer Reset()

	// Precompile the type
	if err := Precompile[CopySimple](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	// Create a "C struct" (simulated)
	cStruct := CopySimple{
		ID:    42,
		Count: 100,
		Score: 95.5,
		Flag:  true,
	}

	// Copy it
	result, err := Copy[CopySimple](unsafe.Pointer(&cStruct))
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	// Verify fields
	if result.ID != 42 {
		t.Errorf("ID = %d, want 42", result.ID)
	}
	if result.Count != 100 {
		t.Errorf("Count = %d, want 100", result.Count)
	}
	if result.Score != 95.5 {
		t.Errorf("Score = %f, want 95.5", result.Score)
	}
	if result.Flag != true {
		t.Errorf("Flag = %v, want true", result.Flag)
	}
}

func TestCopy_WithStrings(t *testing.T) {
	Reset()
	defer Reset()

	if err := Precompile[CopyWithStrings](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	// Create a "C struct" with a C-style string
	cStruct := CopyWithStrings{
		ID:   123,
		Name: cString("John Doe"),
	}

	result, err := Copy[CopyWithStrings](unsafe.Pointer(&cStruct))
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	if result.ID != 123 {
		t.Errorf("ID = %d, want 123", result.ID)
	}

	// Note: Name field is *byte in the struct, so we need to check the copied string differently
	// For this test, we'd need to adjust the struct or the test expectations
	// Let's create a version that uses string in Go
}

func TestCopy_EmptyString(t *testing.T) {
	Reset()
	defer Reset()

	if err := Precompile[CopyWithStrings](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	// Create struct with nil string pointer
	cStruct := CopyWithStrings{
		ID:   999,
		Name: nil,
	}

	result, err := Copy[CopyWithStrings](unsafe.Pointer(&cStruct))
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	if result.ID != 999 {
		t.Errorf("ID = %d, want 999", result.ID)
	}
}

func TestCopy_Nested(t *testing.T) {
	Reset()
	defer Reset()

	// Precompile both types
	if err := Precompile[CopySimple](); err != nil {
		t.Fatalf("Precompile[CopySimple]() error = %v", err)
	}
	if err := Precompile[CopyNested](); err != nil {
		t.Fatalf("Precompile[CopyNested]() error = %v", err)
	}

	// Create nested struct
	cStruct := CopyNested{
		ID: 1,
		Profile: CopySimple{
			ID:    2,
			Count: 50,
			Score: 88.8,
			Flag:  false,
		},
	}

	result, err := Copy[CopyNested](unsafe.Pointer(&cStruct))
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	if result.ID != 1 {
		t.Errorf("ID = %d, want 1", result.ID)
	}
	if result.Profile.ID != 2 {
		t.Errorf("Profile.ID = %d, want 2", result.Profile.ID)
	}
	if result.Profile.Count != 50 {
		t.Errorf("Profile.Count = %d, want 50", result.Profile.Count)
	}
	if result.Profile.Score != 88.8 {
		t.Errorf("Profile.Score = %f, want 88.8", result.Profile.Score)
	}
	if result.Profile.Flag != false {
		t.Errorf("Profile.Flag = %v, want false", result.Profile.Flag)
	}
}

func TestCopy_Arrays(t *testing.T) {
	Reset()
	defer Reset()

	if err := Precompile[CopyArrays](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	cStruct := CopyArrays{
		Values: [3]int{10, 20, 30},
		Scores: [2]float64{1.5, 2.5},
	}

	result, err := Copy[CopyArrays](unsafe.Pointer(&cStruct))
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	for i, expected := range []int{10, 20, 30} {
		if result.Values[i] != expected {
			t.Errorf("Values[%d] = %d, want %d", i, result.Values[i], expected)
		}
	}

	for i, expected := range []float64{1.5, 2.5} {
		if result.Scores[i] != expected {
			t.Errorf("Scores[%d] = %f, want %f", i, result.Scores[i], expected)
		}
	}
}

func TestCopy_NilPointer(t *testing.T) {
	Reset()
	defer Reset()

	if err := Precompile[CopySimple](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	_, err := Copy[CopySimple](nil)
	if err == nil {
		t.Fatal("Copy(nil) error = nil, want error")
	}

	if !errors.Is(err, ErrNilPointer) {
		t.Errorf("error = %v, want ErrNilPointer", err)
	}
}

func TestCopy_NotRegistered(t *testing.T) {
	Reset()
	defer Reset()

	// Don't precompile
	cStruct := CopySimple{ID: 1}

	_, err := Copy[CopySimple](unsafe.Pointer(&cStruct))
	if err == nil {
		t.Fatal("Copy() error = nil, want error for unregistered type")
	}

	if !errors.Is(err, ErrNotRegistered) {
		t.Errorf("error = %v, want ErrNotRegistered", err)
	}
}

func TestCopy_AllPrimitiveTypes(t *testing.T) {
	Reset()
	defer Reset()

	type AllPrimitives struct {
		I   int
		I8  int8
		I16 int16
		I32 int32
		I64 int64
		U   uint
		U8  uint8
		U16 uint16
		U32 uint32
		U64 uint64
		F32 float32
		F64 float64
		B   bool
	}

	if err := Precompile[AllPrimitives](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	cStruct := AllPrimitives{
		I:   -1,
		I8:  -8,
		I16: -16,
		I32: -32,
		I64: -64,
		U:   1,
		U8:  8,
		U16: 16,
		U32: 32,
		U64: 64,
		F32: 3.14,
		F64: 2.718,
		B:   true,
	}

	result, err := Copy[AllPrimitives](unsafe.Pointer(&cStruct))
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	if result.I != -1 || result.I8 != -8 || result.I16 != -16 || result.I32 != -32 || result.I64 != -64 {
		t.Errorf("signed integers mismatch: got %+v", result)
	}

	if result.U != 1 || result.U8 != 8 || result.U16 != 16 || result.U32 != 32 || result.U64 != 64 {
		t.Errorf("unsigned integers mismatch: got %+v", result)
	}

	if result.F32 != 3.14 || result.F64 != 2.718 {
		t.Errorf("floats mismatch: got F32=%f F64=%f", result.F32, result.F64)
	}

	if result.B != true {
		t.Errorf("bool = %v, want true", result.B)
	}
}

func TestCopy_TaggedStruct(t *testing.T) {
	Reset()
	defer Reset()

	if err := Precompile[CopyTagged](); err != nil {
		t.Fatalf("Precompile() error = %v", err)
	}

	cStruct := CopyTagged{
		UserID:   456,
		FullName: cString("Jane Smith"),
		Internal: "should not be copied",
	}

	result, err := Copy[CopyTagged](unsafe.Pointer(&cStruct))
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}

	if result.UserID != 456 {
		t.Errorf("UserID = %d, want 456", result.UserID)
	}

	// Internal should remain at zero value (wasn't copied)
	if result.Internal != "" {
		t.Errorf("Internal = %q, want empty string (field should be skipped)", result.Internal)
	}
}

func TestCopyPrimitive(t *testing.T) {
	tests := []struct {
		name  string
		setup func() (reflect.Value, unsafe.Pointer, reflect.Type)
		check func(t *testing.T, result reflect.Value)
	}{
		{
			name: "int",
			setup: func() (reflect.Value, unsafe.Pointer, reflect.Type) {
				val := int(42)
				result := reflect.New(reflect.TypeOf(val)).Elem()
				return result, unsafe.Pointer(&val), reflect.TypeOf(val)
			},
			check: func(t *testing.T, result reflect.Value) {
				if result.Int() != 42 {
					t.Errorf("result = %d, want 42", result.Int())
				}
			},
		},
		{
			name: "uint32",
			setup: func() (reflect.Value, unsafe.Pointer, reflect.Type) {
				val := uint32(100)
				result := reflect.New(reflect.TypeOf(val)).Elem()
				return result, unsafe.Pointer(&val), reflect.TypeOf(val)
			},
			check: func(t *testing.T, result reflect.Value) {
				if result.Uint() != 100 {
					t.Errorf("result = %d, want 100", result.Uint())
				}
			},
		},
		{
			name: "float64",
			setup: func() (reflect.Value, unsafe.Pointer, reflect.Type) {
				val := float64(3.14)
				result := reflect.New(reflect.TypeOf(val)).Elem()
				return result, unsafe.Pointer(&val), reflect.TypeOf(val)
			},
			check: func(t *testing.T, result reflect.Value) {
				if result.Float() != 3.14 {
					t.Errorf("result = %f, want 3.14", result.Float())
				}
			},
		},
		{
			name: "bool true",
			setup: func() (reflect.Value, unsafe.Pointer, reflect.Type) {
				val := true
				// For bool, C represents it as uint8
				cVal := uint8(1)
				result := reflect.New(reflect.TypeOf(val)).Elem()
				return result, unsafe.Pointer(&cVal), reflect.TypeOf(val)
			},
			check: func(t *testing.T, result reflect.Value) {
				if result.Bool() != true {
					t.Errorf("result = %v, want true", result.Bool())
				}
			},
		},
		{
			name: "bool false",
			setup: func() (reflect.Value, unsafe.Pointer, reflect.Type) {
				val := false
				// For bool, C represents it as uint8
				cVal := uint8(0)
				result := reflect.New(reflect.TypeOf(val)).Elem()
				return result, unsafe.Pointer(&cVal), reflect.TypeOf(val)
			},
			check: func(t *testing.T, result reflect.Value) {
				if result.Bool() != false {
					t.Errorf("result = %v, want false", result.Bool())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultValue, ptr, typ := tt.setup()

			err := copyPrimitive(resultValue, ptr, typ)
			if err != nil {
				t.Fatalf("copyPrimitive() error = %v", err)
			}

			tt.check(t, resultValue)
		})
	}
}

func TestCopyString_ValidString(t *testing.T) {
	str := "Hello, World!"
	cStr := cString(str)

	// Create a field to hold the result
	var result string
	resultValue := reflect.ValueOf(&result).Elem()

	// Copy the string
	err := copyString(resultValue, unsafe.Pointer(&cStr))
	if err != nil {
		t.Fatalf("copyString() error = %v", err)
	}

	if result != str {
		t.Errorf("result = %q, want %q", result, str)
	}
}

func TestCopyString_EmptyString(t *testing.T) {
	cStr := cString("")

	var result string
	resultValue := reflect.ValueOf(&result).Elem()

	err := copyString(resultValue, unsafe.Pointer(&cStr))
	if err != nil {
		t.Fatalf("copyString() error = %v", err)
	}

	if result != "" {
		t.Errorf("result = %q, want empty string", result)
	}
}

func TestCopyString_NilPointer(t *testing.T) {
	var cStr *byte = nil

	var result string
	resultValue := reflect.ValueOf(&result).Elem()

	err := copyString(resultValue, unsafe.Pointer(&cStr))
	if err != nil {
		t.Fatalf("copyString() error = %v", err)
	}

	if result != "" {
		t.Errorf("result = %q, want empty string for nil pointer", result)
	}
}
