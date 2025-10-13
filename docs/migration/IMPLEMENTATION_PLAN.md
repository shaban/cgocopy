# cgocopy2 Implementation Plan

## Overview

This document outlines the step-by-step implementation of cgocopy2 in `pkg/cgocopy2/` with comprehensive testing at each stage.

## Key Implementation Details

### Directory Structure
```
pkg/
├── cgocopy/           # v1 implementation (unchanged)
└── cgocopy2/          # v2 implementation (new)

native/
├── cgocopy_metadata.h      # v1 macros (old BEGIN/END style)
└── metadata_registry.c     # v1 registry

native2/                     # NEW for v2
├── cgocopy_metadata.h      # v2 macros (simplified CGOCOPY_STRUCT style)
└── metadata_registry.c     # v2 registry (same implementation as v1)
```

### Macro Syntax Comparison

**v1 (native/cgocopy_metadata.h):**
```c
CGOCOPY_STRUCT_BEGIN(User)
    CGOCOPY_FIELD_PRIMITIVE(User, id, uint32_t),
    CGOCOPY_FIELD_STRING(User, email),
    CGOCOPY_FIELD_STRUCT(User, details, UserDetails)
CGOCOPY_STRUCT_END(User)
```

**v2 (native2/cgocopy_metadata.h):**
```c
CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, id),
    CGOCOPY_FIELD(User, email),
    CGOCOPY_FIELD(User, details)
)
```

**Key improvements:**
- Single macro `CGOCOPY_STRUCT` (no BEGIN/END)
- Auto-detection via `_Generic` (no need to specify PRIMITIVE/STRING/STRUCT)
- Less repetition, more readable
- Fewer errors possible

---

## Phase 1: Project Setup & Core Types

### Tasks

- [ ] 1.1: Create `pkg/cgocopy2/` directory structure
- [ ] 1.2: Copy base types from cgocopy to cgocopy2
- [ ] 1.3: Create error definitions
- [ ] 1.4: Set up basic test infrastructure

### Files to Create

```
pkg/cgocopy2/
├── types.go           # FieldInfo, StructMapping, etc.
├── errors.go          # Error definitions
├── arch_info.go       # Architecture detection (copy from v1)
├── cstringptr.go      # String converter (copy from v1)
├── metadata_cgo.go    # C metadata lookup (copy from v1)
└── types_test.go      # Basic type tests
```

### Implementation Details

#### 1.1 Directory Structure
```bash
mkdir -p pkg/cgocopy2
```

#### 1.2 types.go
```go
package cgocopy2

import (
    "reflect"
    "sync"
)

// CStringConverter converts C strings to Go strings
type CStringConverter interface {
    CStringToGo(ptr unsafe.Pointer) string
}

// FieldKind classifies field types
type FieldKind uint8

const (
    FieldPrimitive FieldKind = iota
    FieldPointer
    FieldString
    FieldArray
    FieldStruct
)

// FieldInfo describes a C field from metadata
type FieldInfo struct {
    Offset    uintptr
    Size      uintptr
    TypeName  string
    Kind      FieldKind
    ElemType  string
    ElemCount uintptr
    IsString  bool
}

// StructMetadata describes a complete C struct
type StructMetadata struct {
    Name      string
    Size      uintptr
    Alignment uintptr
    Fields    []FieldInfo
}

// FieldMapping maps a C field to its Go counterpart
type FieldMapping struct {
    COffset           uintptr
    GoOffset          uintptr
    Size              uintptr
    CType             string
    GoType            reflect.Type
    Kind              FieldKind
    IsNested          bool
    IsString          bool
    IsArray           bool
    IsSlice           bool
    ArrayLen          uintptr
    ArrayElemSize     uintptr
    ArrayElemGoType   reflect.Type
    ArrayElemIsNested bool
    NestedMapping     *StructMapping
    ArrayElemMapping  *StructMapping
}

// StructMapping stores validated C-to-Go struct mapping
type StructMapping struct {
    CSize           uintptr
    GoSize          uintptr
    Fields          []FieldMapping
    CTypeName       string
    GoTypeName      string
    StringConverter CStringConverter
    CanFastPath     bool
}

// Registry holds thread-safe struct mappings
type Registry struct {
    mu       sync.RWMutex
    mappings map[reflect.Type]*StructMapping
}
```

#### 1.3 errors.go
```go
package cgocopy2

import "errors"

var (
    ErrNotPrecompiled    = errors.New("cgocopy2: struct not precompiled")
    ErrCannotUseFastPath = errors.New("cgocopy2: struct contains non-primitive fields")
    ErrNilDestination    = errors.New("cgocopy2: destination is nil")
    ErrNilSourcePointer  = errors.New("cgocopy2: source pointer is nil")
    ErrNotAStructType    = errors.New("cgocopy2: type is not a struct")
    ErrAnonymousStruct   = errors.New("cgocopy2: anonymous structs not supported")
    ErrMetadataMissing   = errors.New("cgocopy2: C metadata not found")
    ErrLayoutMismatch    = errors.New("cgocopy2: C and Go struct layouts incompatible")
    ErrNilRegistry       = errors.New("cgocopy2: registry is nil")
    ErrFieldNotFound     = errors.New("cgocopy2: C field not found")
)
```

### Testing Phase 1

#### Test Setup
```go
// pkg/cgocopy2/types_test.go
package cgocopy2

import "testing"

func TestErrorDefinitions(t *testing.T) {
    errors := []error{
        ErrNotPrecompiled,
        ErrCannotUseFastPath,
        ErrNilDestination,
        ErrNilSourcePointer,
        ErrNotAStructType,
        ErrMetadataMissing,
    }
    
    for _, err := range errors {
        if err == nil {
            t.Error("error should not be nil")
        }
        if err.Error() == "" {
            t.Error("error message should not be empty")
        }
    }
}

func TestRegistryCreation(t *testing.T) {
    reg := &Registry{
        mappings: make(map[reflect.Type]*StructMapping),
    }
    
    if reg.mappings == nil {
        t.Error("mappings should be initialized")
    }
}
```

#### Validation
```bash
cd /Volumes/Space/Code/cgocopy
go test ./pkg/cgocopy2/... -v
```

**Success Criteria:**
- ✅ All type definitions compile
- ✅ Basic tests pass
- ✅ No import cycles

---

## Phase 2: Registry & Precompile Implementation

### Tasks

- [ ] 2.1: Implement basic Registry methods
- [ ] 2.2: Implement metadata lookup wrapper
- [ ] 2.3: Implement Precompile function
- [ ] 2.4: Add recursive nested struct handling
- [ ] 2.5: Comprehensive precompile tests

### Files to Create/Modify

