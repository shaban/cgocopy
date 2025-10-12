# cgocopy v2 API Improvements

## Overview

This document specifies the refined improvements for cgocopy v2, focusing on:
- Simplified C macros with auto-detection
- Thread-safe registry without explicit finalization
- Precompilation-based workflow
- Tagged struct support for idiomatic Go
- Fast-path copying for primitive structs

## Design Principles

1. **Explicit over implicit**: Require precompilation, no lazy registration
2. **Performance**: Move work to setup time, optimize runtime copies
3. **Safety**: Thread-safe by design with proper locking
4. **Ergonomics**: Simplify C macros, support Go conventions via tags
5. **Compatibility**: Implement in `pkg/cgocopy2` alongside existing `pkg/cgocopy`

---

## 1. Simplified C Macros

### Current State (cgocopy v1)

```c
CGOCOPY_STRUCT_BEGIN(User)
    CGOCOPY_FIELD_PRIMITIVE(User, id, uint32_t),
    CGOCOPY_FIELD_STRING(User, email),
    CGOCOPY_FIELD_STRUCT(User, details, UserDetails),
    CGOCOPY_FIELD_PRIMITIVE(User, account_balance, double)
CGOCOPY_STRUCT_END(User)
```

**Pain points:**
- Verbose: BEGIN/END, repeated struct name
- Manual type specification: `FIELD_PRIMITIVE`, `FIELD_STRING`, etc.
- Error-prone: Easy to mismatch type declarations

### Improved State (cgocopy v2)

```c
CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, id),
    CGOCOPY_FIELD(User, email),
    CGOCOPY_FIELD(User, details),
    CGOCOPY_FIELD(User, account_balance)
)
```

**Improvements:**
- Single macro: `CGOCOPY_STRUCT`
- Auto-detection: `CGOCOPY_FIELD` uses `_Generic` to detect field types
- Less boilerplate: No BEGIN/END, less repetition
- Fewer errors: Type is auto-detected from C struct definition

### Implementation

Add to `native/cgocopy_metadata.h`:

```c
// Auto-detection helper for field kinds
#define CGOCOPY_FIELD(struct_type, field_name) \
    _Generic(((struct_type*)0)->field_name, \
        char*: CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_STRING_KIND, struct_type, field_name, "char*", NULL, 0, true), \
        const char*: CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_STRING_KIND, struct_type, field_name, "char*", NULL, 0, true), \
        default: CGOCOPY_INTERNAL_FIELD_AUTO(struct_type, field_name))

// Helper to auto-detect non-string types
#define CGOCOPY_INTERNAL_FIELD_AUTO(struct_type, field_name) \
    CGOCOPY_INTERNAL_FIELD_INIT( \
        CGOCOPY_INTERNAL_IS_STRUCT(struct_type, field_name) ? CGOCOPY_FIELD_STRUCT_KIND : \
        CGOCOPY_INTERNAL_IS_ARRAY(struct_type, field_name) ? CGOCOPY_FIELD_ARRAY_KIND : \
        CGOCOPY_FIELD_PRIMITIVE_KIND, \
        struct_type, field_name, \
        CGOCOPY_INTERNAL_TYPE_STRING(struct_type, field_name), \
        CGOCOPY_INTERNAL_ELEM_TYPE(struct_type, field_name), \
        CGOCOPY_INTERNAL_ELEM_COUNT(struct_type, field_name), \
        false)

// Simplified single-macro struct definition
#define CGOCOPY_STRUCT(struct_type, ...) \
    static const cgocopy_field_info cgocopy_fields_##struct_type[] = { \
        __VA_ARGS__ \
    }; \
    static const cgocopy_struct_info cgocopy_struct_info_##struct_type = { \
        .name = #struct_type, \
        .size = sizeof(struct_type), \
        .alignment = _Alignof(struct_type), \
        .field_count = sizeof(cgocopy_fields_##struct_type) / sizeof(cgocopy_field_info), \
        .fields = cgocopy_fields_##struct_type, \
    }; \
    static cgocopy_struct_registry_node cgocopy_registry_node_##struct_type; \
    static void cgocopy_register_##struct_type(void) __attribute__((constructor)); \
    static void cgocopy_register_##struct_type(void) { \
        cgocopy_registry_node_##struct_type.info = &cgocopy_struct_info_##struct_type; \
        cgocopy_registry_node_##struct_type.next = NULL; \
        cgocopy_registry_add(&cgocopy_registry_node_##struct_type); \
    }
```

