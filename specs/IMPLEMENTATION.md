# Implementation Specifications

The cgocopy package provides type-safe copying between C and Go structs through two approaches:

### Direct

Zero-overhead copying for structs containing only primitives and fixed-size arrays.

```go
Direct[T any](dst *T, src unsafe.Pointer)  // 0.3ns, compiler-inlined
```

Limitations:
- No string fields (char* → string conversion)
- No dynamic arrays/slices
- No validation

### Registry

Validated copying with support for strings and nested structs.

```go
registry.Register(reflect.TypeOf(Device{}), cSize, layout, converter)
registry.Copy(&goDevice, cDevicePtr)  // ~50ns with validation
```

Features:
- Compile-time validation
- String conversion (char* → string)
- Nested struct support
- Platform-portable offset handling

## Performance Characteristics

**⚠️ Performance Caveats:**
- All benchmarks measured on: **macOS (Darwin), ARM64 architecture, Apple Silicon**
- Performance may vary significantly on different platforms, compilers, and hardware
- String conversion times depend on string length and memory allocation patterns
- Registry validation overhead occurs once at registration time, not per copy
- Direct operations are compiler-inlined and may vary with optimization levels

| Operation | Time | Notes |
|-----------|------|-------|
| Direct | 0.3ns | Compiler-inlined |
| DirectArray (per element) | 0.3ns | Inlined loop |
| Registry.Copy (primitives only) | ~110ns | Includes validation |
| Registry.Copy (with strings) | ~170ns | Includes string allocation |
| String conversion | ~20-85ns | Scales ~0.7ns per character |
| Nested structs (5 levels) | ~640ns | Recursive field-by-field |

## String Handling

Three approaches available:

### 1. StringPtr (Recommended for performance)
```go
type Device struct {
    Name StringPtr  // 8 bytes, lazy conversion
}
// Requires: Explicit C memory management
// Performance: 29ns total (0.3ns copy + 29ns String())
```

### 2. Registry with Converter (Recommended for safety)
```go
type Device struct {
    Name string  // Eager conversion
}
// Requires: CStringConverter implementation
// Performance: ~50ns copy + string allocation
// Benefit: C memory can be freed immediately
```

### 3. Direct (Primitives only)
```go
type Device struct {
    ID uint32
    // No string fields
}
// Performance: 0.3ns
```

## Alignment and Padding

### Standard C Alignment Rules

Primitive types typically align to their size:
- uint8_t: 1-byte alignment
- uint16_t: 2-byte alignment
- uint32_t: 4-byte alignment
- uint64_t: 8-byte alignment
- pointers: 8-byte alignment (64-bit) / 4-byte (32-bit)

The package captures actual alignment at compile time via `ArchInfo`.

### AutoLayout Function

Calculates offsets automatically for standard C structs:

```go
layout := cgocopy.AutoLayout("uint32_t", "char*")
```

Tested accuracy: 100% across 31 fields in 8 struct patterns.

Limitations:
- Does not work with `#pragma pack` directives
- Does not work with `__attribute__((packed))`
- Does not work with bitfields or unions

For these cases, use explicit `offsetof()` helpers.

## Registry Improvements

### Auto-deduction

Two fields are now optional:

1. **IsString**: Automatically deduced from `TypeName == "char*"`
2. **Size**: Automatically deduced from TypeName using ArchInfo

Before:
```go
layout := []FieldInfo{
    {Offset: 0, Size: 4, TypeName: "uint32_t", IsString: false},
    {Offset: 8, Size: 8, TypeName: "char*", IsString: true},
}
```

After:
```go
layout := []FieldInfo{
    {Offset: 0, TypeName: "uint32_t"},
    {Offset: 8, TypeName: "char*"},  // IsString and Size deduced
}
```

Or use AutoLayout:
```go
layout := cgocopy.AutoLayout("uint32_t", "char*")
```

## Validation

The Registry validates at registration time:
- Field count matches
- Field sizes match
- Types are compatible
- Nested structs are registered first
- String fields have a converter

Validation errors occur at init(), not runtime.

## Nested Structs

Register innermost structs first:

```go
registry.Register(reflect.TypeOf(Inner{}), ...)
registry.Register(reflect.TypeOf(Outer{}), ...)  // References Inner
```

Copy handles recursion automatically. Tested up to 5 levels deep.

## Platform Safety

The package handles platform differences through:
1. Captured sizes via `sizeof()` at compile time
2. Captured alignments via test struct with `offsetof()`
3. No hardcoded offsets
4. Works across macOS (x86_64, ARM64), Linux, Windows

## Use Case Selection

### Direct + StringPtr
Use when:
- Performance critical (hot path)
- Can manage C memory lifetime
- Strings accessed rarely

### Registry.Copy
Use when:
- Want automatic memory management
- Need validation
- Working with third-party C libraries
- Platform portability required
- Complex or unpredictable struct layouts

## Test Coverage

- 35+ test cases
- All primitive types
- Nested structs (1-5 levels)
- String conversion edge cases
- Platform safety validation
- Offset prediction accuracy
- Benchmark suite

## Backward Compatibility

All improvements maintain backward compatibility:
- Explicit IsString still works
- Explicit Size still works
- Manual offsetof() still works
- Existing registrations unchanged
