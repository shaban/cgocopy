# cgocopy2 Native C11 Macros

This directory contains simplified C11 macros for defining struct metadata used by cgocopy2.

## ⚠️ Requirements

**C11 or later** is required for `_Generic` support used in automatic type detection.

- GCC 4.9+ (5.0+ recommended)
- Clang 3.3+ (any recent version)
- MSVC 2015+ (VS 2015 Update 3 or later)

## Files

- `cgocopy_macros.h` - Main header with CGOCOPY_STRUCT and CGOCOPY_FIELD macros
- `example.c` - Example usage with various struct types
- `test_macros.sh` - Test script to verify compilation

## Usage

### Basic Example

```c
#include "cgocopy_macros.h"

typedef struct {
    int id;
    char* name;
    double score;
} Person;

// Register the struct (automatic type detection via C11 _Generic)
CGOCOPY_STRUCT(Person,
    CGOCOPY_FIELD(Person, id),
    CGOCOPY_FIELD(Person, name),
    CGOCOPY_FIELD(Person, score)
)
```

### With Arrays

For array fields, use `CGOCOPY_ARRAY_FIELD` and specify the element type:

```c
typedef struct {
    int id;
    int scores[5];
    float values[3];
} Student;

CGOCOPY_STRUCT(Student,
    CGOCOPY_FIELD(Student, id),
    CGOCOPY_ARRAY_FIELD(Student, scores, int),
    CGOCOPY_ARRAY_FIELD(Student, values, float)
)
```

### In Go with CGO

```go
package mypackage

/*
#cgo CFLAGS: -std=c11
#include "native2/cgocopy_macros.h"

// Define your C structs here
typedef struct {
    int id;
    char* name;
} Person;

CGOCOPY_STRUCT(Person,
    CGOCOPY_FIELD(Person, id),
    CGOCOPY_FIELD(Person, name)
)
*/
import "C"

type Person struct {
    ID   int    `cgocopy:"id"`
    Name string `cgocopy:"name"`
}

func init() {
    cgocopy2.Precompile[Person]()
}

func ConvertPerson(cPtr unsafe.Pointer) (Person, error) {
    return cgocopy2.Copy[Person](cPtr)
}
```

## Features

### Automatic Type Detection

The macros use C11 `_Generic` to automatically detect field types:

- `_Bool` → "bool"
- `int8_t`, `char` → "int8"
- `uint8_t`, `unsigned char` → "uint8"
- `int16_t`, `short` → "int16"
- `uint16_t`, `unsigned short` → "uint16"
- `int32_t`, `int` → "int32"
- `uint32_t`, `unsigned int` → "uint32"
- `int64_t`, `long`, `long long` → "int64"
- `uint64_t`, `unsigned long`, `unsigned long long` → "uint64"
- `float` → "float32"
- `double` → "float64"
- `char*`, `const char*` → "string"
- Other → "struct"

### Metadata Generation

Each `CGOCOPY_STRUCT` call generates:

1. `cgocopy_fields_<StructName>[]` - Array of field metadata
2. `cgocopy_metadata_<StructName>` - Struct metadata with:
   - Struct name
   - Total size
   - Field count
   - Pointer to field array

### Field Metadata

Each field includes:

- `name` - Field name as string
- `type` - Type name (auto-detected)
- `offset` - Byte offset in struct
- `size` - Size in bytes
- `is_pointer` - 1 if pointer, 0 otherwise
- `is_array` - 1 if array, 0 otherwise
- `array_len` - Array length (0 if not array)

## Testing

Run the test script to verify everything compiles:

```bash
cd pkg/cgocopy2/native2
./test_macros.sh
```

Expected output shows metadata for all example structs with correct:
- Struct sizes
- Field offsets
- Type detection
- Array length detection

## Limitations

1. **Array Detection**: Arrays require using `CGOCOPY_ARRAY_FIELD` macro instead of `CGOCOPY_FIELD`
2. **Nested Structs**: Nested structs are detected as type "struct" - you need to register them separately
3. **Function Pointers**: Not supported
4. **Unions**: Not supported
5. **Bit Fields**: Not supported

## Implementation Notes

### Why C11?

The `_Generic` keyword provides compile-time type selection, allowing automatic type detection without:
- Manual type strings
- Preprocessor token pasting tricks
- Runtime type checking

This makes the API cleaner and safer:

```c
// ❌ Old way (cgocopy v1)
CGOCOPY_REGISTER_FIELD(MyStruct, id, "int32", offsetof(MyStruct, id))

// ✅ New way (cgocopy v2 with C11)
CGOCOPY_FIELD(MyStruct, id)  // Type auto-detected!
```

### Why Two Macros for Fields?

Array element types cannot be reliably determined via `_Generic` alone, as arrays decay to pointers in many contexts. Rather than complex heuristics, we provide:

- `CGOCOPY_FIELD(StructType, field)` - For primitives, strings, pointers, structs
- `CGOCOPY_ARRAY_FIELD(StructType, field, elemtype)` - For arrays (requires element type)

This is explicit and prevents errors from incorrect array length calculations.

## See Also

- `../README.md` - Main cgocopy2 documentation
- `example.c` - Comprehensive usage examples
- `cgocopy_macros.h` - Full macro implementation with comments

