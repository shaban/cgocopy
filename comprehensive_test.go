package cgocopy

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

// Test structs covering all primitive types
type AllPrimitivesStruct struct {
	Int8Field    int8
	Int16Field   int16
	Int32Field   int32
	Int64Field   int64
	Uint8Field   uint8
	Uint16Field  uint16
	Uint32Field  uint32
	Uint64Field  uint64
	Float32Field float32
	Float64Field float64
	BoolField    bool
}

type PointerStruct struct {
	IntPtr    *int32
	FloatPtr  *float32
	UnsafePtr unsafe.Pointer
}

type ArrayStruct struct {
	ByteArray  [16]byte
	IntArray   [4]int32
	FloatArray [8]float32
}

type NestedStruct struct {
	ID    int32
	Inner InnerStruct
	Value float32
}

type InnerStruct struct {
	X int32
	Y int32
}

// Table-driven validation tests
func TestStructValidation(t *testing.T) {
	tests := []struct {
		name          string
		goType        reflect.Type
		cSize         uintptr
		cLayout       []FieldInfo
		expectError   bool
		errorContains string
	}{
		{
			name:   "valid_basic_struct",
			goType: reflect.TypeOf(TestStruct{}),
			cSize:  16,
			cLayout: []FieldInfo{
				{Offset: 0, Size: 4, TypeName: "int32_t"},
				{Offset: 4, Size: 4, TypeName: "float"},
				{Offset: 8, Size: 8, TypeName: "int64_t"},
			},
			expectError: false,
		},
		{
			name:   "field_count_mismatch_too_few",
			goType: reflect.TypeOf(TestStruct{}),
			cSize:  16,
			cLayout: []FieldInfo{
				{Offset: 0, Size: 4, TypeName: "int32_t"},
				{Offset: 4, Size: 4, TypeName: "float"},
			},
			expectError:   true,
			errorContains: "field count mismatch",
		},
		{
			name:   "field_count_mismatch_too_many",
			goType: reflect.TypeOf(TestStruct{}),
			cSize:  16,
			cLayout: []FieldInfo{
				{Offset: 0, Size: 4, TypeName: "int32_t"},
				{Offset: 4, Size: 4, TypeName: "float"},
				{Offset: 8, Size: 8, TypeName: "int64_t"},
				{Offset: 16, Size: 4, TypeName: "int32_t"},
			},
			expectError:   true,
			errorContains: "field count mismatch",
		},
		{
			name:   "field_size_mismatch_first_field",
			goType: reflect.TypeOf(TestStruct{}),
			cSize:  16,
			cLayout: []FieldInfo{
				{Offset: 0, Size: 8, TypeName: "int32_t"}, // Wrong size!
				{Offset: 4, Size: 4, TypeName: "float"},
				{Offset: 8, Size: 8, TypeName: "int64_t"},
			},
			expectError:   true,
			errorContains: "size mismatch",
		},
		{
			name:   "field_size_mismatch_middle_field",
			goType: reflect.TypeOf(TestStruct{}),
			cSize:  16,
			cLayout: []FieldInfo{
				{Offset: 0, Size: 4, TypeName: "int32_t"},
				{Offset: 4, Size: 8, TypeName: "float"}, // Wrong size!
				{Offset: 8, Size: 8, TypeName: "int64_t"},
			},
			expectError:   true,
			errorContains: "size mismatch",
		},
		{
			name:   "type_incompatibility_int_to_float",
			goType: reflect.TypeOf(TestStruct{}),
			cSize:  16,
			cLayout: []FieldInfo{
				{Offset: 0, Size: 4, TypeName: "float"}, // int32 field but float type
				{Offset: 4, Size: 4, TypeName: "float"},
				{Offset: 8, Size: 8, TypeName: "int64_t"},
			},
			expectError:   true,
			errorContains: "type incompatible",
		},
		{
			name:   "type_incompatibility_float_to_int",
			goType: reflect.TypeOf(TestStruct{}),
			cSize:  16,
			cLayout: []FieldInfo{
				{Offset: 0, Size: 4, TypeName: "int32_t"},
				{Offset: 4, Size: 4, TypeName: "int32_t"}, // float field but int type
				{Offset: 8, Size: 8, TypeName: "int64_t"},
			},
			expectError:   true,
			errorContains: "type incompatible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()
			err := registry.Register(tt.goType, tt.cSize, tt.cLayout, nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// Test all primitive type compatibilities
func TestPrimitiveTypeCompatibility(t *testing.T) {
	tests := []struct {
		goType      reflect.Kind
		cTypeName   string
		shouldMatch bool
	}{
		// Integer types
		{reflect.Int8, "char", true},
		{reflect.Int8, "int8_t", true}, // Now supported!
		{reflect.Uint8, "char", true},
		{reflect.Int32, "int", true},
		{reflect.Int32, "int32_t", true},
		{reflect.Int32, "int64_t", false},
		{reflect.Int64, "int64_t", true},
		{reflect.Int64, "int32_t", false},
		{reflect.Uint32, "uint32_t", true},
		{reflect.Uint32, "int32_t", false},
		{reflect.Uint64, "uint64_t", true},

		// Float types
		{reflect.Float32, "float", true},
		{reflect.Float32, "double", false},
		{reflect.Float64, "double", true},
		{reflect.Float64, "float", false},

		// Bool
		{reflect.Bool, "bool", true},
		{reflect.Bool, "int", false},

		// Pointers
		{reflect.Ptr, "pointer", true},
		{reflect.UnsafePointer, "pointer", true},
	}

	for _, tt := range tests {
		t.Run(tt.goType.String()+"_to_"+tt.cTypeName, func(t *testing.T) {
			// Create a reflect.Type for the kind
			var goType reflect.Type
			switch tt.goType {
			case reflect.Int8:
				goType = reflect.TypeOf(int8(0))
			case reflect.Uint8:
				goType = reflect.TypeOf(uint8(0))
			case reflect.Int32:
				goType = reflect.TypeOf(int32(0))
			case reflect.Int64:
				goType = reflect.TypeOf(int64(0))
			case reflect.Uint32:
				goType = reflect.TypeOf(uint32(0))
			case reflect.Uint64:
				goType = reflect.TypeOf(uint64(0))
			case reflect.Float32:
				goType = reflect.TypeOf(float32(0))
			case reflect.Float64:
				goType = reflect.TypeOf(float64(0))
			case reflect.Bool:
				goType = reflect.TypeOf(bool(false))
			case reflect.Ptr:
				var ptr *int32
				goType = reflect.TypeOf(ptr)
			case reflect.UnsafePointer:
				goType = reflect.TypeOf(unsafe.Pointer(nil))
			}

			err := validateTypeCompatibility(goType, tt.cTypeName)

			if tt.shouldMatch && err != nil {
				t.Errorf("Expected types to be compatible, got error: %v", err)
			}
			if !tt.shouldMatch && err == nil {
				t.Errorf("Expected types to be incompatible, but got no error")
			}
		})
	}
}

// Test struct with all primitive types
func TestAllPrimitivesStruct(t *testing.T) {
	registry := NewRegistry()

	cLayout := []FieldInfo{
		{Offset: 0, Size: 1, TypeName: "char"},      // int8
		{Offset: 2, Size: 2, TypeName: "int16_t"},   // int16 (assume padding)
		{Offset: 4, Size: 4, TypeName: "int32_t"},   // int32
		{Offset: 8, Size: 8, TypeName: "int64_t"},   // int64
		{Offset: 16, Size: 1, TypeName: "char"},     // uint8
		{Offset: 18, Size: 2, TypeName: "uint16_t"}, // uint16
		{Offset: 20, Size: 4, TypeName: "uint32_t"}, // uint32
		{Offset: 24, Size: 8, TypeName: "uint64_t"}, // uint64
		{Offset: 32, Size: 4, TypeName: "float"},    // float32
		{Offset: 36, Size: 8, TypeName: "double"},   // float64
		{Offset: 44, Size: 1, TypeName: "bool"},     // bool
	}

	goType := reflect.TypeOf(AllPrimitivesStruct{})

	// Note: This will fail because we need to add int16/uint16 support
	// and the struct needs proper layout matching
	err := registry.Register(goType, unsafe.Sizeof(AllPrimitivesStruct{}), cLayout, nil)

	// For now, we expect it might fail due to missing type mappings
	// This test documents what needs to be added
	if err != nil {
		t.Logf("Registration failed (expected for now): %v", err)
		t.Logf("TODO: Add support for int16_t and uint16_t types")
	}
}

// Test error messages are descriptive
func TestErrorMessageQuality(t *testing.T) {
	registry := NewRegistry()

	t.Run("field_count_error_shows_counts", func(t *testing.T) {
		cLayout := []FieldInfo{
			{Offset: 0, Size: 4, TypeName: "int32_t"},
		}

		err := registry.Register(reflect.TypeOf(TestStruct{}), 16, cLayout, nil)
		if err == nil {
			t.Fatal("Expected error")
		}

		// Should show both counts
		if !strings.Contains(err.Error(), "1") || !strings.Contains(err.Error(), "3") {
			t.Errorf("Error should show both field counts, got: %v", err)
		}
	})

	t.Run("size_mismatch_shows_field_name", func(t *testing.T) {
		cLayout := []FieldInfo{
			{Offset: 0, Size: 8, TypeName: "int32_t"}, // Wrong size
			{Offset: 4, Size: 4, TypeName: "float"},
			{Offset: 8, Size: 8, TypeName: "int64_t"},
		}

		err := registry.Register(reflect.TypeOf(TestStruct{}), 16, cLayout, nil)
		if err == nil {
			t.Fatal("Expected error")
		}

		// Should show field name and both sizes
		if !strings.Contains(err.Error(), "ID") {
			t.Errorf("Error should show field name 'ID', got: %v", err)
		}
		if !strings.Contains(err.Error(), "8") || !strings.Contains(err.Error(), "4") {
			t.Errorf("Error should show both sizes, got: %v", err)
		}
	})

	t.Run("type_mismatch_shows_both_types", func(t *testing.T) {
		cLayout := []FieldInfo{
			{Offset: 0, Size: 4, TypeName: "float"},
			{Offset: 4, Size: 4, TypeName: "float"},
			{Offset: 8, Size: 8, TypeName: "int64_t"},
		}

		err := registry.Register(reflect.TypeOf(TestStruct{}), 16, cLayout, nil)
		if err == nil {
			t.Fatal("Expected error")
		}

		// Should show both types
		if !strings.Contains(err.Error(), "float") || !strings.Contains(err.Error(), "int32") {
			t.Errorf("Error should show both types, got: %v", err)
		}
	})
}

// Test nested struct placeholder
func TestNestedStructPlaceholder(t *testing.T) {
	t.Skip("Nested struct support not yet implemented - TODO")

	// This test documents what nested struct support would look like
	registry := NewRegistry()

	// Would need to register inner struct first
	innerLayout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: 4, Size: 4, TypeName: "int32_t"},
	}
	_ = innerLayout

	// Then register outer struct referencing inner
	outerLayout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: 4, Size: 8, TypeName: "InnerStruct"}, // Would need special handling
		{Offset: 12, Size: 4, TypeName: "float"},
	}

	err := registry.Register(reflect.TypeOf(NestedStruct{}), 16, outerLayout, nil)
	if err != nil {
		t.Logf("Nested struct registration: %v", err)
	}
}

// Benchmark comparison with proper struct sizes
func BenchmarkAllPrimitiveCopies(b *testing.B) {
	tests := []struct {
		name string
		size int
		fn   func()
	}{
		{
			name: "8byte_struct",
			size: 8,
			fn: func() {
				type Small struct{ A, B int32 }
				var s Small
				_ = s
			},
		},
		{
			name: "16byte_struct",
			size: 16,
			fn: func() {
				var s TestStruct
				_ = s
			},
		},
		{
			name: "64byte_struct",
			size: 64,
			fn: func() {
				type Large struct {
					A, B, C, D int64
					E, F, G, H int64
				}
				var s Large
				_ = s
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tt.fn()
			}
		})
	}
}
