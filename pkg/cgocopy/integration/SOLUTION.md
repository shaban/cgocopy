# CGO Integration - Problem & Solution Summary

## The Problem: Static Symbols Aren't Visible to CGO

When using C macros to generate metadata, the `CGOCOPY_STRUCT` macro creates **static** variables:

```c
#define CGOCOPY_STRUCT(structtype, ...) \
    static cgocopy_field_info cgocopy_fields_##structtype[] = { ... }; \
    static cgocopy_struct_info cgocopy_metadata_##structtype = { ... };
```

**Why static?** This is a C best practice:
- Encapsulates metadata within the compilation unit
- Avoids namespace pollution
- Prevents symbol conflicts when multiple .c files define metadata

**The Challenge:** Static symbols are **not visible** across compilation units or to Go via CGO!

```go
// ❌ THIS DOESN'T WORK
extractCMetadata(&C.cgocopy_metadata_SimplePerson)  // Error: undefined
```

## Solutions Attempted

### ❌ Attempt 1: Use extern declarations in header

```c
// structs.h
extern cgocopy_field_info cgocopy_fields_SimplePerson[];
extern cgocopy_struct_info cgocopy_metadata_SimplePerson;
```

**Problem:** Creates conflict with static declarations from the macro.

### ❌ Attempt 2: Override macro with CGOCOPY_IMPLEMENTATION pattern

```c
// cgocopy_export.h - stb-style header-only pattern
#ifdef CGOCOPY_IMPLEMENTATION
  #undef CGOCOPY_STRUCT
  #define CGOCOPY_STRUCT(structtype, ...) \
      cgocopy_field_info cgocopy_fields_##structtype[] = { ... };  // Non-static
#endif
```

**Problem:** Complex include ordering issues. CGO compiles all files together and the macro redefinition timing becomes fragile.

### ❌ Attempt 3: Override macro in CGO preamble

```go
/*
#undef CGOCOPY_STRUCT
#define CGOCOPY_STRUCT(structtype, ...) \
    cgocopy_field_info cgocopy_fields_##structtype[] = { ... };  // Non-static
#include "native/structs.c"
*/
```

**Status:** ✅ This actually WORKS but has drawbacks:
- CGO-specific hack
- Doesn't work with traditional C compilation
- Not production-ready
- Violates C encapsulation principles

## ✅ The Correct Solution: Metadata API with Getter Functions

This is the **standard C library pattern** for exposing internal static data.

### Architecture

```
native/
├── structs.h            # Public API: struct definitions, functions
├── structs.c            # Implementation: CGOCOPY_STRUCT creates static metadata
└── metadata_api.h       # Getter functions: bridge from static to public
```

### Implementation

**1. structs.c - Generate static metadata**
```c
#include "../../native2/cgocopy_macros.h"
#include "metadata_api.h"

typedef struct {
    int id;
    double score;
} SimplePerson;

// Macro creates STATIC metadata
CGOCOPY_STRUCT(SimplePerson,
    CGOCOPY_FIELD(SimplePerson, id),
    CGOCOPY_FIELD(SimplePerson, score)
)

// Getter function is NON-STATIC and visible to CGO
const cgocopy_struct_info* get_SimplePerson_metadata(void) {
    return &cgocopy_metadata_SimplePerson;  // Access static metadata
}
```

**2. metadata_api.h - Declare getters**
```c
#include "../../native2/cgocopy_macros.h"

const cgocopy_struct_info* get_SimplePerson_metadata(void);
const cgocopy_struct_info* get_User_metadata(void);
// ... one getter per struct type
```

**3. integration_cgo.go - Call getters**
```go
/*
#cgo CFLAGS: -I${SRCDIR}/../native2
#include "native/metadata_api.h"
#include "native/structs.c"  // CGO requires including .c files
*/
import "C"

func init() {
    // ✅ Call getter function to access static metadata
    cgocopy2.PrecompileWithC[SimplePerson](
        extractCMetadata(C.get_SimplePerson_metadata()),
    )
}
```

