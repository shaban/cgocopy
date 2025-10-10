package cgocopy

import (
	"strings"
	"testing"
)

/*
#include <stddef.h>
#include <stdint.h>

// ============================================================================
// TEST STRUCTS FOR OFFSET PREDICTION EXPERIMENT
// ============================================================================

// Test Case 1: Simple struct (no padding expected)
typedef struct {
    uint32_t id;
    uint32_t value;
} SimpleStruct;

// Test Case 2: Struct with padding (small type followed by large)
typedef struct {
    uint8_t  flag;
    uint32_t id;
    uint64_t timestamp;
} PaddedStruct;

// Test Case 3: Struct with pointer (string field)
typedef struct {
    uint32_t id;
    char*    name;
    float    value;
} PointerStruct;

// Test Case 4: Mixed types
typedef struct {
    uint8_t   byte1;
    uint16_t  short1;
    uint32_t  int1;
    uint64_t  long1;
    float     float1;
    double    double1;
    char*     ptr1;
} MixedStruct;

// Test Case 5: Multiple small types
typedef struct {
    uint8_t  a;
    uint8_t  b;
    uint8_t  c;
    uint8_t  d;
    uint32_t e;
} MultiSmallStruct;

// Test Case 6: Reverse order (large to small)
typedef struct {
    uint64_t timestamp;
    uint32_t id;
    uint8_t  flag;
} ReverseStruct;

// Test Case 7: All pointers
typedef struct {
    char*  name;
    void*  data;
    char*  description;
} AllPointersStruct;

// Test Case 8: Complex nesting simulation
typedef struct {
    uint8_t  header;
    uint32_t id;
    char*    name;
    double   value;
    uint8_t  footer;
} ComplexStruct;

// ============================================================================
// OFFSET HELPER FUNCTIONS
// ============================================================================

// SimpleStruct offsets
size_t simpleIdOffset() { return offsetof(SimpleStruct, id); }
size_t simpleValueOffset() { return offsetof(SimpleStruct, value); }
size_t simpleSize() { return sizeof(SimpleStruct); }

// PaddedStruct offsets
size_t paddedFlagOffset() { return offsetof(PaddedStruct, flag); }
size_t paddedIdOffset() { return offsetof(PaddedStruct, id); }
size_t paddedTimestampOffset() { return offsetof(PaddedStruct, timestamp); }
size_t paddedSize() { return sizeof(PaddedStruct); }

// PointerStruct offsets
size_t pointerIdOffset() { return offsetof(PointerStruct, id); }
size_t pointerNameOffset() { return offsetof(PointerStruct, name); }
size_t pointerValueOffset() { return offsetof(PointerStruct, value); }
size_t pointerSize() { return sizeof(PointerStruct); }

// MixedStruct offsets
size_t mixedByte1Offset() { return offsetof(MixedStruct, byte1); }
size_t mixedShort1Offset() { return offsetof(MixedStruct, short1); }
size_t mixedInt1Offset() { return offsetof(MixedStruct, int1); }
size_t mixedLong1Offset() { return offsetof(MixedStruct, long1); }
size_t mixedFloat1Offset() { return offsetof(MixedStruct, float1); }
size_t mixedDouble1Offset() { return offsetof(MixedStruct, double1); }
size_t mixedPtr1Offset() { return offsetof(MixedStruct, ptr1); }
size_t mixedSize() { return sizeof(MixedStruct); }

// MultiSmallStruct offsets
size_t multiAOffset() { return offsetof(MultiSmallStruct, a); }
size_t multiBOffset() { return offsetof(MultiSmallStruct, b); }
size_t multiCOffset() { return offsetof(MultiSmallStruct, c); }
size_t multiDOffset() { return offsetof(MultiSmallStruct, d); }
size_t multiEOffset() { return offsetof(MultiSmallStruct, e); }
size_t multiSize() { return sizeof(MultiSmallStruct); }

// ReverseStruct offsets
size_t reverseTimestampOffset() { return offsetof(ReverseStruct, timestamp); }
size_t reverseIdOffset() { return offsetof(ReverseStruct, id); }
size_t reverseFlagOffset() { return offsetof(ReverseStruct, flag); }
size_t reverseSize() { return sizeof(ReverseStruct); }

// AllPointersStruct offsets
size_t allPtrsNameOffset() { return offsetof(AllPointersStruct, name); }
size_t allPtrsDataOffset() { return offsetof(AllPointersStruct, data); }
size_t allPtrsDescOffset() { return offsetof(AllPointersStruct, description); }
size_t allPtrsSize() { return sizeof(AllPointersStruct); }

// ComplexStruct offsets
size_t complexHeaderOffset() { return offsetof(ComplexStruct, header); }
size_t complexIdOffset() { return offsetof(ComplexStruct, id); }
size_t complexNameOffset() { return offsetof(ComplexStruct, name); }
size_t complexValueOffset() { return offsetof(ComplexStruct, value); }
size_t complexFooterOffset() { return offsetof(ComplexStruct, footer); }
size_t complexSize() { return sizeof(ComplexStruct); }

*/
import "C"

