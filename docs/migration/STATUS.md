# cgocopy v2 Migration Status

## Overview
This document tracks the status of the cgocopy v2 implementation based on the 8-phase plan detailed in `IMPLEMENTATION_PLAN.md`.

## Design Phase ✅ COMPLETE

### Completed Documents
- ✅ **API_SUGGESTIONS.md** - Original improvement ideas (archived as reference)
- ✅ **API_IMPROVEMENTS.md** - Refined v2 specification with agreed-upon improvements
- ✅ **IMPLEMENTATION_PLAN.md** - Detailed 8-phase implementation roadmap with test strategies

### Key Design Decisions
1. **Separate Package**: Implement in `pkg/cgocopy2/` to avoid breaking v1
2. **Simplified Macros**: Use C11 `_Generic` for auto-detection in `native2/` directory
   - `CGOCOPY_STRUCT(type, ...)` replaces `CGOCOPY_STRUCT_BEGIN/END`
   - `CGOCOPY_FIELD(type, field)` auto-detects field types
3. **Thread Safety**: `sync.RWMutex` instead of atomic finalization pattern
4. **Precompilation Only**: No lazy loading, explicit `Precompile[T any]()` at init time
5. **Tag Support**: Idiomatic Go with `cgocopy:"field_name"` struct tags
6. **FastCopy**: Zero-allocation copying for primitives via `FastCopy[T any]()`
7. **Validation**: `ValidateStruct[T any]()` helper for debugging

## Implementation Status

### Phase 1: Setup & Basic Types ✅ COMPLETE
**Status**: Complete  
**Target**: Basic package structure with type system and error handling

**Completed Tasks**:
- ✅ Create `pkg/cgocopy2/` directory
- ✅ Implement `types.go` with core types (FieldType, FieldInfo, StructMetadata, Registry)
- ✅ Implement `errors.go` with error constants (8 error types + 3 error structs)
- ✅ Create `types_test.go` with 17 tests (all passing)
- ✅ Create `errors_test.go` with 10 tests (all passing)
- ✅ Thread-safe Registry with RWMutex implementation
- ✅ Package README with architecture overview

**Test Results**: 27 tests, 0 failures, 100% pass rate

### Phase 2: Registry & Precompile � NEXT
**Status**: Ready to Start  
**Target**: Thread-safe registry with RWMutex and Precompile function

**Tasks**:
- [ ] Implement `Precompile[T any]()` function
- [ ] C metadata loader integration
- [ ] Type validation logic
- [ ] Precompile tests

### Phase 3: Mapping with Tags 📋 PLANNED
**Status**: Not Started  
**Target**: Field mapping using reflection and struct tags

### Phase 4: Copy Implementation 📋 PLANNED
**Status**: Not Started  
**Target**: Generic Copy[T any]() function with C integration

### Phase 5: FastCopy 📋 PLANNED
**Status**: Not Started  
**Target**: Zero-allocation primitive copying

### Phase 6: Validation & Error Handling 📋 PLANNED
**Status**: Not Started  
**Target**: ValidateStruct helper and comprehensive error reporting

### Phase 7: C Macro Implementation 📋 PLANNED
**Status**: Not Started  
**Target**: Create `native2/cgocopy_metadata.h` with simplified macros

### Phase 8: Integration & Migration 📋 PLANNED
**Status**: Not Started  
**Target**: End-to-end testing, benchmarks, and migration guide

## Macro Syntax Comparison

### v1 (Current - `native/`)
```c
CGOCOPY_STRUCT_BEGIN(User)
FIELD_PRIMITIVE(User, int, id)
FIELD_STRING(User, name)
CGOCOPY_STRUCT_END(User)
```

### v2 (Planned - `native2/`)
```c
CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, id),
    CGOCOPY_FIELD(User, name)
)
```

## API Comparison

### v1 (Current)
```go
cgocopy.RegisterStruct[User]()
cgocopy.Finalize() // atomic.Bool flag
result := cgocopy.Copy[User](cUser)
```

### v2 (Planned)
```go
cgocopy2.Precompile[User]() // explicit, thread-safe
result := cgocopy2.Copy[User](cUser)

// Tagged structs work automatically
type UserDTO struct {
    UserID   int    `cgocopy:"id"`
    FullName string `cgocopy:"name"`
}
result := cgocopy2.Copy[UserDTO](cUser)

// Primitives get zero-allocation fast path
score := cgocopy2.FastCopy[int](cScore)

// Debugging helper
if err := cgocopy2.ValidateStruct[User](); err != nil {
    log.Fatal(err)
}
```

## Next Steps

1. **Immediate**: Start Phase 1 implementation
   - Create `pkg/cgocopy2/` directory structure
   - Define core types and errors
   - Set up basic test framework

2. **Short Term**: Complete Phases 2-4
   - Build registry with thread safety
   - Implement tag-aware field mapping
   - Create core Copy[T] function

3. **Medium Term**: Complete Phases 5-7
   - Add FastCopy optimization
   - Build validation helpers
   - Implement new C macros in `native2/`

4. **Long Term**: Phase 8
   - End-to-end integration tests
   - Performance benchmarking
   - Write migration guide
   - Update documentation

## Notes
- v1 (`pkg/cgocopy/`) remains stable and untouched
- v2 development happens in parallel in `pkg/cgocopy2/`
- Both versions can coexist in the same codebase
- Migration path will be clearly documented