## Why This Is The Right Solution

### Benefits

1. **Follows C Best Practices**
   - Static metadata stays encapsulated
   - Clean public API surface
   - No namespace pollution

2. **Works with Traditional C Compilation**
   - Can compile C library separately
   - Can link with LDFLAGS
   - Not CGO-specific

3. **Type Safe**
   - Getter returns `const` pointer
   - Can't accidentally modify metadata
   - Clear ownership model

4. **Standard Pattern**
   - Same approach as OpenSSL, SQLite, etc.
   - Familiar to C developers
   - Well-documented pattern

5. **Maintainable**
   - Clear separation of concerns
   - Easy to add new types (just add getter)
   - No macro magic in CGO preamble

### Comparison

| Approach | Works? | Production-Ready? | C-Compatible? | Maintainable? |
|----------|--------|-------------------|---------------|---------------|
| Direct static access | ❌ No | - | - | - |
| extern declarations | ❌ No | - | - | - |
| CGOCOPY_IMPLEMENTATION | ❌ No | - | - | - |
| Macro override in CGO | ✅ Yes | ❌ No | ❌ No | ⚠️ Fragile |
| **Getter functions** | **✅ Yes** | **✅ Yes** | **✅ Yes** | **✅ Yes** |

## Symbol Visibility Rules

| Symbol Type | Static? | Visible in .c? | Visible in Go? | Solution |
|-------------|---------|----------------|----------------|----------|
| Metadata structs | Yes | ❌ No | ❌ No | Need getter |
| Getter functions | No | ✅ Yes | ✅ Yes | ✅ Use this! |
| Types in headers | - | ✅ Yes | ✅ Yes | ✅ Good |
| Macros | - | ⚠️ .h only | ⚠️ .h only | Include properly |

## Usage Pattern

### Adding a New Type

**Step 1:** Define struct and metadata in C
```c
// native/structs.c
typedef struct {
    int field1;
    float field2;
} MyNewType;

CGOCOPY_STRUCT(MyNewType,
    CGOCOPY_FIELD(MyNewType, field1),
    CGOCOPY_FIELD(MyNewType, field2)
)

const cgocopy_struct_info* get_MyNewType_metadata(void) {
    return &cgocopy_metadata_MyNewType;
}
```

**Step 2:** Declare getter in API
```c
// native/metadata_api.h
const cgocopy_struct_info* get_MyNewType_metadata(void);
```

**Step 3:** Register in Go
```go
// integration_cgo.go
func init() {
    cgocopy2.PrecompileWithC[MyNewType](
        extractCMetadata(C.get_MyNewType_metadata()),
    )
}
```

**That's it!** No macro overrides, no CGO hacks, just clean API usage.

## Test Results

✅ **93 core tests passing** (cgocopy2 package)
✅ **9 integration tests passing**:
- TestIntegration_SimplePerson
- TestIntegration_User
- TestIntegration_Student  
- TestIntegration_GameObject
- TestIntegration_AllTypes
- TestIntegration_FastCopy_Int32
- TestIntegration_FastCopy_Float64
- TestIntegration_Validation
- TestIntegration_GetRegisteredTypes

## Key Insight

> **Static metadata needs getter functions because static symbols aren't visible across compilation units or to CGO.**

This is not a workaround or hack - it's the **standard C library pattern** for exposing internal implementation details through a clean API. Libraries like OpenSSL, SQLite, and many others use this exact pattern.

## References

- C standard: static storage duration and linkage
- CGO documentation: symbol visibility rules
- stb libraries: header-only pattern (attempted but not needed)
- OpenSSL, SQLite: real-world examples of getter function patterns

## Conclusion

The getter function pattern is:
- ✅ The correct solution
- ✅ Production-ready
- ✅ Following C best practices
- ✅ Maintainable and clear
- ✅ What experienced C developers would expect

**Do not** try to make static symbols visible. **Do** provide an API layer with getter functions.
