package cgocopy

import (
	"fmt"
	"unsafe"
)

// Example 1: Using StringPtr with DirectCopy (ultra-fast, lazy strings)
func Example_cstringPtrDirect() {
	// Simulate C struct with char* pointer
	cName := testCreateCString("Test Device")
	defer testFreeCString(cName)

	type CDevice struct {
		ID   uint32
		Name unsafe.Pointer // C: char* name
		Size uint32
	}

	// Go struct using StringPtr
	type Device struct {
		ID   uint32
		Name StringPtr // 8-byte pointer
		Size uint32
	}

	// Create a "C" device
	cDevice := CDevice{
		ID:   42,
		Name: cName,
		Size: 1024,
	}

	// Ultra-fast direct copy (0.31ns, just copies pointer)
	var device Device
	Direct(&device, unsafe.Pointer(&cDevice))

	// Access fields directly
	fmt.Printf("ID: %d\n", device.ID)
	fmt.Printf("Name: %s\n", device.Name.String()) // Lazy conversion (29ns)
	fmt.Printf("Size: %d\n", device.Size)

	// Output:
	// ID: 42
	// Name: Test Device
	// Size: 1024
}

// Example 2: Iterating over C array with cleanup pattern
func Example_cstringPtrArrayCleanup() {
	// Simulate C array of structs with char* pointers
	type CDevice struct {
		ID   uint32
		Name unsafe.Pointer // C: char* name
		Size uint32
	}

	// Go equivalent
	type Device struct {
		ID   uint32
		Name StringPtr
		Size uint32
	}

	// Create C strings (in real code, these come from C)
	cNames := []unsafe.Pointer{
		testCreateCString("Device 1"),
		testCreateCString("Device 2"),
		testCreateCString("Device 3"),
	}

	// Cleanup function frees all C memory
	cleanup := func() {
		for _, name := range cNames {
			testFreeCString(name)
		}
	}
	defer cleanup()

	// Create array of C devices
	cDevices := []CDevice{
		{ID: 1, Name: cNames[0], Size: 512},
		{ID: 2, Name: cNames[1], Size: 1024},
		{ID: 3, Name: cNames[2], Size: 2048},
	}

	// Convert to Go structs using DirectCopy (0.31ns each)
	devices := make([]Device, len(cDevices))
	for i := range cDevices {
		Direct(&devices[i], unsafe.Pointer(&cDevices[i]))
	}

	// Use the devices (lazy string conversion only when needed)
	for _, d := range devices {
		fmt.Printf("Device %d: %s (%d bytes)\n", d.ID, d.Name.String(), d.Size)
	}

	// Output:
	// Device 1: Device 1 (512 bytes)
	// Device 2: Device 2 (1024 bytes)
	// Device 3: Device 3 (2048 bytes)
}

// Example 3: DirectCopyArray - copy entire C arrays efficiently
func Example_directCopyArray() {
	// Simulate C array of structs with char* pointers
	type CDevice struct {
		ID   uint32
		Name unsafe.Pointer // C: char* name
		Size uint32
	}

	type Device struct {
		ID   uint32
		Name StringPtr
		Size uint32
	}

	// Create C strings
	cNames := []unsafe.Pointer{
		testCreateCString("Device 1"),
		testCreateCString("Device 2"),
		testCreateCString("Device 3"),
	}
	defer func() {
		for _, name := range cNames {
			testFreeCString(name)
		}
	}()

	// Create C array
	cDevices := []CDevice{
		{ID: 1, Name: cNames[0], Size: 512},
		{ID: 2, Name: cNames[1], Size: 1024},
		{ID: 3, Name: cNames[2], Size: 2048},
	}

	// Allocate Go slice
	devices := make([]Device, len(cDevices))

	// Copy entire array in one call (compiler-inlined)
	cSize := unsafe.Sizeof(CDevice{})
	DirectArray(devices, unsafe.Pointer(&cDevices[0]), cSize)

	// Use the devices
	for _, d := range devices {
		fmt.Printf("Device %d: %s (%d bytes)\n", d.ID, d.Name.String(), d.Size)
	}

	// Output:
	// Device 1: Device 1 (512 bytes)
	// Device 2: Device 2 (1024 bytes)
	// Device 3: Device 3 (2048 bytes)
}

// Example 4: Composition with StringPtr - build complex structures in Go
func Example_cstringPtrComposition() {
	// C structures with char* pointers
	type CEngine struct {
		ID         uint32
		Name       unsafe.Pointer // C: char* name
		SampleRate float64
	}

	type CDevice struct {
		EngineID uint32
		ID       uint32
		Name     unsafe.Pointer // C: char* name
		Channels uint32
	}

	// Go structures using StringPtr
	type Engine struct {
		ID         uint32
		Name       StringPtr
		SampleRate float64
	}

	type Device struct {
		EngineID uint32
		ID       uint32
		Name     StringPtr
		Channels uint32
	}

	// Create C strings
	engineName := testCreateCString("Main Engine")
	device1Name := testCreateCString("Speakers")
	device2Name := testCreateCString("Audio Interface")

	cleanup := func() {
		testFreeCString(engineName)
		testFreeCString(device1Name)
		testFreeCString(device2Name)
	}
	defer cleanup()

	// Simulate getting data from C
	cEngine := CEngine{
		ID:         1,
		Name:       engineName,
		SampleRate: 48000.0,
	}

	cDevices := []CDevice{
		{EngineID: 1, ID: 101, Name: device1Name, Channels: 2},
		{EngineID: 1, ID: 102, Name: device2Name, Channels: 8},
	}

	// Copy engine (0.31ns)
	var engine Engine
	Direct(&engine, unsafe.Pointer(&cEngine))

	// Copy devices array (0.31ns each)
	devices := make([]Device, len(cDevices))
	for i := range cDevices {
		Direct(&devices[i], unsafe.Pointer(&cDevices[i]))
	}

	// Compose in Go!
	type EngineState struct {
		Engine  Engine
		Devices []Device
	}

	state := EngineState{
		Engine:  engine,
		Devices: devices,
	}

	// Use composed structure (lazy string conversion only when needed)
	fmt.Printf("Engine: %s (%.0f Hz)\n", state.Engine.Name.String(), state.Engine.SampleRate)
	fmt.Printf("Devices (%d):\n", len(state.Devices))
	for _, device := range state.Devices {
		fmt.Printf("  - %s: %d channels\n", device.Name.String(), device.Channels)
	}

	// Output:
	// Engine: Main Engine (48000 Hz)
	// Devices (2):
	//   - Speakers: 2 channels
	//   - Audio Interface: 8 channels
}
