package cgocopy

import (
	"testing"
	"unsafe"
)

func TestStringPtr(t *testing.T) {
	cStr := testCreateCString("Hello, World!")
	defer testFreeCString(cStr)

	ptr := NewStringPtr(cStr)

	if ptr.IsNil() {
		t.Error("Pointer should not be nil")
	}

	s := ptr.String()
	if s != "Hello, World!" {
		t.Errorf("String() = %q, want %q", s, "Hello, World!")
	}
}

func TestStringPtrNil(t *testing.T) {
	ptr := NewStringPtr(nil)

	if !ptr.IsNil() {
		t.Error("Pointer should be nil")
	}

	s := ptr.String()
	if s != "" {
		t.Errorf("String() on nil = %q, want empty string", s)
	}
}

func TestStringPtrDirect(t *testing.T) {
	// Create C struct with char* field
	cDevice, cleanup := testCreateCDevicePtr(42, "Test Device", 8)
	defer cleanup()

	// Go struct with StringPtr
	type Device struct {
		ID       uint32
		Name     StringPtr
		Channels uint32
	}

	// Verify sizes match (should be identical)
	cSize := unsafe.Sizeof(cDevice)
	goSize := unsafe.Sizeof(Device{})
	if cSize != goSize {
		t.Fatalf("Size mismatch: C=%d, Go=%d", cSize, goSize)
	}

	// DirectCopy - ultra fast!
	var device Device
	Direct(&device, unsafe.Pointer(&cDevice))

	// Verify fields
	if device.ID != 42 {
		t.Errorf("ID = %d, want 42", device.ID)
	}

	if device.Channels != 8 {
		t.Errorf("Channels = %d, want 8", device.Channels)
	}

	// Lazy string conversion
	name := device.Name.String()
	if name != "Test Device" {
		t.Errorf("Name = %q, want %q", name, "Test Device")
	}
}

func BenchmarkStringPtrDirect(b *testing.B) {
	type Device struct {
		ID       uint32
		Name     StringPtr
		Channels uint32
	}

	cDevice, cleanup := testCreateCDevicePtr(42, "Benchmark Device", 8)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var device Device
		Direct(&device, unsafe.Pointer(&cDevice))
	}
}

func BenchmarkStringPtrConversion(b *testing.B) {
	cStr := testCreateCString("Benchmark String")
	defer testFreeCString(cStr)

	ptr := NewStringPtr(cStr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ptr.String()
	}
}

func BenchmarkStringPtrFullCycle(b *testing.B) {
	// Benchmark: DirectCopy + String conversion
	type Device struct {
		ID       uint32
		Name     StringPtr
		Channels uint32
	}

	cDevice, cleanup := testCreateCDevicePtr(42, "Benchmark Device", 8)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var device Device
		Direct(&device, unsafe.Pointer(&cDevice))
		_ = device.Name.String()
	}
}
