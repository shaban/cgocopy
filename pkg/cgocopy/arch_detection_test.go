package cgocopy

import "testing"

func TestActualAlignmentDetection(t *testing.T) {
	ai := GetArchInfo()

	t.Log("Raw offset information from C test struct:")

	// We need to see the actual raw data to understand alignment
	// Let's create a simpler test that shows what's happening

	t.Logf("  Expected natural alignment (typical):")
	t.Logf("    int32:  4-byte alignment")
	t.Logf("    int64:  8-byte alignment")
	t.Logf("    double: 8-byte alignment")
	t.Logf("    pointer: %d-byte alignment", ai.PointerSize)

	t.Logf("\n  Detected alignment:")
	t.Logf("    int32:  %d-byte alignment", ai.Int32Align)
	t.Logf("    int64:  %d-byte alignment", ai.Int64Align)
	t.Logf("    double: %d-byte alignment", ai.Float64Align)
	t.Logf("    pointer: %d-byte alignment", ai.PointerAlign)

	if ai.Int32Align == 1 && ai.Int64Align == 1 {
		t.Log("\n⚠️  All alignments are 1 - struct might be getting packed!")
		t.Log("   This could mean:")
		t.Log("   1. The test struct fields happen to align naturally")
		t.Log("   2. Compiler is auto-packing")
		t.Log("   3. Our alignment detection logic needs improvement")
	}

	// The key insight: if things are naturally aligned, we might not see padding!
	// Let's add a test that FORCES padding
}
