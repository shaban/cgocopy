//go:build disable_no_reflection_tests

package cgocopy

import (
	"reflect"
	"testing"
)

func TestCopyNoReflection_PrimitiveStruct(t *testing.T) {
	resetDefaultRegistry(t)
	if err := RegisterStruct[V2Sample](nil); err != nil {
		t.Fatalf("register V2Sample: %v", err)
	}
	Finalize()

	cPtr, cleanup := newSamplePointer(42, "Ada")
	if cPtr == nil {
		t.Fatal("newSamplePointer returned nil")
	}
	t.Cleanup(cleanup)

	var withReflection, withoutReflection V2Sample
	if err := Copy(&withReflection, cPtr); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if err := CopyNoReflection(&withoutReflection, cPtr); err != nil {
		t.Fatalf("CopyNoReflection: %v", err)
	}

	if !reflect.DeepEqual(withReflection, withoutReflection) {
		t.Fatalf("mismatch between copy implementations: %+v vs %+v", withReflection, withoutReflection)
	}
}

func TestCopyNoReflection_PrimitiveSlice(t *testing.T) {
	resetDefaultRegistry(t)
	if err := RegisterStruct[V2BenchSlice](nil); err != nil {
		t.Fatalf("register V2BenchSlice: %v", err)
	}
	Finalize()

	values := make([]uint16, 16)
	for i := range values {
		values[i] = uint16(i * 7)
	}

	cPtr, cleanup := newBenchSlicePointer(values)
	if cPtr == nil {
		t.Fatal("newBenchSlicePointer returned nil")
	}
	t.Cleanup(cleanup)

	var withReflection, withoutReflection V2BenchSlice
	if err := Copy(&withReflection, cPtr); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if err := CopyNoReflection(&withoutReflection, cPtr); err != nil {
		t.Fatalf("CopyNoReflection: %v", err)
	}

	if !reflect.DeepEqual(withReflection, withoutReflection) {
		t.Fatalf("mismatch between copy implementations: %+v vs %+v", withReflection, withoutReflection)
	}
}

func TestCopyNoReflection_NestedArrays(t *testing.T) {
	t.Skip("no-reflection nested array support under development")

	resetDefaultRegistry(t)
	if err := RegisterStruct[V2BenchTeam](DefaultCStringConverter); err != nil {
		t.Fatalf("register V2BenchTeam: %v", err)
	}
	Finalize()

	ids := []uint32{101, 102, 103, 104}
	names := []string{"Ada", "Alan", "Grace", "Linus"}
	cPtr, cleanup := newBenchTeamPointer(ids, names)
	if cPtr == nil {
		t.Fatal("newBenchTeamPointer returned nil")
	}
	t.Cleanup(cleanup)

	var withReflection, withoutReflection V2BenchTeam
	if err := Copy(&withReflection, cPtr); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if err := CopyNoReflection(&withoutReflection, cPtr); err != nil {
		t.Fatalf("CopyNoReflection: %v", err)
	}

	if !reflect.DeepEqual(withReflection, withoutReflection) {
		t.Fatalf("mismatch between copy implementations: %+v vs %+v", withReflection, withoutReflection)
	}
}