// OffsetPredictor predicts C struct field offsets based on arch info
type OffsetPredictor struct {
	archInfo *ArchInfo
}

// NewOffsetPredictor creates a predictor using architecture info
func NewOffsetPredictor(archInfo *ArchInfo) *OffsetPredictor {
	return &OffsetPredictor{archInfo: archInfo}
}

// alignOffset rounds up offset to the next alignment boundary
func (p *OffsetPredictor) alignOffset(offset, align uintptr) uintptr {
	if align == 0 || align == 1 {
		return offset
	}
	return ((offset + align - 1) / align) * align
}

// getTypeInfo returns size and alignment for a C type name
func (p *OffsetPredictor) getTypeInfo(typeName string) (size uintptr, align uintptr) {
	switch typeName {
	case "int8_t", "uint8_t", "char":
		return p.archInfo.Int8Size, p.archInfo.Int8Align
	case "int16_t", "uint16_t", "short":
		return p.archInfo.Int16Size, p.archInfo.Int16Align
	case "int32_t", "uint32_t", "int":
		return p.archInfo.Int32Size, p.archInfo.Int32Align
	case "int64_t", "uint64_t", "long", "long long":
		return p.archInfo.Int64Size, p.archInfo.Int64Align
	case "float":
		return p.archInfo.Float32Size, p.archInfo.Float32Align
	case "double":
		return p.archInfo.Float64Size, p.archInfo.Float64Align
	case "char*", "void*", "pointer":
		return p.archInfo.PointerSize, p.archInfo.PointerAlign
	default:
		return 0, 0
	}
}

// PredictOffsets calculates struct field offsets from type names
func (p *OffsetPredictor) PredictOffsets(typeNames []string) []uintptr {
	offsets := make([]uintptr, len(typeNames))
	currentOffset := uintptr(0)

	for i, typeName := range typeNames {
		size, align := p.getTypeInfo(typeName)
		if size == 0 {
			// Unknown type - can't predict
			offsets[i] = 0
			continue
		}

		// Align current offset to field's alignment requirement
		currentOffset = p.alignOffset(currentOffset, align)
		offsets[i] = currentOffset

		// Move to next field
		currentOffset += size
	}

	return offsets
}

