# Registry API Guide

The Registry provides validated, safe copying between C and Go structs with automatic string conversion and nested struct support. Use it when yWhen DirectCopy becomes insufficient:

```go
// Before (Direct - primitives only)
type Device struct {
    ID   uint32
    Size uint32
}
Direct(&device, cPtr)

// After (Registry - with strings)
type Device struct {
    ID   uint32
    Name string  // Now supports strings!
    Size uint32
}
layout := cgocopy.AutoLayout("uint32_t", "char*", "uint32_t")
registry.MustRegister(Device{}, cSize, layout, stringConverter)
registry.Copy(&device, cPtr)  // Strings converted automatically
```st primitive copying.

## Quick Start

```go
import "github.com/shaban/cgocopy"

// 1. Create registry
registry := cgocopy.NewRegistry()

// 2. Register your struct (choose one method)
layout := cgocopy.AutoLayout("uint32_t", "char*", "float")
registry.MustRegister(Device{}, cSize, layout, stringConverter)

// 3. Copy safely
var device Device
err := registry.Copy(&device, unsafe.Pointer(cDevice))
```

## When to Use Registry

Choose Registry when you need:
- ✅ **String conversion** (char* → string)
- ✅ **Validation** of struct compatibility
- ✅ **Automatic memory management** for strings
- ✅ **Nested structs** with complex layouts
- ✅ **Platform portability** across different architectures
- ✅ **Third-party C libraries** with unknown layouts

## Registration Methods

### Method 1: AutoLayout (Recommended)

**Best for:** Standard C structs with normal alignment.

```go
layout := cgocopy.AutoLayout("uint32_t", "char*", "float")
registry.MustRegister(Device{}, cSize, layout, stringConverter)
```

**Pros:** Zero C code, automatic offset calculation, 100% accurate for standard layouts.

**Cons:** Doesn't work with `#pragma pack` or custom alignment.

### Method 2: Manual Layout with Auto-Deduction

**Best for:** When you need explicit control but want some automation.

```go
layout := []cgocopy.FieldInfo{
    {Offset: 0, TypeName: "uint32_t"},     // Size auto-deduced
    {Offset: 8, TypeName: "char*"},        // IsString auto-deduced
    {Offset: 16, TypeName: "float"},
}
registry.MustRegister(Device{}, cSize, layout, stringConverter)
```

**Pros:** Explicit offsets, automatic size/string detection.

**Cons:** Still need to calculate offsets manually.

### Method 3: Full Manual Layout

**Best for:** Custom alignment, third-party libraries, complex layouts.

```go
layout := []cgocopy.FieldInfo{
    {Offset: 0, Size: 4, TypeName: "uint32_t", IsString: false},
    {Offset: 8, Size: 8, TypeName: "char*", IsString: true},
    {Offset: 16, Size: 4, TypeName: "float", IsString: false},
}
registry.MustRegister(Device{}, cSize, layout, stringConverter)
```

**Pros:** Full control, works with any layout.

**Cons:** Most verbose, error-prone.

## String Conversion

Registry automatically handles char* → string conversion:

```go
type SimpleConverter struct{}
func (c SimpleConverter) CStringToGo(ptr unsafe.Pointer) string {
    if ptr == nil { return "" }
    return C.GoString((*C.char)(ptr))
}

converter := SimpleConverter{}
registry.MustRegister(Device{}, cSize, layout, converter)
```

**Key Benefit:** C memory can be freed immediately after Copy().

## Nested Structs

Registry handles nested structs automatically:

```go
// Register innermost first
registry.MustRegister(InnerStruct{}, innerSize, innerLayout, nil)
registry.MustRegister(OuterStruct{}, outerSize, outerLayout, stringConverter)

// Copy handles nesting recursively
var outer OuterStruct
registry.Copy(&outer, cPtr)  // Inner structs copied automatically
```

## Error Handling

Registry validates everything at registration time:

```go
// Method 1: MustRegister panics on error (use in init())
registry.MustRegister(Device{}, cSize, layout, converter)

// Method 2: Register returns error (use in main())
if err := registry.Register(reflect.TypeOf(Device{}), cSize, layout, converter); err != nil {
    log.Fatal("Registration failed:", err)
}
```

