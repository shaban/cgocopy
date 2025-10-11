# API Design Exploration: Maximum Inference, Zero Compromise

## Current State Analysis

### ✅ What's Working Well
- `Direct()` - 0.3ns, zero-overhead for primitives
- `AutoLayout()` - Automatic offset calculation for standard C structs
- `CustomLayout()` - Generates C code for non-standard layouts
- Nested struct support with proper validation
- String conversion via `CStringConverter`

### ❌ Current Pain Points
1. **Manual layout specification** still required for nested structs
2. **CustomLayout requires copy-paste** of generated C code
3. **String converters** must be explicitly provided
4. **Multiple APIs** for different use cases (Direct vs Registry)
5. **Registration complexity** for simple cases

## Vision: Single API with Maximum Inference

### Ideal End State
```go
// One API that handles everything automatically
cgocopy.Copy(&goStruct, cPtr)

// Infers:
// - Whether to use Direct or Registry
// - Struct layout (recursively analyzes C structs)
// - String conversion (automatic)
// - Nested struct handling
// - Platform-specific alignment
```

### Key Insights

1. **Registration time is irrelevant** - We can do expensive analysis at init time
2. **C struct analysis is possible** - We can generate C code that reports actual layouts
3. **Reflection gives us Go struct info** - We can infer much from Go struct tags/types
4. **String conversion can be automatic** - Default converter for `char*` → `string`

## Proposed API Evolution

### Phase 1: Enhanced Auto-Registration

```go
// Current: Manual registration
registry := cgocopy.NewRegistry()
layout := cgocopy.AutoLayout("uint32_t", "char*", "float")
registry.MustRegister(Device{}, cSize, layout, converter)

// Proposed: Auto-registration with inference
registry := cgocopy.NewRegistry()
registry.MustRegister(Device{}) // Infers everything from Go struct!
```

**How it works:**
- Analyze Go struct via reflection
- Generate C code automatically (like CustomLayout)
- Compile-time C analysis provides exact layout
- Automatic string converter for `char*` fields
- Recursive handling of nested structs

### Phase 2: Unified Copy API

```go
// Current: Different APIs for different cases
cgocopy.Direct(&simpleStruct, cPtr)        // Primitives only
registry.Copy(&complexStruct, cPtr)        // With strings/nesting

// Proposed: Single API with automatic dispatch
cgocopy.Copy(&anyStruct, cPtr) // Automatically chooses best method
```

**Dispatch logic:**
- If struct contains only primitives: use `Direct` (0.3ns)
- If struct has strings/nested: use `Registry` with auto-generated layout
- Runtime: identical performance to current optimized paths

### Phase 3: Recursive C Struct Analysis

**Core Innovation:** Generate comprehensive C analysis code that reports everything about C structs.

Instead of user copy-pasting, the library could:
1. Generate complete C analysis file
2. Auto-include it in build
3. Extract all struct information at compile time

**Generated C code would provide:**
```c
// For each registered struct
StructInfo getDeviceInfo() {
    return (StructInfo){
        .name = "Device",
        .size = sizeof(Device),
        .fields = {
            {"id", offsetof(Device, id), sizeof(uint32_t), FIELD_UINT32},
            {"name", offsetof(Device, name), sizeof(char*), FIELD_CHAR_PTR},
            {"value", offsetof(Device, value), sizeof(float), FIELD_FLOAT},
        },
        .field_count = 3,
        .has_strings = true,
        .alignment = _Alignof(Device),
    };
}
```

### Metadata Integration: `cgocopy_metadata.h`

To guarantee we always interrogate the caller’s toolchain (headers, compiler flags, packing pragmas), we ship a tiny helper header that the user includes next to their C structs:

```c
#include "cgocopy_metadata.h"

CGOCOPY_STRUCT(Device)
   CGOCOPY_FIELD_PRIMITIVE(Device, id, uint32_t)
   CGOCOPY_FIELD_STRING(Device, name)
   CGOCOPY_FIELD_ARRAY(Device, offsets, float, 8)
   CGOCOPY_FIELD_NESTED(Device, info, DeviceInfo)
CGOCOPY_STRUCT_END(Device)
```

The macros expand to static metadata that uses `sizeof`, `_Alignof`, and `offsetof` from the project’s own compiler invocation. At registration time the Go side calls the generated `cgocopy_get_Device_info()` symbol to retrieve exact layout/array/string facts. Benefits:

- ✅ Zero copy/paste of generated code
- ✅ Works with any compiler flags, packing pragmas, or platform quirks
- ✅ Plays nicely with build systems: users only add one header and a few macro blocks
- ✅ Backward-compatible — manual layouts remain available for edge cases

We can also ship a `cgocopy generate` CLI that emits these macro blocks automatically by parsing headers (nice-to-have, not required for the first release).

## Critical Feature Gap: Fixed-Size Arrays Inside Structs

Today `Direct()` can copy structs that contain fixed-size arrays, but the registry path cannot validate or transform them. This is a blocker for real-world interop: audio APIs, GPU descriptors, and networking stacks routinely embed `float coefficients[32]` or `char name[64]` in their structs.

### Goals