```
pkg/cgocopy2/
├── registry.go        # Registry implementation
├── precompile.go      # Precompile function
├── metadata.go        # Metadata lookup helpers
├── registry_test.go   # Registry tests
└── precompile_test.go # Precompile tests
```

### Implementation Details

#### 2.1 registry.go
```go
package cgocopy2

import (
    "fmt"
    "reflect"
    "sync"
)

var defaultRegistry = &Registry{
    mappings: make(map[reflect.Type]*StructMapping),
}

// Reset clears the default registry (testing only)
func Reset() {
    defaultRegistry.mu.Lock()
    defaultRegistry.mappings = make(map[reflect.Type]*StructMapping)
    defaultRegistry.mu.Unlock()
}

// GetMapping retrieves a precompiled mapping (thread-safe read)
func (r *Registry) GetMapping(t reflect.Type) (*StructMapping, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    mapping, exists := r.mappings[t]
    return mapping, exists
}

// setMapping stores a mapping (must be called with write lock held)
func (r *Registry) setMapping(t reflect.Type, mapping *StructMapping) {
    r.mappings[t] = mapping
}

// hasMapping checks if type is registered (must be called with lock held)
func (r *Registry) hasMapping(t reflect.Type) bool {
    _, exists := r.mappings[t]
    return exists
}
```

#### 2.2 precompile.go
```go
package cgocopy2

import (
    "fmt"
    "reflect"
)

// Precompile registers T and all nested structs recursively.
// Thread-safe and idempotent.
func Precompile[T any](converter ...CStringConverter) error {
    var zero T
    t := reflect.TypeOf(zero)
    
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    
    if t.Kind() != reflect.Struct {
        return ErrNotAStructType
    }
    
    if t.Name() == "" {
        return ErrAnonymousStruct
    }
    
    conv := getConverter(converter...)
    visited := make(map[reflect.Type]bool)
    
    return defaultRegistry.precompileType(t, conv, visited)
}

func (r *Registry) precompileType(t reflect.Type, conv CStringConverter, visited map[reflect.Type]bool) error {
    // Acquire write lock
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Check if already compiled (idempotent)
    if r.hasMapping(t) {
        return nil
    }
    
    // Prevent infinite recursion
    if visited[t] {
        return nil
    }
    visited[t] = true
    
    // Lookup C metadata
    metadata, err := lookupStructMetadata(t.Name())
    if err != nil {
        return fmt.Errorf("%w for type %s", ErrMetadataMissing, t.Name())
    }
    
    // Build mapping
    mapping, err := r.buildMapping(t, metadata, conv)
    if err != nil {
        return fmt.Errorf("build mapping for %s: %w", t.Name(), err)
    }
    
    r.setMapping(t, mapping)
    
    // Recursively precompile nested types
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fieldType := field.Type
        
        // Handle struct fields
        if fieldType.Kind() == reflect.Struct {
            if err := r.precompileType(fieldType, conv, visited); err != nil {
                return err
            }
        }
        
        // Handle array/slice of structs
        if fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice {
            elemType := fieldType.Elem()
            if elemType.Kind() == reflect.Struct {
                if err := r.precompileType(elemType, conv, visited); err != nil {
                    return err
                }
            }
        }
    }
    
    return nil
}

func getConverter(converters ...CStringConverter) CStringConverter {
    if len(converters) > 0 && converters[0] != nil {
        return converters[0]
    }
    return DefaultCStringConverter
}
```

#### 2.3 buildMapping stub (to be completed in Phase 3)
```go
func (r *Registry) buildMapping(t reflect.Type, metadata StructMetadata, conv CStringConverter) (*StructMapping, error) {
    // Stub implementation - to be completed in Phase 3
    return &StructMapping{
        CSize:      metadata.Size,
        GoSize:     t.Size(),
        CTypeName:  metadata.Name,
        GoTypeName: t.Name(),
        StringConverter: conv,
        CanFastPath: false,
        Fields: []FieldMapping{},
    }, nil
}
```

### Testing Phase 2

#### 2.4 precompile_test.go
```go
package cgocopy2

import (
    "testing"
)

// Simple test struct (no C metadata yet)
type SimpleStruct struct {
    ID    uint32
    Value float64
}

func TestPrecompile_NotAStruct(t *testing.T) {
    Reset()
    
    // Should fail for non-struct types
    err := Precompile[int]()
    if err != ErrNotAStructType {
        t.Errorf("expected ErrNotAStructType, got %v", err)
    }
}

func TestPrecompile_Idempotent(t *testing.T) {
    Reset()
    defer Reset()
    
    // Note: This will fail until we have actual C metadata
    // For now, test the error path
    
    err1 := Precompile[SimpleStruct]()
    if err1 == nil {
        // If metadata exists, try again
        err2 := Precompile[SimpleStruct]()
        if err2 != nil {
            t.Errorf("second Precompile should succeed (idempotent): %v", err2)
        }
    } else if err1 != ErrMetadataMissing {
        t.Errorf("expected ErrMetadataMissing or success, got %v", err1)
    }
}

func TestReset(t *testing.T) {
    Reset()
    
    // After reset, registry should be empty
    if len(defaultRegistry.mappings) != 0 {
        t.Error("registry should be empty after Reset")
    }
}

func TestRegistry_ThreadSafety(t *testing.T) {
    Reset()
    defer Reset()
    
    // Test concurrent access doesn't panic
    done := make(chan bool, 10)
    
    for i := 0; i < 10; i++ {
        go func() {
            _, _ = defaultRegistry.GetMapping(reflect.TypeOf(SimpleStruct{}))
            done <- true
        }()
    }
    
    for i := 0; i < 10; i++ {
        <-done
    }
}
```

#### Validation
```bash
go test ./pkg/cgocopy2/... -v -race
```

**Success Criteria:**
- ✅ Precompile signature compiles
- ✅ Thread-safety tests pass (no race conditions)
- ✅ Idempotency works
- ✅ Error handling correct
- ⚠️ Metadata lookup may fail (acceptable at this stage)

---

## Phase 3: Build Mapping with Tag Support

### Tasks

- [ ] 3.1: Implement field matching with tag support
- [ ] 3.2: Implement buildMapping core logic
- [ ] 3.3: Add field validation
- [ ] 3.4: Detect fast-path eligibility
- [ ] 3.5: Test with C metadata fixtures

### Files to Create/Modify

```
pkg/cgocopy2/
├── mapping.go           # buildMapping implementation
├── field_matcher.go     # Tag-aware field matching
├── validation.go        # Field validation logic
├── mapping_test.go      # Mapping tests
└── sample_cgo.go        # Test fixtures with C metadata
```

