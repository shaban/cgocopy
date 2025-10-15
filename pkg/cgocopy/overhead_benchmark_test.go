package cgocopy2

import (
	"testing"
	"unsafe"
)

// ============================================================================
// Microbenchmark 1: String Conversion Overhead
// ============================================================================

func BenchmarkStringConversion_UnsafeSlice(b *testing.B) {
	// Simulate C string
	testStr := "Hello, World! This is a test string.\x00"
	cBytes := []byte(testStr)
	ptr := unsafe.Pointer(&cBytes[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Manual conversion using unsafe (faster method)
		length := 0
		for length < 64 {
			if *(*byte)(unsafe.Add(ptr, length)) == 0 {
				break
			}
			length++
		}
		bytes := unsafe.Slice((*byte)(ptr), length)
		_ = string(bytes)
	}
}

// ============================================================================
// Microbenchmark 2: Dynamic Array Creation
// ============================================================================

func BenchmarkDynamicArray_UnsafeSlice_Small(b *testing.B) {
	// Simulate C array
	arr := make([]int32, 10)
	ptr := unsafe.Pointer(&arr[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unsafe.Slice((*int32)(ptr), 10)
	}
}

func BenchmarkDynamicArray_UnsafeSlice_Medium(b *testing.B) {
	arr := make([]int32, 100)
	ptr := unsafe.Pointer(&arr[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unsafe.Slice((*int32)(ptr), 100)
	}
}

func BenchmarkDynamicArray_UnsafeSlice_Large(b *testing.B) {
	arr := make([]int32, 1000)
	ptr := unsafe.Pointer(&arr[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unsafe.Slice((*int32)(ptr), 1000)
	}
}

// ============================================================================
// Microbenchmark 3: Type Switching Overhead
// ============================================================================

type fieldType int

const (
	typeInt32 fieldType = iota
	typeFloat64
	typeString
	typeArray
	typeStruct
)

func BenchmarkTypeSwitch_FourCases(b *testing.B) {
	types := []fieldType{typeInt32, typeFloat64, typeString, typeArray}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := types[i%4]
		var dummy int
		switch t {
		case typeInt32:
			dummy = 1
		case typeFloat64:
			dummy = 2
		case typeString:
			dummy = 3
		case typeArray:
			dummy = 4
		}
		_ = dummy
	}
}

func BenchmarkTypeSwitch_StringBased(b *testing.B) {
	types := []string{"int32", "float64", "string", "array"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := types[i%4]
		var dummy int
		switch t {
		case "int32":
			dummy = 1
		case "float64":
			dummy = 2
		case "string":
			dummy = 3
		case "array":
			dummy = 4
		}
		_ = dummy
	}
}

// ============================================================================
// Microbenchmark 4: Memory Access Pattern
// ============================================================================

type testStruct struct {
	id      int32
	count   int32
	dataPtr unsafe.Pointer
	name    [64]byte
}

func BenchmarkMemoryAccess_ReadCountField(b *testing.B) {
	s := testStruct{
		id:    42,
		count: 100,
		name:  [64]byte{},
	}
	ptr := unsafe.Pointer(&s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate reading count field at known offset
		countOffset := unsafe.Offsetof(s.count)
		_ = *(*int32)(unsafe.Add(ptr, countOffset))
	}
}

func BenchmarkMemoryAccess_ReadMultipleFields(b *testing.B) {
	s := testStruct{
		id:    42,
		count: 100,
		name:  [64]byte{},
	}
	ptr := unsafe.Pointer(&s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate reading multiple fields (like we'd do in real code)
		idOffset := unsafe.Offsetof(s.id)
		countOffset := unsafe.Offsetof(s.count)
		dataPtrOffset := unsafe.Offsetof(s.dataPtr)

		_ = *(*int32)(unsafe.Add(ptr, idOffset))
		_ = *(*int32)(unsafe.Add(ptr, countOffset))
		_ = *(*unsafe.Pointer)(unsafe.Add(ptr, dataPtrOffset))
	}
}

// ============================================================================
// Microbenchmark 5: Combined Overhead (Realistic scenario)
// ============================================================================

func BenchmarkRealisticOverhead_WithString(b *testing.B) {
	// Simulate struct with: 2 primitives + 1 string + 1 dynamic array
	testStr := "Hello, World! This is a test string.\x00"
	strBytes := []byte(testStr)
	cStr := unsafe.Pointer(&strBytes[0])

	arr := make([]int32, 50)
	cArr := unsafe.Pointer(&arr[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Base copy (simulated with dummy values)
		var id int32 = 42
		var count int32 = 50

		// String conversion (manual, like we'd do)
		length := 0
		for length < 64 {
			if *(*byte)(unsafe.Add(cStr, length)) == 0 {
				break
			}
			length++
		}
		bytes := unsafe.Slice((*byte)(cStr), length)
		_ = string(bytes)

		// Dynamic array creation
		_ = unsafe.Slice((*int32)(cArr), count)

		// Type switch overhead (simulated)
		_ = id + count
	}
}

func BenchmarkRealisticOverhead_NoString(b *testing.B) {
	// Simulate struct with: 4 primitives + 1 dynamic array (no string)
	arr := make([]int32, 50)
	cArr := unsafe.Pointer(&arr[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Base copy (simulated with dummy values)
		var id int32 = 42
		var count int32 = 50

		// Dynamic array creation
		_ = unsafe.Slice((*int32)(cArr), count)

		// Type switch overhead (simulated)
		_ = id + count
	}
}
