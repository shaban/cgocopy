package cgocopy2

import (
	"reflect"
	"unsafe"
)

// FastCopy performs zero-allocation copying for primitive types.
// This function bypasses reflection entirely and does a direct memory copy,
// making it significantly faster than Copy[T]() for simple primitive values.
//
// FastCopy only works with primitive types:
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - bool
//
// For struct types, use Copy[T]() instead.
//
// Example:
//
//	// Copying a primitive from C
//	var cInt C.int = 42
//	goInt := cgocopy2.FastCopy[int32](unsafe.Pointer(&cInt))
//
//	// Copying a bool
//	var cBool C.bool = 1
//	goBool := cgocopy2.FastCopy[bool](unsafe.Pointer(&cBool))
func FastCopy[T any](cPtr unsafe.Pointer) T {
	var zero T
	typ := reflect.TypeOf(zero)

	// Validate it's a primitive type
	if !isPrimitiveKind(typ.Kind()) {
		panic("FastCopy only works with primitive types (int, uint, float, bool)")
	}

	// Perform direct memory copy based on size
	switch typ.Size() {
	case 1:
		return *(*T)(cPtr)
	case 2:
		return *(*T)(cPtr)
	case 4:
		return *(*T)(cPtr)
	case 8:
		return *(*T)(cPtr)
	default:
		panic("unsupported primitive size")
	}
}

// FastCopyInt copies an int from C memory (direct memory access).
func FastCopyInt(cPtr unsafe.Pointer) int {
	return *(*int)(cPtr)
}

// FastCopyInt8 copies an int8 from C memory.
func FastCopyInt8(cPtr unsafe.Pointer) int8 {
	return *(*int8)(cPtr)
}

// FastCopyInt16 copies an int16 from C memory.
func FastCopyInt16(cPtr unsafe.Pointer) int16 {
	return *(*int16)(cPtr)
}

// FastCopyInt32 copies an int32 from C memory.
func FastCopyInt32(cPtr unsafe.Pointer) int32 {
	return *(*int32)(cPtr)
}

// FastCopyInt64 copies an int64 from C memory.
func FastCopyInt64(cPtr unsafe.Pointer) int64 {
	return *(*int64)(cPtr)
}

// FastCopyUint copies a uint from C memory.
func FastCopyUint(cPtr unsafe.Pointer) uint {
	return *(*uint)(cPtr)
}

// FastCopyUint8 copies a uint8 from C memory.
func FastCopyUint8(cPtr unsafe.Pointer) uint8 {
	return *(*uint8)(cPtr)
}

// FastCopyUint16 copies a uint16 from C memory.
func FastCopyUint16(cPtr unsafe.Pointer) uint16 {
	return *(*uint16)(cPtr)
}

// FastCopyUint32 copies a uint32 from C memory.
func FastCopyUint32(cPtr unsafe.Pointer) uint32 {
	return *(*uint32)(cPtr)
}

// FastCopyUint64 copies a uint64 from C memory.
func FastCopyUint64(cPtr unsafe.Pointer) uint64 {
	return *(*uint64)(cPtr)
}

// FastCopyFloat32 copies a float32 from C memory.
func FastCopyFloat32(cPtr unsafe.Pointer) float32 {
	return *(*float32)(cPtr)
}

// FastCopyFloat64 copies a float64 from C memory.
func FastCopyFloat64(cPtr unsafe.Pointer) float64 {
	return *(*float64)(cPtr)
}

// FastCopyBool copies a bool from C memory.
// Note: C bool is typically represented as uint8.
func FastCopyBool(cPtr unsafe.Pointer) bool {
	return *(*uint8)(cPtr) != 0
}

// MustFastCopy is like FastCopy but panics if the type is not a primitive.
// This is useful when you know the type is primitive at compile time.
func MustFastCopy[T any](cPtr unsafe.Pointer) T {
	return FastCopy[T](cPtr)
}

// CanFastCopy returns true if type T can use FastCopy.
// Only primitive types return true.
func CanFastCopy[T any]() bool {
	var zero T
	typ := reflect.TypeOf(zero)
	return isPrimitiveKind(typ.Kind())
}