### Implementation Details

#### 3.1 field_matcher.go
```go
package cgocopy2

import (
    "fmt"
    "reflect"
)

// fieldMatcher handles tag-aware field matching
type fieldMatcher struct {
    cFields []FieldInfo
}

func newFieldMatcher(cFields []FieldInfo) *fieldMatcher {
    return &fieldMatcher{cFields: cFields}
}

// findCField finds C field by name (with tag support)
func (fm *fieldMatcher) findCField(goField reflect.StructField) (FieldInfo, int, error) {
    // Check for cgocopy tag
    tag := goField.Tag.Get("cgocopy")
    
    if tag == "-" {
        return FieldInfo{}, -1, nil // Skip field marker
    }
    
    // Use tag as C field name, or default to Go field name
    searchName := goField.Name
    if tag != "" {
        searchName = tag
    }
    
    // Find matching C field
    for idx, cField := range fm.cFields {
        if cField.TypeName == searchName {
            return cField, idx, nil
        }
    }
    
    return FieldInfo{}, -1, fmt.Errorf("%w: %s", ErrFieldNotFound, searchName)
}

// shouldSkipField checks if field should be skipped
func shouldSkipField(goField reflect.StructField) bool {
    tag := goField.Tag.Get("cgocopy")
    return tag == "-"
}
```

#### 3.2 mapping.go
```go
package cgocopy2

import (
    "fmt"
    "reflect"
)

func (r *Registry) buildMapping(goType reflect.Type, metadata StructMetadata, conv CStringConverter) (*StructMapping, error) {
    mapping := &StructMapping{
        CSize:           metadata.Size,
        GoSize:          goType.Size(),
        CTypeName:       metadata.Name,
        GoTypeName:      goType.Name(),
        StringConverter: conv,
        CanFastPath:     true, // Assume true, set false if needed
        Fields:          make([]FieldMapping, 0, goType.NumField()),
    }
    
    matcher := newFieldMatcher(metadata.Fields)
    
    for i := 0; i < goType.NumField(); i++ {
        goField := goType.Field(i)
        
        // Check if field should be skipped
        if shouldSkipField(goField) {
            continue
        }
        
        // Find matching C field
        cField, cIdx, err := matcher.findCField(goField)
        if err != nil {
            return nil, fmt.Errorf("field %s: %w", goField.Name, err)
        }
        
        // Build field mapping
        fieldMapping, err := r.buildFieldMapping(goField, cField, conv)
        if err != nil {
            return nil, fmt.Errorf("field %s: %w", goField.Name, err)
        }
        
        // Disable fast path if field is complex
        if fieldMapping.IsString || fieldMapping.IsNested || fieldMapping.IsArray {
            mapping.CanFastPath = false
        }
        
        mapping.Fields = append(mapping.Fields, fieldMapping)
    }
    
    // Validate string fields have converter
    if !mapping.CanFastPath && hasStringFields(mapping) && conv == nil {
        return nil, fmt.Errorf("struct %s has string fields but no converter provided", goType.Name())
    }
    
    return mapping, nil
}

func (r *Registry) buildFieldMapping(goField reflect.StructField, cField FieldInfo, conv CStringConverter) (FieldMapping, error) {
    fm := FieldMapping{
        COffset:  cField.Offset,
        GoOffset: goField.Offset,
        Size:     goField.Type.Size(),
        CType:    cField.TypeName,
        GoType:   goField.Type,
        Kind:     cField.Kind,
        IsString: cField.IsString || cField.Kind == FieldString,
    }
    
    // Handle different field kinds
    switch cField.Kind {
    case FieldString:
        if goField.Type.Kind() != reflect.String {
            return fm, fmt.Errorf("C string field must map to Go string, got %v", goField.Type.Kind())
        }
        
    case FieldStruct:
        fm.IsNested = true
        nestedMapping, exists := r.GetMapping(goField.Type)
        if !exists {
            return fm, fmt.Errorf("nested struct %s not precompiled", goField.Type.Name())
        }
        fm.NestedMapping = nestedMapping
        
    case FieldArray:
        fm.IsArray = true
        fm.IsSlice = (goField.Type.Kind() == reflect.Slice)
        fm.ArrayLen = cField.ElemCount
        fm.ArrayElemSize = cField.Size / cField.ElemCount
        fm.ArrayElemGoType = goField.Type.Elem()
        
        if fm.ArrayElemGoType.Kind() == reflect.Struct {
            fm.ArrayElemIsNested = true
            elemMapping, exists := r.GetMapping(fm.ArrayElemGoType)
            if !exists {
                return fm, fmt.Errorf("array element struct %s not precompiled", fm.ArrayElemGoType.Name())
            }
            fm.ArrayElemMapping = elemMapping
        }
        
    case FieldPrimitive:
        // Validate size match
        if cField.Size != goField.Type.Size() {
            return fm, fmt.Errorf("size mismatch: C=%d bytes, Go=%d bytes", cField.Size, goField.Type.Size())
        }
    }
    
    return fm, nil
}

func hasStringFields(mapping *StructMapping) bool {
    for _, field := range mapping.Fields {
        if field.IsString {
            return true
        }
    }
    return false
}
```

### Testing Phase 3

#### 3.3 Create test fixtures with C metadata

**Note:** Uses new simplified macros from `native2/cgocopy_metadata.h`

```go
// pkg/cgocopy2/sample_cgo.go
package cgocopy2

/*
#include "../../native2/cgocopy_metadata.h"

typedef struct {
    uint32_t id;
    double value;
} SamplePrimitive;

typedef struct {
    uint32_t user_id;
    char* email;
} SampleWithString;

// New simplified macro syntax
CGOCOPY_STRUCT(SamplePrimitive,
    CGOCOPY_FIELD(SamplePrimitive, id),
    CGOCOPY_FIELD(SamplePrimitive, value)
)

CGOCOPY_STRUCT(SampleWithString,
    CGOCOPY_FIELD(SampleWithString, user_id),
    CGOCOPY_FIELD(SampleWithString, email)
)
*/
import "C"
```

