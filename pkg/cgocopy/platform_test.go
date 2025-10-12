package cgocopy

import "testing"

// TestPrimitiveSizes verifies that C primitive type sizes are captured at init time
// This is critical for cross-platform compatibility
func TestPrimitiveSizes(t *testing.T) {
	tests := []struct {
		name     string
		size     uintptr
		expected uintptr
	}{
		{"char", CCharSize, 1},
		{"int8_t", CInt8Size, 1},
		{"uint8_t", CUInt8Size, 1},
		{"int16_t", CInt16Size, 2},
		{"uint16_t", CUInt16Size, 2},
		{"int32_t", CInt32Size, 4},
		{"uint32_t", CUInt32Size, 4},
		{"int64_t", CInt64Size, 8},
		{"uint64_t", CUInt64Size, 8},
		{"float", CFloatSize, 4},
		{"double", CDoubleSize, 8},
	}

	t.Log("C primitive type sizes on this platform:")
	for _, tt := range tests {
		t.Logf("  %-12s: %d bytes", tt.name, tt.size)
		if tt.size != tt.expected {
			t.Errorf("%s: expected %d bytes, got %d bytes", tt.name, tt.expected, tt.size)
		}
	}

	// Platform-dependent sizes (may vary)
	t.Logf("\nPlatform-dependent C types:")
	t.Logf("  short       : %d bytes", CShortSize)
	t.Logf("  int         : %d bytes", CIntSize)
	t.Logf("  long        : %d bytes", CLongSize)
	t.Logf("  long long   : %d bytes", CLongLongSize)
	t.Logf("  void*       : %d bytes", CPointerSize)
	t.Logf("  size_t      : %d bytes", CSizeTSize)

	// Sanity checks
	if CPointerSize != 8 && CPointerSize != 4 {
		t.Errorf("Unexpected pointer size: %d (expected 4 or 8)", CPointerSize)
	}
}

// TestStructLayoutCapture verifies struct layouts are captured correctly
func TestStructLayoutCapture(t *testing.T) {
	t.Log("C struct layouts on this platform:")

	t.Logf("\nTestStruct (size=%d):", testStructSize)
	t.Logf("  id        : offset=%d", testStructIdOffset)
	t.Logf("  value     : offset=%d", testStructValueOffset)
	t.Logf("  timestamp : offset=%d", testStructTimestampOffset)

	t.Logf("\nInnerStruct (size=%d):", innerStructSize)
	t.Logf("  x         : offset=%d", innerStructXOffset)
	t.Logf("  y         : offset=%d", innerStructYOffset)

	t.Logf("\nMiddleStruct (size=%d):", middleStructSize)
	t.Logf("  id        : offset=%d", middleStructIdOffset)
	t.Logf("  inner     : offset=%d", middleStructInnerOffset)
	t.Logf("  value     : offset=%d", middleStructValueOffset)

	t.Logf("\nOuterStruct (size=%d):", outerStructSize)
	t.Logf("  outerID   : offset=%d", outerStructOuterIDOffset)
	t.Logf("  middle    : offset=%d", outerStructMiddleOffset)
	t.Logf("  timestamp : offset=%d", outerStructTimestampOffset)

	// Sanity check: timestamp should be at offset > 20 due to padding
	if outerStructTimestampOffset <= 20 {
		t.Errorf("Expected timestamp offset > 20 (due to padding), got %d", outerStructTimestampOffset)
	}
}

// TestPrimitiveSizeUsage demonstrates using captured primitive sizes
func TestPrimitiveSizeUsage(t *testing.T) {
	// Example: You could use these sizes to validate your FieldInfo structs
	// or to build layouts programmatically

	t.Log("Example: Building a layout using captured primitive sizes")

	// Instead of hardcoding sizes, use the captured values:
	layout := []FieldInfo{
		{Offset: 0, Size: CInt32Size, TypeName: "int32_t"},
		{Offset: CInt32Size, Size: CFloatSize, TypeName: "float"},
		{Offset: CInt32Size + CFloatSize, Size: CInt64Size, TypeName: "int64_t"},
	}

	t.Logf("Layout built with platform-specific sizes:")
	for i, field := range layout {
		t.Logf("  Field %d: offset=%d, size=%d (%s)", i, field.Offset, field.Size, field.TypeName)
	}

	// This would be the same as our TestStruct layout
	expectedTotalSize := CInt32Size + CFloatSize + CInt64Size
	t.Logf("Expected size: %d, Actual TestStruct size: %d", expectedTotalSize, testStructSize)

	// Note: They might not match exactly due to padding/alignment!
	// That's why we need offsetof() for struct fields
	if testStructSize != expectedTotalSize {
		t.Logf("⚠️  Sizes differ due to struct padding/alignment rules")
	}
}
