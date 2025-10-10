package cgocopy

import (
	"testing"
	"unsafe"
)

type GoSimpleStruct struct {
	ID        int32
	Value     float32
	Timestamp int64
}

// DirectCopyInline - inline candidate (simple, small function)
//
//go:inline
func DirectCopyInline(dst *GoSimpleStruct, src unsafe.Pointer) {
	*dst = *(*GoSimpleStruct)(src)
}

// DirectCopyGeneric - generic version for any struct
func DirectCopyGeneric[T any](dst *T, src unsafe.Pointer) {
	*dst = *(*T)(src)
}

// Manual copy without function (baseline)
func BenchmarkDirectCopy_Inline(b *testing.B) {
	cStruct := CreateTestStruct()
	defer FreeTestStruct(cStruct)

	var goStruct GoSimpleStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Using inline function
		DirectCopyInline(&goStruct, cStruct)
	}
}

func BenchmarkDirectCopy_Generic(b *testing.B) {
	cStruct := CreateTestStruct()
	defer FreeTestStruct(cStruct)

	var goStruct GoSimpleStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Using generic function
		DirectCopyGeneric(&goStruct, cStruct)
	}
}

func BenchmarkDirectCopy_Manual(b *testing.B) {
	cStruct := CreateTestStruct()
	defer FreeTestStruct(cStruct)

	var goStruct GoSimpleStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Direct pointer cast (no function)
		goStruct = *(*GoSimpleStruct)(cStruct)
	}

	// Use goStruct to prevent optimization
	_ = goStruct
}

// Test to verify they all work correctly
func TestDirectCopyMethods(t *testing.T) {
	cStruct := CreateTestStruct()
	defer FreeTestStruct(cStruct)

	// Test inline version
	var goStruct1 GoSimpleStruct
	DirectCopyInline(&goStruct1, cStruct)
	if goStruct1.ID != 42 || goStruct1.Value != 3.14 || goStruct1.Timestamp != 1234567890 {
		t.Errorf("Inline copy failed: %+v", goStruct1)
	}

	// Test generic version
	var goStruct2 GoSimpleStruct
	DirectCopyGeneric(&goStruct2, cStruct)
	if goStruct2.ID != 42 || goStruct2.Value != 3.14 || goStruct2.Timestamp != 1234567890 {
		t.Errorf("Generic copy failed: %+v", goStruct2)
	}

	// Test manual version

	goStruct3 := *(*GoSimpleStruct)(cStruct)
	if goStruct3.ID != 42 || goStruct3.Value != 3.14 || goStruct3.Timestamp != 1234567890 {
		t.Errorf("Manual copy failed: %+v", goStruct3)
	}

	t.Logf("✅ All three methods produce identical results")
}
