# Phase 9: Code Generation Tool

## Problem

Currently, integrating a C struct with cgocopy requires **manual boilerplate in 5 places**:

```c
// 1. structs.h - Define struct
typedef struct {
    int id;
    double score;
} SimplePerson;

// 2. structs.c - Use CGOCOPY_STRUCT macro
CGOCOPY_STRUCT(SimplePerson,
    CGOCOPY_FIELD(SimplePerson, id),
    CGOCOPY_FIELD(SimplePerson, score)
)

// 3. structs.c - Implement getter
const cgocopy_struct_info* get_SimplePerson_metadata(void) {
    return &cgocopy_metadata_SimplePerson;
}

// 4. metadata_api.h - Declare getter
const cgocopy_struct_info* get_SimplePerson_metadata(void);

// 5. integration_cgo.go - Register in init()
cgocopy2.PrecompileWithC[SimplePerson](
    extractCMetadata(C.get_SimplePerson_metadata()),
)
```

**Pain points:**
- ❌ Repetitive and error-prone
- ❌ Easy to forget a step
- ❌ Must update 4 files when changing struct
- ❌ Manual field listing in CGOCOPY_STRUCT

## Solution: `cgocopy-generate` Tool

A lightweight Go code generator that:
1. Parses C struct definitions from `.h` files
2. Generates `CGOCOPY_STRUCT` metadata in `.c` file
3. Generates getter functions in `.c` file
4. Generates API header with declarations

### New Workflow

```c
// 1. structs.h - Define struct (ONLY MANUAL STEP)
typedef struct {
    int id;
    double score;
} SimplePerson;
```

```go
// 2. integration_cgo.go - Add go:generate directive
//go:generate cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h
```

```bash
# 3. Run generator
$ go generate ./...
Found 6 struct(s): SimplePerson, User, Student, Point3D, GameObject, AllTypes
Generated: native/structs_meta.c
Generated: native/metadata_api.h
```

**All boilerplate is now generated!**

## Tool Implementation

### Command Line Interface

```bash
cgocopy-generate -input=FILE.h [-output=FILE_meta.c] [-api=API.h]
```

**Flags:**
- `-input` (required) - Input C header file with struct definitions
- `-output` (optional) - Output C file with metadata (default: `{input}_meta.c`)
- `-api` (optional) - Output header file with getter declarations

### Parser

**Regex-based parser** (sufficient for 90% of use cases):

```go
// Match: struct Name { ... }
structRegex := regexp.MustCompile(`(?s)struct\s+(\w+)\s*\{([^}]+)\}`)

// Match fields: type name; or type name[size];
fieldRegex := regexp.MustCompile(`([a-zA-Z_][\w\s\*]+?)\s+(\w+)(?:\[(\d+)\])?\s*;`)
```

**Handles:**
- ✅ Primitive types (`int`, `float`, `double`, `_Bool`)
- ✅ Pointers (`char*`, `int*`)
- ✅ Arrays (`int arr[10]`, `char name[32]`)
- ✅ Nested structs (`Point3D position`)
- ✅ C/C++ comments
- ✅ Multiple structs per file

**Limitations** (acceptable for v1):
- ❌ Complex pointer types (`int *const *`)
- ❌ Anonymous nested structs
- ❌ Macros in struct definitions
- ❌ Cross-file type resolution

### Output Templates

#### Metadata Implementation (`structs_meta.c`)

```c
// GENERATED CODE - DO NOT EDIT
// Generated from: native/structs.h

#include "../../native2/cgocopy_macros.h"

// Metadata for SimplePerson
CGOCOPY_STRUCT(SimplePerson,
    CGOCOPY_FIELD(SimplePerson, id),
    CGOCOPY_FIELD(SimplePerson, score)
)

const cgocopy_struct_info* get_SimplePerson_metadata(void) {
    return &cgocopy_metadata_SimplePerson;
}

// Metadata for User
CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, user_id),
    CGOCOPY_FIELD(User, username),
    CGOCOPY_FIELD(User, email)
)

const cgocopy_struct_info* get_User_metadata(void) {
    return &cgocopy_metadata_User;
}

// ... more structs
```

#### API Header (`metadata_api.h`)

```c
// GENERATED CODE - DO NOT EDIT
// Generated from: native/structs.h

#ifndef METADATA_API_H
#define METADATA_API_H

#include "../../native2/cgocopy_macros.h"

// Getter functions for each struct
const cgocopy_struct_info* get_SimplePerson_metadata(void);
const cgocopy_struct_info* get_User_metadata(void);
const cgocopy_struct_info* get_Student_metadata(void);
const cgocopy_struct_info* get_Point3D_metadata(void);
const cgocopy_struct_info* get_GameObject_metadata(void);
const cgocopy_struct_info* get_AllTypes_metadata(void);

#endif // METADATA_API_H
```

### Integration with CGO