#### 3.4 mapping_test.go
```go
package cgocopy2

import (
    "testing"
)

type SamplePrimitive struct {
    ID    uint32
    Value float64
}

type SampleWithString struct {
    UserID uint32 `cgocopy:"user_id"`
    Email  string
}

type SampleSkipField struct {
    ID       uint32
    Password string `cgocopy:"-"`
    Email    string
}

func TestPrecompile_PrimitiveStruct(t *testing.T) {
    Reset()
    defer Reset()
    
    err := Precompile[SamplePrimitive]()
    if err != nil {
        t.Fatalf("Precompile failed: %v", err)
    }
    
    mapping, exists := defaultRegistry.GetMapping(reflect.TypeOf(SamplePrimitive{}))
    if !exists {
        t.Fatal("mapping should exist after Precompile")
    }
    
    if !mapping.CanFastPath {
        t.Error("primitive-only struct should allow fast path")
    }
    
    if len(mapping.Fields) != 2 {
        t.Errorf("expected 2 fields, got %d", len(mapping.Fields))
    }
}

func TestPrecompile_WithStringConverter(t *testing.T) {
    Reset()
    defer Reset()
    
    err := Precompile[SampleWithString](DefaultCStringConverter)
    if err != nil {
        t.Fatalf("Precompile failed: %v", err)
    }
    
    mapping, exists := defaultRegistry.GetMapping(reflect.TypeOf(SampleWithString{}))
    if !exists {
        t.Fatal("mapping should exist")
    }
    
    if mapping.CanFastPath {
        t.Error("struct with strings should not allow fast path")
    }
    
    if mapping.StringConverter == nil {
        t.Error("converter should be set")
    }
}

func TestPrecompile_TagSupport(t *testing.T) {
    Reset()
    defer Reset()
    
    // SampleWithString uses tag: UserID maps to "user_id"
    err := Precompile[SampleWithString](DefaultCStringConverter)
    if err != nil {
        t.Fatalf("Precompile with tags failed: %v", err)
    }
    
    mapping, _ := defaultRegistry.GetMapping(reflect.TypeOf(SampleWithString{}))
    
    // First field should map to "user_id"
    if mapping.Fields[0].CType != "user_id" {
        t.Errorf("expected C field 'user_id', got '%s'", mapping.Fields[0].CType)
    }
}

func TestPrecompile_SkipField(t *testing.T) {
    Reset()
    defer Reset()
    
    err := Precompile[SampleSkipField](DefaultCStringConverter)
    if err != nil {
        t.Fatalf("Precompile failed: %v", err)
    }
    
    mapping, _ := defaultRegistry.GetMapping(reflect.TypeOf(SampleSkipField{}))
    
    // Should have 2 fields (Password skipped)
    if len(mapping.Fields) != 2 {
        t.Errorf("expected 2 fields (1 skipped), got %d", len(mapping.Fields))
    }
}
```

#### Validation
```bash
go test ./pkg/cgocopy2/... -v -run TestPrecompile
```

**Success Criteria:**
- ✅ Primitive struct precompiles successfully
- ✅ Fast path detection works
- ✅ Tag support maps fields correctly
- ✅ Skip field tag works
- ✅ String converter validation works

---

## Phase 4: Copy Implementation

### Tasks

- [ ] 4.1: Implement Copy function
- [ ] 4.2: Implement copyWithMapping
- [ ] 4.3: Handle nested structs
- [ ] 4.4: Handle arrays/slices
- [ ] 4.5: Handle strings
- [ ] 4.6: Comprehensive copy tests

### Files to Create/Modify

```
pkg/cgocopy2/
├── copy.go           # Copy implementation
├── copy_helpers.go   # copyWithMapping, field copiers
├── copy_test.go      # Copy tests
└── sample_test.go    # Integration tests with C data
```

### Implementation Details

#### 4.1 copy.go
```go
package cgocopy2

import (
    "reflect"
    "unsafe"
)

// Copy copies data from C memory to Go struct using precompiled mapping
func Copy[T any](dst *T, cPtr unsafe.Pointer) error {
    if dst == nil {
        return ErrNilDestination
    }
    if cPtr == nil {
        return ErrNilSourcePointer
    }
    
    t := reflect.TypeOf(*dst)
    
    // Thread-safe read
    defaultRegistry.mu.RLock()
    mapping, exists := defaultRegistry.mappings[t]
    defaultRegistry.mu.RUnlock()
    
    if !exists {
        return ErrNotPrecompiled
    }
    
    // Fast path for primitive-only structs
    if mapping.CanFastPath {
        *dst = *(*T)(cPtr)
        return nil
    }
    
    // Standard path with field-by-field copy
    return copyWithMapping(defaultRegistry, mapping, unsafe.Pointer(dst), cPtr)
}
```

#### 4.2 copy_helpers.go
```go
package cgocopy2

import (
    "fmt"
    "reflect"
    "unsafe"
)

func copyWithMapping(reg *Registry, mapping *StructMapping, dstPtr, srcPtr unsafe.Pointer) error {
    for i := range mapping.Fields {
        field := &mapping.Fields[i]
        
        fieldDst := unsafe.Add(dstPtr, field.GoOffset)
        fieldSrc := unsafe.Add(srcPtr, field.COffset)
        
        if err := copyField(reg, field, fieldDst, fieldSrc, mapping.StringConverter); err != nil {
            return fmt.Errorf("field %d: %w", i, err)
        }
    }
    
    return nil
}

func copyField(reg *Registry, field *FieldMapping, dstPtr, srcPtr unsafe.Pointer, conv CStringConverter) error {
    switch {
    case field.IsString:
        return copyStringField(dstPtr, srcPtr, conv)
        
    case field.IsNested:
        return copyWithMapping(reg, field.NestedMapping, dstPtr, srcPtr)
        
    case field.IsArray:
        return copyArrayField(reg, field, dstPtr, srcPtr, conv)
        
    default:
        // Primitive: direct memory copy
        copyBytes(dstPtr, srcPtr, field.Size)
        return nil
    }
}

func copyStringField(dstPtr, srcPtr unsafe.Pointer, conv CStringConverter) error {
    if conv == nil {
        return fmt.Errorf("string converter required")
    }
    
    cStrPtr := *(*unsafe.Pointer)(srcPtr)
    goStr := ""
    if cStrPtr != nil {
        goStr = conv.CStringToGo(cStrPtr)
    }
    
    *(*string)(dstPtr) = goStr
    return nil
}

func copyArrayField(reg *Registry, field *FieldMapping, dstPtr, srcPtr unsafe.Pointer, conv CStringConverter) error {
    if field.IsSlice {
        return copySliceField(reg, field, dstPtr, srcPtr, conv)
    }
    
    // Fixed-size array
    if field.ArrayElemIsNested {
        return copyArrayOfStructs(reg, field, dstPtr, srcPtr)
    }
    
    // Array of primitives: bulk copy
    copyBytes(dstPtr, srcPtr, field.Size)
    return nil
}

func copySliceField(reg *Registry, field *FieldMapping, dstPtr, srcPtr unsafe.Pointer, conv CStringConverter) error {
    length := int(field.ArrayLen)
    
    // Create slice with reflect
    sliceVal := reflect.NewAt(field.GoType, dstPtr).Elem()
    sliceVal.Set(reflect.MakeSlice(field.GoType, length, length))
    
    if length == 0 {
        return nil
    }
    
    // Get slice data pointer
    sliceData := sliceVal.Index(0).Addr().UnsafePointer()
    
    if field.ArrayElemIsNested {
        return copyArrayOfStructs(reg, field, sliceData, srcPtr)
    }
    
    // Bulk copy primitives
    totalSize := uintptr(length) * field.ArrayElemSize
    copyBytes(sliceData, srcPtr, totalSize)
    return nil
}

func copyArrayOfStructs(reg *Registry, field *FieldMapping, dstPtr, srcPtr unsafe.Pointer) error {
    for i := 0; i < int(field.ArrayLen); i++ {
        elemDst := unsafe.Add(dstPtr, uintptr(i)*field.ArrayElemSize)
        elemSrc := unsafe.Add(srcPtr, uintptr(i)*field.ArrayElemSize)
        
        if err := copyWithMapping(reg, field.ArrayElemMapping, elemDst, elemSrc); err != nil {
            return fmt.Errorf("array element %d: %w", i, err)
        }
    }
    return nil
}

func copyBytes(dst, src unsafe.Pointer, size uintptr) {
    if size == 0 {
        return
    }
    dstSlice := unsafe.Slice((*byte)(dst), size)
    srcSlice := unsafe.Slice((*byte)(src), size)
    copy(dstSlice, srcSlice)
}
```

