# Us```
Do you have char* fields that need string## Direct + CStringPtr

### When to Use
- Struct has char* fields
- Performance critical
- Can manage C memory explicitly
- Strings accessed rarely or conditionally

### Example
```go
type Device struct {
    ID   uint32
    Name cgocopy.StringPtr  // 8-byte pointer wrapper
}

devices := make([]Device, count)
cgocopy.DirectArray(devices, unsafe.Pointer(cDevices), cSize)ES → Can you manage C memory lifetime explicitly?
│   ├─ YES → Direct + CStringPtr (29ns, manual cleanup)
│   └─ NO  → Registry.Copy (50ns, automatic)
└─ NO → Direct (0.3ns, primitives only)
```ision Guide

## Quick Selection

```
Do you have char* fields that need string conversion?
├─ YES → Can you manage C memory lifetime explicitly?
│   ├─ YES → Direct + StringPtr (29ns, manual cleanup)
│   └─ NO  → Registry.Copy (110-170ns, automatic)
└─ NO → Direct (0.3ns, primitives only)
```

**⚠️ Performance Caveats:**
- All benchmarks measured on: **macOS (Darwin), ARM64 architecture, Apple Silicon**
- Performance may vary significantly across platforms, compilers, and hardware configurations
- String conversion times depend on string length and memory allocation patterns
- Registry validation overhead occurs once at registration time, not per copy operation

## Direct

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
cgocopy.DirectArray(devices, unsafe.Pointer(cDevices), cSize)
```

### Performance
- 0.3ns per struct (compiler-inlined)
- Zero allocations
- No validation overhead

### Limitations
- No char* → string conversion
- No validation
- Silent corruption if layouts mismatch

## Direct + StringPtr

### When to Use
- Struct has char* fields
- Performance critical
- Can manage C memory explicitly
- Strings accessed rarely or conditionally

### Example
```go
type Device struct {
    ID   uint32
    Name StringPtr  // 8-byte pointer wrapper
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
layout := cgocopy.AutoLayout("uint32_t", "char*")
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
- ~110ns per struct (includes validation)
- String conversion: ~20-85ns per string
- Total: ~110-170ns typical

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

| Aspect | Direct | Direct + StringPtr | Registry.Copy |
|--------|-----------|------------------------|---------------|
| Speed | 0.3ns | 29ns (with access) | 110-170ns |
| Strings | No | Yes (lazy) | Yes (eager) |
| Validation | No | No | Yes |
| Memory | Manual | Manual | Automatic |
| Safety | Lowest | Medium | Highest |
| C Lifetime | N/A | Must keep alive | Free immediately |
| Use-after-free | N/A | Possible | Impossible |

## Migration Strategy

### From Registry to Direct
If Registry.Copy is a bottleneck:

1. Measure: Confirm it's actually the bottleneck
2. Assess: Can you manage C memory lifetime?
3. Change struct: string → StringPtr
4. Change copy: registry.Copy() → Direct()
5. Add cleanup: Implement cleanup function

### From Direct to Registry
If you need validation or automatic memory management:

1. Change struct: StringPtr → string
2. Implement CStringConverter
3. Register struct with layout
4. Change copy: Direct() → registry.Copy()
5. Remove cleanup: No longer needed

## Real-World Examples

### Audio Device Enumeration (Performance-critical)
```go
// Hot path: enumerate 100+ devices frequently
// Solution: Direct + StringPtr
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
// Solution: Direct (primitives only)
// Reason: No strings, pure data, nanoseconds matter
samples := GetAudioBuffer()  // 0.3ns per sample struct
ProcessAudio(samples)
```
