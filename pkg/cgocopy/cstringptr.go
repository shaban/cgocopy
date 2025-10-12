package cgocopy

import "unsafe"

// UTF8Converter converts C char* pointers into Go strings.
type UTF8Converter struct{}

// CStringToGo converts a C string pointer into a Go string, returning an empty string for nil.
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

// DefaultCStringConverter provides a ready-to-use converter instance.
var DefaultCStringConverter UTF8Converter
