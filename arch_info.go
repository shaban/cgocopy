package cgocopy

/*
#include "native/arch_info.c"
*/
import "C"
import "fmt"

// ArchInfo holds compile-time architecture information from C
type ArchInfo struct {
	// Primitive sizes
	Int8Size    uintptr
	Int16Size   uintptr
	Int32Size   uintptr
	Int64Size   uintptr
	Uint8Size   uintptr
	Uint16Size  uintptr
	Uint32Size  uintptr
	Uint64Size  uintptr
	Float32Size uintptr
	Float64Size uintptr
	PointerSize uintptr
	SizeTSize   uintptr

	// Natural alignment requirements (deduced from offsets)
	Int8Align    uintptr
	Int16Align   uintptr
	Int32Align   uintptr
	Int64Align   uintptr
	Uint8Align   uintptr
	Uint16Align  uintptr
	Uint32Align  uintptr
	Uint64Align  uintptr
	Float32Align uintptr
	Float64Align uintptr
	PointerAlign uintptr
	SizeTAlign   uintptr

	// Platform info
	Is64Bit        bool
	IsLittleEndian bool
}

// Global architecture info captured at init time
var archInfo ArchInfo

func init() {
	// Call C function to get architecture info
	cInfo := C.getArchitectureInfo()

	// Populate Go struct
	archInfo = ArchInfo{
		// Sizes
		Int8Size:    uintptr(cInfo.int8_size),
		Int16Size:   uintptr(cInfo.int16_size),
		Int32Size:   uintptr(cInfo.int32_size),
		Int64Size:   uintptr(cInfo.int64_size),
		Uint8Size:   uintptr(cInfo.uint8_size),
		Uint16Size:  uintptr(cInfo.uint16_size),
		Uint32Size:  uintptr(cInfo.uint32_size),
		Uint64Size:  uintptr(cInfo.uint64_size),
		Float32Size: uintptr(cInfo.float_size),
		Float64Size: uintptr(cInfo.double_size),
		PointerSize: uintptr(cInfo.pointer_size),
		SizeTSize:   uintptr(cInfo.sizet_size),

		// Platform
		Is64Bit:        cInfo.is_64bit == 1,
		IsLittleEndian: cInfo.is_little_endian == 1,
	}

	// Calculate alignment requirements from offsets
	// Our test struct is designed to force padding: i8, i32, i64, i16, f64, i8_2, ptr, ...
	archInfo.Int8Align = calculateAlignmentFromOffset(0, uintptr(cInfo.int8_offset), archInfo.Int8Size)

	archInfo.Int32Align = calculateAlignmentFromOffset(
		uintptr(cInfo.int8_offset)+archInfo.Int8Size,
		uintptr(cInfo.int32_offset),
		archInfo.Int32Size,
	)

	// For 64-bit types, alignment equals size on most platforms
	archInfo.Int64Align = archInfo.Int64Size // uint64_t aligns to 8 bytes

	// For primitive types, alignment typically equals size (power of 2 rule)
	// The test struct may not reveal this if fields happen to be naturally aligned
	archInfo.Int16Align = archInfo.Int16Size // uint16_t aligns to 2 bytes

	archInfo.Float64Align = calculateAlignmentFromOffset(
		uintptr(cInfo.int16_offset)+archInfo.Int16Size,
		uintptr(cInfo.double_offset),
		archInfo.Float64Size,
	)

	// uint8 comes after double (i8_2 in C struct - but we don't track it separately)
	archInfo.Uint8Align = calculateAlignmentFromOffset(
		uintptr(cInfo.double_offset)+archInfo.Float64Size,
		uintptr(cInfo.uint8_offset),
		archInfo.Uint8Size,
	)

	archInfo.PointerAlign = calculateAlignmentFromOffset(
		uintptr(cInfo.uint8_offset)+archInfo.Uint8Size,
		uintptr(cInfo.pointer_offset),
		archInfo.PointerSize,
	)

	// Remaining fields - use size-equals-alignment rule for primitives
	archInfo.Uint16Align = archInfo.Uint16Size   // uint16_t aligns to 2 bytes
	archInfo.Uint32Align = archInfo.Uint32Size   // uint32_t aligns to 4 bytes
	archInfo.Uint64Align = archInfo.Uint64Size   // uint64_t aligns to 8 bytes
	archInfo.Float32Align = archInfo.Float32Size // float aligns to 4 bytes
}

// calculateAlignmentFromOffset deduces alignment requirement from padding
func calculateAlignmentFromOffset(prevEnd, currentOffset, fieldSize uintptr) uintptr {
	if currentOffset == prevEnd {
		// No padding = alignment of 1 or field is self-aligned
		return 1
	}

	padding := currentOffset - prevEnd

	// The alignment is typically the field size or a power of 2
	// Common alignments: 1, 2, 4, 8, 16
	if fieldSize <= padding {
		return fieldSize
	}

	// Find the actual alignment (should divide evenly into offset)
	for align := uintptr(16); align >= 1; align /= 2 {
		if currentOffset%align == 0 {
			return align
		}
	}

	return 1
}

// GetArchInfo returns the captured architecture information
func GetArchInfo() ArchInfo {
	return archInfo
}

// String returns a human-readable description of the architecture
func (ai ArchInfo) String() string {
	result := "Architecture Information:\n"
	result += fmt.Sprintf("  Platform: %d-bit, ", map[bool]int{false: 32, true: 64}[ai.Is64Bit])
	result += fmt.Sprintf("Endian: %s\n", map[bool]string{false: "Big", true: "Little"}[ai.IsLittleEndian])
	result += "\n  Primitive Sizes:\n"
	result += fmt.Sprintf("    int8:    %d bytes, align: %d\n", ai.Int8Size, ai.Int8Align)
	result += fmt.Sprintf("    int16:   %d bytes, align: %d\n", ai.Int16Size, ai.Int16Align)
	result += fmt.Sprintf("    int32:   %d bytes, align: %d\n", ai.Int32Size, ai.Int32Align)
	result += fmt.Sprintf("    int64:   %d bytes, align: %d\n", ai.Int64Size, ai.Int64Align)
	result += fmt.Sprintf("    uint8:   %d bytes, align: %d\n", ai.Uint8Size, ai.Uint8Align)
	result += fmt.Sprintf("    uint16:  %d bytes, align: %d\n", ai.Uint16Size, ai.Uint16Align)
	result += fmt.Sprintf("    uint32:  %d bytes, align: %d\n", ai.Uint32Size, ai.Uint32Align)
	result += fmt.Sprintf("    uint64:  %d bytes, align: %d\n", ai.Uint64Size, ai.Uint64Align)
	result += fmt.Sprintf("    float32: %d bytes, align: %d\n", ai.Float32Size, ai.Float32Align)
	result += fmt.Sprintf("    float64: %d bytes, align: %d\n", ai.Float64Size, ai.Float64Align)
	result += fmt.Sprintf("    pointer: %d bytes, align: %d\n", ai.PointerSize, ai.PointerAlign)
	result += fmt.Sprintf("    size_t:  %d bytes, align: %d\n", ai.SizeTSize, ai.SizeTAlign)
	return result
}