**Note:** Keep old macros for backward compatibility in v1.

---

## 2. Thread-Safe Registry (Remove Finalize Pattern)

### Current State (cgocopy v1)

```go
// Setup phase
cgocopy.RegisterStruct[User](converter)
cgocopy.RegisterStruct[Address](converter)
cgocopy.Finalize() // Explicit finalization required

// Runtime - Finalize() must be called first
cgocopy.Copy(&user, cPtr)
```

**Pain points:**
- Two-phase initialization (register + finalize)
- Easy to forget Finalize() call
- Race conditions if Finalize() called at wrong time
- Can't register after Finalize()

### Improved State (cgocopy v2)

```go
// Setup phase - no finalize needed
cgocopy2.Precompile[User](converter)
cgocopy2.Precompile[Address](converter)
// Can call Precompile concurrently - thread-safe

// Runtime - immediately ready
cgocopy2.Copy(&user, cPtr)
```

**Improvements:**
- Single-step registration: `Precompile` does everything
- Thread-safe by design: Uses `sync.RWMutex`
- No finalization step: Registry always ready
- Can precompile at any time (even concurrently)

### Implementation

```go
package cgocopy2

import (
    "reflect"
    "sync"
    "unsafe"
)

// Registry is thread-safe and always ready for use
type Registry struct {
    mu       sync.RWMutex
    mappings map[reflect.Type]*StructMapping
}

var defaultRegistry = &Registry{
    mappings: make(map[reflect.Type]*StructMapping),
}

// Precompile registers T and all nested structs recursively.
// Thread-safe and idempotent - can be called multiple times.
func Precompile[T any](converter ...CStringConverter) error {
    var zero T
    return defaultRegistry.precompile(reflect.TypeOf(zero), converter...)
}

func (r *Registry) precompile(t reflect.Type, converter ...CStringConverter) error {
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    if t.Kind() != reflect.Struct {
        return ErrNotAStructType
    }
    
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Check if already compiled (idempotent)
    if _, exists := r.mappings[t]; exists {
        return nil
    }
    
    // Lookup C metadata
    metadata, err := lookupStructMetadata(t.Name())
    if err != nil {
        return err
    }
    
    // Build and cache mapping
    conv := getConverter(converter...)
    mapping, err := r.buildMapping(t, metadata, conv)
    if err != nil {
        return err
    }
    
    r.mappings[t] = mapping
    
    // Recursively precompile nested structs
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        if field.Type.Kind() == reflect.Struct {
            if err := r.precompile(field.Type, converter...); err != nil {
                return err
            }
        }
        if field.Type.Kind() == reflect.Array || field.Type.Kind() == reflect.Slice {
            if field.Type.Elem().Kind() == reflect.Struct {
                if err := r.precompile(field.Type.Elem(), converter...); err != nil {
                    return err
                }
            }
        }
    }
    
    return nil
}

// Copy uses precompiled mappings (thread-safe read)
func Copy[T any](dst *T, cPtr unsafe.Pointer) error {
    if dst == nil {
        return ErrNilDestination
    }
    if cPtr == nil {
        return ErrNilSourcePointer
    }
    
    t := reflect.TypeOf(*dst)
    
    defaultRegistry.mu.RLock()
    mapping, exists := defaultRegistry.mappings[t]
    defaultRegistry.mu.RUnlock()
    
    if !exists {
        return ErrNotPrecompiled
    }
    
    return copyWithMapping(defaultRegistry, mapping, dst, cPtr)
}

// Reset clears registry (testing only)
func Reset() {
    defaultRegistry.mu.Lock()
    defaultRegistry.mappings = make(map[reflect.Type]*StructMapping)
    defaultRegistry.mu.Unlock()
}
```

