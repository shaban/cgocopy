package cgocopy

import (
	"encoding/json"
	"sync"
	"testing"
	"unsafe"
)

type V2BenchPrimitive struct {
	ID    uint32  `json:"id"`
	Value float64 `json:"value"`
}

type V2BenchSlice struct {
	Count    uint32   `json:"count"`
	Readings []uint16 `json:"readings"`
}

type V2BenchMember struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

type V2BenchTeam struct {
	Count   uint32          `json:"count"`
	Members []V2BenchMember `json:"members"`
}

var (
	benchmarkRegistryOnce sync.Once
	benchmarkRegistry     *Registry
	benchmarkInitErr      error
)

func getBenchmarkRegistry(tb testing.TB) *Registry {
	benchmarkRegistryOnce.Do(func() {
		Reset()
		if err := RegisterStruct[V2BenchPrimitive](nil); err != nil {
			benchmarkInitErr = err
			return
		}
		if err := RegisterStruct[V2BenchSlice](nil); err != nil {
			benchmarkInitErr = err
			return
		}
		if err := RegisterStruct[V2BenchTeam](DefaultCStringConverter); err != nil {
			benchmarkInitErr = err
			return
		}
		Finalize()
		benchmarkRegistry = defaultRegistry
		Reset()
	})

	if benchmarkInitErr != nil {
		tb.Fatalf("benchmark setup failed: %v", benchmarkInitErr)
	}
	return benchmarkRegistry
}

func BenchmarkCopy_PrimitiveFastPath(b *testing.B) {
	registry := getBenchmarkRegistry(b)
	cPtr, cleanup := newBenchPrimitivePointer(42, 99.5)
	if cPtr == nil {
		b.Fatal("primitive pointer is nil")
	}
	b.Cleanup(cleanup)

	var out V2BenchPrimitive
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&out, cPtr); err != nil {
			b.Fatalf("Copy failed: %v", err)
		}
	}
}

func BenchmarkFastCopier_Primitive(b *testing.B) {
	registry := getBenchmarkRegistry(b)
	fastCopy, err := FastCopierFor[V2BenchPrimitive](registry)
	if err != nil {
		b.Fatalf("FastCopierFor failed: %v", err)
	}

	cPtr, cleanup := newBenchPrimitivePointer(42, 99.5)
	if cPtr == nil {
		b.Fatal("primitive pointer is nil")
	}
	b.Cleanup(cleanup)

	var out V2BenchPrimitive
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := fastCopy(&out, cPtr); err != nil {
			b.Fatalf("fast copy failed: %v", err)
		}
	}
}

func BenchmarkCopy_PrimitiveSlice(b *testing.B) {
	registry := getBenchmarkRegistry(b)
	values := make([]uint16, 64)
	for i := range values {
		values[i] = uint16(i * 3)
	}
	cPtr, cleanup := newBenchSlicePointer(values)
	if cPtr == nil {
		b.Fatal("slice pointer is nil")
	}
	b.Cleanup(cleanup)

	var out V2BenchSlice
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&out, cPtr); err != nil {
			b.Fatalf("Copy failed: %v", err)
		}
	}
}

func BenchmarkCopy_NestedStructSlice(b *testing.B) {
	registry := getBenchmarkRegistry(b)
	ids := []uint32{1001, 1002, 1003, 1004}
	names := []string{"Ada", "Alan", "Grace", "Linus"}
	cPtr, cleanup := newBenchTeamPointer(ids, names)
	if cPtr == nil {
		b.Fatal("team pointer is nil")
	}
	b.Cleanup(cleanup)

	var out V2BenchTeam
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&out, cPtr); err != nil {
			b.Fatalf("Copy failed: %v", err)
		}
	}
}

func BenchmarkCopy_NilSafety(b *testing.B) {
	registry := getBenchmarkRegistry(b)
	var out V2BenchPrimitive
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&out, unsafe.Pointer(uintptr(0))); err == nil {
			b.Fatalf("expected error for nil source")
		}
	}
}

func BenchmarkCopyNoReflection_PrimitiveSlice(b *testing.B) {
	registry := getBenchmarkRegistry(b)
	values := make([]uint16, 64)
	for i := range values {
		values[i] = uint16(i * 5)
	}
	cPtr, cleanup := newBenchSlicePointer(values)
	if cPtr == nil {
		b.Fatal("slice pointer is nil")
	}
	b.Cleanup(cleanup)

	var out V2BenchSlice
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.CopyNoReflection(&out, cPtr); err != nil {
			b.Fatalf("CopyNoReflection failed: %v", err)
		}
	}
}

func BenchmarkJSONRoundTrip_Team(b *testing.B) {
	ids := []uint32{1001, 1002, 1003, 1004}
	names := []string{"Ada", "Alan", "Grace", "Linus"}

	var out V2BenchTeam
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonStr, cleanup := newBenchTeamJSON(ids, names)
		if jsonStr == "" {
			b.Fatal("json baseline returned empty string")
		}
		if err := json.Unmarshal([]byte(jsonStr), &out); err != nil {
			cleanup()
			b.Fatalf("json unmarshal failed: %v", err)
		}
		cleanup()
	}
}
