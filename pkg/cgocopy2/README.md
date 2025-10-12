# cgocopy2

Package cgocopy2 provides improved type-safe copying between C and Go structures with simplified macros, thread-safe registry, and struct tag support.

## Status

✅ **Phase 1-6 Complete** - Core functionality implemented and tested (93 tests, 100% pass rate).  
🚧 **Phase 7-8 Next** - C macro integration and migration planning.

## Features (Planned)

- **Simplified C Macros**: Use `CGOCOPY_STRUCT(type, ...)` with `_Generic` auto-detection
- **Thread-Safe Registry**: `sync.RWMutex` for concurrent type registration
- **Precompilation**: Explicit `Precompile[T any]()` for type registration at init time
- **Tagged Structs**: Support for `cgocopy:"field_name"` and `cgocopy:"-"` tags
- **FastCopy**: Zero-allocation copying for primitive types
- **Validation**: `ValidateStruct[T any]()` helper for debugging

## Current Implementation

### Phase 1: Basic Types ✅

#### Types (`types.go`)
- `FieldType`: Enum for field categories (Primitive, String, Struct, Array, Slice, Pointer)
- `FieldInfo`: Metadata for individual struct fields
- `StructMetadata`: Complete metadata for a struct type
- `Registry`: Thread-safe registry using `sync.RWMutex`

#### Errors (`errors.go`)
- Common errors: `ErrNotRegistered`, `ErrNilPointer`, `ErrInvalidType`, etc.
- `ValidationError`: Field-level validation failures
- `RegistrationError`: Type registration failures with cause wrapping
- `CopyError`: Runtime copy errors with field context

#### Tests
- `types_test.go`: 17 tests covering all type system functionality
- `errors_test.go`: 10 tests covering all error types and helpers
- All tests passing ✅

### Phase 2: Registry & Precompile ✅

#### Registry (`registry.go`)
- `Precompile[T any]()`: Analyzes and registers struct types at initialization
- `analyzeStruct()`: Reflection-based field metadata extraction
- `parseTag()`: Struct tag parsing for `cgocopy:"field_name"` and `cgocopy:"-"`
- `categorizeFieldType()`: Type validation and categorization
- `IsRegistered[T]()`: Check if type is precompiled
- `GetMetadata[T]()`: Retrieve precompiled metadata
- `Reset()`: Clear registry (for testing)

#### Features
- Automatic skipping of unexported fields
- Support for all Go types: primitives, strings, structs, arrays, slices, pointers
- Tag-based field name mapping
- Nested struct support
- Error detection for unsupported types (func, map, chan, interface)

#### Tests
- `registry_test.go`: 17 tests covering all precompile scenarios
- Tests for tagged structs, nested structs, arrays, slices, pointers
- Validation of unsupported types
- All 44 tests passing ✅

### Phase 4: Copy Implementation ✅

#### Copy (`copy.go`)
- `Copy[T any](cPtr unsafe.Pointer) (T, error)`: Main copying function
- `copyField()`: Type dispatcher for different field kinds
- `copyPrimitive()`: All primitive types (int, uint, float, bool variants)
- `copyString()`: C string (char*) to Go string with null-termination handling
- `copyStruct()`: Recursive copying for nested structs
- `copyArray()`: Fixed-size array copying
- `copySlice()`: Dynamic slice copying (assumes C struct: {void* data; size_t len})
- `copyPointer()`: Pointer field copying with nil handling

#### Features
- Zero-copy field addressing using unsafe.Pointer arithmetic
- Nil pointer safety checks
- Error context with field names
- Support for all precompiled types
- Recursive nested struct copying
- C string conversion with proper null-termination
- Array and slice element-by-element copying

#### Tests
- `copy_test.go`: 12 comprehensive test cases
- Simple structs with all primitive types
- String handling (including nil and empty strings)
- Nested struct copying
- Array copying with multiple element types
- Tagged struct field mapping validation
- Error cases (nil pointer, unregistered types)
- All 56 tests passing ✅

### Phase 5: FastCopy ✅

#### FastCopy (`fastcopy.go`)
- `FastCopy[T any](cPtr unsafe.Pointer) T`: Zero-allocation generic primitive copying
- Type-specific functions for direct access:
  - `FastCopyInt`, `FastCopyInt8`, `FastCopyInt16`, `FastCopyInt32`, `FastCopyInt64`
  - `FastCopyUint`, `FastCopyUint8`, `FastCopyUint16`, `FastCopyUint32`, `FastCopyUint64`
  - `FastCopyFloat32`, `FastCopyFloat64`
  - `FastCopyBool`
