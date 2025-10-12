# cgocopy2

Package cgocopy2 provides improved type-safe copying between C and Go structures with simplified macros, thread-safe registry, and struct tag support.

## Status

✅ **Phase 1 Complete** - Basic types and error handling implemented and tested.

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

## Next Steps

### Phase 2: Registry & Precompile
- Implement `Precompile[T any]()` function
- C metadata integration
- Type validation

### Phase 3: Mapping with Tags
- Struct tag parsing (`cgocopy:"field_name"`)
- Field mapping logic
- Nested struct support

### Phase 4: Copy Implementation
- Generic `Copy[T any]()` function
- C integration for actual copying
- Field-by-field copying logic

### Phase 5-8
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
├── types_test.go     # Type system tests ✅
├── errors.go         # Error types ✅
├── errors_test.go    # Error tests ✅
├── registry.go       # [Phase 2] Precompile implementation
├── registry_test.go  # [Phase 2] Registry tests
├── mapping.go        # [Phase 3] Field mapping with tags
├── mapping_test.go   # [Phase 3] Mapping tests
├── copy.go           # [Phase 4] Copy implementation
├── copy_test.go      # [Phase 4] Copy tests
├── fastcopy.go       # [Phase 5] FastCopy optimization
├── validation.go     # [Phase 6] ValidateStruct helper
└── README.md         # This file
```

## Documentation

- **API Specification**: `docs/migration/API_IMPROVEMENTS.md`
- **Implementation Plan**: `docs/migration/IMPLEMENTATION_PLAN.md`
- **Migration Status**: `docs/migration/STATUS.md`

## License

See LICENSE file in repository root.