```go
package integration

//go:generate cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h

/*
#cgo CFLAGS: -I${SRCDIR}/../native2
#include "native/metadata_api.h"
#include "native/structs_meta.c"  // Generated file
*/
import "C"

func init() {
    // Same as before - call getters
    cgocopy2.PrecompileWithC[SimplePerson](
        extractCMetadata(C.get_SimplePerson_metadata()),
    )
}
```

## Benefits

### For Users

1. **80% less boilerplate** - only write struct definition
2. **Type-safe** - parses actual C code
3. **No manual synchronization** - regenerate after changes
4. **CI verification** - `go generate` + `git diff` catches stale code
5. **Fast iteration** - change struct, regenerate, test

### For cgocopy

1. **Lower barrier to entry** - easier onboarding
2. **Fewer user errors** - no manual field listing
3. **Better documentation** - clear workflow to follow
4. **Professional polish** - expected feature for modern Go tools

### Technical

1. **Fast** - stdlib only, < 10ms for typical files
2. **Portable** - pure Go, no external dependencies
3. **Maintainable** - ~200 lines of straightforward code
4. **Extensible** - easy to add features

## Implementation Plan

### Phase 9a: Build Tool

1. Create `tools/cgocopy-generate/`
2. Implement parser with regex
3. Implement templates for `.c` and `.h` generation
4. Add tests with sample C headers
5. Document tool in README

**Estimated:** 2-3 hours

### Phase 9b: Refactor Integration Tests

1. Delete manual `structs.c` and `metadata_api.h`
2. Add `//go:generate` directive
3. Run generator
4. Verify all 9 tests still pass

**Estimated:** 1 hour

### Phase 9c: Documentation

1. Update main README with generator workflow
2. Create tutorial: "Adding Your First Struct"
3. Add to `integration/README.md`

**Estimated:** 1 hour

## Example: Before vs After

### Before (Manual)

**Files to edit:** 4
**Lines to write:** ~30
**Time:** 5-10 minutes
**Risk:** Typos, missing steps

```c
// structs.h
typedef struct { int id; } Person;

// structs.c
CGOCOPY_STRUCT(Person, CGOCOPY_FIELD(Person, id))
const cgocopy_struct_info* get_Person_metadata(void) {
    return &cgocopy_metadata_Person;
}

// metadata_api.h
const cgocopy_struct_info* get_Person_metadata(void);

// integration_cgo.go
cgocopy2.PrecompileWithC[Person](extractCMetadata(C.get_Person_metadata()))
```

### After (Generated)

**Files to edit:** 1
**Lines to write:** 3
**Time:** 30 seconds
**Risk:** None

```c
// structs.h
typedef struct { int id; } Person;
```

```bash
$ go generate ./...
Generated: native/structs_meta.c
Generated: native/metadata_api.h
```

**That's it!**

## Comparison with Alternatives

| Approach | Setup | Speed | Dependencies | Maintainability |
|----------|-------|-------|--------------|-----------------|
| Manual | None | N/A | None | ❌ High burden |
| libclang | Complex | Slow | LLVM | ⚠️ Heavy |
| pycparser | Medium | Medium | Python | ⚠️ External |
| **Regex parser** | **Simple** | **Fast** | **None** | **✅ Easy** |

## Risks & Mitigation

**Risk:** Parser can't handle complex C syntax
**Mitigation:** Document limitations, start simple, extend as needed

**Risk:** Generated code gets out of sync
**Mitigation:** CI check with `go generate` + `git diff --exit-code`

**Risk:** Users don't run generator
**Mitigation:** Clear error messages, pre-commit hooks, CI enforcement

## Success Metrics

- ✅ Tool generates correct metadata for all integration test structs
- ✅ All 9 integration tests pass with generated code
- ✅ Tool runs in < 100ms
- ✅ Zero external dependencies
- ✅ < 300 lines of code
- ✅ Works with `go generate`

## Future Enhancements (v2)

1. **Auto-registration in Go** - generate Go init() code too
2. **Type validation** - verify Go struct matches C struct
3. **Better error messages** - line numbers, suggestions
4. **IDE integration** - VS Code extension for live generation
5. **Advanced parsing** - handle preprocessor, macros, typedefs

## Conclusion

This tool would transform cgocopy from "powerful but verbose" to "powerful and effortless."

**Recommendation:** Build it! This is the missing piece that makes cgocopy v2 truly production-ready.

## Questions for Discussion

1. Should we auto-generate Go registration code too?
2. Should we validate Go struct matches C struct?
3. Should we support multiple input files?
4. Should we add a `--watch` mode for development?

## Next Steps

1. ✅ Get approval for Phase 9
2. 🚧 Create `tools/cgocopy-generate/`
3. 🚧 Implement parser + templates
4. 🚧 Add tests
5. 🚧 Refactor integration tests
6. 🚧 Update documentation
7. 🚧 Release cgocopy v2.0.0!
