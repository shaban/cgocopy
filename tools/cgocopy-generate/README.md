# cgocopy-generate

Code generation tool for `cgocopy` - automatically generates C metadata from C struct definitions.

## Purpose

Eliminates 80% of boilerplate code when integrating C structs with cgocopy. Instead of manually writing:
- `CGOCOPY_STRUCT` macros for each struct
- `CGOCOPY_FIELD` entries for each field
- Getter functions for metadata access

You simply write your C struct definitions and run `go generate`.

## Installation

```bash
cd tools/cgocopy-generate
go build
```

## Usage

### Command Line

```bash
cgocopy-generate -input=structs.h [-output=structs_meta.c] [-api=metadata_api.h] [-header-path=path]
```

**Flags:**
- `-input` (required): Path to C header file containing struct definitions
- `-output` (optional): Path to generated metadata file (default: `<input>_meta.c`)
- `-api` (optional): Path to generated API header file with getter declarations
- `-header-path` (optional): Path to cgocopy_macros.h (default: auto-detected from go.mod)

### With go generate

Add a directive to your CGO file:

```go
//go:generate ../../../tools/cgocopy-generate/cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h

/*
#include "native/metadata_api.h"
#include "native/structs_meta.c"  // Generated metadata
*/
import "C"
```

Then run:
```bash
go generate ./...
```

## Supported C Syntax

The tool parses standard C struct definitions:

```c
// Simple struct
struct Point {
    double x;
    double y;
};

// Typedef struct
typedef struct {
    int id;
    char* name;
} Person;

// Named typedef
typedef struct User {
    int user_id;
    char* email;
} User;

// Arrays
struct Scores {
    int grades[5];
    float values[10];
};

// Nested structs
struct GameObject {
    char* name;
    Point position;    // Nested struct
    Point velocity;
};

// Pointers
struct Node {
    int value;
    struct Node* next;
};
```

**Supported:**
- Primitive types (int, float, double, char, etc.)
- Pointers (`char*`, `void*`, etc.)
- Fixed-size arrays (`int arr[10]`)
- Nested structs
- C and C++ style comments
- Multiple structs in one file
- `typedef struct` patterns

**Not Supported:**
- Flexible array members (`int arr[]`)
- Function pointers
- Bitfields
- Unions
- Variable-length arrays

## Generated Output

### Metadata File (`_meta.c`)

Contains:
1. `CGOCOPY_STRUCT` macros for each struct
2. `get_TypeName_metadata()` getter functions

Example:
```c
// GENERATED CODE - DO NOT EDIT
// Generated from: structs.h

#include <stdlib.h>
#include "../../../pkg/cgocopy/native/cgocopy_macros.h"
#include "structs.h"

// Metadata for Person
CGOCOPY_STRUCT(Person,
    CGOCOPY_FIELD(Person, id),
    CGOCOPY_FIELD(Person, name)
)

const cgocopy_struct_info* get_Person_metadata(void) {
    return &cgocopy_metadata_Person;
}
```

### API Header (`_api.h`)

Contains getter function declarations:
```c
// GENERATED CODE - DO NOT EDIT
// Generated from: structs.h

#ifndef METADATA_API_H
#define METADATA_API_H

#include "../../../pkg/cgocopy/native/cgocopy_macros.h"

// Getter functions for each struct
const cgocopy_struct_info* get_Person_metadata(void);

#endif // METADATA_API_H
```

## Header Path Auto-Detection

The tool automatically detects the path to `cgocopy_macros.h`:

1. Finds `go.mod` by walking up from the output directory
2. Locates `pkg/cgocopy/native/cgocopy_macros.h` from module root
3. Calculates relative path from output location to macros file

**Example:**
- Output: `examples/users/native/structs_meta.c`
- Module root: `/path/to/project/`
- Macros: `/path/to/project/pkg/cgocopy/native/cgocopy_macros.h`
- Generated include: `#include "../../../pkg/cgocopy/native/cgocopy_macros.h"`

Override auto-detection with `-header-path` flag if needed:
```bash
cgocopy-generate -input=structs.h -output=metadata.c -header-path=custom/path/cgocopy_macros.h
```

## Performance

- Fast: < 10ms for typical header files
- Zero external dependencies (stdlib only)
- Regex-based parser (simple, no complex AST)
- Auto-detects header paths (no hardcoded paths)

## Example Workflow

**Before** (manual - 5-10 minutes per struct):
1. Write C struct definition
2. Manually write `CGOCOPY_STRUCT` macro
3. Manually add `CGOCOPY_FIELD` for each field
4. Write getter function
5. Add declaration to header
6. Update CGO imports

**After** (automatic - 30 seconds):
1. Write C struct definition in `structs.h`
2. Run `go generate`

## Complete Examples

See `examples/users/` for a complete working example with:
- 6 different struct types (simple, nested, arrays, all types)
- Generated metadata (`native/structs_meta.c`)
- Generated API header (`native/metadata_api.h`)
- Helper functions (`native/helpers.c`) - NOT auto-generated
- 9 passing integration tests

## Architecture

```
tools/cgocopy-generate/
├── main.go           # CLI interface
├── parser.go         # Regex-based C struct parser
├── generator.go      # Template-based code generation
├── parser_test.go    # 11 comprehensive tests
└── README.md         # This file
```

## Testing

```bash
go test -v
```

11 tests covering:
- Simple structs
- Pointers
- Arrays
- Nested structs
- Multiple structs
- Comments (line and block)
- Edge cases
- All primitive types

## Design Rationale

**Why regex instead of a full C parser?**
- Simplicity: 280 lines of code vs thousands for a full parser
- Speed: < 10ms vs seconds for complex parsers
- Dependencies: Zero external deps vs clang/libclang
- Sufficient: Handles 99% of common struct patterns
- Maintainable: Easy to understand and modify

**Why generate C code instead of Go?**
- C11 `_Generic` provides type-safe macros (Phase 7)
- Compile-time validation of field names/types
- Works with any C library (not Go-specific)
- Matches industry patterns (similar to C++ template generation)

## Limitations

- Only parses struct definitions, not macros or preprocessor directives
- Assumes standard C formatting (works with most code formatters)
- Does not validate that referenced types exist
- No support for C++ classes or templates

## Future Enhancements

Potential improvements:
- [ ] Support for union types
- [ ] Bitfield handling
- [ ] Function pointer fields
- [ ] Custom template support
- [ ] Multiple input files
- [ ] Watch mode for development

## License

Same as cgocopy project.