Common errors:
- Field count mismatch
- Size incompatibility
- Type mismatches
- Missing string converters

## Performance

| Operation | Time | When |
|-----------|------|------|
| Registration | ~127ns | Once at startup |
| AutoLayout() | ~127ns | Once at startup |
| Copy (primitives) | ~110ns | Per copy operation |
| Copy (with strings) | ~170ns | Per copy operation |

**Note:** Registration overhead is negligible since it happens once at startup.

## Platform Safety

Registry works across platforms by:
- Using captured architecture info at compile time
- Validating layouts before runtime
- Handling endianness and alignment differences
- Supporting both 32-bit and 64-bit architectures

## Best Practices

### ✅ Do:
- Use `AutoLayout()` for standard C structs
- Register in `init()` functions with `MustRegister()`
- Handle registration errors in development/testing
- Free C memory immediately after Registry.Copy()

### ❌ Don't:
- Use Registry for primitives-only structs (use Direct instead)
- Register the same type multiple times
- Forget string converters for char* fields
- Use Registry in performance-critical hot paths (consider StringPtr + Direct)

## Migration from Direct

When DirectCopy becomes insufficient:

```go
// Before (DirectCopy - primitives only)
type Device struct {
    ID   uint32
    Size uint32
}
DirectCopy(&device, cPtr)

// After (Registry - with strings)
type Device struct {
    ID   uint32
    Name string  // Now supports strings!
    Size uint32
}
layout := cgocopy.AutoLayout("uint32_t", "char*", "uint32_t")
registry.MustRegister(Device{}, cSize, layout, converter)
registry.Copy(&device, cPtr)  // Strings converted automatically
```

## Complete Example

```go
package main

import (
    "unsafe"
    "github.com/shaban/cgocopy"
)

/*
#include <stdint.h>
typedef struct {
    uint32_t id;
    char* name;
    float value;
} CDevice;
*/
import "C"

type Device struct {
    ID    uint32
    Name  string
    Value float32
}

type StringConverter struct{}
func (c StringConverter) CStringToGo(ptr unsafe.Pointer) string {
    if ptr == nil { return "" }
    return C.GoString((*C.char)(ptr))
}

func init() {
    registry := cgocopy.NewRegistry()
    layout := cgocopy.AutoLayout("uint32_t", "char*", "float")
    registry.MustRegister(Device{}, unsafe.Sizeof(C.CDevice{}), layout, StringConverter{})
}

func main() {
    // Your registry is ready to use!
    var device Device
    registry.Copy(&device, cPtr)
    // device.Name is now a safe Go string
}
```

## Troubleshooting

### "Field count mismatch"
- Check that your Go struct has the same number of fields as your C layout

### "Size incompatibility"
- Verify field sizes match between C and Go types
- Use `AutoLayout()` to avoid manual size errors

### "Missing string converter"
- Add a `CStringConverter` implementation for any `char*` fields
- Ensure `IsString: true` for string fields in manual layouts

### AutoLayout doesn't work
- Your C struct might use custom alignment (`#pragma pack`)
- Switch to manual layout with explicit `offsetof()` calls

## API Reference

```go
// Core functions
func NewRegistry() *Registry
func (r *Registry) Register(goType reflect.Type, cSize uintptr, layout []FieldInfo, converter ...CStringConverter) error
func (r *Registry) MustRegister(goType interface{}, cSize uintptr, layout []FieldInfo, converter ...CStringConverter)
func (r *Registry) Copy(dst interface{}, cPtr unsafe.Pointer) error

// Layout helpers
func AutoLayout(typeNames ...string) []FieldInfo

// Types
type FieldInfo struct {
    Offset   uintptr  // Byte offset in C struct
    Size     uintptr  // Size in bytes (0 = auto-deduce)
    TypeName string   // C type name ("uint32_t", "char*", etc.)
    IsString bool     // true for char* fields (auto-deduced if omitted)
}

type CStringConverter interface {
    CStringToGo(ptr unsafe.Pointer) string
}
```
