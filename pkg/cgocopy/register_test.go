package cgocopy

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"unsafe"
)

type V2Sample struct {
	ID   uint32
	Name string
}

type MissingMetadata struct {
	Value int32
}

type V2ArrayHolder struct {
	Values []uint16
}

type V2NestedInner struct {
	Code  uint32
	Label string
}

type V2NestedOuter struct {
	Inner V2NestedInner
	Data  V2ArrayHolder
}

func resetDefaultRegistry(tb testing.TB) {
	tb.Helper()
	Reset()
	tb.Cleanup(Reset)
}

func TestRegisterStruct_Success(t *testing.T) {
	resetDefaultRegistry(t)

	if err := RegisterStruct[V2Sample](nil); err != nil {
		t.Fatalf("RegisterStruct returned error: %v", err)
	}

	mapping, ok := defaultRegistry.GetMapping(reflect.TypeOf(V2Sample{}))
	if !ok {
		t.Fatalf("expected mapping to be registered")
	}

	if mapping.StringConverter == nil {
		t.Fatalf("expected string converter to be configured")
	}
}

func TestRegisterStruct_RepeatedCallNoop(t *testing.T) {
	resetDefaultRegistry(t)

	if err := RegisterStruct[V2Sample](nil); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	if err := RegisterStruct[V2Sample](nil); err != nil {
		t.Fatalf("second registration should noop, got error: %v", err)
	}
}

func TestRegisterStruct_AfterFinalize(t *testing.T) {
	resetDefaultRegistry(t)
	Finalize()

	err := RegisterStruct[V2Sample](nil)
	if !errors.Is(err, ErrRegistryFinalized) {
		t.Fatalf("expected ErrRegistryFinalized, got %v", err)
	}
}

func TestRegisterStruct_MetadataMissing(t *testing.T) {
	resetDefaultRegistry(t)

	err := RegisterStruct[MissingMetadata](nil)
	if !errors.Is(err, ErrMetadataMissing) {
		t.Fatalf("expected ErrMetadataMissing, got %v", err)
	}
}

func TestRegisterStruct_InvalidType(t *testing.T) {
	resetDefaultRegistry(t)

	err := RegisterStruct[int](nil)
	if !errors.Is(err, ErrNotAStructType) {
		t.Fatalf("expected ErrNotAStructType, got %v", err)
	}
}

func TestRegisterStruct_AnonymousStruct(t *testing.T) {
	type anon struct {
		Value int
	}

	resetDefaultRegistry(t)

	err := RegisterStruct[struct {
		Value int
	}](nil)
	if !errors.Is(err, ErrAnonymousStruct) {
		t.Fatalf("expected ErrAnonymousStruct, got %v", err)
	}
}

func TestCopy_BeforeFinalize(t *testing.T) {
	resetDefaultRegistry(t)
	if err := RegisterStruct[V2Sample](nil); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	ptr, cleanup := newSamplePointer(42, "Before Finalize")
	if ptr == nil {
		t.Fatalf("expected non-nil sample pointer")
	}
	defer cleanup()

	var out V2Sample
	err := Copy(&out, ptr)
	if !errors.Is(err, ErrRegistryNotFinalized) {
		t.Fatalf("expected ErrRegistryNotFinalized, got %v", err)
	}
}

func TestCopy_AfterFinalizeCopies(t *testing.T) {
	resetDefaultRegistry(t)
	if err := RegisterStruct[V2Sample](nil); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	Finalize()

	ptr, cleanup := newSamplePointer(77, "Test User")
	if ptr == nil {
		t.Fatalf("expected non-nil sample pointer")
	}
	defer cleanup()

	var out V2Sample
	if err := Copy(&out, ptr); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}

	if out.ID != 77 {
		t.Fatalf("expected ID 77, got %d", out.ID)
	}
	if out.Name != "Test User" {
		t.Fatalf("expected Name 'Test User', got %q", out.Name)
	}

	mapping, ok := defaultRegistry.GetMapping(reflect.TypeOf(V2Sample{}))
	if !ok {
		t.Fatalf("expected mapping to be cached")
	}
	if mapping.StringConverter == nil {
		t.Fatalf("expected string converter to be configured in mapping")
	}
}

func TestCopy_NilDestination(t *testing.T) {
	resetDefaultRegistry(t)
	if err := RegisterStruct[V2Sample](nil); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	Finalize()

	ptr, cleanup := newSamplePointer(1, "Sample")
	if ptr != nil {
		defer cleanup()
	}

	err := Copy[V2Sample](nil, ptr)
	if !errors.Is(err, ErrNilDestination) {
		t.Fatalf("expected ErrNilDestination, got %v", err)
	}
}

func TestCopy_NilSourcePointer(t *testing.T) {
	resetDefaultRegistry(t)
	Finalize()

	var out V2Sample
	err := Copy(&out, nil)
	if !errors.Is(err, ErrNilSourcePointer) {
		t.Fatalf("expected ErrNilSourcePointer, got %v", err)
	}
}

func TestCopy_TypeNotRegistered(t *testing.T) {
	resetDefaultRegistry(t)
	Finalize()

	var out MissingMetadata
	dummy := unsafe.Pointer(&out)

	err := Copy(&out, dummy)
	if !errors.Is(err, ErrStructNotRegistered) {
		t.Fatalf("expected ErrStructNotRegistered, got %v", err)
	}
}

func TestCopy_DestinationNotStructPointer(t *testing.T) {
	resetDefaultRegistry(t)
	Finalize()

	var out int
	dummy := unsafe.Pointer(&out)

	err := Copy(&out, dummy)
	if !errors.Is(err, ErrDestinationNotStructPointer) {
		t.Fatalf("expected ErrDestinationNotStructPointer, got %v", err)
	}
}

func TestCopy_NestedStructsAndArray(t *testing.T) {
	resetDefaultRegistry(t)
	if err := RegisterStruct[V2NestedOuter](nil); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	Finalize()

	values := [4]uint16{11, 22, 33, 44}
	ptr, cleanup := newNestedSamplePointer(123, "Nested Label", values)
	if ptr == nil {
		t.Fatalf("expected non-nil nested sample pointer")
	}
	defer cleanup()

	var out V2NestedOuter
	if err := Copy(&out, ptr); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}

	if out.Inner.Code != 123 {
		t.Fatalf("expected inner code 123, got %d", out.Inner.Code)
	}
	if out.Inner.Label != "Nested Label" {
		t.Fatalf("expected inner label 'Nested Label', got %q", out.Inner.Label)
	}
	expected := []uint16{11, 22, 33, 44}
	if !slices.Equal(out.Data.Values, expected) {
		t.Fatalf("expected values %v, got %v", expected, out.Data.Values)
	}
}
