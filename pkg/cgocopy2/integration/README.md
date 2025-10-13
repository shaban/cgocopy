# CGO Integration Tests - Proper C Library Structure

This directory demonstrates the **correct way** to integrate cgocopy with C libraries using C11 macros.

## Directory Structure

```
integration/
├── native/                    # C library code (like a real project)
│   ├── structs.h             # Struct definitions and function declarations
│   ├── structs.c             # Implementations with CGOCOPY_STRUCT macros
│   └── metadata_api.h        # API for Go to access metadata
├── integration_cgo.go        # Go CGO bridge
└── integration_test.go       # Integration tests
```

## Key Design Principles

### 1. Static Metadata Problem

The `CGOCOPY_STRUCT` macro (defined in `native2/cgocopy_macros.h`) creates **static** metadata:

```c
// From cgocopy_macros.h
#define CGOCOPY_STRUCT(structtype, ...) \
    static cgocopy_field_info cgocopy_fields_##structtype[] = { ... }; \
    static cgocopy_struct_info cgocopy_metadata_##structtype = { ... };
```

**Problem:** Static symbols are **not visible** to Go via CGO!

### 2. Solution: Getter Functions

We provide a **metadata API** layer with getter functions:

```c
// metadata_api.h - Declares getters
const cgocopy_struct_info* get_SimplePerson_metadata(void);
const cgocopy_struct_info* get_User_metadata(void);
// ...

// structs.c - Implements getters
const cgocopy_struct_info* get_SimplePerson_metadata(void) {
    return &cgocopy_metadata_SimplePerson;  // Access static metadata
}
```

These getter functions are **non-static** and therefore visible to CGO!

### 3. CGO Integration

```go
/*
#cgo CFLAGS: -I${SRCDIR}/../native2
#include "native/metadata_api.h"
#include "native/structs.c"  // CGO requires including .c files
*/
import "C"

func init() {
    // Call getter function to access static metadata
    cgocopy2.PrecompileWithC[SimplePerson](
        extractCMetadata(C.get_SimplePerson_metadata()),
    )
}
```

## Why This Pattern?

### What Doesn't Work

❌ **Direct access to static symbols:**
```go
// FAILS: static symbols not visible to CGO
extractCMetadata(&C.cgocopy_metadata_SimplePerson)
```

❌ **Overriding macros in CGO preamble:**
```go
// WORKS but not production-ready - too CGO-specific
#undef CGOCOPY_STRUCT
#define CGOCOPY_STRUCT(structtype, ...) \
    cgocopy_field_info cgocopy_fields_##structtype[] = { ... };
```

### What Works

✅ **Getter functions API:**
```go
// CORRECT: Call getter function that returns pointer to static metadata
extractCMetadata(C.get_SimplePerson_metadata())
```

**Benefits:**
- Works with standard C compilation
- Follows C library best practices
- Static metadata stays encapsulated in .c file
- API is clean and explicit
- Can be compiled separately and linked (not just CGO)

## File Responsibilities

### `native/structs.h`
- Struct type definitions
- Constructor/destructor function declarations
- **Does NOT include cgocopy headers** (keeps C API clean)

### `native/structs.c`
- Includes cgocopy_macros.h
- Uses `CGOCOPY_STRUCT` macros to generate metadata
- Implements constructor/destructor functions
- **Implements getter functions** for metadata access

### `native/metadata_api.h`
- Declares getter functions like `get_XXX_metadata()`
- This is what **Go imports** via CGO
- Small, focused API surface

### `integration_cgo.go`
- CGO preamble includes `metadata_api.h`
- Includes `.c` file(s) for CGO compilation
- Calls getter functions to access metadata
- Registers types with cgocopy2

## CGO Build Rules

1. **Include .c files in CGO preamble:**
   ```go
   /*
   #include "native/structs.c"
   */
   ```
   CGO doesn't automatically compile `.c` files in subdirectories.

2. **Or use LDFLAGS approach:**
   ```go
   /*
   #cgo LDFLAGS: -L${SRCDIR}/native -lmystruct
   #include "native/metadata_api.h"
   */
   ```
   Then provide a Makefile to build `libmystruct.a` separately.

3. **Static symbols need getters:**
   - Cannot access: `C.Point_info` (static, invisible)
   - Must call: `C.get_point_metadata()` (function, visible)

## Usage Example

### Step 1: Define struct in C

```c
// native/mystruct.c
#include "../native2/cgocopy_macros.h"
#include "metadata_api.h"

typedef struct {
    int id;
    float value;
} MyStruct;

// Use macro to generate metadata
CGOCOPY_STRUCT(MyStruct,
    CGOCOPY_FIELD(MyStruct, id),
    CGOCOPY_FIELD(MyStruct, value)
)

// Implement getter
const cgocopy_struct_info* get_MyStruct_metadata(void) {
    return &cgocopy_metadata_MyStruct;
}
```

### Step 2: Declare getter in API header

```c
// native/metadata_api.h
const cgocopy_struct_info* get_MyStruct_metadata(void);
```

### Step 3: Register in Go

```go
// integration_cgo.go
func init() {
    cgocopy2.PrecompileWithC[MyStruct](
        extractCMetadata(C.get_MyStruct_metadata()),
    )
}
```

### Step 4: Use cgocopy

```go
// your_code.go
cStruct := C.create_my_struct(42, 3.14)
defer C.free(unsafe.Pointer(cStruct))

goStruct := cgocopy2.Copy[MyStruct](unsafe.Pointer(cStruct))
```

## Symbol Visibility Table

| Symbol Type | Visible in other .c? | Visible in Go? | Solution |
|-------------|---------------------|----------------|----------|
| `static const struct_info_t` | ❌ No | ❌ No | Need getter |
| Function returning pointer | ✅ Yes | ✅ Yes | ✅ Use this! |
| Macro definitions | ❌ No* | ❌ No* | Must be in .h |
| Header-defined types | ✅ Yes | ✅ Yes | ✅ Good |

\* Unless in an included header file

## Testing

```bash
# Run integration tests
cd pkg/cgocopy2/integration
go test -v

# All 9 tests should pass:
# ✅ TestIntegration_SimplePerson
# ✅ TestIntegration_User
# ✅ TestIntegration_Student
# ✅ TestIntegration_GameObject
# ✅ TestIntegration_AllTypes
# ✅ TestIntegration_FastCopy_Int32
# ✅ TestIntegration_FastCopy_Float64
# ✅ TestIntegration_Validation
# ✅ TestIntegration_GetRegisteredTypes
```

## Summary

This integration pattern demonstrates:

1. ✅ **Clean C library structure** - native/ directory with proper .h/.c separation
2. ✅ **C11 macros working automatically** - no manual field info writing
3. ✅ **Proper CGO integration** - using getter functions for static metadata access
4. ✅ **Production-ready pattern** - follows C best practices
5. ✅ **Type safety** - Go generics + C metadata = compile-time checks

The key insight: **Static metadata needs getter functions** because static symbols aren't visible across compilation units or to CGO. This is the standard solution used in real C libraries.
