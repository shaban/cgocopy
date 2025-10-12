package cgocopy

import "unsafe"

// UTF8Converter converts C char* pointers into Go strings without invoking
// cgo helpers. It walks the byte sequence until the terminating NUL byte and
// then builds the resulting Go string.
type UTF8Converter struct{}

// CStringToGo converts a C string pointer into a Go string. It returns an
// empty string if the pointer is nil.
func (UTF8Converter) CStringToGo(ptr unsafe.Pointer) string {
	if ptr == nil {
		return ""
	}
	var buf []byte
	for i := uintptr(0); ; i++ {
		b := *(*byte)(unsafe.Add(ptr, i))
		if b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf)
}

// DefaultCStringConverter provides a ready-to-use converter instance that can
// be shared across registrations.
var DefaultCStringConverter UTF8Converter