- `CanFastCopy[T]()`: Check if type can use FastCopy
- `MustFastCopy[T]()`: Panic-on-non-primitive variant

#### Features
- **Zero allocations**: Direct memory access without heap allocation
- **15x faster** than Copy for primitives (3.5ns vs 52ns)
- **No reflection overhead**: Compile-time type checking
- Panic protection for non-primitive types
- Works with all Go primitive types

#### Performance
```
BenchmarkFastCopy_Int32      332M ops    3.5 ns/op    0 B/op    0 allocs
BenchmarkCopy_Int32           22M ops   52.0 ns/op    8 B/op    2 allocs

BenchmarkFastCopy_Float64    421M ops    2.8 ns/op    0 B/op    0 allocs
BenchmarkCopy_Float64         25M ops   48.1 ns/op   16 B/op    2 allocs

BenchmarkFastCopyInt32_NonGeneric    1B ops    0.3 ns/op    0 B/op    0 allocs
```

#### Tests
- `fastcopy_test.go`: 20 test cases + 9 benchmarks
- All 13 primitive type variants tested
- Panic test for non-primitive types
- CanFastCopy validation tests
- Performance benchmarks vs Copy
- All 76 tests passing ✅

### Phase 6: Validation ✅

#### Validation (`validation.go`)
- `ValidateStruct[T any]() error`: Check if type is properly registered
- `validateMetadata()`: Check metadata completeness
- `validateField()`: Validate individual field (including recursive nested checks)
- `validateNestedStructs()`: Ensure all nested types are registered
- `ValidateAll() []error`: Validate all registered types at once
- `MustValidateStruct[T]()`: Panic-on-error variant
- `GetRegisteredTypes() []string`: List all registered type names

#### Features
- Registration verification before use
- Nested struct registration validation
- Array/slice/pointer element type checks
- Detailed error messages with field names and types
- Helps catch registration issues at initialization time
- Useful for debugging complex struct hierarchies

#### Tests
- `validation_test.go`: 17 comprehensive test cases
- Valid struct validation
- Unregistered type detection
- Nested struct validation (single and multiple levels)
- Array/slice element type validation
- Pointer target type validation
- Complex nested scenarios
- ValidateAll() for batch validation
- MustValidateStruct() panic behavior
- GetRegisteredTypes() introspection
- All 93 tests passing ✅

## Next Steps

### Phase 7: C Macro Implementation (Optional)
- Create `native2/cgocopy_metadata.h`
- Simplified `CGOCOPY_STRUCT` and `CGOCOPY_FIELD` macros
- C11 `_Generic` for auto-detection
- Integration tests with real C code

**Note**: Current Go-only implementation is fully functional. C macros are optional for integration with existing C projects.

### Phase 8: Integration & Migration
- Comprehensive integration tests
- Performance benchmarks vs v1
- Migration guide and examples
- Documentation updates

See `docs/migration/IMPLEMENTATION_PLAN.md` for complete roadmap.

## Testing

```bash
# Run all tests
go test ./pkg/cgocopy2/...

# Run with coverage
go test -cover ./pkg/cgocopy2/...

# Run specific test
go test -v -run TestRegistry ./pkg/cgocopy2/
```

## Architecture

```
pkg/cgocopy2/
├── types.go          # Core type definitions ✅
├── types_test.go     # Type system tests (17) ✅
├── errors.go         # Error types ✅
├── errors_test.go    # Error tests (10) ✅
├── registry.go       # Precompile & registry ✅
├── registry_test.go  # Registry tests (17) ✅
├── copy.go           # Copy implementation ✅
├── copy_test.go      # Copy tests (12) ✅
├── fastcopy.go       # FastCopy optimization ✅
├── fastcopy_test.go  # FastCopy tests (20) ✅
├── validation.go     # Validation helpers ✅
├── validation_test.go # Validation tests (17) ✅
└── README.md         # This file

Total: 93 tests, 100% pass rate
```

## Documentation

- **API Specification**: `docs/migration/API_IMPROVEMENTS.md`
- **Implementation Plan**: `docs/migration/IMPLEMENTATION_PLAN.md`
- **Migration Status**: `docs/migration/STATUS.md`

## License

See LICENSE file in repository root.
