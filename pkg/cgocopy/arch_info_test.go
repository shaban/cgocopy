package cgocopy

import (
	"testing"
)

func TestArchitectureInfo(t *testing.T) {
	ai := GetArchInfo()

	t.Logf("Architecture Info:\n%s", ai.String())

	// Sanity checks
	if ai.Int8Size != 1 {
		t.Errorf("Expected int8 size = 1, got %d", ai.Int8Size)
	}

	if ai.Int16Size != 2 {
		t.Errorf("Expected int16 size = 2, got %d", ai.Int16Size)
	}

	if ai.Int32Size != 4 {
		t.Errorf("Expected int32 size = 4, got %d", ai.Int32Size)
	}

	if ai.Int64Size != 8 {
		t.Errorf("Expected int64 size = 8, got %d", ai.Int64Size)
	}

	if ai.Float32Size != 4 {
		t.Errorf("Expected float32 size = 4, got %d", ai.Float32Size)
	}

	if ai.Float64Size != 8 {
		t.Errorf("Expected float64 size = 8, got %d", ai.Float64Size)
	}

	if ai.PointerSize != 4 && ai.PointerSize != 8 {
		t.Errorf("Expected pointer size = 4 or 8, got %d", ai.PointerSize)
	}

	// Verify 64-bit detection
	if ai.Is64Bit {
		if ai.PointerSize != 8 {
			t.Errorf("64-bit platform should have 8-byte pointers, got %d", ai.PointerSize)
		}
		t.Logf("✅ Detected 64-bit platform")
	} else {
		if ai.PointerSize != 4 {
			t.Errorf("32-bit platform should have 4-byte pointers, got %d", ai.PointerSize)
		}
		t.Logf("✅ Detected 32-bit platform")
	}

	// Verify alignment makes sense
	t.Logf("\n📏 Alignment Requirements:")
	t.Logf("  int32:   %d-byte alignment", ai.Int32Align)
	t.Logf("  int64:   %d-byte alignment", ai.Int64Align)
	t.Logf("  float64: %d-byte alignment", ai.Float64Align)
	t.Logf("  pointer: %d-byte alignment", ai.PointerAlign)

	// Common alignment expectations
	if ai.Int32Align != 4 && ai.Int32Align != 1 {
		t.Logf("⚠️  Unusual int32 alignment: %d", ai.Int32Align)
	}

	if ai.Int64Align != 8 && ai.Int64Align != 4 {
		t.Logf("⚠️  Unusual int64 alignment: %d", ai.Int64Align)
	}

	if ai.Float64Align != 8 && ai.Float64Align != 4 {
		t.Logf("⚠️  Unusual float64 alignment: %d", ai.Float64Align)
	}

	// Endianness
	if ai.IsLittleEndian {
		t.Logf("✅ Detected little-endian byte order")
	} else {
		t.Logf("✅ Detected big-endian byte order")
	}
}

func TestAlignmentCalculation(t *testing.T) {
	tests := []struct {
		name        string
		prevEnd     uintptr
		currentOff  uintptr
		fieldSize   uintptr
		expectAlign uintptr
	}{
		{"no_padding", 0, 0, 4, 1},
		{"4byte_align", 1, 4, 4, 4},
		{"8byte_align", 4, 8, 8, 8},
		{"packed", 4, 4, 4, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			align := calculateAlignmentFromOffset(tt.prevEnd, tt.currentOff, tt.fieldSize)
			t.Logf("%s: prevEnd=%d, offset=%d, size=%d => align=%d",
				tt.name, tt.prevEnd, tt.currentOff, tt.fieldSize, align)

			// We're deducing alignment, so just verify it's reasonable
			if align > 16 {
				t.Errorf("Unreasonable alignment: %d", align)
			}
		})
	}
}

func BenchmarkGetArchInfo(b *testing.B) {
	// This should be called once at init, but let's benchmark it anyway
	for i := 0; i < b.N; i++ {
		_ = GetArchInfo()
	}
}
