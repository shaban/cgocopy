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
registry.Copy(&goDevice, cDevicePtr)  // ~110ns with validation
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

Two primary strategies remain after removing `StringPtr`:

### 1. Registry with UTF8Converter (Recommended default)
```go
type Device struct {
    Name string
}

registry.MustRegister(Device{}, cSize, layout, cgocopy.DefaultCStringConverter)
```

- Strings are copied eagerly into Go memory.
- `UTF8Converter` avoids cgo helpers while keeping ownership clear.
- C memory can be freed the moment `Copy` returns.
- Performance: ~110–170ns per struct depending on string count/length.

### 2. Direct (Primitives only)
```go
type Device struct {
    ID uint32
    // No string fields
}
```

- Fast path (~0.3ns) for pure primitive layouts.
- Any pointers copied remain owned by C – keep the backing memory alive yourself.

Custom `CStringConverter` implementations are still supported for non-UTF8 encodings or advanced scenarios (lazy conversion, arena allocators, etc.).

## Alignment and Padding

### Standard C Alignment Rules

Primitive types typically align to their size:
- uint8_t: 1-byte alignment
- uint16_t: 2-byte alignment
- uint32_t: 4-byte alignment
- uint64_t: 8-byte alignment
- pointers: 8-byte alignment (64-bit) / 4-byte (32-bit)

The package captures actual alignment at compile time via `ArchInfo`.

### CustomLayout Function

For structs that don't follow standard C alignment rules, use `CustomLayout` which generates layout information using your C compiler's actual `offsetof()` calculations:

```go
// 1. Copy this C code into your .c file:
/*
#include <stddef.h>

// Define your struct exactly as it appears in your C code
typedef struct {
    uint32_t id;
    char* name;
    float value;
} MyDevice;

// Generate layout info using actual C compiler
cgocopy_FieldInfo* getMyDeviceLayout() {
    static cgocopy_FieldInfo layout[] = {
        {.offset = offsetof(MyDevice, id), .size = sizeof(uint32_t), .type_name = "uint32_t"},
        {.offset = offsetof(MyDevice, name), .size = sizeof(char*), .type_name = "char*", .is_string = true},
        {.offset = offsetof(MyDevice, value), .size = sizeof(float), .type_name = "float"},
    };
    return layout;
}
*/

// 2. Call from Go:
layout := cgocopy.CustomLayout("MyDevice", "uint32_t", "char*", "float")
registry.MustRegister(Device{}, cSize, layout, converter)
```

**Benefits:**
- ✅ Works with `#pragma pack(1)`
- ✅ Works with `__attribute__((packed))`
- ✅ Works with third-party libraries
- ✅ Uses actual C compiler layout decisions
- ✅ Zero runtime overhead (calculated at compile time)

**When to use CustomLayout:**
- Your C struct uses custom packing (`#pragma pack`)
- You're interfacing with third-party C libraries
- AutoLayout fails due to complex alignment
- You need 100% accuracy for critical systems

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
    {Offset: 0, TypeName: "uint32_t"},  // Size and IsString auto-deduced
    {Offset: 8, TypeName: "char*"},     // IsString auto-deduced from "char*"
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

### Direct
Use when:
- Hot paths require the absolute minimum overhead
- Struct contains only primitives or fixed-size arrays
- No automatic memory management is needed

### Registry.Copy + UTF8Converter
Use when:
- Need automatic string conversion or nested metadata validation
- Working with third-party C libraries where layouts may drift
- Platform portability/defensive checks are desired
- Offloading memory ownership to Go matters

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