---

## 3. Tagged Struct Support

### Problem

C and Go naming conventions often differ:

```c
// C side - snake_case
typedef struct {
    uint32_t user_id;
    char* email_address;
} user_account_t;
```

```go
// Current - forced to match C names (ugly in Go)
type UserAccount struct {
    User_id       uint32
    Email_address string
}
```

### Solution

Support struct tags for field mapping:

```go
// Improved - idiomatic Go with tags
type UserAccount struct {
    UserID uint32  `cgocopy:"user_id"`
    Email  string  `cgocopy:"email_address"`
}

// Skip fields
type UserAccount struct {
    UserID   uint32  `cgocopy:"user_id"`
    Password string  `cgocopy:"-"`        // Not copied from C
}
```

### Implementation

Modify `buildMapping` to check tags:

```go
func (r *Registry) buildMapping(goType reflect.Type, metadata StructMetadata, conv CStringConverter) (*StructMapping, error) {
    mapping := &StructMapping{
        CSize:    metadata.Size,
        GoSize:   goType.Size(),
        Fields:   make([]FieldMapping, 0),
        // ...
    }
    
    for i := 0; i < goType.NumField(); i++ {
        goField := goType.Field(i)
        
        // Check for cgocopy tag
        tag := goField.Tag.Get("cgocopy")
        if tag == "-" {
            continue // Skip this field
        }
        
        // Use tag as C field name, or default to Go field name
        cFieldName := goField.Name
        if tag != "" {
            cFieldName = tag
        }
        
        // Find matching C field by name
        cField, err := findCFieldByName(metadata.Fields, cFieldName)
        if err != nil {
            return nil, fmt.Errorf("field %q: %w", goField.Name, err)
        }
        
        // Validate and build field mapping
        fieldMapping, err := r.buildFieldMapping(goField, cField, conv)
        if err != nil {
            return nil, err
        }
        
        mapping.Fields = append(mapping.Fields, fieldMapping)
    }
    
    return mapping, nil
}

func findCFieldByName(cFields []FieldInfo, name string) (FieldInfo, error) {
    for _, field := range cFields {
        // Match by TypeName which contains field name in metadata
        // This may need adjustment based on actual metadata structure
        if /* field name matches */ {
            return field, nil
        }
    }
    return FieldInfo{}, fmt.Errorf("C field %q not found", name)
}
```

**Critical:** Offset calculation must use C field order, not Go field order:

```go
// Build offset mapping based on C metadata order
type FieldMapping struct {
    COffset  uintptr // From C metadata
    GoOffset uintptr // From Go reflect
    // ...
}
```

---

## 4. Fast-Path Copying

### Use Case

Primitive-only structs can use direct `memcpy`:

```go
type SimpleStruct struct {
    ID    uint32
    Value float64
    Count int64
}

// Can use fast copy (no pointers, strings, or nested structs)
cgocopy2.FastCopy(&simple, cPtr)
```

### Implementation

```go
// FastCopy performs direct memory copy for primitive-only structs
func FastCopy[T any](dst *T, cPtr unsafe.Pointer) error {
    if dst == nil {
        return ErrNilDestination
    }
    if cPtr == nil {
        return ErrNilSourcePointer
    }
    
    t := reflect.TypeOf(*dst)
    
    defaultRegistry.mu.RLock()
    mapping, exists := defaultRegistry.mappings[t]
    defaultRegistry.mu.RUnlock()
    
    if !exists {
        return ErrNotPrecompiled
    }
    
    if !mapping.CanFastPath {
        return ErrCannotUseFastPath
    }
    
    // Direct memory copy
    *dst = *(*T)(cPtr)
    return nil
}
```

**Detection in buildMapping:**

```go
func (r *Registry) buildMapping(...) (*StructMapping, error) {
    // ...
    
    canFastPath := true
    for _, field := range metadata.Fields {
        switch field.Kind {
        case FieldString, FieldPointer, FieldStruct, FieldArray:
            canFastPath = false
        }
    }
    
    mapping.CanFastPath = canFastPath
    // ...
}
```