// TestOffsetPrediction tests if we can predict C struct offsets
func TestOffsetPrediction(t *testing.T) {
	archInfo := GetArchInfo()
	predictor := NewOffsetPredictor(&archInfo)

	t.Logf("Architecture Info:")
	t.Logf("  int8: size=%d, align=%d", archInfo.Int8Size, archInfo.Int8Align)
	t.Logf("  int16: size=%d, align=%d", archInfo.Int16Size, archInfo.Int16Align)
	t.Logf("  int32: size=%d, align=%d", archInfo.Int32Size, archInfo.Int32Align)
	t.Logf("  int64: size=%d, align=%d", archInfo.Int64Size, archInfo.Int64Align)
	t.Logf("  float32: size=%d, align=%d", archInfo.Float32Size, archInfo.Float32Align)
	t.Logf("  float64: size=%d, align=%d", archInfo.Float64Size, archInfo.Float64Align)
	t.Logf("  pointer: size=%d, align=%d", archInfo.PointerSize, archInfo.PointerAlign)
	t.Logf("")

	tests := []struct {
		name      string
		typeNames []string
		actual    []uintptr
		totalSize uintptr
	}{
		{
			name:      "SimpleStruct",
			typeNames: []string{"uint32_t", "uint32_t"},
			actual: []uintptr{
				uintptr(C.simpleIdOffset()),
				uintptr(C.simpleValueOffset()),
			},
			totalSize: uintptr(C.simpleSize()),
		},
		{
			name:      "PaddedStruct",
			typeNames: []string{"uint8_t", "uint32_t", "uint64_t"},
			actual: []uintptr{
				uintptr(C.paddedFlagOffset()),
				uintptr(C.paddedIdOffset()),
				uintptr(C.paddedTimestampOffset()),
			},
			totalSize: uintptr(C.paddedSize()),
		},
		{
			name:      "PointerStruct",
			typeNames: []string{"uint32_t", "char*", "float"},
			actual: []uintptr{
				uintptr(C.pointerIdOffset()),
				uintptr(C.pointerNameOffset()),
				uintptr(C.pointerValueOffset()),
			},
			totalSize: uintptr(C.pointerSize()),
		},
		{
			name: "MixedStruct",
			typeNames: []string{
				"uint8_t", "uint16_t", "uint32_t", "uint64_t",
				"float", "double", "char*",
			},
			actual: []uintptr{
				uintptr(C.mixedByte1Offset()),
				uintptr(C.mixedShort1Offset()),
				uintptr(C.mixedInt1Offset()),
				uintptr(C.mixedLong1Offset()),
				uintptr(C.mixedFloat1Offset()),
				uintptr(C.mixedDouble1Offset()),
				uintptr(C.mixedPtr1Offset()),
			},
			totalSize: uintptr(C.mixedSize()),
		},
		{
			name:      "MultiSmallStruct",
			typeNames: []string{"uint8_t", "uint8_t", "uint8_t", "uint8_t", "uint32_t"},
			actual: []uintptr{
				uintptr(C.multiAOffset()),
				uintptr(C.multiBOffset()),
				uintptr(C.multiCOffset()),
				uintptr(C.multiDOffset()),
				uintptr(C.multiEOffset()),
			},
			totalSize: uintptr(C.multiSize()),
		},
		{
			name:      "ReverseStruct",
			typeNames: []string{"uint64_t", "uint32_t", "uint8_t"},
			actual: []uintptr{
				uintptr(C.reverseTimestampOffset()),
				uintptr(C.reverseIdOffset()),
				uintptr(C.reverseFlagOffset()),
			},
			totalSize: uintptr(C.reverseSize()),
		},
		{
			name:      "AllPointersStruct",
			typeNames: []string{"char*", "void*", "char*"},
			actual: []uintptr{
				uintptr(C.allPtrsNameOffset()),
				uintptr(C.allPtrsDataOffset()),
				uintptr(C.allPtrsDescOffset()),
			},
			totalSize: uintptr(C.allPtrsSize()),
		},
		{
			name:      "ComplexStruct",
			typeNames: []string{"uint8_t", "uint32_t", "char*", "double", "uint8_t"},
			actual: []uintptr{
				uintptr(C.complexHeaderOffset()),
				uintptr(C.complexIdOffset()),
				uintptr(C.complexNameOffset()),
				uintptr(C.complexValueOffset()),
				uintptr(C.complexFooterOffset()),
			},
			totalSize: uintptr(C.complexSize()),
		},
	}

	successCount := 0
	totalFields := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicted := predictor.PredictOffsets(tt.typeNames)

			t.Logf("Struct: %s (total size: %d bytes)", tt.name, tt.totalSize)
			allMatch := true

			for i, typeName := range tt.typeNames {
				match := predicted[i] == tt.actual[i]
				status := "✅"
				if !match {
					status = "❌"
					allMatch = false
				}

				t.Logf("  Field %d (%s): predicted=%d, actual=%d %s",
					i, typeName, predicted[i], tt.actual[i], status)

				totalFields++
				if match {
					successCount++
				}
			}

			if allMatch {
				t.Logf("  🎉 ALL OFFSETS MATCH!")
			} else {
				t.Logf("  ⚠️  Some offsets don't match")
			}
		})
	}

	t.Logf("\n%s", strings.Repeat("=", 60))
	t.Logf("EXPERIMENT RESULTS:")
	t.Logf("  Total fields tested: %d", totalFields)
	t.Logf("  Successful predictions: %d", successCount)
	t.Logf("  Success rate: %.1f%%", float64(successCount)/float64(totalFields)*100)
	t.Logf("%s", strings.Repeat("=", 60))

	if successCount == totalFields {
		t.Logf("🎉 PERFECT SCORE! Offset prediction is 100%% accurate!")
		t.Logf("✅ We can eliminate offsetof() helpers for standard structs!")
	} else if float64(successCount)/float64(totalFields) >= 0.95 {
		t.Logf("✅ Excellent! >95%% accuracy - production ready with documented edge cases")
	} else if float64(successCount)/float64(totalFields) >= 0.80 {
		t.Logf("⚠️  Good but not great - keep as convenience feature with explicit fallback")
	} else {
		t.Logf("❌ Too unreliable - stick with offsetof() helpers")
	}
}

// BenchmarkOffsetPrediction benchmarks the prediction algorithm
func BenchmarkOffsetPrediction(b *testing.B) {
	archInfo := GetArchInfo()
	predictor := NewOffsetPredictor(&archInfo)
	typeNames := []string{"uint8_t", "uint32_t", "char*", "double", "uint8_t"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = predictor.PredictOffsets(typeNames)
	}
}