### Testing Phase 4

#### 4.3 copy_test.go

**Note:** Uses new simplified macros from `native2/cgocopy_metadata.h`

```go
package cgocopy2

/*
#include <stdlib.h>
#include "../../native2/cgocopy_metadata.h"

typedef struct {
    uint32_t id;
    double value;
} CPrimitive;

typedef struct {
    uint32_t user_id;
    char* email;
} CWithString;

// Register with new simplified macros
CGOCOPY_STRUCT(CPrimitive,
    CGOCOPY_FIELD(CPrimitive, id),
    CGOCOPY_FIELD(CPrimitive, value)
)

CGOCOPY_STRUCT(CWithString,
    CGOCOPY_FIELD(CWithString, user_id),
    CGOCOPY_FIELD(CWithString, email)
)

CPrimitive* createPrimitive() {
    CPrimitive* p = (CPrimitive*)malloc(sizeof(CPrimitive));
    p->id = 42;
    p->value = 3.14;
    return p;
}

CWithString* createWithString() {
    CWithString* s = (CWithString*)malloc(sizeof(CWithString));
    s->user_id = 100;
    s->email = strdup("test@example.com");
    return s;
}
*/
import "C"
import (
    "testing"
    "unsafe"
)

func TestCopy_Primitive(t *testing.T) {
    Reset()
    defer Reset()
    
    if err := Precompile[SamplePrimitive](); err != nil {
        t.Fatalf("Precompile failed: %v", err)
    }
    
    cPtr := C.createPrimitive()
    defer C.free(unsafe.Pointer(cPtr))
    
    var result SamplePrimitive
    if err := Copy(&result, unsafe.Pointer(cPtr)); err != nil {
        t.Fatalf("Copy failed: %v", err)
    }
    
    if result.ID != 42 {
        t.Errorf("expected ID=42, got %d", result.ID)
    }
    if result.Value != 3.14 {
        t.Errorf("expected Value=3.14, got %f", result.Value)
    }
}

func TestCopy_WithString(t *testing.T) {
    Reset()
    defer Reset()
    
    if err := Precompile[SampleWithString](DefaultCStringConverter); err != nil {
        t.Fatalf("Precompile failed: %v", err)
    }
    
    cPtr := C.createWithString()
    defer func() {
        C.free(unsafe.Pointer((*C.CWithString)(cPtr).email))
        C.free(unsafe.Pointer(cPtr))
    }()
    
    var result SampleWithString
    if err := Copy(&result, unsafe.Pointer(cPtr)); err != nil {
        t.Fatalf("Copy failed: %v", err)
    }
    
    if result.UserID != 100 {
        t.Errorf("expected UserID=100, got %d", result.UserID)
    }
    if result.Email != "test@example.com" {
        t.Errorf("expected email='test@example.com', got '%s'", result.Email)
    }
}

func TestCopy_NotPrecompiled(t *testing.T) {
    Reset()
    defer Reset()
    
    type Unregistered struct {
        X int
    }
    
    var result Unregistered
    err := Copy(&result, unsafe.Pointer(&result))
    
    if err != ErrNotPrecompiled {
        t.Errorf("expected ErrNotPrecompiled, got %v", err)
    }
}

func TestCopy_NilPointers(t *testing.T) {
    Reset()
    defer Reset()
    
    Precompile[SamplePrimitive]()
    
    var result SamplePrimitive
    
    // Nil destination
    err := Copy((*SamplePrimitive)(nil), unsafe.Pointer(&result))
    if err != ErrNilDestination {
        t.Errorf("expected ErrNilDestination, got %v", err)
    }
    
    // Nil source
    err = Copy(&result, nil)
    if err != ErrNilSourcePointer {
        t.Errorf("expected ErrNilSourcePointer, got %v", err)
    }
}
```

#### Validation
```bash
go test ./pkg/cgocopy2/... -v -run TestCopy
```

**Success Criteria:**
- ✅ Primitive struct copies correctly
- ✅ String fields copy correctly
- ✅ Error handling works
- ✅ Nil pointer checks work
- ✅ Fast path optimization triggers

---

## Phase 5: FastCopy Implementation

### Tasks

- [ ] 5.1: Implement FastCopy function
- [ ] 5.2: Add fast path validation
- [ ] 5.3: FastCopy tests
- [ ] 5.4: Benchmark Fast vs Standard copy

### Implementation

#### 5.1 fastcopy.go
```go
package cgocopy2

import (
    "reflect"
    "unsafe"
)

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

#### Testing
```go
func TestFastCopy_Primitive(t *testing.T) {
    Reset()
    defer Reset()
    
    Precompile[SamplePrimitive]()
    
    cPtr := C.createPrimitive()
    defer C.free(unsafe.Pointer(cPtr))
    
    var result SamplePrimitive
    if err := FastCopy(&result, unsafe.Pointer(cPtr)); err != nil {
        t.Fatalf("FastCopy failed: %v", err)
    }
    
    if result.ID != 42 {
        t.Errorf("expected ID=42, got %d", result.ID)
    }
}

