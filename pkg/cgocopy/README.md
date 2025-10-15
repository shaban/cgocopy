# cgocopy

High-performance library for copying data between C and Go structs with automatic type conversion and field mapping.

## Features

- ✅ **Zero-Copy Performance**: FastCopy achieves 15-17x speedup over reflection-based Copy
- ✅ **Type Safety**: Go generics ensure compile-time type checking
- ✅ **Automatic Field Mapping**: Position-based or tag-based field matching
- ✅ **Code Generation**: `cgocopy-generate` tool eliminates 80% of boilerplate
- ✅ **C11 Macros**: Automatic type detection in C with `_Generic`
- ✅ **Comprehensive Types**: All Go primitives, strings, arrays, nested structs
- ✅ **Thread-Safe**: Concurrent-safe registry for metadata
- ✅ **Production-Ready**: 100+ tests, benchmarked, zero external dependencies

## Quick Start

### 1. Install cgocopy-generate

```bash
cd tools/cgocopy-generate
go build
```

### 2. Define C Structs

```c
// native/structs.h
typedef struct {
    int id;
    char* name;
    double score;
} Person;
```

### 3. Add go:generate Directive

```go
// main.go
//go:generate cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h

/*
#include "native/metadata_api.h"
#include "native/structs_meta.c"
*/
import "C"
import "github.com/shaban/cgocopy/pkg/cgocopy"
```

### 4. Define Go Struct

```go
// Match by position (no tags needed if fields are in same order)
type Person struct {
    ID    int32
    Name  string
    Score float64
}
```

### 5. Register Metadata

```go
func init() {
    cgocopy.PrecompileWithC[Person](extractCMetadata(C.get_Person_metadata()))
}
```

### 6. Use Copy Functions

```go
// Slow but flexible (uses reflection)
goStruct, err := cgocopy.Copy[Person](cPtr)

// Fast (15-17x faster, zero allocations)
goStruct, err := cgocopy.FastCopy[Person](cPtr)
```

## API Reference

### Copy[T](ptr unsafe.Pointer) (T, error)

Copies data from C struct to Go struct using reflection. Works with any registered type.

**Performance**: ~50ns per call, 2 allocations

**Use When**:
- You need maximum flexibility
- Performance is not critical
- Working with dynamic types

```go
person, err := cgocopy.Copy[Person](cPersonPtr)
if err != nil {
    log.Fatal(err)
}
```

### FastCopy[T](ptr unsafe.Pointer) (T, error)

High-performance copy using pre-compiled memory operations. Requires PrecompileWithC registration.

**Performance**: ~3ns per call, 0 allocations (15-17x faster than Copy)

**Use When**:
- Performance is critical
- Type is known at compile time
- You've precompiled with PrecompileWithC

```go
person, err := cgocopy.FastCopy[Person](cPersonPtr)
if err != nil {
    log.Fatal(err)
}
```

### PrecompileWithC[T](cInfo CStructInfo) error

Registers C struct metadata and precompiles memory copy operations for FastCopy.

**Required for**: FastCopy to work

```go
func init() {
    if err := cgocopy.PrecompileWithC[Person](
        extractCMetadata(C.get_Person_metadata()),
    ); err != nil {
        panic(err)
    }
}
```

### Precompile[T]() error

Registers Go struct for reflection-based Copy (without C metadata).

**Use When**: You want Copy to work but don't have C metadata

```go
if err := cgocopy.Precompile[Person](); err != nil {
    log.Fatal(err)
}
```

## Field Mapping

cgocopy supports two field matching modes:

### Position-Based Matching (Recommended)

Fields match by their order in the struct. **No tags needed!**

```go
// C struct
typedef struct {
    int id;
    char* name;
    double score;
} Person;

// Go struct - matches by position
type Person struct {
    ID    int32   // matches 'id' (field 0)
    Name  string  // matches 'name' (field 1)
    Score float64 // matches 'score' (field 2)
}
```

### Tag-Based Matching

Use `cgocopy:"field_name"` tags when field names differ:

```go
type Person struct {
    UserID   int32  `cgocopy:"id"`       // explicit mapping
    FullName string `cgocopy:"name"`     // explicit mapping
    Score    float64                     // matches by position
}
```

### Skip Fields

Use `cgocopy:"-"` to skip fields:

```go
type Person struct {
    ID       int32
    Name     string
    Password string `cgocopy:"-"` // not copied from C
}
```

## Code Generation Workflow

The `cgocopy-generate` tool automates metadata generation:

### Manual Workflow (Old - 5-10 minutes)

1. Write C struct
2. Write CGOCOPY_STRUCT macro
3. Write CGOCOPY_FIELD for each field
4. Write getter function
5. Update header file
6. Update CGO imports

### Automated Workflow (New - 30 seconds)

1. Write C struct in `.h` file
2. Add `//go:generate` directive
3. Run `go generate`

**Example:**

```go
//go:generate cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h
```

Generates:
- `structs_meta.c`: CGOCOPY_STRUCT macros and getter functions
- `metadata_api.h`: Getter function declarations

See `examples/users/` for a complete working example.

## Performance Benchmarks

Benchmarks on Apple M1 Pro:

