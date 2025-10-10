# Usage Decision Guide

## Quick Selection

```
Do you have char* fields that need string conversion?
├─ YES → Can you manage C memory lifetime explicitly?
│   ├─ YES → DirectCopy + CStringPtr (29ns, manual cleanup)
│   └─ NO  → Registry.Copy + Converter (50ns, automatic)
└─ NO → DirectCopy (0.3ns, primitives only)
```

## DirectCopy

### When to Use
- Struct contains only primitives and fixed-size arrays
- Performance is critical
- No string conversion needed
- You control both C and Go struct definitions

### Example
```go
type Device struct {
    ID       uint32
    Channels uint32
    Rate     float64
}

devices := make([]Device, count)
cSize := unsafe.Sizeof(C.Device{})
structcopy.DirectCopyArray(devices, unsafe.Pointer(cDevices), cSize)
```

### Performance
- 0.3ns per struct (compiler-inlined)
- Zero allocations
- No validation overhead

### Limitations
- No char* → string conversion
- No validation
- Silent corruption if layouts mismatch

## DirectCopy + CStringPtr

### When to Use
- Struct has char* fields
- Performance critical
- Can manage C memory explicitly
- Strings accessed rarely or conditionally

### Example
```go
type Device struct {
    ID   uint32
    Name CStringPtr  // 8-byte pointer wrapper
}

devices := make([]Device, count)
structcopy.DirectCopyArray(devices, unsafe.Pointer(cDevices), cSize)

// Cleanup function required
cleanup := func() {
    for i := range devices {
        C.freeDeviceName(unsafe.Pointer(devices[i].Name))
    }
}
defer cleanup()

// Lazy string access
if devices[0].ID == targetID {
    name := devices[0].Name.String()  // 29ns conversion
}
```

### Performance
- 0.3ns copy + 29ns String() (lazy)
- Only pay conversion cost when accessed
- Zero allocations during copy

### Requirements
- Explicit cleanup function
- C memory must stay valid
- Risk of use-after-free if mismanaged

## Registry.Copy

### When to Use
- Want automatic memory management
- Strings accessed multiple times
- Need validation
- Working with third-party C libraries
- Platform-portable code required
- Complex or unpredictable struct layouts

### Example
```go
type Device struct {
    ID   uint32
    Name string  // Real Go string
}

// Register once at init
layout := structcopy.AutoLayout("uint32_t", "char*")
registry.Register(reflect.TypeOf(Device{}), cSize, layout, converter)

// Use many times
func GetDevice() (Device, error) {
    cDevice := C.getDevice()
    defer C.freeDevice(cDevice)  // Safe to free immediately
    
    var device Device
    err := registry.Copy(&device, unsafe.Pointer(cDevice))
    return device, err
}
```

### Performance
- ~50ns per struct (includes validation)
- String conversion: ~20-85ns per string
- Total: ~50-100ns typical

### Benefits
- C memory freed immediately
- No use-after-free bugs possible
- Runtime validation
- Most idiomatic Go code
- Platform-portable

## Registry Use Cases

### 1. Eager String Conversion
Strings needed immediately and accessed multiple times.

### 2. Complex Struct Layouts
Platform-specific padding, third-party libraries, unpredictable alignment.

### 3. Dynamic Registration
Struct mappings determined at runtime, plugin systems.

### 4. Nested Structs
Mix of primitives and strings at different levels, automatic recursive handling.

## Comparison

| Aspect | DirectCopy | DirectCopy + CStringPtr | Registry.Copy |
|--------|-----------|------------------------|---------------|
| Speed | 0.3ns | 29ns (with access) | 50-100ns |
| Strings | No | Yes (lazy) | Yes (eager) |
| Validation | No | No | Yes |
| Memory | Manual | Manual | Automatic |
| Safety | Lowest | Medium | Highest |
| C Lifetime | N/A | Must keep alive | Free immediately |
| Use-after-free | N/A | Possible | Impossible |

## Migration Strategy

### From Registry to DirectCopy
If Registry.Copy is a bottleneck:

1. Measure: Confirm it's actually the bottleneck
2. Assess: Can you manage C memory lifetime?
3. Change struct: string → CStringPtr
4. Change copy: registry.Copy() → DirectCopy()
5. Add cleanup: Implement cleanup function

### From DirectCopy to Registry
If you need validation or automatic memory management:

1. Change struct: CStringPtr → string
2. Implement CStringConverter
3. Register struct with layout
4. Change copy: DirectCopy() → registry.Copy()
5. Remove cleanup: No longer needed

## Real-World Examples

### Audio Device Enumeration (Performance-critical)
```go
// Hot path: enumerate 100+ devices frequently
// Solution: DirectCopy + CStringPtr
// Reason: Names rarely accessed (filter by channels first)
devices := EnumerateDevices()  // 0.3ns per device
filtered := FilterByChannels(devices, 2)  // No string access
name := filtered[0].Name.String()  // Only convert when needed
```

### Configuration Loading (Safety-first)
```go
// One-time: load config at startup
// Solution: Registry.Copy
// Reason: Strings accessed multiple times, want validation
config := LoadConfig()  // 50-100ns acceptable
log.Printf("Server: %s", config.ServerName)  // String already converted
log.Printf("Port: %d", config.Port)
```

### Real-time Audio Processing (Maximum performance)
```go
// Hot path: process audio samples every millisecond
// Solution: DirectCopy (primitives only)
// Reason: No strings, pure data, nanoseconds matter
samples := GetAudioBuffer()  // 0.3ns per sample struct
ProcessAudio(samples)
```