func TestFastCopy_NotEligible(t *testing.T) {
    Reset()
    defer Reset()
    
    Precompile[SampleWithString](DefaultCStringConverter)
    
    cPtr := C.createWithString()
    defer C.free(unsafe.Pointer(cPtr))
    
    var result SampleWithString
    err := FastCopy(&result, unsafe.Pointer(cPtr))
    
    if err != ErrCannotUseFastPath {
        t.Errorf("expected ErrCannotUseFastPath, got %v", err)
    }
}
```

#### 5.4 Benchmarks
```go
func BenchmarkCopy_Primitive(b *testing.B) {
    Reset()
    Precompile[SamplePrimitive]()
    
    cPtr := C.createPrimitive()
    defer C.free(unsafe.Pointer(cPtr))
    
    var result SamplePrimitive
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        Copy(&result, unsafe.Pointer(cPtr))
    }
}

func BenchmarkFastCopy_Primitive(b *testing.B) {
    Reset()
    Precompile[SamplePrimitive]()
    
    cPtr := C.createPrimitive()
    defer C.free(unsafe.Pointer(cPtr))
    
    var result SamplePrimitive
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        FastCopy(&result, unsafe.Pointer(cPtr))
    }
}
```

---

## Phase 6: Validation Helper

### Tasks

- [ ] 6.1: Implement ValidateStruct
- [ ] 6.2: Add detailed error reporting
- [ ] 6.3: Validation tests

### Implementation

```go
// validate.go
package cgocopy2

import (
    "fmt"
    "reflect"
)

type ValidationReport struct {
    TypeName        string
    IsPrecompiled   bool
    CanUseFastPath  bool
    SizeMatch       bool
    CSize           uintptr
    GoSize          uintptr
    FieldCount      int
    FieldMismatches []FieldMismatch
    Errors          []string
}

type FieldMismatch struct {
    GoFieldName string
    CFieldName  string
    Issue       string
    GoType      string
    CType       string
    GoSize      uintptr
    CSize       uintptr
}

func ValidateStruct[T any]() (*ValidationReport, error) {
    var zero T
    t := reflect.TypeOf(zero)
    
    report := &ValidationReport{
        TypeName: t.Name(),
    }
    
    // Check if precompiled
    defaultRegistry.mu.RLock()
    mapping, exists := defaultRegistry.mappings[t]
    defaultRegistry.mu.RUnlock()
    
    report.IsPrecompiled = exists
    if exists {
        report.CanUseFastPath = mapping.CanUseFastPath
    }
    
    // Lookup metadata
    metadata, err := lookupStructMetadata(t.Name())
    if err != nil {
        report.Errors = append(report.Errors, err.Error())
        return report, err
    }
    
    report.CSize = metadata.Size
    report.GoSize = t.Size()
    report.SizeMatch = (metadata.Size == t.Size())
    report.FieldCount = len(metadata.Fields)
    
    // Field-by-field validation
    // ... (detailed field checking)
    
    return report, nil
}
```

---

## Phase 7: C Macro Improvements

**Important:** Create `native2/` directory for improved macros to avoid breaking v1.

### Directory Structure
```
native/
├── cgocopy_metadata.h      # v1 macros (old style)
└── metadata_registry.c     # v1 registry

native2/
├── cgocopy_metadata.h      # v2 macros (new simplified style)
└── metadata_registry.c     # v2 registry (copy from native/)
```

### Tasks

- [ ] 7.1: Create `native2/` directory
- [ ] 7.2: Copy `metadata_registry.c` from `native/` to `native2/`
- [ ] 7.3: Create new `native2/cgocopy_metadata.h` with simplified macros
- [ ] 7.4: Add CGOCOPY_FIELD macro with _Generic auto-detection
- [ ] 7.5: Add CGOCOPY_STRUCT single-macro definition
- [ ] 7.6: Test new macros with sample structs
- [ ] 7.7: Verify old macros in `native/` still work (v1 compatibility)

### Implementation Details

#### 7.3 native2/cgocopy_metadata.h

```c
#ifndef CGOCOPY_METADATA_V2_H
#define CGOCOPY_METADATA_V2_H

#include <stddef.h>
#include <stdbool.h>
#include <string.h>

// Keep same type definitions as v1
typedef enum {
    CGOCOPY_FIELD_PRIMITIVE_KIND = 0,
    CGOCOPY_FIELD_POINTER_KIND = 1,
    CGOCOPY_FIELD_STRING_KIND = 2,
    CGOCOPY_FIELD_ARRAY_KIND = 3,
    CGOCOPY_FIELD_STRUCT_KIND = 4,
} cgocopy_field_kind;

typedef struct {
    size_t offset;
    size_t size;
    const char *type_name;
    cgocopy_field_kind kind;
    const char *elem_type;
    size_t elem_count;
    bool is_string;
} cgocopy_field_info;

typedef struct {
    const char *name;
    size_t size;
    size_t alignment;
    size_t field_count;
    const cgocopy_field_info *fields;
} cgocopy_struct_info;

typedef struct cgocopy_struct_registry_node {
    const cgocopy_struct_info *info;
    struct cgocopy_struct_registry_node *next;
} cgocopy_struct_registry_node;

void cgocopy_registry_add(cgocopy_struct_registry_node *node);
const cgocopy_struct_info *cgocopy_lookup_struct_info(const char *name);

// Helper macros (same as v1)
#define CGOCOPY_INTERNAL_STRINGIFY(x) CGOCOPY_INTERNAL_STRINGIFY_IMPL(x)
#define CGOCOPY_INTERNAL_STRINGIFY_IMPL(x) #x

#define CGOCOPY_INTERNAL_FIELD_INIT(kind_value, struct_type, field_name, type_literal, elem_literal, elem_count_value, string_flag) \
    { \
        .offset = offsetof(struct_type, field_name), \
        .size = sizeof(((struct_type *)0)->field_name), \
        .type_name = (type_literal), \
        .kind = (kind_value), \
        .elem_type = (elem_literal), \
        .elem_count = (elem_count_value), \
        .is_string = (string_flag) \
    }

// NEW: Auto-detection macro using _Generic
#define CGOCOPY_FIELD(struct_type, field_name) \
    _Generic(((struct_type*)0)->field_name, \
        char*: CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_STRING_KIND, struct_type, field_name, #field_name, NULL, 0, true), \
        const char*: CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_STRING_KIND, struct_type, field_name, #field_name, NULL, 0, true), \
        default: CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_PRIMITIVE_KIND, struct_type, field_name, #field_name, NULL, 0, false))

// NEW: Single-macro struct definition
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
    } \
    static inline const cgocopy_struct_info *cgocopy_get_##struct_type##_info(void) { \
        return &cgocopy_struct_info_##struct_type; \
    }