| Operation | Time | Allocations | vs Copy |
|-----------|------|-------------|---------|
| FastCopy[Int32] | 3.47 ns | 0 | **15x faster** |
| Copy[Int32] | 51.86 ns | 2 | baseline |
| FastCopy[Int64] | 3.15 ns | 0 | **16x faster** |
| Copy[Int64] | 49.25 ns | 2 | baseline |
| FastCopy[Float64] | 2.84 ns | 0 | **17x faster** |
| Copy[Float64] | 48.29 ns | 2 | baseline |
| Non-generic FastCopy | 0.31 ns | 0 | **155x faster** |

**Key Findings:**
- FastCopy is 15-17x faster than Copy
- FastCopy has zero heap allocations
- Non-generic FastCopy is fastest but requires manual setup
- Copy provides flexibility with reasonable performance

**Recommendation**: Use FastCopy for hot paths, Copy for convenience.

## Supported Types

### Primitives
- `int8`, `uint8`, `int16`, `uint16`, `int32`, `uint32`, `int64`, `uint64`
- `float32`, `float64`
- `bool`
- `string` (auto-converts C `char*` to Go string)

### Complex Types
- Fixed-size arrays: `[N]T`
- Nested structs: `Point3D`
- Pointers: `*T` (basic support)

### Type Conversions

cgocopy automatically handles:
- C `int` ↔ Go `int32`
- C `char*` ↔ Go `string` (with proper memory management)
- C `double` ↔ Go `float64`
- C fixed arrays ↔ Go arrays

## Limitations

### Not Supported

❌ **Arrays of structs with strings**: Known issue causing segfaults due to complex memory management

```go
// This may crash:
type User struct {
    Name string
}
type Users struct {
    Items [10]User  // Array of structs with strings
}
```

**Workaround**: Use slices and iterate manually, or simplify struct to primitives only.

❌ **Dynamic arrays**: C arrays without known size at compile time

❌ **Function pointers**: No support for copying function pointers

❌ **Unions**: C unions are not supported

❌ **Bit fields**: C bit fields cannot be properly mapped

### Partially Supported

⚠️ **Nested structs**: Supported but must register each nested type separately

```go
// Both Point3D and GameObject must be registered
cgocopy.PrecompileWithC[Point3D](...)
cgocopy.PrecompileWithC[GameObject](...)
```

⚠️ **Pointers**: Basic pointer support, but not for complex nested pointer structures

## Project Structure

```
pkg/cgocopy/
├── cgocopy.go           # Main API (Copy, FastCopy, Precompile, PrecompileWithC)
├── registry.go          # Thread-safe metadata registry
├── types.go             # Core types (CStructInfo, CFieldInfo, TypedCopier)
├── native/              # C11 macros
│   ├── cgocopy_macros.h # Core CGOCOPY_STRUCT/FIELD macros
│   └── README.md        # C macro documentation
├── integration/         # Integration tests (9 tests)
├── examples/
│   └── users/           # Complete working example
└── tools/
    └── cgocopy-generate/  # Code generation tool
```

## Examples

### Complete Example: examples/users

See `examples/users/` for a full working example with:
- C struct definitions (`native/structs.h`)
- Auto-generated metadata (`native/structs_meta.c`)
- Go structs matching C structs
- CGO integration with `go:generate`
- Working test demonstrating Copy

### Integration Tests: pkg/cgocopy/integration

See `pkg/cgocopy/integration/` for comprehensive tests covering:
- Simple types (SimplePerson)
- Strings (User)
- Arrays (Student)
- Nested structs (GameObject with Point3D)
- All primitive types (AllTypes)
- FastCopy primitives
- Validation

## Requirements

- **Go**: 1.21+ (for generics)
- **C Compiler**: GCC 4.9+, Clang 3.3+, or MSVC 2015+
- **C Standard**: C11 or later (for `_Generic` support)

## Testing

```bash
# Run all tests
go test ./... -v

# Run benchmarks
go test -bench=. -benchmem

# Run specific test
go test -run TestCopy_TaggedStruct -v
```

## Best Practices

1. **Use FastCopy for hot paths**: 15-17x performance improvement
2. **Use position-based matching**: Avoid tags unless names differ
3. **Register with PrecompileWithC**: Required for FastCopy
4. **Use cgocopy-generate**: Eliminate boilerplate, reduce errors
5. **Keep structs simple**: Avoid arrays-of-structs-with-strings
6. **Register nested types**: Each nested struct needs separate registration
7. **Use go:generate**: Standard Go workflow, reproducible builds

## Troubleshooting

### "type not registered"

**Solution**: Call `PrecompileWithC` or `Precompile` in `init()`

### "C field not found"

**Solution**: Check field matching - use tags if names differ, or ensure fields are in same order

### FastCopy returns zero values

**Solution**: Use `PrecompileWithC` (not `Precompile`) to register C metadata

### Segfault with arrays

**Solution**: Avoid arrays-of-structs-with-strings. Use simpler types or iterate manually.

### Header path not found

**Solution**: Use `-header-path` flag with cgocopy-generate:
```bash
cgocopy-generate -input=structs.h -output=metadata.c -header-path=path/to/cgocopy_macros.h
```

## License

MIT License - See LICENSE file for details

## See Also

- [cgocopy-generate Tool](../../tools/cgocopy-generate/README.md) - Code generation documentation
- [Integration Tests](integration/README.md) - Complete integration examples
- [C Macros Guide](native/README.md) - C11 macro documentation
- [Examples](../../examples/users/) - Working example code
