# cgocopy

Fast and safe C to Go struct copying.

## Installation

```bash
go get github.com/shaban/cgocopy
```

## Quick Start

### Simple Structs (Primitives Only)

```go
import "github.com/shaban/cgocopy"

// C struct
typedef struct {
    uint32_t id;
    float value;
} Sensor;

// Go struct (identical layout)
type Sensor struct {
    ID    uint32
    Value float32
}

// Copy (0.3ns, compiler-inlined)
var sensor Sensor
cgocopy.Direct(&sensor, unsafe.Pointer(cSensor))
```

### Arrays

```go
sensors := make([]Sensor, count)
cSize := unsafe.Sizeof(C.Sensor{})
cgocopy.DirectArray(sensors, unsafe.Pointer(cSensors), cSize)
```

### With Strings

Two approaches available:

#### 1. StringPtr (Fast, Manual Memory)

```go
type Device struct {
    ID   uint32
    Name cgocopy.StringPtr  // 8-byte pointer
}

// Copy structs (0.3ns each)
devices := make([]Device, count)
cgocopy.DirectArray(devices, unsafe.Pointer(cDevices), cSize)

// Must keep C memory alive
cleanup := func() { C.freeDevices(cDevices) }
defer cleanup()

// Lazy string access (29ns when called)
name := devices[0].Name.String()
```

#### 2. Registry (Safe, Automatic Memory)

```go
type Device struct {
    ID   uint32
    Name string  // Real Go string
}

// Register once (at init) - streamlined!
registry := cgocopy.NewRegistry()
layout := cgocopy.AutoLayout("uint32_t", "char*")  // Auto-deduces sizes & IsString
registry.MustRegister(Device{}, cSize, layout, converter)

// Use many times
var device Device
registry.Copy(&device, unsafe.Pointer(cDevice))
C.freeDevice(cDevice)  // ✅ C memory can be freed immediately
```

## Performance

| Method | Time | Use Case |
|--------|------|----------|
| Direct | 0.3ns | Primitives only |
| Direct + StringPtr | 29ns/string | Lazy string access |
| Direct + 3 strings | ~65ns | Manual string conversion |
| Registry.Copy | ~110ns primitives, ~170ns with strings | Automatic strings, validation |

**⚠️ Performance Notes:**
- Benchmarks measured on: **macOS (Darwin), x86_64 architecture, Apple Silicon**
- Performance may vary significantly across platforms, compilers, and hardware configurations
- String conversion times depend on string length and memory allocation patterns
- Registry validation overhead occurs once at registration time, not per copy operation

## Documentation

- [CSTRINGPTR.md](CSTRINGPTR.md) - StringPtr detailed guide
- [specs/USAGE_GUIDE.md](specs/USAGE_GUIDE.md) - When to use which approach
- [specs/IMPLEMENTATION.md](specs/IMPLEMENTATION.md) - Technical details
- [specs/REGISTRY_IMPROVEMENTS.md](specs/REGISTRY_IMPROVEMENTS.md) - AutoLayout and improvements
- [specs/WHEN_TO_USE_REGISTRY.md](specs/WHEN_TO_USE_REGISTRY.md) - Registry detailed guide

## API Reference

### Direct

```go
func Direct[T any](dst *T, src unsafe.Pointer)
```

Zero-overhead copy. Compiler-inlined. Use for primitives and fixed-size arrays only.

### DirectArray

```go
func DirectArray[T any](dst []T, src unsafe.Pointer, cElemSize uintptr)
```

Efficient array copying. Fully inlined.

### StringPtr

```go
type StringPtr uintptr

func (p StringPtr) String() string
func (p StringPtr) IsNil() bool
```

Lazy char* wrapper. Requires C memory to stay alive.

### Registry

```go
type Registry struct { ... }

func NewRegistry() *Registry
func (r *Registry) Register(goType reflect.Type, cSize uintptr, 
                            layout []FieldInfo, 
                            converter ...CStringConverter) error
func (r *Registry) Copy(dst interface{}, cPtr unsafe.Pointer) error

func AutoLayout(typeNames ...string) []FieldInfo
```

Validated copying with string conversion and nested struct support.

## Memory Safety

### Direct + StringPtr
- Requires explicit memory management
- Must keep C memory alive during string access
- Use cleanup functions

### Registry.Copy
- Automatic memory management
- Copies strings immediately
- C memory can be freed after Copy()

## Examples

See test files for complete examples:
- Basic struct copying
- Array handling
- String conversion
- Nested structs

## Testing

```bash
cd cgocopy
go test -v
go test -bench .
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

Copyright (c) 2025 shaban