#endif // CGOCOPY_METADATA_V2_H
```

### Testing Phase 7

#### Test new macros
```c
// test_new_macros.c
#include "../../native2/cgocopy_metadata.h"

typedef struct {
    uint32_t id;
    char* name;
    double value;
} TestStruct;

// New simplified syntax!
CGOCOPY_STRUCT(TestStruct,
    CGOCOPY_FIELD(TestStruct, id),
    CGOCOPY_FIELD(TestStruct, name),
    CGOCOPY_FIELD(TestStruct, value)
)
```

#### Comparison test
Create side-by-side comparison:
- Old macros in `test_old_macros.c` using `native/cgocopy_metadata.h`
- New macros in `test_new_macros.c` using `native2/cgocopy_metadata.h`
- Verify both produce identical metadata structures
- Confirm field offsets, sizes, and kinds match

---

## Phase 8: Integration & Documentation

### Tasks

- [ ] 8.1: Port users example to cgocopy2
- [ ] 8.2: Comprehensive benchmark suite
- [ ] 8.3: Update documentation
- [ ] 8.4: Migration guide
- [ ] 8.5: Performance comparison report

---

## Testing Strategy Summary

### Unit Tests
- Each phase has dedicated unit tests
- Test error paths explicitly
- Use table-driven tests where appropriate

### Integration Tests
- Full end-to-end with real C data
- Nested structs
- Arrays and slices
- String handling

### Race Detection
```bash
go test -race ./pkg/cgocopy2/...
```

### Benchmarks
```bash
go test -bench=. -benchmem ./pkg/cgocopy2/...
```

### Coverage
```bash
go test -cover ./pkg/cgocopy2/...
```

---

## Success Criteria (Final)

- [ ] All tests pass
- [ ] No race conditions detected
- [ ] Coverage > 80%
- [ ] FastCopy measurably faster than Copy for primitives
- [ ] Examples work with new API
- [ ] Documentation complete
- [ ] cgocopy v1 still works (no breakage)

---

## Phase 9: Code Generation Tool

### Overview

Eliminate manual boilerplate by auto-generating metadata code from C header files.

**Current Pain:** 5 manual steps per struct across 4 files
**Solution:** 1 step (write struct) + `go generate`

### Tasks

- [ ] 9.1: Create `tools/cgocopy-generate/` directory
- [ ] 9.2: Implement C struct parser (regex-based)
- [ ] 9.3: Implement code generator with templates
- [ ] 9.4: Add CLI with flags
- [ ] 9.5: Test with sample C headers
- [ ] 9.6: Refactor integration tests to use generator
- [ ] 9.7: Document tool usage
- [ ] 9.8: Add to CI pipeline

### Directory Structure

```
tools/
└── cgocopy-generate/
    ├── main.go          # CLI and orchestration
    ├── parser.go        # C struct parser
    ├── generator.go     # Code generation
    ├── templates.go     # Output templates
    ├── parser_test.go   # Parser tests
    └── generator_test.go # Generator tests
```

### Implementation Details

#### 9.2 Parser (parser.go)

```go
package main

import (
    "regexp"
    "strings"
)

type Field struct {
    Name      string
    Type      string
    ArraySize string // empty if not array
}

type Struct struct {
    Name   string
    Fields []Field
}

// parseStructs extracts struct definitions from C code
func parseStructs(content string) ([]Struct, error) {
    var structs []Struct
    
    // Remove comments
    content = removeComments(content)
    
    // Match struct definitions: typedef struct { ... } Name; or struct Name { ... };
    structRegex := regexp.MustCompile(`(?:typedef\s+)?struct(?:\s+(\w+))?\s*\{([^}]+)\}(?:\s*(\w+))?`)
    matches := structRegex.FindAllStringSubmatch(content, -1)
    
    for _, match := range matches {
        // Extract struct name (either before or after body)
        name := match[1]
        if name == "" {
            name = match[3]
        }
        if name == "" {
            continue // Anonymous struct, skip
        }
        
        body := match[2]
        s := Struct{Name: name}
        
        // Parse fields: type name; or type name[size];
        fieldRegex := regexp.MustCompile(`([a-zA-Z_][\w\s\*]+?)\s+(\w+)(?:\[(\d+)\])?\s*;`)
        fieldMatches := fieldRegex.FindAllStringSubmatch(body, -1)
        
        for _, fm := range fieldMatches {
            fieldType := strings.TrimSpace(fm[1])
            fieldName := fm[2]
            arraySize := fm[3]
            
            s.Fields = append(s.Fields, Field{
                Name:      fieldName,
                Type:      fieldType,
                ArraySize: arraySize,
            })
        }
        
        if len(s.Fields) > 0 {
            structs = append(structs, s)
        }
    }
    
    return structs, nil
}

func removeComments(content string) string {
    // Remove // comments
    lineComment := regexp.MustCompile(`//.*`)
    content = lineComment.ReplaceAllString(content, "")
    
    // Remove /* */ comments
    blockComment := regexp.MustCompile(`(?s)/\*.*?\*/`)
    content = blockComment.ReplaceAllString(content, "")
    
    return content
}
```

#### 9.3 Generator (generator.go)

```go
package main

import (
    "os"
    "text/template"
)

type TemplateData struct {
    InputFile string
    Structs   []Struct
}

var metadataTemplate = `// GENERATED CODE - DO NOT EDIT
// Generated from: {{.InputFile}}

#include "../../native2/cgocopy_macros.h"
#include "structs.h"

{{range .Structs}}
// Metadata for {{.Name}}
CGOCOPY_STRUCT({{.Name}},
{{- range $idx, $field := .Fields}}
    CGOCOPY_FIELD({{$.Name}}, {{$field.Name}}){{if ne $idx (sub1 (len $.Fields))}},{{end}}
{{- end}}
)

const cgocopy_struct_info* get_{{.Name}}_metadata(void) {
    return &cgocopy_metadata_{{.Name}};
}
{{end}}
`

var apiHeaderTemplate = `// GENERATED CODE - DO NOT EDIT
// Generated from: {{.InputFile}}

#ifndef METADATA_API_H
#define METADATA_API_H

#include "../../native2/cgocopy_macros.h"

// Getter functions for each struct
{{range .Structs}}
const cgocopy_struct_info* get_{{.Name}}_metadata(void);
{{end}}

#endif // METADATA_API_H
`

func generateMetadata(data TemplateData, outputFile string) error {
    tmpl := template.Must(template.New("metadata").Funcs(template.FuncMap{
        "sub1": func(n int) int { return n - 1 },
    }).Parse(metadataTemplate))
    
    f, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer f.Close()
    
    return tmpl.Execute(f, data)
}

