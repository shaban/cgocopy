package cgocopy

import (
	"testing"
	"unsafe"
)

func TestDirectArray(t *testing.T) {
	// Create C array
	type CDevice struct {
		ID   uint32
		Name unsafe.Pointer
		Size uint32
	}

	// Go equivalent
	type Device struct {
		ID   uint32
		Name StringPtr
		Size uint32
	}

	// Create C strings
	name1 := testCreateCString("Device 1")
	name2 := testCreateCString("Device 2")
	name3 := testCreateCString("Device 3")
	defer func() {
		testFreeCString(name1)
		testFreeCString(name2)
		testFreeCString(name3)
	}()

	// Create C array
	cDevices := []CDevice{
		{ID: 1, Name: name1, Size: 512},
		{ID: 2, Name: name2, Size: 1024},
		{ID: 3, Name: name3, Size: 2048},
	}

	// Allocate Go slice
	devices := make([]Device, len(cDevices))

	// Copy using DirectCopyArray
	cSize := unsafe.Sizeof(CDevice{})
	DirectArray(devices, unsafe.Pointer(&cDevices[0]), cSize)

	// Verify results
	expected := []struct {
		id   uint32
		name string
		size uint32
	}{
		{1, "Device 1", 512},
		{2, "Device 2", 1024},
		{3, "Device 3", 2048},
	}

	for i, exp := range expected {
		if devices[i].ID != exp.id {
			t.Errorf("Device %d: expected ID %d, got %d", i, exp.id, devices[i].ID)
		}
		if devices[i].Name.String() != exp.name {
			t.Errorf("Device %d: expected name %s, got %s", i, exp.name, devices[i].Name.String())
		}
		if devices[i].Size != exp.size {
			t.Errorf("Device %d: expected size %d, got %d", i, exp.size, devices[i].Size)
		}
	}

	t.Logf("✅ Successfully copied %d devices using DirectCopyArray", len(devices))
}

func BenchmarkDirectArray(b *testing.B) {
	type CDevice struct {
		ID   uint32
		Name unsafe.Pointer
		Size uint32
	}

	type Device struct {
		ID   uint32
		Name StringPtr
		Size uint32
	}

	// Create test data
	name := testCreateCString("Test Device")
	defer testFreeCString(name)

	cDevices := make([]CDevice, 100)
	for i := range cDevices {
		cDevices[i] = CDevice{
			ID:   uint32(i),
			Name: name,
			Size: 1024,
		}
	}

	devices := make([]Device, len(cDevices))
	cSize := unsafe.Sizeof(CDevice{})
	src := unsafe.Pointer(&cDevices[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DirectArray(devices, src, cSize)
	}
}

func BenchmarkDirectCopyArray_Single(b *testing.B) {
	type CDevice struct {
		ID   uint32
		Name unsafe.Pointer
		Size uint32
	}

	type Device struct {
		ID   uint32
		Name StringPtr
		Size uint32
	}

	name := testCreateCString("Test")
	defer testFreeCString(name)

	cDevice := CDevice{ID: 42, Name: name, Size: 1024}
	devices := make([]Device, 1)
	cSize := unsafe.Sizeof(CDevice{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DirectArray(devices, unsafe.Pointer(&cDevice), cSize)
	}
}

func BenchmarkDirectCopyArray_vs_Manual(b *testing.B) {
	type CDevice struct {
		ID   uint32
		Name unsafe.Pointer
		Size uint32
	}

	type Device struct {
		ID   uint32
		Name StringPtr
		Size uint32
	}

	name := testCreateCString("Test")
	defer testFreeCString(name)

	cDevices := make([]CDevice, 100)
	for i := range cDevices {
		cDevices[i] = CDevice{ID: uint32(i), Name: name, Size: 1024}
	}

	devices := make([]Device, len(cDevices))
	cSize := unsafe.Sizeof(CDevice{})
	src := unsafe.Pointer(&cDevices[0])

	b.Run("DirectCopyArray", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DirectArray(devices, src, cSize)
		}
	})

	b.Run("Manual_Loop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := range devices {
				cPtr := unsafe.Pointer(uintptr(src) + uintptr(j)*cSize)
				Direct(&devices[j], cPtr)
			}
		}
	})

	b.Run("Manual_unsafe.Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := range devices {
				cPtr := unsafe.Add(src, j*int(cSize))
				Direct(&devices[j], cPtr)
			}
		}
	})
}
