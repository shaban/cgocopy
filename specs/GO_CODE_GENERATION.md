# Go Code Generation Specification

**Version:** 1.0  
**Date:** October 15, 2025  
**Status:** Draft  
**Author:** cgocopy team

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Workflow](#workflow)
4. [Command-Line Interface](#command-line-interface)
5. [Type Mapping](#type-mapping)
6. [Name Conversion](#name-conversion)
7. [Code Generation Details](#code-generation-details)
8. [Validation Strategy](#validation-strategy)
9. [Edge Cases](#edge-cases)
10. [Testing Strategy](#testing-strategy)
11. [Implementation Phases](#implementation-phases)

---

## 1. Overview

### 1.1 Purpose

Extend `cgocopy-generate` to automatically generate **both C metadata and Go bridge code** from a single C header file, eliminating 90% of manual boilerplate.

### 1.2 Goals

- ✅ Generate Go struct definitions from C struct definitions
- ✅ Generate complete CGO bridge with import "C" preamble
- ✅ Generate automatic type registration in `init()` function
- ✅ Support arrays, nested structs, and all primitive types
- ✅ Validate type compatibility at runtime
- ✅ Maintain backward compatibility (optional flag)

### 1.3 Performance Validation ✅

**Benchmark Results (Apple M1 Pro):**

| Test Case | cgocopy | JSON | Speedup | Memory |
|-----------|---------|------|---------|---------|
| SimplePerson (4 fields) | 45.15ns | 950.8ns | **21.1x faster** | 64B vs 336B |
| GameObject (6 fields + nested) | 50.38ns | 2601ns | **51.6x faster** | 64B vs 496B |

**Key Findings:**
- cgocopy is **21-52x faster** than JSON for real-world structs
- Constant ~50ns performance regardless of struct complexity
- 5-7x less memory allocation (1 alloc vs 7-9 allocs)
- Complexity is **absolutely justified** for FFI use cases

**When to Use cgocopy vs JSON:**
- Use cgocopy: High-frequency C↔Go data transfer (game loops, audio processing, real-time systems)
- Use JSON: Occasional serialization, human-readable data, network protocols

### 1.4 Non-Goals

- ❌ Parse preprocessor directives (#define, #ifdef)
- ❌ Support C unions
- ❌ Support function pointers
- ❌ Support bitfields
- ❌ Follow typedef chains (require explicit types)
- ❌ Support pointer-to-struct fields (except opaque handles)

---

## 2. Architecture

### 2.1 Conceptual Model

The bridge package is **separate and independent** from both the user's C library and their Go application:

```
┌─────────────────────────────────────────────────────────┐
│  External C Library (Unchanged)                          │
│  ├── audio_engine.c                                     │
│  ├── audio_engine.h                                     │
│  └── struct AudioData { int rate; char* name; }         │
└─────────────────────────────────────────────────────────┘
                   ↕ Same memory layout
┌─────────────────────────────────────────────────────────┐
│  Generated Bridge Package (pkg/audio_bridge/)           │
│  ├── bridge.h          ← User creates (blueprint)       │
│  ├── bridge_meta.c     ← Generated (C metadata)         │
│  └── bridge_meta.go    ← Generated (Go structs + init)  │
│                                                          │
│  Purpose: Define struct layout for bridging             │
│  NOT included in external C library!                    │
└─────────────────────────────────────────────────────────┘
                   ↕ Import
┌─────────────────────────────────────────────────────────┐
│  User's Go Application                                  │
│  ├── main.go                                            │
│  └── import "myproject/pkg/audio_bridge"                │
│      data := cgocopy.Copy[audio_bridge.AudioData](cPtr) │
└─────────────────────────────────────────────────────────┘
```

### 2.2 Key Insight

The bridge header (`bridge.h`) is a **pure blueprint** that:
1. Defines struct layout for type safety
2. Is compiled with Go code (via CGO)
3. Never needs to be included in the external C library
4. Must match the memory layout of structs in the external library

### 2.3 Why This Works

- C structs with same layout = binary compatible
- `sizeof()` and `offsetof()` detect actual layout at compile time
- Runtime validation catches mismatches immediately
- No modifications to external C code required

---

## 3. Workflow

### 3.1 Step-by-Step Process

#### Step 1: User Creates Bridge Header

```c
// pkg/audio_bridge/bridge.h
#ifndef AUDIO_BRIDGE_H
#define AUDIO_BRIDGE_H

#include <stdint.h>

// Blueprint: matches layout of struct in audio_engine.c
typedef struct {
    int32_t sample_rate;  // Use explicit types!
    char* filename;
} AudioData;

typedef struct {
    uint32_t id;
    AudioData* data;  // Pointer to AudioData
} AudioStream;

#endif
```

**Best Practice:** Use explicit sized types (`int32_t`, `uint64_t`) instead of platform-dependent types (`int`, `long`).

#### Step 2: Add go:generate Directive

```go
// pkg/audio_bridge/doc.go
package audio_bridge

//go:generate cgocopy-generate -input=bridge.h -go=bridge_meta.go -package=audio_bridge
```

#### Step 3: Run Generator

```bash
cd pkg/audio_bridge
go generate
```

#### Step 4: Generated Files

Two files are created:
1. `bridge_meta.c` - C metadata with `CGOCOPY_STRUCT` macros
2. `bridge_meta.go` - Go structs with automatic registration

#### Step 5: Use in Application

```go
package main

/*
#cgo LDFLAGS: -laudio
#include "audio_engine.h"  // External C library
*/
import "C"
import (
    "myproject/pkg/audio_bridge"
    cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

func main() {
    cData := C.get_audio_data()
    goData := cgocopy.Copy[audio_bridge.AudioData](unsafe.Pointer(cData))
    // Use goData...
}
```

---

## 4. Command-Line Interface

### 4.1 New Flags

```bash
cgocopy-generate \
  -input=<file.h>              # Required: C header to parse
  -output=<file_meta.c>        # Optional: C metadata output (default: {input}_meta.c)
  -api=<api.h>                 # Optional: C API header (default: none)
  -go=<file.go>                # NEW: Go code output (enables Go generation)
  -package=<name>              # NEW: Go package name (default: auto-detect)
  -header-path=<path>          # Optional: Path to cgocopy_macros.h (default: auto-detect)
```

### 4.2 Flag Behavior

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-input` | ✅ Yes | - | C header file to parse |
| `-output` | ❌ No | `{input}_meta.c` | Generated C metadata file |
| `-api` | ❌ No | none | Generated C API header |
| `-go` | ❌ No | none | If set, enables Go code generation |
| `-package` | ❌ No | Auto-detect from directory | Go package name |
| `-header-path` | ❌ No | Auto-detect from go.mod | Path to cgocopy_macros.h |

### 4.3 Auto-Detection Rules

#### Package Name Detection

1. If `-package` provided → use it
2. Else if `-go` path is `pkg/audio_bridge/bridge.go` → extract `audio_bridge`
3. Else if current directory is `audio_bridge/` → use `audio_bridge`
4. Else → error, require explicit `-package`

**Sanitization:**
- Remove hyphens: `audio-bridge` → `audio_bridge`
- Lowercase: `AudioBridge` → `audiobridge`
- Remove dots: `audio.bridge` → `audio_bridge`

#### Header Path Detection

1. If `-header-path` provided → use it
2. Else find go.mod by walking up from output directory
3. Calculate relative path to `pkg/cgocopy/native/cgocopy_macros.h`
4. If not found → use default `../../native/cgocopy_macros.h`

### 4.4 Examples

**Minimal (Go generation):**
```bash
cgocopy-generate -input=bridge.h -go=bridge_meta.go -package=audio
```

**Full (C + Go generation):**
```bash
cgocopy-generate \
  -input=native/structs.h \
  -output=native/structs_meta.c \
  -api=native/metadata_api.h \
  -go=structs.go \
  -package=mybridge
```

**With explicit paths:**
```bash
cgocopy-generate \
  -input=bridge.h \
  -go=bridge.go \
  -package=audio \
  -header-path=../../../vendor/cgocopy/native/cgocopy_macros.h
```

---

## 5. Type Mapping

### 5.1 Three-Tier Type System

Our approach uses three validation layers:

1. **Generation Time (Pessimistic):** Assume common types
2. **Compile Time (Accurate):** C macros detect actual sizes
3. **Runtime (Validation):** Verify assumptions were correct

### 5.2 Primitive Type Mapping

#### Explicit Sized Types (Recommended)

| C Type | Go Type | Notes |
|--------|---------|-------|
| `int8_t` | `int8` | Always 8-bit |
| `uint8_t` | `uint8` | Always 8-bit |
| `int16_t` | `int16` | Always 16-bit |
| `uint16_t` | `uint16` | Always 16-bit |
| `int32_t` | `int32` | Always 32-bit |
| `uint32_t` | `uint32` | Always 32-bit |
| `int64_t` | `int64` | Always 64-bit |
| `uint64_t` | `uint64` | Always 64-bit |
| `float` | `float32` | IEEE 754 single |
| `double` | `float64` | IEEE 754 double |
| `_Bool` | `bool` | C99 boolean |

#### Platform-Dependent Types (Use with Caution)

| C Type | Go Type | Assumption | Runtime Check |
|--------|---------|------------|---------------|
| `char` | `int8` | 8-bit signed | ✅ Validated |
| `unsigned char` | `uint8` | 8-bit unsigned | ✅ Validated |
| `short` | `int16` | 16-bit | ✅ Validated |
| `unsigned short` | `uint16` | 16-bit | ✅ Validated |
| `int` | `int32` | 32-bit (most common) | ✅ Validated |
| `unsigned int` | `uint32` | 32-bit (most common) | ✅ Validated |
| `long` | `int64` | 64-bit (conservative) | ✅ Validated |
| `unsigned long` | `uint64` | 64-bit (conservative) | ✅ Validated |
| `long long` | `int64` | 64-bit | ✅ Validated |
| `size_t` | `uint64` | 64-bit (conservative) | ✅ Validated |

#### Special Types

| C Type | Go Type | Notes |
|--------|---------|-------|
| `char*` | `string` | CGO converts automatically |
| `const char*` | `string` | Treated same as `char*` |
| `void*` | `unsafe.Pointer` | Opaque pointer |
| `SomeStruct*` | `unsafe.Pointer` | Opaque handle only |

### 5.3 Composite Types

#### Fixed-Size Arrays

```c
// C
int grades[5];
float scores[3];
char buffer[256];
```

```go
// Go
Grades [5]int32
Scores [3]float32
Buffer [256]int8
```

**Rule:** `type name[N]` → `Name [N]GoType`

#### Nested Structs (By Value)

```c
// C
typedef struct {
    double x, y, z;
} Point3D;

typedef struct {
    Point3D position;
    Point3D velocity;
} GameObject;
```

```go
// Go
type Point3D struct {
    X float64
    Y float64
    Z float64
}

type GameObject struct {
    Position Point3D
    Velocity Point3D
}
```

**Rule:** Nested structs must be defined in same header. Generate in dependency order (Point3D before GameObject).

#### Pointers

**Supported:**
- ✅ `char*` → `string` (automatic conversion)
- ✅ `void*` → `unsafe.Pointer` (opaque handle)
- ✅ `OpaqueType*` → `unsafe.Pointer` (opaque handle)

**Not Supported:**
- ❌ `int*` → Pointer to primitive (memory safety issue)
- ❌ `Point3D*` → Pointer to struct (ownership ambiguity)

**Exception:** If struct type is not found in parsed structs, assume opaque handle:

```c
typedef struct AudioEngine AudioEngine;  // Opaque

typedef struct {
    AudioEngine* engine;  // OK: opaque handle
} AudioContext;
```

```go
type AudioContext struct {
    Engine unsafe.Pointer  `cgocopy:"engine"`
}
```

### 5.4 Unsupported Types

**Error immediately:**
- Function pointers: `void (*callback)(int)`
- Unions: `union { int i; float f; }`
- Bitfields: `unsigned int flag : 1;`
- Flexible arrays: `int data[];`

**Warning + Skip field:**
- Unknown typedefs: `UserID id;` (if UserID not in header)
- Pointer to known struct: `Point3D* pos;`

---

## 6. Name Conversion

### 6.1 Package Names

**Rule:** Header filename (without extension) → Go package name

**Algorithm:**
```go
func sanitizePackageName(headerFile string) string {
    // Remove extension
    name := strings.TrimSuffix(headerFile, ".h")
    
    // Replace invalid chars with underscore
    name = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(name, "_")
    
    // Lowercase
    name = strings.ToLower(name)
    
    // Remove leading digits
    name = regexp.MustCompile(`^[0-9]+`).ReplaceAllString(name, "")
    
    // Must start with letter
    if !regexp.MustCompile(`^[a-z]`).MatchString(name) {
        name = "pkg_" + name
    }
    
    return name
}
```

**Examples:**
- `audio_bridge.h` → `audio_bridge`
- `AudioBridge.h` → `audiobridge`
- `audio-bridge.h` → `audio_bridge`
- `audio.bridge.h` → `audio_bridge`
- `123audio.h` → `pkg_123audio`

### 6.2 Struct Names

**Rule:** Keep C struct name as-is (already PascalCase by convention)

```c
typedef struct {
    int id;
} User;
```

```go
type User struct {
    ID int32
}
```

**Edge case:** If struct name conflicts with Go keyword, append underscore:

```c
typedef struct {
    int value;
} type;  // 'type' is Go keyword
```

```go
type Type_ struct {  // Renamed
    Value int32
}
```

### 6.3 Field Names

**Rule:** Convert snake_case → PascalCase

**Algorithm:**
```go
func toPascalCase(snakeCase string) string {
    // Split on underscore
    parts := strings.Split(snakeCase, "_")
    
    // Capitalize each part
    for i, part := range parts {
        if len(part) > 0 {
            parts[i] = strings.ToUpper(part[:1]) + part[1:]
        }
    }
    
    // Join
    pascalCase := strings.Join(parts, "")
    
    // Handle special cases
    pascalCase = handleAcronyms(pascalCase)
    
    return pascalCase
}

func handleAcronyms(name string) string {
    // Common acronyms that should stay uppercase
    replacements := map[string]string{
        "Id":   "ID",
        "Url":  "URL",
        "Http": "HTTP",
        "Api":  "API",
        "Db":   "DB",
        "Sql":  "SQL",
        "Uuid": "UUID",
    }
    
    for old, new := range replacements {
        name = strings.ReplaceAll(name, old, new)
    }
    
    return name
}
```

**Examples:**
- `user_id` → `UserID`
- `sample_rate` → `SampleRate`
- `full_name` → `FullName`
- `http_status` → `HTTPStatus`
- `api_key` → `APIKey`

### 6.4 Go Keyword Conflicts

**Rule:** If field name conflicts with Go keyword, append underscore

**Go keywords to check:**
```go
var goKeywords = map[string]bool{
    "break": true, "case": true, "chan": true, "const": true,
    "continue": true, "default": true, "defer": true, "else": true,
    "fallthrough": true, "for": true, "func": true, "go": true,
    "goto": true, "if": true, "import": true, "interface": true,
    "map": true, "package": true, "range": true, "return": true,
    "select": true, "struct": true, "switch": true, "type": true,
    "var": true,
}
```

**Examples:**

```c
typedef struct {
    int type;      // Go keyword
    int range;     // Go keyword
    int count;     // Not a keyword
} Metadata;
```

```go
type Metadata struct {
    Type_  int32 `cgocopy:"type"`   // Renamed + tag
    Range_ int32 `cgocopy:"range"`  // Renamed + tag
    Count  int32 `cgocopy:"count"`  // Normal
}
```

### 6.5 Name Collision Detection

**Within same struct:**

```c
typedef struct {
    int type;      // C field
    int Type;      // Different in C (case-sensitive)
} Data;
```

**Problem:** Both become `Type` in Go!

**Solution:** Append numeric suffix:

```go
type Data struct {
    Type1 int32 `cgocopy:"type"`   // First occurrence
    Type2 int32 `cgocopy:"Type"`   // Second occurrence
}
```

**Algorithm:**
```go
func resolveCollisions(fields []Field) []Field {
    seen := make(map[string]int)  // Go name → count
    
    for i := range fields {
        goName := fields[i].GoName
        
        if count, exists := seen[goName]; exists {
            // Collision! Append number
            count++
            fields[i].GoName = goName + strconv.Itoa(count)
            seen[goName] = count
        } else {
            seen[goName] = 1
        }
    }
    
    return fields
}
```

---

## 7. Code Generation Details

### 7.1 Generated File Structure

#### C Metadata File (`{input}_meta.c`)

```c
// GENERATED CODE - DO NOT EDIT
// Generated from: bridge.h
// Generator: cgocopy-generate v1.0.0
// Date: 2025-10-15 12:34:56

#include <stdlib.h>
#include "../../../pkg/cgocopy/native/cgocopy_macros.h"
#include "bridge.h"

// Metadata for AudioData
CGOCOPY_STRUCT(AudioData,
    CGOCOPY_FIELD(AudioData, sample_rate),
    CGOCOPY_FIELD(AudioData, filename)
)

const cgocopy_struct_info* get_AudioData_metadata(void) {
    return &cgocopy_metadata_AudioData;
}

// Metadata for AudioStream
CGOCOPY_STRUCT(AudioStream,
    CGOCOPY_FIELD(AudioStream, id),
    CGOCOPY_FIELD(AudioStream, data)
)

const cgocopy_struct_info* get_AudioStream_metadata(void) {
    return &cgocopy_metadata_AudioStream;
}
```

#### Go Bridge File (`{go}`)

```go
// GENERATED CODE - DO NOT EDIT
// Generated from: bridge.h
// Generator: cgocopy-generate v1.0.0
// Date: 2025-10-15 12:34:56

package audio_bridge

//go:generate cgocopy-generate -input=bridge.h -go=bridge_meta.go -package=audio_bridge

/*
#cgo CFLAGS: -I${SRCDIR}/../../../pkg/cgocopy
#include "bridge.h"
#include "bridge_meta.c"
*/
import "C"

import (
    "fmt"
    "unsafe"
    
    cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

// AudioData matches the C AudioData struct
type AudioData struct {
    SampleRate int32  `cgocopy:"sample_rate"`
    Filename   string `cgocopy:"filename"`
}

// AudioStream matches the C AudioStream struct
type AudioStream struct {
    ID   uint32         `cgocopy:"id"`
    Data unsafe.Pointer `cgocopy:"data"`  // Pointer to AudioData
}

func init() {
    // Register AudioData
    if err := cgocopy.PrecompileWithC[AudioData](
        extractCMetadata(C.get_AudioData_metadata()),
    ); err != nil {
        panic(fmt.Sprintf("Failed to register AudioData: %v", err))
    }
    
    // Register AudioStream
    if err := cgocopy.PrecompileWithC[AudioStream](
        extractCMetadata(C.get_AudioStream_metadata()),
    ); err != nil {
        panic(fmt.Sprintf("Failed to register AudioStream: %v", err))
    }
}

// extractCMetadata converts C struct metadata to Go format
func extractCMetadata(cStructInfoPtr *C.cgocopy_struct_info) cgocopy.CStructInfo {
    if cStructInfoPtr == nil {
        panic("nil C struct info pointer")
    }
    
    fieldCount := int(cStructInfoPtr.field_count)
    fields := make([]cgocopy.CFieldInfo, fieldCount)
    
    cFieldsSlice := (*[1 << 30]C.cgocopy_field_info)(unsafe.Pointer(cStructInfoPtr.fields))[:fieldCount:fieldCount]
    
    for i := 0; i < fieldCount; i++ {
        cField := &cFieldsSlice[i]
        fields[i] = cgocopy.CFieldInfo{
            Name:      C.GoString(cField.name),
            Type:      C.GoString(cField._type),
            Offset:    uintptr(cField.offset),
            Size:      uintptr(cField.size),
            IsPointer: cField.is_pointer != 0,
            IsArray:   cField.is_array != 0,
            ArrayLen:  int(cField.array_len),
        }
    }
    
    return cgocopy.CStructInfo{
        Name:   C.GoString(cStructInfoPtr.name),
        Size:   uintptr(cStructInfoPtr.size),
        Fields: fields,
    }
}
```

### 7.2 Template Variables

#### For C Metadata Template

```go
type CMetadataTemplateData struct {
    InputFile   string    // Original header filename
    GeneratorVersion string
    Timestamp   string    // RFC3339 format
    MacrosPath  string    // Relative path to cgocopy_macros.h
    Structs     []Struct
}

type Struct struct {
    Name   string
    Fields []Field
}

type Field struct {
    Name      string  // C field name
    ArraySize string  // Empty if not array
}
```

#### For Go Template

```go
type GoTemplateData struct {
    InputFile   string
    GeneratorVersion string
    Timestamp   string
    PackageName string
    CGOCFlags   string    // Auto-calculated relative path
    Structs     []GoStruct
    GenerateDirective string  // The go:generate line
}

type GoStruct struct {
    Name        string      // C struct name (PascalCase)
    GoName      string      // Go type name (may differ if keyword conflict)
    Comment     string      // Doc comment
    Fields      []GoField
}

type GoField struct {
    Name        string      // C field name (snake_case)
    GoName      string      // Go field name (PascalCase)
    GoType      string      // Go type (int32, string, [5]int32, etc)
    Tag         string      // cgocopy tag (if needed)
    Comment     string      // Inline comment
}
```

### 7.3 Dependency Ordering

**Problem:** Nested structs must be defined before they're used:

```c
typedef struct {
    Point3D position;  // Requires Point3D to be defined first
} GameObject;

typedef struct {
    double x, y, z;
} Point3D;  // Should come first!
```

**Solution:** Topological sort

**Algorithm:**
```go
func sortStructsByDependency(structs []Struct) ([]Struct, error) {
    // Build dependency graph
    deps := make(map[string][]string)  // struct -> []dependencies
    
    for _, s := range structs {
        deps[s.Name] = []string{}
        for _, field := range s.Fields {
            // Check if field type is another struct
            if isStructType(field.Type, structs) {
                deps[s.Name] = append(deps[s.Name], field.Type)
            }
        }
    }
    
    // Topological sort using Kahn's algorithm
    sorted := []Struct{}
    inDegree := make(map[string]int)
    
    // Calculate in-degrees
    for _, dependencies := range deps {
        for _, dep := range dependencies {
            inDegree[dep]++
        }
    }
    
    // Queue nodes with no dependencies
    queue := []string{}
    for _, s := range structs {
        if inDegree[s.Name] == 0 {
            queue = append(queue, s.Name)
        }
    }
    
    // Process queue
    for len(queue) > 0 {
        name := queue[0]
        queue = queue[1:]
        
        // Add to sorted list
        for _, s := range structs {
            if s.Name == name {
                sorted = append(sorted, s)
                break
            }
        }
        
        // Reduce in-degree of dependent nodes
        for dependent, dependencies := range deps {
            for _, dep := range dependencies {
                if dep == name {
                    inDegree[dependent]--
                    if inDegree[dependent] == 0 {
                        queue = append(queue, dependent)
                    }
                }
            }
        }
    }
    
    // Check for cycles
    if len(sorted) != len(structs) {
        return nil, errors.New("circular dependency detected in struct definitions")
    }
    
    return sorted, nil
}
```

**Error handling:** If cycle detected, fail with clear error:
```
Error: Circular dependency in struct definitions:
  GameObject depends on Transform
  Transform depends on GameObject
  
Solution: Remove circular reference or use pointers
```

---

## 8. Validation Strategy

### 8.1 Three-Layer Validation

#### Layer 1: Parse-Time Validation (Generator)

**Validates:**
- ✅ Header file syntax is parseable
- ✅ All struct names are valid C identifiers
- ✅ All field types are recognized or warn
- ✅ No circular dependencies
- ✅ No duplicate struct names
- ✅ No duplicate field names within struct

**Errors immediately:**
```
Error: Unknown type 'UserID' in struct User, field id
Hint: Use explicit types like uint32_t, or define UserID in same header

Error: Circular dependency detected:
  GameObject -> Transform -> GameObject
```

#### Layer 2: Compile-Time Validation (C Compiler)

**Validates:**
- ✅ C syntax is correct
- ✅ All types exist and have correct sizes
- ✅ `sizeof()` and `offsetof()` produce valid values
- ✅ C macros expand correctly

**C compiler will error if:**
- Header has syntax errors
- Types are undefined
- Macros fail to expand

#### Layer 3: Runtime Validation (Go init)

**Validates:**
- ✅ C field count matches Go field count
- ✅ C field sizes match Go field sizes
- ✅ C struct size >= Go struct size
- ✅ Platform assumptions were correct

**Generated validation code:**

```go
func init() {
    // Register AudioData
    cInfo := extractCMetadata(C.get_AudioData_metadata())
    
    // Validation 1: Field count
    goFieldCount := 2  // Generated: reflect on AudioData
    if len(cInfo.Fields) != goFieldCount {
        panic(fmt.Sprintf(
            "AudioData field count mismatch: C has %d fields, Go has %d fields",
            len(cInfo.Fields), goFieldCount,
        ))
    }
    
    // Validation 2: Struct size
    goSize := unsafe.Sizeof(AudioData{})
    if cInfo.Size < uintptr(goSize) {
        panic(fmt.Sprintf(
            "AudioData size mismatch: C size=%d bytes, Go size=%d bytes (C must be >= Go)",
            cInfo.Size, goSize,
        ))
    }
    
    // Validation 3: Field sizes (for platform-dependent types)
    validateField := func(cField cgocopy.CFieldInfo, expectedSize uintptr, fieldName string) {
        if cField.Size != expectedSize {
            panic(fmt.Sprintf(
                "AudioData.%s size mismatch: C=%d bytes, Go=%d bytes\n"+
                "Hint: Use explicit types like int32_t instead of int",
                fieldName, cField.Size, expectedSize,
            ))
        }
    }
    
    // Only validate platform-dependent types
    if cInfo.Fields[0].Type == "int32" {  // Assumed int -> int32
        validateField(cInfo.Fields[0], 4, "SampleRate")
    }
    
    // Register
    if err := cgocopy.PrecompileWithC[AudioData](cInfo); err != nil {
        panic(fmt.Sprintf("Failed to register AudioData: %v", err))
    }
}
```

### 8.2 Validation Error Messages

**Good error messages include:**
1. What failed
2. Expected vs actual values
3. Hint for fix

**Examples:**

```
panic: AudioData field count mismatch: C has 3 fields, Go has 2 fields
Hint: Did you modify bridge.h without regenerating? Run: go generate

panic: AudioData.SampleRate size mismatch: C=8 bytes, Go=4 bytes
Hint: C 'int' is 64-bit on this platform. Use 'int32_t' in bridge.h for portability

panic: Circular dependency detected:
  GameObject -> Transform -> GameObject
Hint: Use pointers to break cycle, or separate into different structs
```

---

## 9. Edge Cases

### 9.1 Empty Structs

```c
typedef struct {
} Empty;
```

**Behavior:** 
- ❌ **Error:** "Struct Empty has no fields"
- **Rationale:** Empty structs not useful for bridging

### 9.2 Anonymous Structs

```c
typedef struct {
    int x;
} Point;  // Named

struct {
    int y;
};  // Anonymous
```

**Behavior:**
- ✅ **Parse:** Named structs only
- ❌ **Skip:** Anonymous structs with warning

### 9.3 Forward Declarations

```c
typedef struct GameObject GameObject;  // Forward declaration

typedef struct {
    GameObject* parent;  // Uses forward declaration
} Transform;

typedef struct GameObject {  // Actual definition
    Transform transform;
} GameObject;
```

**Behavior:**
- ✅ **Parse:** Actual definitions only
- ✅ **Pointer to forward-declared type** → `unsafe.Pointer`
- ⚠️ **Warning:** If forward-declared type never defined

### 9.4 Multiple Type Names (Typedef)

```c
typedef struct Point {
    double x, y;
} Point, Point2D, Coordinate;  // Three names!
```

**Behavior:**
- ✅ **Generate:** All three types as aliases
- **Implementation:** Use Go type alias

```go
type Point struct {
    X float64
    Y float64
}

type Point2D = Point      // Alias
type Coordinate = Point   // Alias
```

### 9.5 Packed Structs

```c
#pragma pack(push, 1)
typedef struct {
    char a;
    int b;  // No padding!
} Packed;
#pragma pack(pop)
```

**Behavior:**
- ⚠️ **Warning:** "Struct Packed may use #pragma pack - cannot detect packing"
- ✅ **Continue:** Generate normally, validation will catch mismatches
- **Runtime:** If sizes don't match, init() panics with clear error

### 9.6 Comments in Structs

```c
typedef struct {
    int id;          // User ID
    char* name;      /* User's full name */
    double balance;  // Account balance in USD
} User;
```

**Behavior:**
- ✅ **Parse:** Extract comments
- ✅ **Include in Go:** Add as inline comments

```go
type User struct {
    ID      int32   `cgocopy:"id"`      // User ID
    Name    string  `cgocopy:"name"`    // User's full name
    Balance float64 `cgocopy:"balance"` // Account balance in USD
}
```

### 9.7 Very Long Field Names

```c
typedef struct {
    int this_is_an_extremely_long_field_name_that_exceeds_normal_limits;
} LongNames;
```

**Behavior:**
- ✅ **Allow:** No length limit
- ✅ **Convert:** `this_is_an_extremely_long_field_name_that_exceeds_normal_limits` → `ThisIsAnExtremelyLongFieldNameThatExceedsNormalLimits`

### 9.8 Special Characters in Comments

```c
typedef struct {
    int value;  // Price in $ (USD)
} Product;
```

**Behavior:**
- ✅ **Preserve:** Keep special chars in comments
- ✅ **Escape:** Only escape comment delimiters if needed

```go
type Product struct {
    Value int32 `cgocopy:"value"`  // Price in $ (USD)
}
```

### 9.9 Preprocessor Directives

```c
#ifdef DEBUG
typedef struct {
    int debug_info;
} DebugData;
#endif
```

**Behavior:**
- ❌ **Cannot handle:** Preprocessor directives not parsed
- **Workaround:** User must preprocess header first or provide clean version

### 9.10 Mixed C/C++ Headers

```c
#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    int value;
} Data;

#ifdef __cplusplus
}
#endif
```

**Behavior:**
- ✅ **Parse:** Extract struct definition regardless of C++ markers
- ⚠️ **Warning:** If C++ features detected (classes, namespaces)

---

## 10. Testing Strategy

### 10.1 Unit Tests (Generator)

Test parsing and code generation in isolation:

```go
// tools/cgocopy-generate/generator_test.go

func TestGenerateGoStruct_SimplePrimitive(t *testing.T) {
    input := `
    typedef struct {
        int id;
        double score;
    } Person;
    `
    
    structs, _ := parseStructs(input)
    goCode := generateGoCode(structs, "testpkg")
    
    assert.Contains(t, goCode, "type Person struct")
    assert.Contains(t, goCode, "ID int32")
    assert.Contains(t, goCode, "Score float64")
}

func TestGenerateGoStruct_Arrays(t *testing.T) {
    input := `
    typedef struct {
        int grades[5];
    } Student;
    `
    
    structs, _ := parseStructs(input)
    goCode := generateGoCode(structs, "testpkg")
    
    assert.Contains(t, goCode, "Grades [5]int32")
}

func TestGenerateGoStruct_NestedStruct(t *testing.T) {
    input := `
    typedef struct {
        double x, y;
    } Point;
    
    typedef struct {
        Point pos;
    } GameObject;
    `
    
    structs, _ := parseStructs(input)
    structs, _ = sortStructsByDependency(structs)
    goCode := generateGoCode(structs, "testpkg")
    
    // Point must come before GameObject
    pointIdx := strings.Index(goCode, "type Point struct")
    objectIdx := strings.Index(goCode, "type GameObject struct")
    assert.True(t, pointIdx < objectIdx)
}

func TestNameConversion_SnakeToPascal(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"user_id", "UserID"},
        {"sample_rate", "SampleRate"},
        {"http_status", "HTTPStatus"},
        {"full_name", "FullName"},
    }
    
    for _, tt := range tests {
        actual := toPascalCase(tt.input)
        assert.Equal(t, tt.expected, actual)
    }
}

func TestKeywordConflict(t *testing.T) {
    input := `
    typedef struct {
        int type;
        int range;
    } Metadata;
    `
    
    structs, _ := parseStructs(input)
    goCode := generateGoCode(structs, "testpkg")
    
    assert.Contains(t, goCode, "Type_ int32 `cgocopy:\"type\"`")
    assert.Contains(t, goCode, "Range_ int32 `cgocopy:\"range\"`")
}
```

### 10.2 Integration Tests (End-to-End)

Test complete workflow with actual compilation:

```
// tests/codegen/
├── simple/
│   ├── bridge.h           (Input)
│   ├── bridge_meta.c      (Expected output)
│   ├── bridge_meta.go     (Expected output)
│   └── test.sh            (Run generator + compile)
├── arrays/
│   └── ...
├── nested/
│   └── ...
└── validation/
    └── ...
```

**Test script:**
```bash
#!/bin/bash
# tests/codegen/simple/test.sh

# Generate code
cgocopy-generate -input=bridge.h -go=bridge_meta.go -package=simple

# Verify files created
[ -f bridge_meta.c ] || exit 1
[ -f bridge_meta.go ] || exit 1

# Compile Go code
go build -o /dev/null . || exit 1

# Run validation test
go test -v || exit 1

echo "✓ Test passed"
```

### 10.3 Validation Tests

Test runtime validation catches errors:

```go
// tests/codegen/validation/mismatch_test.go

func TestValidation_SizeMismatch(t *testing.T) {
    // Manually create mismatched metadata
    badInfo := cgocopy.CStructInfo{
        Name: "BadStruct",
        Size: 8,  // C says 8 bytes
        Fields: []cgocopy.CFieldInfo{
            {Name: "value", Type: "int32", Size: 4, Offset: 0},
        },
    }
    
    type BadStruct struct {
        Value int64  // Go says 8 bytes for this field alone!
    }
    
    // Should panic
    assert.Panics(t, func() {
        cgocopy.PrecompileWithC[BadStruct](badInfo)
    })
}
```

### 10.4 Platform Tests

Test across different platforms:

```yaml
# .github/workflows/test-platforms.yml
name: Platform Tests

on: [push]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        arch: [amd64, arm64]
    
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test ./... -v
      
      - name: Test code generation
        run: |
          cd tests/codegen
          for dir in */; do
            echo "Testing $dir"
            cd "$dir"
            ./test.sh
            cd ..
          done
```

### 10.5 Regression Tests

Keep examples of previously-fixed bugs:

```
tests/regression/
├── issue_001_pointer_to_struct/
├── issue_002_circular_dependency/
├── issue_003_keyword_collision/
└── ...
```

Each with:
- Input C header
- Expected error (if applicable)
- Expected generated code
- Test that verifies fix

---

## 11. Implementation Phases

### Phase 1: Core Go Generation (Week 1)
**Goal:** Generate basic Go structs from simple C structs

**Tasks:**
1. Add `-go` and `-package` flags to CLI
2. Implement Go template (basic version)
3. Implement type mapping (primitives only)
4. Implement name conversion (snake_case → PascalCase)
5. Generate Go struct definitions
6. Generate `import "C"` preamble
7. Generate `extractCMetadata()` helper

**Test criteria:**
- ✅ Parse simple struct with primitives
- ✅ Generate valid Go code
- ✅ Go code compiles
- ✅ Can import generated package

**Deliverable:** Can generate Go code for:
```c
typedef struct {
    int32_t id;
    double score;
    char* name;
} Person;
```

### Phase 2: Registration & Validation (Week 1-2)
**Goal:** Auto-register types and validate at runtime

**Tasks:**
1. Generate `init()` function
2. Generate metadata extraction calls
3. Implement field count validation
4. Implement size validation
5. Add helpful error messages
6. Test validation catches mismatches

**Test criteria:**
- ✅ `init()` registers types automatically
- ✅ Validation detects size mismatches
- ✅ Error messages are clear
- ✅ Can use `cgocopy.Copy[T]()` immediately

**Deliverable:** Generated code includes validation

### Phase 3: Arrays & Nested Structs (Week 2)
**Goal:** Support composite types

**Tasks:**
1. Implement array field generation: `[N]Type`
2. Implement nested struct detection
3. Implement dependency sorting (topological)
4. Handle circular dependency errors
5. Test with complex nested structures

**Test criteria:**
- ✅ Arrays generate correctly: `Grades [5]int32`
- ✅ Nested structs in correct order
- ✅ Circular dependencies detected
- ✅ Complex nesting works

**Deliverable:** Can generate code for:
```c
typedef struct {
    int grades[5];
} Student;

typedef struct {
    double x, y;
} Point;

typedef struct {
    Point position;
    Point velocity;
} GameObject;
```

### Phase 4: Edge Cases & Polish (Week 3)
**Goal:** Handle edge cases gracefully

**Tasks:**
1. Implement Go keyword detection & renaming
2. Handle name collisions
3. Support comments in generated code
4. Handle pointers (strings + opaque)
5. Add warnings for unsupported features
6. Improve error messages

**Test criteria:**
- ✅ Go keywords renamed with underscore
- ✅ Name collisions resolved
- ✅ Comments preserved
- ✅ Pointers handled correctly
- ✅ Clear warnings for unsupported types

**Deliverable:** Robust generator that handles edge cases

### Phase 5: Documentation & Examples (Week 3-4)
**Goal:** Complete documentation and examples

**Tasks:**
1. Write comprehensive README
2. Create example: simple bridge package
3. Create example: real-world C library
4. Document conventions and best practices
5. Add troubleshooting guide
6. Update main cgocopy docs

**Test criteria:**
- ✅ README covers all features
- ✅ Examples work and demonstrate patterns
- ✅ Troubleshooting guide helpful
- ✅ All edge cases documented

**Deliverable:** Complete documentation

### Phase 6: Testing & Validation (Week 4)
**Goal:** Comprehensive test coverage

**Tasks:**
1. Write unit tests for all functions
2. Create integration test suite
3. Add platform tests (Linux, Mac, Windows)
4. Create regression test suite
5. Add CI/CD pipeline
6. Performance testing

**Test criteria:**
- ✅ 90%+ code coverage
- ✅ All integration tests pass
- ✅ All platform tests pass
- ✅ CI/CD green
- ✅ No performance regressions

**Deliverable:** Production-ready code

### Phase 7: v1.1.0 Release (Week 4)
**Goal:** Release Go code generation feature

**Tasks:**
1. Final code review
2. Update version numbers
3. Write release notes
4. Tag release
5. Update pkg.go.dev
6. Announce on forums/social media

**Deliverable:** Public release

---

## 12. Success Criteria

### 12.1 Functional Requirements

- ✅ **F1:** Generate valid Go code from C headers
- ✅ **F2:** Support all primitive types
- ✅ **F3:** Support arrays (fixed-size)
- ✅ **F4:** Support nested structs
- ✅ **F5:** Auto-register types in init()
- ✅ **F6:** Validate compatibility at runtime
- ✅ **F7:** Handle Go keyword conflicts
- ✅ **F8:** Preserve comments
- ✅ **F9:** Generate correct cgocopy tags
- ✅ **F10:** Work with existing C libraries (no modifications needed)

### 12.2 Non-Functional Requirements

- ✅ **NF1:** Generated code compiles without errors
- ✅ **NF2:** Generated code passes `go vet`
- ✅ **NF3:** Generated code follows Go conventions
- ✅ **NF4:** Error messages are clear and actionable
- ✅ **NF5:** Generation completes in < 1 second for typical headers
- ✅ **NF6:** Works on Linux, macOS, Windows
- ✅ **NF7:** Backward compatible (optional flag)

### 12.3 User Experience Goals

- ✅ **UX1:** User writes only C header (no manual Go code)
- ✅ **UX2:** One command generates everything
- ✅ **UX3:** Types work immediately (no manual registration)
- ✅ **UX4:** Validation catches errors early (init time)
- ✅ **UX5:** Error messages include fix suggestions
- ✅ **UX6:** Documentation covers common use cases

---

## 13. Open Questions

### 13.1 Resolved

- ✅ How to handle platform-dependent types? → Three-tier validation
- ✅ How to handle nested structs? → Topological sort
- ✅ How to handle pointers? → Only strings + opaque handles
- ✅ How to handle typedefs? → Require explicit types, document convention
- ✅ Package naming strategy? → Header filename → package name

### 13.2 To Be Decided

1. **Should we support C++ headers?**
   - Pros: More users
   - Cons: Much more complex (classes, templates, namespaces)
   - **Recommendation:** No, document workaround (extern "C")

2. **Should we generate tests automatically?**
   - Pros: Catch errors early
   - Cons: Extra complexity
   - **Recommendation:** Phase 8 (future enhancement)

3. **Should we support custom type mappings?**
   - Example: User wants `time_t` → `time.Time` instead of `int64`
   - Pros: More flexible
   - Cons: Requires config file or annotations
   - **Recommendation:** Phase 8 (future enhancement)

4. **Should we support multiple input files?**
   - Example: `structs.h` includes `types.h`
   - Pros: Handles includes
   - Cons: Need to follow #include directives
   - **Recommendation:** Phase 8, for now document single-file convention

---

## 14. Related Documents

- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - Original v1.0.0 plan
- [REGISTRY_IMPROVEMENTS.md](./REGISTRY_IMPROVEMENTS.md) - Registry design
- [USAGE_GUIDE.md](./USAGE_GUIDE.md) - User documentation
- [WHEN_TO_USE_REGISTRY.md](./WHEN_TO_USE_REGISTRY.md) - Registry patterns

---

## Appendix A: Complete Example

### Input: bridge.h

```c
#ifndef AUDIO_BRIDGE_H
#define AUDIO_BRIDGE_H

#include <stdint.h>

// Audio sample format
typedef struct {
    char* name;
    uint32_t bits_per_sample;
    uint32_t sample_rate;
} AudioFormat;

// Audio buffer with samples
typedef struct {
    int32_t id;
    AudioFormat format;
    float samples[1024];
    uint64_t timestamp;
} AudioBuffer;

#endif
```

### Command

```bash
cgocopy-generate -input=bridge.h -go=bridge.go -package=audio
```

### Output: bridge_meta.c

```c
// GENERATED CODE - DO NOT EDIT
// Generated from: bridge.h

#include <stdlib.h>
#include "../../../pkg/cgocopy/native/cgocopy_macros.h"
#include "bridge.h"

// Metadata for AudioFormat
CGOCOPY_STRUCT(AudioFormat,
    CGOCOPY_FIELD(AudioFormat, name),
    CGOCOPY_FIELD(AudioFormat, bits_per_sample),
    CGOCOPY_FIELD(AudioFormat, sample_rate)
)

const cgocopy_struct_info* get_AudioFormat_metadata(void) {
    return &cgocopy_metadata_AudioFormat;
}

// Metadata for AudioBuffer
CGOCOPY_STRUCT(AudioBuffer,
    CGOCOPY_FIELD(AudioBuffer, id),
    CGOCOPY_FIELD(AudioBuffer, format),
    CGOCOPY_ARRAY_FIELD(AudioBuffer, samples, float),
    CGOCOPY_FIELD(AudioBuffer, timestamp)
)

const cgocopy_struct_info* get_AudioBuffer_metadata(void) {
    return &cgocopy_metadata_AudioBuffer;
}
```

### Output: bridge.go

```go
// GENERATED CODE - DO NOT EDIT
// Generated from: bridge.h

package audio

//go:generate cgocopy-generate -input=bridge.h -go=bridge.go -package=audio

/*
#cgo CFLAGS: -I${SRCDIR}/../../../pkg/cgocopy
#include "bridge.h"
#include "bridge_meta.c"
*/
import "C"

import (
    "fmt"
    "unsafe"
    
    cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

// AudioFormat matches the C AudioFormat struct
type AudioFormat struct {
    Name           string `cgocopy:"name"`
    BitsPerSample  uint32 `cgocopy:"bits_per_sample"`
    SampleRate     uint32 `cgocopy:"sample_rate"`
}

// AudioBuffer matches the C AudioBuffer struct
type AudioBuffer struct {
    ID        int32         `cgocopy:"id"`
    Format    AudioFormat   `cgocopy:"format"`
    Samples   [1024]float32 `cgocopy:"samples"`
    Timestamp uint64        `cgocopy:"timestamp"`
}

func init() {
    // Register AudioFormat
    if err := cgocopy.PrecompileWithC[AudioFormat](
        extractCMetadata(C.get_AudioFormat_metadata()),
    ); err != nil {
        panic(fmt.Sprintf("Failed to register AudioFormat: %v", err))
    }
    
    // Register AudioBuffer
    if err := cgocopy.PrecompileWithC[AudioBuffer](
        extractCMetadata(C.get_AudioBuffer_metadata()),
    ); err != nil {
        panic(fmt.Sprintf("Failed to register AudioBuffer: %v", err))
    }
}

func extractCMetadata(cStructInfoPtr *C.cgocopy_struct_info) cgocopy.CStructInfo {
    if cStructInfoPtr == nil {
        panic("nil C struct info pointer")
    }
    
    fieldCount := int(cStructInfoPtr.field_count)
    fields := make([]cgocopy.CFieldInfo, fieldCount)
    
    cFieldsSlice := (*[1 << 30]C.cgocopy_field_info)(unsafe.Pointer(cStructInfoPtr.fields))[:fieldCount:fieldCount]
    
    for i := 0; i < fieldCount; i++ {
        cField := &cFieldsSlice[i]
        fields[i] = cgocopy.CFieldInfo{
            Name:      C.GoString(cField.name),
            Type:      C.GoString(cField._type),
            Offset:    uintptr(cField.offset),
            Size:      uintptr(cField.size),
            IsPointer: cField.is_pointer != 0,
            IsArray:   cField.is_array != 0,
            ArrayLen:  int(cField.array_len),
        }
    }
    
    return cgocopy.CStructInfo{
        Name:   C.GoString(cStructInfoPtr.name),
        Size:   uintptr(cStructInfoPtr.size),
        Fields: fields,
    }
}
```

### Usage: main.go

```go
package main

/*
#cgo LDFLAGS: -laudio
#include "audio_engine.h"  // External C library
*/
import "C"

import (
    "fmt"
    "unsafe"
    
    "myproject/pkg/audio"
    cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

func main() {
    // Get buffer from C library
    cBuffer := C.get_audio_buffer()
    defer C.free_audio_buffer(cBuffer)
    
    // Copy to Go (automatically registered!)
    goBuffer := cgocopy.Copy[audio.AudioBuffer](unsafe.Pointer(cBuffer))
    
    // Use in Go
    fmt.Printf("Buffer ID: %d\n", goBuffer.ID)
    fmt.Printf("Format: %s @ %d Hz\n", 
        goBuffer.Format.Name, 
        goBuffer.Format.SampleRate)
    fmt.Printf("First sample: %f\n", goBuffer.Samples[0])
}
```

---

## Appendix B: Type Mapping Reference

### Complete Type Table

| C Type | Size (typical) | Go Type | Tag Needed? | Notes |
|--------|----------------|---------|-------------|-------|
| `_Bool` | 1 | `bool` | No | C99 boolean |
| `char` | 1 | `int8` | No | Signed by default |
| `signed char` | 1 | `int8` | No | Explicitly signed |
| `unsigned char` | 1 | `uint8` | No | Unsigned |
| `short` | 2 | `int16` | No | |
| `unsigned short` | 2 | `uint16` | No | |
| `int` | 4 | `int32` | No | ⚠️ Platform-dependent |
| `unsigned int` | 4 | `uint32` | No | ⚠️ Platform-dependent |
| `long` | 4/8 | `int64` | No | ⚠️ 32-bit Windows vs 64-bit Unix |
| `unsigned long` | 4/8 | `uint64` | No | ⚠️ Platform-dependent |
| `long long` | 8 | `int64` | No | |
| `unsigned long long` | 8 | `uint64` | No | |
| `int8_t` | 1 | `int8` | No | ✅ Recommended |
| `uint8_t` | 1 | `uint8` | No | ✅ Recommended |
| `int16_t` | 2 | `int16` | No | ✅ Recommended |
| `uint16_t` | 2 | `uint16` | No | ✅ Recommended |
| `int32_t` | 4 | `int32` | No | ✅ Recommended |
| `uint32_t` | 4 | `uint32` | No | ✅ Recommended |
| `int64_t` | 8 | `int64` | No | ✅ Recommended |
| `uint64_t` | 8 | `uint64` | No | ✅ Recommended |
| `float` | 4 | `float32` | No | IEEE 754 |
| `double` | 8 | `float64` | No | IEEE 754 |
| `char*` | 8 | `string` | No | Auto-converted |
| `const char*` | 8 | `string` | No | Same as char* |
| `void*` | 8 | `unsafe.Pointer` | No | Opaque |
| `TypeName*` | 8 | `unsafe.Pointer` | No | If TypeName is struct |
| `TypeName` | Varies | `TypeName` | No | Nested struct |
| `type[N]` | Varies | `[N]Type` | No | Fixed array |

---

**End of Specification**