func generateAPIHeader(data TemplateData, outputFile string) error {
    tmpl := template.Must(template.New("api").Parse(apiHeaderTemplate))
    
    f, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer f.Close()
    
    return tmpl.Execute(f, data)
}
```

#### 9.4 CLI (main.go)

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "strings"
)

func main() {
    input := flag.String("input", "", "Input C header file")
    output := flag.String("output", "", "Output C file (default: input_meta.c)")
    apiHeader := flag.String("api", "", "Output API header file (optional)")
    flag.Parse()
    
    if *input == "" {
        fmt.Fprintln(os.Stderr, "Usage: cgocopy-generate -input=file.h [-output=file_meta.c] [-api=api.h]")
        os.Exit(1)
    }
    
    // Read input file
    content, err := os.ReadFile(*input)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", *input, err)
        os.Exit(1)
    }
    
    // Parse structs
    structs, err := parseStructs(string(content))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing structs: %v\n", err)
        os.Exit(1)
    }
    
    if len(structs) == 0 {
        fmt.Fprintf(os.Stderr, "Warning: No structs found in %s\n", *input)
        os.Exit(0)
    }
    
    fmt.Printf("Found %d struct(s): ", len(structs))
    for i, s := range structs {
        if i > 0 {
            fmt.Print(", ")
        }
        fmt.Print(s.Name)
    }
    fmt.Println()
    
    data := TemplateData{
        InputFile: *input,
        Structs:   structs,
    }
    
    // Generate metadata implementation
    if *output == "" {
        *output = strings.TrimSuffix(*input, ".h") + "_meta.c"
    }
    
    if err := generateMetadata(data, *output); err != nil {
        fmt.Fprintf(os.Stderr, "Error generating metadata: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Generated: %s\n", *output)
    
    // Generate API header if requested
    if *apiHeader != "" {
        if err := generateAPIHeader(data, *apiHeader); err != nil {
            fmt.Fprintf(os.Stderr, "Error generating API header: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("Generated: %s\n", *apiHeader)
    }
}
```

### Testing Phase 9

#### 9.5 Parser Tests

```go
// parser_test.go
package main

import (
    "testing"
)

func TestParseStructs_Simple(t *testing.T) {
    input := `
typedef struct {
    int id;
    double value;
} SimplePerson;
`
    
    structs, err := parseStructs(input)
    if err != nil {
        t.Fatalf("parseStructs failed: %v", err)
    }
    
    if len(structs) != 1 {
        t.Fatalf("expected 1 struct, got %d", len(structs))
    }
    
    s := structs[0]
    if s.Name != "SimplePerson" {
        t.Errorf("expected name 'SimplePerson', got '%s'", s.Name)
    }
    
    if len(s.Fields) != 2 {
        t.Fatalf("expected 2 fields, got %d", len(s.Fields))
    }
}

func TestParseStructs_WithArrays(t *testing.T) {
    input := `
struct Student {
    int id;
    char name[32];
    int grades[5];
};
`
    
    structs, err := parseStructs(input)
    if err != nil {
        t.Fatalf("parseStructs failed: %v", err)
    }
    
    if len(structs[0].Fields) != 3 {
        t.Fatalf("expected 3 fields, got %d", len(structs[0].Fields))
    }
    
    nameField := structs[0].Fields[1]
    if nameField.ArraySize != "32" {
        t.Errorf("expected array size '32', got '%s'", nameField.ArraySize)
    }
}

func TestRemoveComments(t *testing.T) {
    input := `
// This is a comment
struct Test { // inline comment
    int x; /* block comment */
};
/* Multi-line
   comment */
`
    
    result := removeComments(input)
    
    if strings.Contains(result, "//") || strings.Contains(result, "/*") {
        t.Error("comments should be removed")
    }
}
```

#### 9.6 Integration Test

Update `pkg/cgocopy2/integration/` to use generator:

```go
// integration/integration_cgo.go
package integration

//go:generate cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h

/*
#cgo CFLAGS: -I${SRCDIR}/../native2
#include "native/metadata_api.h"
#include "native/structs_meta.c"
*/
import "C"
```

Then run:
```bash
cd pkg/cgocopy2/integration
go generate ./...
go test -v
```

**Expected:**
- ✅ `native/structs_meta.c` generated
- ✅ `native/metadata_api.h` generated
- ✅ All 9 integration tests pass
- ✅ Generated code compiles without errors

#### 9.7 Documentation

Create `tools/cgocopy-generate/README.md`:

```markdown
# cgocopy-generate

Automatic metadata generation for cgocopy v2.

## Installation

```bash
go install github.com/shaban/cgocopy/tools/cgocopy-generate@latest
```

## Usage

```bash
cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h
```

## Integration with go generate

Add to your Go file:

```go
//go:generate cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h
```

Run:
```bash
go generate ./...
```

## What It Does

1. Parses C struct definitions from header file
2. Generates `CGOCOPY_STRUCT` macro calls
3. Generates getter functions
4. Generates API header with declarations

## Example

**Input (structs.h):**
```c
typedef struct {
    int id;
    double score;
} SimplePerson;
```

**Generated (structs_meta.c):**
```c
CGOCOPY_STRUCT(SimplePerson,
    CGOCOPY_FIELD(SimplePerson, id),
    CGOCOPY_FIELD(SimplePerson, score)
)

const cgocopy_struct_info* get_SimplePerson_metadata(void) {
    return &cgocopy_metadata_SimplePerson;
}
```

## Limitations

- Simple struct definitions only
- No preprocessor macro expansion
- No cross-file type resolution

Works for 90% of use cases!
```

#### 9.8 CI Integration

Add to `.github/workflows/test.yml`:

```yaml
- name: Check generated code is up to date
  run: |
    go generate ./...
    git diff --exit-code || (echo "Generated code is out of date. Run 'go generate ./...' and commit changes." && exit 1)
```

### Success Criteria Phase 9

- [ ] Parser handles common struct patterns
- [ ] Generator produces valid C code
- [ ] CLI flags work correctly
- [ ] Integration tests pass with generated code
- [ ] Tool runs in < 100ms
- [ ] Zero external dependencies
- [ ] Documentation complete
- [ ] CI catches stale generated code

### Benefits Summary

**Before:**
- ✍️ Write struct in `.h`
- ✍️ Write `CGOCOPY_STRUCT` in `.c`
- ✍️ Write getter in `.c`
- ✍️ Declare getter in `.h`
- ✍️ Register in Go
- ⏱️ **5-10 minutes per struct**

**After:**
- ✍️ Write struct in `.h`
- 🤖 Run `go generate`
- ⏱️ **30 seconds per struct**

**Impact:** 80% reduction in manual work, zero chance of typos, always in sync!