---

## 5. Validation Helper

```go
// ValidationReport contains compatibility check results
type ValidationReport struct {
    TypeName         string
    IsPrecompiled    bool
    CanUseFastPath   bool
    SizeMatch        bool
    CSize            uintptr
    GoSize           uintptr
    FieldCount       int
    FieldMismatches  []FieldMismatch
    Errors           []string
}

type FieldMismatch struct {
    GoFieldName   string
    CFieldName    string
    Issue         string
    GoType        string
    CType         string
    GoSize        uintptr
    CSize         uintptr
}

// ValidateStruct checks struct compatibility without precompiling
func ValidateStruct[T any]() (*ValidationReport, error) {
    var zero T
    t := reflect.TypeOf(zero)
    
    report := &ValidationReport{
        TypeName: t.Name(),
    }
    
    // Check if already precompiled
    defaultRegistry.mu.RLock()
    _, exists := defaultRegistry.mappings[t]
    defaultRegistry.mu.RUnlock()
    
    report.IsPrecompiled = exists
    
    // Lookup metadata
    metadata, err := lookupStructMetadata(t.Name())
    if err != nil {
        report.Errors = append(report.Errors, err.Error())
        return report, err
    }
    
    report.CSize = metadata.Size
    report.GoSize = t.Size()
    report.SizeMatch = (metadata.Size == t.Size())
    
    // Check field compatibility
    // ... detailed field-by-field validation
    
    return report, nil
}
```

---

## API Summary

### cgocopy2 Public API

```go
// Registration
func Precompile[T any](converter ...CStringConverter) error
func Reset() // Testing only

// Copying
func Copy[T any](dst *T, cPtr unsafe.Pointer) error
func FastCopy[T any](dst *T, cPtr unsafe.Pointer) error

// Validation
func ValidateStruct[T any]() (*ValidationReport, error)

// Types
type CStringConverter interface {
    CStringToGo(ptr unsafe.Pointer) string
}

// Errors
var (
    ErrNotPrecompiled     = errors.New("cgocopy2: struct not precompiled")
    ErrCannotUseFastPath  = errors.New("cgocopy2: struct contains non-primitive fields")
    ErrNilDestination     = errors.New("cgocopy2: destination is nil")
    ErrNilSourcePointer   = errors.New("cgocopy2: source pointer is nil")
    ErrNotAStructType     = errors.New("cgocopy2: type is not a struct")
    ErrMetadataMissing    = errors.New("cgocopy2: C metadata not found")
    ErrLayoutMismatch     = errors.New("cgocopy2: C and Go struct layouts incompatible")
)
```

### Usage Example

```go
package main

import (
    cgocopy2 "github.com/shaban/cgocopy/pkg/cgocopy2"
)

func init() {
    // Precompile all types at startup
    if err := cgocopy2.Precompile[User](); err != nil {
        panic(err)
    }
    // Nested structs auto-precompiled
}

func main() {
    cUsersPtr, count := createUsers()
    defer freeUsers(cUsersPtr, count)
    
    users := make([]User, count)
    for i := range users {
        if err := cgocopy2.Copy(&users[i], userAt(cUsersPtr, i)); err != nil {
            panic(err)
        }
    }
    
    // Use fast path for simple structs
    var stats Stats
    if err := cgocopy2.FastCopy(&stats, cStatsPtr); err != nil {
        // Fall back to normal copy if needed
        cgocopy2.Copy(&stats, cStatsPtr)
    }
}
```

---

## Migration Path

1. Implement cgocopy2 in `pkg/cgocopy2/`
2. Keep cgocopy v1 in `pkg/cgocopy/` untouched
3. Users can import both for gradual migration
4. Benchmark cgocopy vs cgocopy2
5. Eventually deprecate v1 (major version bump)

---

## Testing Strategy

See `IMPLEMENTATION_PLAN.md` for detailed testing approach.
