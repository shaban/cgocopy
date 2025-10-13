package cgocopy2

import (
	"testing"
	"unsafe"
)

func TestFastCopy_Int(t *testing.T) {
	val := int(42)
	result := FastCopy[int](unsafe.Pointer(&val))
	if result != 42 {
		t.Errorf("FastCopy[int] = %d, want 42", result)
	}
}

func TestFastCopy_Int8(t *testing.T) {
	val := int8(-8)
	result := FastCopy[int8](unsafe.Pointer(&val))
	if result != -8 {
		t.Errorf("FastCopy[int8] = %d, want -8", result)
	}
}

func TestFastCopy_Int16(t *testing.T) {
	val := int16(1600)
	result := FastCopy[int16](unsafe.Pointer(&val))
	if result != 1600 {
		t.Errorf("FastCopy[int16] = %d, want 1600", result)
	}
}

func TestFastCopy_Int32(t *testing.T) {
	val := int32(320000)
	result := FastCopy[int32](unsafe.Pointer(&val))
	if result != 320000 {
		t.Errorf("FastCopy[int32] = %d, want 320000", result)
	}
}

func TestFastCopy_Int64(t *testing.T) {
	val := int64(9223372036854775807)
	result := FastCopy[int64](unsafe.Pointer(&val))
	if result != 9223372036854775807 {
		t.Errorf("FastCopy[int64] = %d, want 9223372036854775807", result)
	}
}

func TestFastCopy_Uint(t *testing.T) {
	val := uint(100)
	result := FastCopy[uint](unsafe.Pointer(&val))
	if result != 100 {
		t.Errorf("FastCopy[uint] = %d, want 100", result)
	}
}

func TestFastCopy_Uint8(t *testing.T) {
	val := uint8(255)
	result := FastCopy[uint8](unsafe.Pointer(&val))
	if result != 255 {
		t.Errorf("FastCopy[uint8] = %d, want 255", result)
	}
}

func TestFastCopy_Uint16(t *testing.T) {
	val := uint16(65535)
	result := FastCopy[uint16](unsafe.Pointer(&val))
	if result != 65535 {
		t.Errorf("FastCopy[uint16] = %d, want 65535", result)
	}
}

func TestFastCopy_Uint32(t *testing.T) {
	val := uint32(4294967295)
	result := FastCopy[uint32](unsafe.Pointer(&val))
	if result != 4294967295 {
		t.Errorf("FastCopy[uint32] = %d, want 4294967295", result)
	}
}

func TestFastCopy_Uint64(t *testing.T) {
	val := uint64(18446744073709551615)
	result := FastCopy[uint64](unsafe.Pointer(&val))
	if result != 18446744073709551615 {
		t.Errorf("FastCopy[uint64] = %d, want 18446744073709551615", result)
	}
}

func TestFastCopy_Float32(t *testing.T) {
	val := float32(3.14159)
	result := FastCopy[float32](unsafe.Pointer(&val))
	if result != 3.14159 {
		t.Errorf("FastCopy[float32] = %f, want 3.14159", result)
	}
}

func TestFastCopy_Float64(t *testing.T) {
	val := float64(2.718281828459045)
	result := FastCopy[float64](unsafe.Pointer(&val))
	if result != 2.718281828459045 {
		t.Errorf("FastCopy[float64] = %f, want 2.718281828459045", result)
	}
}

func TestFastCopy_BoolTrue(t *testing.T) {
	// Simulate C bool (uint8)
	val := uint8(1)
	result := FastCopy[bool](unsafe.Pointer(&val))
	if result != true {
		t.Errorf("FastCopy[bool] = %v, want true", result)
	}
}

func TestFastCopy_BoolFalse(t *testing.T) {
	// Simulate C bool (uint8)
	val := uint8(0)
	result := FastCopy[bool](unsafe.Pointer(&val))
	if result != false {
		t.Errorf("FastCopy[bool] = %v, want false", result)
	}
}

func TestFastCopy_NonPrimitive_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("FastCopy with struct should panic, but didn't")
		}
	}()

	type TestStruct struct {
		X int
	}
	val := TestStruct{X: 42}
	_ = FastCopy[TestStruct](unsafe.Pointer(&val))
}

