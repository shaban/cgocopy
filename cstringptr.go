package cgocopy

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import "unsafe"

// StringPtr wraps a C char* pointer for lazy string conversion.
//
// CRITICAL: The underlying C memory MUST remain valid for the lifetime
// of any struct containing StringPtr fields. The caller is responsible
// for memory management.
//
// Use this when:
//   - You want ultra-fast struct copying (0.31ns Direct)
//   - Strings are accessed rarely or conditionally
//   - You're willing to manage C memory lifetime
//
// Pattern:
//
//	type Device struct {
//	    ID   uint32
//	    Name cgocopy.StringPtr  // Lazy string
//	}
//
//	cDevice := C.getDevice()
//	defer C.freeDevice(cDevice)  // Keep C memory alive!
//
//	var device Device
//	cgocopy.Direct(&device, unsafe.Pointer(cDevice))  // 0.31ns
//
//	// String conversion only happens when called
//	fmt.Println(device.Name.String())  // ~29ns
type StringPtr struct {
	ptr unsafe.Pointer // Points to C char*
}

// String converts the C string to a Go string.
// Allocates a new Go string each time it's called.
// Returns empty string if pointer is nil.
func (c StringPtr) String() string {
	if c.ptr == nil {
		return ""
	}
	return C.GoString((*C.char)(c.ptr))
}

// IsNil returns true if the pointer is nil
func (c StringPtr) IsNil() bool {
	return c.ptr == nil
}

// Ptr returns the raw unsafe.Pointer (for advanced use)
func (c StringPtr) Ptr() unsafe.Pointer {
	return c.ptr
}

// NewStringPtr creates a StringPtr from an unsafe.Pointer
// (for testing or advanced use cases)
func NewStringPtr(ptr unsafe.Pointer) StringPtr {
	return StringPtr{ptr: ptr}
}
