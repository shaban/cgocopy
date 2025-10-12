# cgocopy

Fast and safe C → Go struct copying with zero fuss.

## Installation

go get github.com/shaban/cgocopy

## Quick Start

### Direct copy (primitives only)

```go
// C struct
typedef struct {
    uint32_t id;
    float value;
} Sensor;

// Matching Go struct
type Sensor struct {
    ID    uint32
    Value float32
}

var s Sensor
cgocopy.Direct(&s, unsafe.Pointer(cSensor))
```

### Registry copy (strings & nested fields)

```go
type CDevice struct {
    uint32_t id;
    char*    name;
    uint32_t channels;
};

type Device struct {
    ID       uint32
    Name     string
    Channels uint32
}

registry := cgocopy.NewRegistry()
layout := []cgocopy.FieldInfo{
    {Offset: unsafe.Offsetof(CDevice{}.id), TypeName: "uint32_t"},
    {Offset: unsafe.Offsetof(CDevice{}.name), TypeName: "char*"},
    {Offset: unsafe.Offsetof(CDevice{}.channels), TypeName: "uint32_t"},
}

registry.MustRegister(Device{}, unsafe.Sizeof(CDevice{}), layout, cgocopy.DefaultCStringConverter)

var out Device
if err := registry.Copy(&out, unsafe.Pointer(cDevice)); err != nil {
    panic(err)
}

C.freeDevice(cDevice) // safe immediately – strings already copied
```

## Performance snapshot

Benchmarks collected with `go test -benchmem` on Apple M1 Pro, macOS 26.0.1 (25A362), Go 1.25.2 `darwin/arm64`. Expect variation on other hardware, OS versions, and Go releases.

### Initialization cost

| Benchmark | Time (ns/op) | Allocations | Notes |
|-----------|--------------|-------------|-------|
| `BenchmarkAutoLayout` | 144.8 | 384 B / 1 alloc | Registry metadata validation + converter lookup |
| `BenchmarkManualLayout` | 0.314 | 0 | Inline unsafe copy when layout already known |

### Copy throughput

| Benchmark | Time (ns/op) | Throughput | Allocations |
|-----------|--------------|------------|-------------|
| `BenchmarkMemCopyPureGo` | 389.2 | 84.2 GB/s | 0 |
| `BenchmarkMemCopyCgo` | 958.5 | 34.2 GB/s | 0 |
| `BenchmarkDirectCopy*` (inline/generic/manual) | 0.31–0.36 | — | 0 |
| `BenchmarkRegistryCopyLargeSensorBlock` | 1,407 | — | 0 |

### Nested struct handling

| Benchmark | Time (ns/op) | Allocations | Notes |
|-----------|--------------|-------------|-------|
| `BenchmarkNestedStructCopy` | 167.7 | 0 | Registry copy of nested primitives |
| `BenchmarkDeepNesting5Levels` | 239.8 | 0 | Five-level tree of nested arrays |

### String conversion

| Benchmark | Time (ns/op) | Allocations | Notes |
|-----------|--------------|-------------|-------|
| `BenchmarkUTF8Converter` | 36.6 | 24 B / 2 alloc | Pure-Go UTF-8 copy via `UTF8Converter` |
| `BenchmarkCStringConversion` | 19.2 | 80 B / 1 alloc | Baseline CGO conversion for comparison |
| `BenchmarkCStringByLength/short` | 12.0 | 5 B / 1 alloc | Direct copy of 5 characters |
| `BenchmarkCStringByLength/medium` | 14.7 | 32 B / 1 alloc | Direct copy of 30 characters |
| `BenchmarkCStringByLength/long` | 45.0 | 288 B / 1 alloc | Direct copy of 200 characters |

## Documentation

- [docs/guides/usage_guide.md](docs/guides/usage_guide.md) – choosing between Direct and Registry
- [docs/references/registry.md](docs/references/registry.md) – Registry internals and metadata helpers
- [docs/references/implementation.md](docs/references/implementation.md) – deep dive into validation & copying
- [docs/guides/when_to_use_registry.md](docs/guides/when_to_use_registry.md) – real-world scenarios

## API overview

### Direct

```go
func Direct[T any](dst *T, src unsafe.Pointer)
```

Compiler-inlined byte copy – ideal for structs containing only primitives or fixed-size arrays.

### DirectArray

```go
func DirectArray[T any](dst []T, src unsafe.Pointer, cElemSize uintptr)
```

Copies entire C arrays into Go slices with a single call.

### Registry

```go
type Registry struct { /* ... */ }

func NewRegistry() *Registry
func (r *Registry) Register(goType reflect.Type, cSize uintptr, layout []FieldInfo, converter ...CStringConverter) error
func (r *Registry) Copy(dst any, cPtr unsafe.Pointer) error
func (r *Registry) MustRegister(goStruct any, cSize uintptr, layout []FieldInfo, converter ...CStringConverter)
```

Validates layouts, handles nested structs/arrays, and performs optional string conversion.

### UTF8Converter

```go
type UTF8Converter struct{}

func (UTF8Converter) CStringToGo(ptr unsafe.Pointer) string

var DefaultCStringConverter UTF8Converter
```

Zero-Cgo string conversion. Pass `DefaultCStringConverter` (or your own implementation) when registering structs with `char*` fields.

## Memory model

- **Direct**: No allocations, but you own the lifetime of any pointers copied into Go structs.
- **Registry**: Performs deep copies of supported types (including strings). The C memory can be freed immediately after `Copy` returns.

## Examples

See the Go example tests for end-to-end samples covering direct copies, registry string handling, and nested arrays.

## Testing

```bash
go test ./...
go test -bench .
```

## License

MIT License – see [LICENSE](LICENSE).