func TestFastCopyInt(t *testing.T) {
	val := int(999)
	result := FastCopyInt(unsafe.Pointer(&val))
	if result != 999 {
		t.Errorf("FastCopyInt = %d, want 999", result)
	}
}

func TestFastCopyInt32(t *testing.T) {
	val := int32(-12345)
	result := FastCopyInt32(unsafe.Pointer(&val))
	if result != -12345 {
		t.Errorf("FastCopyInt32 = %d, want -12345", result)
	}
}

func TestFastCopyUint64(t *testing.T) {
	val := uint64(123456789)
	result := FastCopyUint64(unsafe.Pointer(&val))
	if result != 123456789 {
		t.Errorf("FastCopyUint64 = %d, want 123456789", result)
	}
}

func TestFastCopyFloat64(t *testing.T) {
	val := float64(99.99)
	result := FastCopyFloat64(unsafe.Pointer(&val))
	if result != 99.99 {
		t.Errorf("FastCopyFloat64 = %f, want 99.99", result)
	}
}

func TestFastCopyBool(t *testing.T) {
	// Test true
	valTrue := uint8(1)
	resultTrue := FastCopyBool(unsafe.Pointer(&valTrue))
	if resultTrue != true {
		t.Errorf("FastCopyBool(1) = %v, want true", resultTrue)
	}

	// Test false
	valFalse := uint8(0)
	resultFalse := FastCopyBool(unsafe.Pointer(&valFalse))
	if resultFalse != false {
		t.Errorf("FastCopyBool(0) = %v, want false", resultFalse)
	}
}

func TestCanFastCopy(t *testing.T) {
	tests := []struct {
		name string
		test func() bool
		want bool
	}{
		{"int", CanFastCopy[int], true},
		{"int32", CanFastCopy[int32], true},
		{"uint64", CanFastCopy[uint64], true},
		{"float64", CanFastCopy[float64], true},
		{"bool", CanFastCopy[bool], true},
		{"string", CanFastCopy[string], false},
		{"struct", func() bool {
			type S struct{ X int }
			return CanFastCopy[S]()
		}, false},
		{"slice", CanFastCopy[[]int], false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.test()
			if got != tt.want {
				t.Errorf("CanFastCopy = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustFastCopy(t *testing.T) {
	val := int32(777)
	result := MustFastCopy[int32](unsafe.Pointer(&val))
	if result != 777 {
		t.Errorf("MustFastCopy = %d, want 777", result)
	}
}

// Benchmarks comparing FastCopy vs Copy

func BenchmarkFastCopy_Int32(b *testing.B) {
	val := int32(42)
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FastCopy[int32](ptr)
	}
}

func BenchmarkCopy_Int32(b *testing.B) {
	Reset()
	type SingleInt32 struct {
		Value int32
	}
	Precompile[SingleInt32]()

	val := SingleInt32{Value: 42}
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Copy[SingleInt32](ptr)
	}
}

func BenchmarkFastCopy_Int64(b *testing.B) {
	val := int64(9223372036854775807)
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FastCopy[int64](ptr)
	}
}

func BenchmarkCopy_Int64(b *testing.B) {
	Reset()
	type SingleInt64 struct {
		Value int64
	}
	Precompile[SingleInt64]()

	val := SingleInt64{Value: 9223372036854775807}
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Copy[SingleInt64](ptr)
	}
}

func BenchmarkFastCopy_Float64(b *testing.B) {
	val := float64(3.14159)
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FastCopy[float64](ptr)
	}
}

func BenchmarkCopy_Float64(b *testing.B) {
	Reset()
	type SingleFloat64 struct {
		Value float64
	}
	Precompile[SingleFloat64]()

	val := SingleFloat64{Value: 3.14159}
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Copy[SingleFloat64](ptr)
	}
}

func BenchmarkFastCopyInt32_NonGeneric(b *testing.B) {
	val := int32(42)
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FastCopyInt32(ptr)
	}
}

func BenchmarkFastCopyInt64_NonGeneric(b *testing.B) {
	val := int64(9223372036854775807)
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FastCopyInt64(ptr)
	}
}

func BenchmarkFastCopyFloat64_NonGeneric(b *testing.B) {
	val := float64(3.14159)
	ptr := unsafe.Pointer(&val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FastCopyFloat64(ptr)
	}
}
