# cgocopy C11 Macros

C11 macros for defining struct metadata used by cgocopy.

## Requirements

**C11 or later** required for `_Generic` support:
- GCC 4.9+ (5.0+ recommended)
- Clang 3.3+
- MSVC 2015+

## Files

- `cgocopy_macros.h` - Core macros for struct metadata generation

## Usage

### With cgocopy-generate (Recommended)

The `cgocopy-generate` tool automatically creates metadata from C struct definitions:

```go
//go:generate cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h
```

See `examples/users/` for a complete working example.

### Manual Usage

If you prefer to write metadata manually:

```c
#include "cgocopy/native/cgocopy_macros.h"

typedef struct {
    int id;
    char* name;
    double score;
} Person;

CGOCOPY_STRUCT(Person,
    CGOCOPY_FIELD(Person, id),
    CGOCOPY_FIELD(Person, name),
    CGOCOPY_FIELD(Person, score)
)
```

### With Arrays

Use `CGOCOPY_ARRAY_FIELD` for array fields:

```c
typedef struct {
    int id;
    int scores[5];
} Student;

CGOCOPY_STRUCT(Student,
    CGOCOPY_FIELD(Student, id),
    CGOCOPY_ARRAY_FIELD(Student, scores, int)
)
```

## Automatic Type Detection

The macros use C11 `_Generic` to detect types automatically:

- `_Bool` → "bool"
- `int8_t`, `char` → "int8"  
- `uint8_t` → "uint8"
- `int16_t`, `short` → "int16"
- `uint16_t` → "uint16"
- `int32_t`, `int` → "int32"
- `uint32_t` → "uint32"
- `int64_t`, `long`, `long long` → "int64"
- `uint64_t` → "uint64"
- `float` → "float32"
- `double` → "float64"
- `char*`, `const char*` → "string"
- Other → "struct"

## Generated Metadata

Each `CGOCOPY_STRUCT` generates:

- `cgocopy_fields_<StructName>[]` - Field metadata array
- `cgocopy_metadata_<StructName>` - Struct metadata with:
  - Struct name
  - Total size
  - Field count
  - Fields array pointer

Field metadata includes:
- `name` - Field name
- `type` - Type string (auto-detected)
- `offset` - Byte offset
- `size` - Size in bytes  
- `is_pointer` - Pointer flag
- `is_array` - Array flag
- `array_len` - Array length

## Limitations

- Arrays require `CGOCOPY_ARRAY_FIELD` (cannot auto-detect element type)
- Nested structs detected as "struct" (register separately)
- Function pointers not supported
- Unions not supported
- Bit fields not supported

## See Also

- `../../README.md` - Main cgocopy documentation
- `../../examples/users/` - Complete working example
- `../../tools/cgocopy-generate/` - Code generation tool