1. **Full coverage of C fixed-size arrays** for registry copies (primitives, bytes, nested structs)
2. **Optional smart handling for char buffers** (`char[32]` → Go string or `[32]byte`)
3. **Shape validation** — ensure Go field length & element type match the C declaration
4. **Preserve speed** — array copy should still be a single `memcpy` when possible

### Proposed Representation

Extend `FieldInfo` (and the C metadata) with array metadata:

```go
type FieldKind byte

const (
   FieldPrimitive FieldKind = iota
   FieldPointer
   FieldString
   FieldArray
   FieldStruct
)

type FieldInfo struct {
   Offset      uintptr
   Size        uintptr
   TypeName    string
   Kind        FieldKind
   ElemType    string  // for arrays
   ElemCount   uintptr // for arrays
   IsString    bool    // kept for backward compatibility
}
```

`cgocopy_metadata.h` will fill these fields using helpers like `CGOCOPY_FIELD_ARRAY(Device, offsets, float, 8)`. On the Go side:

- Arrays of primitives → single `memcpy`, after checking `GoField.Len()` and element size
- Arrays of nested structs → loop and recursively call `Copy`
- `char[N]` buffers → two modes: as `[N]byte` (default) or opt-in eager string conversion via `CGOCOPY_FIELD_CSTRING(Device, name, 64)`

### Copy-Time Behaviour

1. **Zero-copy path** when array element type is primitive and Go field is `[N]T`
2. **Recursive copy** for nested struct elements (respecting previously registered mappings)
3. **Optional converters** for byte buffers → Go `string`/`[]byte`
4. **Validation** ensures Go array length equals `ElemCount` and element size matches `ElemType`

### Roadmap Additions

- Update metadata header macros to emit `FieldKind`, element type, and count
- Teach `Registry.Register` to understand the enriched `FieldInfo`
- Extend `Copy` fast path: treat “no nested + no strings + primitive arrays” as memcpy-compatible
- Add tests covering:
  - Primitive arrays (`float[8]`, `uint8_t[16]`)
  - Char buffers both as `[N]byte` and eager strings
  - Nested struct arrays (copy + validation)
- Document migration path for existing manual layouts (old API continues to work; `Kind` defaults to primitive when omitted)

## Technical Feasibility

### ✅ Possible with Current Architecture

1. **Enhanced Reflection Analysis**
   - Go struct tags for C type hints: `cgocopy:"uint32_t"`
   - Automatic string converter generation
   - Nested struct recursive analysis

2. **Build-Time C Code Generation**
   - Generate complete C analysis file
   - Auto-integrate into CGO build process
   - Extract all layout info at compile time

3. **Automatic Converter Generation**
   - Default `char*` → `string` converter
   - Custom converter support for edge cases
   - Memory management automation

### 🤔 Open Questions

1. **Build Integration Complexity**
   - How to auto-include generated C code in build?
   - Cross-platform C compilation issues?

2. **Go Struct Tag Requirements**
   - Should users add `cgocopy:"uint32_t"` tags?
   - Or can we infer from Go types + naming conventions?

3. **Fallback Mechanisms**
   - What if C analysis fails?
   - Graceful degradation to manual methods?

## Implementation Strategy

### Step 1: Enhanced Auto-Registration (Week 1-2)
- Ship `cgocopy_metadata.h` macros and runtime loader
- Extend `MustRegister()` to accept just the Go struct (plus optional struct tags)
- Generate C analysis metadata automatically (either via macros or `cgocopy generate`)
- Basic string conversion automation (default converter for `char*` fields)

### Step 2: Unified Copy API (Week 3)
- Single `cgocopy.Copy()` function
- Automatic dispatch based on struct analysis
- Maintain performance equivalence

### Step 3: Recursive Analysis & Array Support (Week 4-6)
- Full C struct recursion (nested struct auto-registration)
- First-class fixed-size array support (primitive, strings, nested structs)
- Build integration for generated/inline metadata code

### Step 4: Polish & Documentation (Week 7-8)
- Comprehensive tests
- Migration guides
- Performance validation

## Success Metrics

### ✅ Correctness
- Zero incorrect copies (current: 100% correct)
- Handles all C struct variations
- Platform-independent results

### ✅ Simplicity
- Single API call for all use cases
- Zero manual layout specification
- Automatic string handling

### ✅ Performance
- Direct: 0.3ns (unchanged)
- Registry: ~110-170ns (unchanged)
- Registration: <100ms (acceptable)

### ✅ Compatibility
- Backward compatible with existing code
- Gradual migration path
- No breaking changes

## Risk Assessment

### Low Risk ✅
- Enhanced reflection analysis
- Automatic converter generation
- Unified API dispatch

### Medium Risk 🟡
- Build-time C code generation
- Cross-platform C compilation
- Complex nested struct recursion

### Mitigation Strategies
- Fallback to manual methods if auto-generation fails
- Comprehensive test coverage
- User feedback during development

## Conclusion

**This is absolutely feasible and would dramatically simplify the API surface while maintaining correctness and performance.**

The key insight: *Registration time being irrelevant means we can do arbitrarily complex analysis at build/init time to enable dead-simple runtime APIs.*

The result would be the most developer-friendly C/Go interop library possible, with one API that "just works" for any struct complexity.</content>
<parameter name="filePath">/Volumes/Space/Code/cgocopy/specs/API.md