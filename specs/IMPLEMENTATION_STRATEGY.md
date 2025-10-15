# Implementation Strategy: Go Code Generation Feature

**Date:** October 15, 2025  
**Decision:** In-place evolution vs separate package

---

## TL;DR Recommendation

**✅ EVOLVE IN-PLACE** - Extend existing `pkg/cgocopy` incrementally

**Why:**
1. Most changes are **tool-only** (cgocopy-generate)
2. Runtime changes are **minimal and non-breaking**
3. Can roll out incrementally without breaking v1.0.0
4. Users get automatic benefits without migration

---

## Analysis

### Current State (v1.0.0)

**Package:** `pkg/cgocopy` (currently named `cgocopy2` internally, but that's just an artifact)

**Core files:**
- `copy.go` - Main Copy/FastCopy/Direct APIs
- `registry.go` - Type registration system
- `validation.go` - Runtime validation
- `types.go` - Core types (CStructInfo, CFieldInfo, etc.)
- `errors.go` - Error types

**Tool:** `tools/cgocopy-generate/`
- Currently generates only C metadata
- **Needs extension** to generate Go code

### Proposed Changes Breakdown

#### 1. Tool Changes (cgocopy-generate) - 90% of work

**Location:** `tools/cgocopy-generate/`

**What changes:**
- ✅ Add Go code generation (NEW feature)
- ✅ Parse C structs (already exists!)
- ✅ Type mapping (NEW logic)
- ✅ Name conversion (NEW logic)
- ✅ Template system (extend existing)

**Impact on users:**
- ZERO breaking changes
- New optional `-go` flag
- Existing C-only generation still works

**Risk:** LOW - Tool is isolated, can iterate freely

---

#### 2. Runtime Changes (pkg/cgocopy) - 10% of work

**Existing code that works perfectly:**
```go
// These APIs stay exactly the same!
Copy[T](cPtr) (T, error)           // ✅ No changes needed
FastCopy[T](cPtr) T                 // ✅ No changes needed
Precompile[T]()                     // ✅ No changes needed
PrecompileWithC[T](info)            // ✅ No changes needed
```

**What needs to change:**

##### A. String Handling (ENHANCEMENT, not breaking)

**Current:** Users manually convert strings
```go
// Current approach (still works)
type User struct {
    Name [64]byte  // User manually converts to string
}
```

**Proposed:** Automatic string conversion
```go
// Generated code does this automatically
type User struct {
    Name string  // Auto-converted from char[64] or char*
}
```

**Implementation:** Add to `copy.go`
```go
// NEW internal helper (doesn't break existing API)
func convertCString(ptr unsafe.Pointer, maxLen int) string {
    length := 0
    for length < maxLen {
        if *(*byte)(unsafe.Add(ptr, length)) == 0 {
            break
        }
        length++
    }
    bytes := unsafe.Slice((*byte)(ptr), length)
    return string(bytes)
}
```

**Files to modify:**
- `copy.go` - Add string conversion helper (~20 lines)

**Risk:** LOW - New internal function, doesn't touch existing code paths

---

##### B. Dynamic Array Support (NEW feature, optional)

**Current:** Fixed arrays only
```go
type Student struct {
    Grades [5]int32  // Works today
}
```

**Proposed:** Dynamic arrays via {data, count} pattern
```go
type DeviceList struct {
    DeviceCount int32      // Count field
    Devices     []Device   // Slice created from count
}
```

**Implementation:** NEW feature path
```go
// NEW code path, doesn't affect existing usage
func copyWithDynamicArrays[T any](cPtr unsafe.Pointer, metadata ArrayInfo) (T, error) {
    // ... new logic for {data, count} pattern
}
```

**Files to modify:**
- `types.go` - Add ArrayInfo struct (~10 lines)
- `copy.go` - Add dynamic array helper (~30 lines)

**Risk:** LOW - Separate code path, opt-in via generated code

---

##### C. Enhanced Metadata (EXTENSION, not breaking)

**Current CFieldInfo:**
```go
type CFieldInfo struct {
    Name     string
    Type     string
    Offset   uintptr
    Size     uintptr
    IsArray  bool
    ArrayLen int
}
```

**Proposed enhancement:**
```go
type CFieldInfo struct {
    Name        string
    Type        string
    Offset      uintptr
    Size        uintptr
    IsArray     bool
    ArrayLen    int
    
    // NEW fields (backward compatible - zero values work)
    IsString    bool       // NEW: char* or char[] → string
    CountField  string     // NEW: For {data, count} pattern
    IsDynamic   bool       // NEW: Dynamic array flag
}
```

**Files to modify:**
- `types.go` - Add 3 optional fields (~3 lines)

**Risk:** ZERO - Adding fields with zero values = backward compatible

---

### Change Impact Summary

| Component | Changes | Breaking? | Risk | LOC |
|-----------|---------|-----------|------|-----|
| **cgocopy-generate tool** | Major (Go generation) | NO (new flag) | LOW | ~800 |
| **copy.go** | Minor (string helper) | NO | LOW | ~50 |
| **types.go** | Minimal (3 fields) | NO | ZERO | ~10 |
| **registry.go** | None | NO | ZERO | 0 |
| **validation.go** | None | NO | ZERO | 0 |
| **Existing tests** | None | NO | ZERO | 0 |
| **NEW tests** | Add Go codegen tests | N/A | LOW | ~200 |

**Total new code:** ~1060 lines  
**Modified existing code:** ~60 lines  
**Breaking changes:** ZERO

---

## Why In-Place Evolution Wins

### ✅ Advantages

#### 1. **Zero Breaking Changes**
```go
// v1.0.0 code continues to work EXACTLY as-is
user, err := cgocopy.Copy[User](cPtr)

// v1.1.0 adds optional features
// Old code: Still works!
// New code: Can use auto-generated structs + string conversion
```

#### 2. **Incremental Rollout**

**Week 1-2:** Tool extension only
- Add `-go` flag to cgocopy-generate
- Users can try new feature without touching runtime
- If bugs found, easy to roll back (just don't use `-go`)

**Week 3-4:** Runtime helpers
- Add string conversion
- Add dynamic array support
- Still backward compatible (unused helpers don't affect old code)

**Week 5:** Integration & testing
- Users can gradually adopt new features
- Can mix old and new code in same project

#### 3. **User Experience**

**Current v1.0.0 user:**
```bash
# Their existing code
go generate ./...  # Generates C metadata only
```

**After v1.1.0 upgrade:**
```bash
# Option A: Keep using old way (works!)
go generate ./...

# Option B: Opt into new features
cgocopy-generate -input bridge.h -go -package bridge
```

**No forced migration!**

#### 4. **Development Velocity**

- Can commit tool changes independently
- Can test in isolation
- Can iterate on Go generation without breaking runtime
- Can ship partial features (e.g., string support first, arrays later)

#### 5. **Maintenance**

- Single codebase to maintain
- Single test suite
- Single documentation
- Single release process
- Bug fixes benefit all users

---

### ❌ Why Separate Package Would Be Worse

**Hypothetical: `pkg/cgocopy_v2/`**

#### Problems:

1. **Code Duplication**
```go
// cgocopy/copy.go
func Copy[T any](cPtr unsafe.Pointer) (T, error) { ... }

// cgocopy_v2/copy.go
func Copy[T any](cPtr unsafe.Pointer) (T, error) { 
    // 95% same code + string conversion
}
```
**Maintenance nightmare!**

2. **Ecosystem Fragmentation**
```go
// User confusion
import "github.com/shaban/cgocopy"           // Which one?
import "github.com/shaban/cgocopy_v2"        // Do I need both?
```

3. **Migration Pain**
```go
// Users must rewrite ALL code
- cgocopy.Copy[User](cPtr)
+ cgocopy_v2.Copy[User](cPtr)  // Change 100s of imports
```

4. **Double Testing**
- Need to test both packages
- Bug fixes must be applied twice
- Performance regressions in both

5. **Documentation Split**
- Two READMEs
- Two sets of examples
- Users don't know which to use

---

## Implementation Plan: In-Place Evolution

### Phase 0: Preparation (Week 0 - Now)

✅ **Status: DONE**
- [x] Finalize spec (GO_CODE_GENERATION.md)
- [x] Performance validation (vs JSON benchmarks)
- [x] Overhead analysis (microbenchmarks)
- [x] Decision: In-place vs separate package

---

### Phase 1: Tool Extension - Go Struct Generation (Week 1)

**Goal:** Generate basic Go structs from C headers

**Location:** `tools/cgocopy-generate/`

**Tasks:**
1. Add CLI flags:
   ```bash
   -go               # Enable Go generation
   -go-output DIR    # Output directory for .go files
   -go-package NAME  # Package name
   ```

2. Extend `parser.go`:
   ```go
   // Already has C parsing - reuse it!
   type ParsedStruct struct {
       Name   string
       Fields []Field  // Already populated
   }
   ```

3. Create `go_generator.go`:
   ```go
   func GenerateGoStruct(parsed ParsedStruct) string {
       // Type mapping: int32_t → int32
       // Name conversion: user_id → UserID
       // Template rendering
   }
   ```

4. Create Go templates:
   ```go
   const goStructTemplate = `
   type {{.Name}} struct {
       {{range .Fields}}
       {{.GoName}} {{.GoType}} {{.Tags}}
       {{end}}
   }
   `
   ```

**Test criteria:**
- ✅ Parse simple struct (reuse existing parser)
- ✅ Generate valid Go code
- ✅ Code compiles
- ✅ Type mapping correct (int32_t → int32, etc.)

**Deliverable:**
```bash
$ cgocopy-generate -input bridge.h -go -go-package bridge
Generated: bridge_types.go (Go structs)
Generated: bridge_meta.c (C metadata)
```

**Files created:**
- `tools/cgocopy-generate/go_generator.go` (~200 LOC)
- `tools/cgocopy-generate/go_templates.go` (~100 LOC)
- `tools/cgocopy-generate/type_mapping.go` (~50 LOC)

**Files modified:**
- `tools/cgocopy-generate/main.go` (add flags, ~30 LOC)
- `tools/cgocopy-generate/generator.go` (orchestration, ~20 LOC)

**Risk:** LOW - Tool changes only, no runtime impact

---

### Phase 2: Init Function Generation (Week 1-2)

**Goal:** Auto-generate init() for type registration

**Location:** `tools/cgocopy-generate/`

**Tasks:**
1. Extend Go template:
   ```go
   func init() {
       cgocopy.PrecompileWithC[User](CStructInfo{
           Name: "User",
           Size: unsafe.Sizeof(User{}),
           Fields: []CFieldInfo{
               {Name: "user_id", Type: "int32", Offset: 0, Size: 4},
               // ...
           },
       })
   }
   ```

2. Generate metadata extraction:
   ```go
   func extractCMetadata() CStructInfo {
       cInfo := C.cgocopy_get_User_info()
       // Convert C metadata to CStructInfo
   }
   ```

**Test criteria:**
- ✅ init() registers types automatically
- ✅ Can call cgocopy.Copy[T]() immediately
- ✅ Validation works (size/field count)

**Deliverable:** Generated code includes full init()

**Files modified:**
- `tools/cgocopy-generate/go_templates.go` (~80 LOC added)

**Risk:** LOW - Still just code generation

---

### Phase 3: Runtime String Support (Week 2)

**Goal:** Add string conversion helpers to runtime

**Location:** `pkg/cgocopy/`

**Tasks:**
1. Add string conversion helper:
   ```go
   // NEW file: pkg/cgocopy/string_helpers.go
   func convertCString(ptr unsafe.Pointer, maxLen int) string {
       // Find null terminator
       length := 0
       for length < maxLen {
           if *(*byte)(unsafe.Add(ptr, length)) == 0 {
               break
           }
           length++
       }
       // Convert to Go string
       bytes := unsafe.Slice((*byte)(ptr), length)
       return string(bytes)
   }
   ```

2. Extend CFieldInfo:
   ```go
   // pkg/cgocopy/types.go
   type CFieldInfo struct {
       // ... existing fields ...
       IsString bool  // NEW: Indicates string field
   }
   ```

3. Update Copy to handle strings:
   ```go
   // pkg/cgocopy/copy.go
   if field.IsString {
       str := convertCString(fieldPtr, field.Size)
       // ... set field value ...
   }
   ```

**Test criteria:**
- ✅ String conversion works (char[] → string)
- ✅ String conversion works (char* → string)
- ✅ Null strings handled gracefully
- ✅ Unicode strings work

**Deliverable:** Runtime supports string fields

**Files created:**
- `pkg/cgocopy/string_helpers.go` (~30 LOC)

**Files modified:**
- `pkg/cgocopy/types.go` (~3 LOC)
- `pkg/cgocopy/copy.go` (~20 LOC)

**Risk:** LOW - New code path, doesn't affect existing users

---

### Phase 4: Arrays & Nested Structs (Week 2-3)

**Goal:** Generate code for arrays and nested structs

**Location:** `tools/cgocopy-generate/` (mostly tool changes)

**Tasks:**
1. Dependency sorting (topological):
   ```go
   func sortStructsByDependency(structs []Struct) []Struct {
       // Kahn's algorithm for topological sort
   }
   ```

2. Generate nested struct support:
   ```go
   type GameObject struct {
       Position Point3D  // Nested struct
       Velocity Point3D  // Must be defined first!
   }
   ```

3. Array generation:
   ```go
   type Student struct {
       Grades [5]int32  // Fixed array
   }
   ```

**Test criteria:**
- ✅ Arrays generate correctly
- ✅ Nested structs in correct order
- ✅ Circular dependencies detected
- ✅ Complex nesting works

**Deliverable:** Full support for composite types

**Files modified:**
- `tools/cgocopy-generate/go_generator.go` (~100 LOC)
- `tools/cgocopy-generate/dependency_sort.go` (~80 LOC, NEW)

**Risk:** LOW - Still tool-only changes

---

### Phase 5: Dynamic Arrays (Week 3 - OPTIONAL)

**Goal:** Support {data, count} pattern for dynamic arrays

**Location:** Both tool and runtime

**Tasks:**
1. Runtime support:
   ```go
   // pkg/cgocopy/array_helpers.go (NEW)
   func createDynamicSlice[T any](dataPtr unsafe.Pointer, count int32) []T {
       return unsafe.Slice((*T)(dataPtr), count)
   }
   ```

2. Tool generation:
   ```go
   // Detect pattern: int32_t count + T* data
   type DeviceList struct {
       DeviceCount int32    // Count field
       Devices     []Device // Auto-populated from count
   }
   ```

3. Extend CFieldInfo:
   ```go
   type CFieldInfo struct {
       // ... existing ...
       IsDynamic  bool   // NEW
       CountField string // NEW: Name of count field
   }
   ```

**Test criteria:**
- ✅ {data, count} pattern detected
- ✅ Slice created correctly
- ✅ Zero-copy (unsafe.Slice)
- ✅ Performance: 0.32ns overhead

**Deliverable:** Dynamic array support

**Files created:**
- `pkg/cgocopy/array_helpers.go` (~40 LOC)

**Files modified:**
- `pkg/cgocopy/types.go` (~2 LOC)
- `tools/cgocopy-generate/pattern_detection.go` (~60 LOC, NEW)

**Risk:** MEDIUM - New runtime feature, but isolated

**Priority:** Can defer to Phase 6+ if needed

---

### Phase 6: Edge Cases & Polish (Week 3-4)

**Goal:** Handle edge cases robustly

**Location:** Mostly tool

**Tasks:**
1. Go keyword detection:
   ```go
   func sanitizeGoName(name string) string {
       if isGoKeyword(name) {
           return name + "_"  // type → type_
       }
       return name
   }
   ```

2. Name collision detection:
   ```go
   // Warn if multiple C structs map to same Go name
   ```

3. Comment preservation:
   ```go
   // Copy C comments to Go code
   ```

4. Better error messages:
   ```go
   // "Circular dependency: A → B → A"
   // "Unsupported type: function pointer in field 'callback'"
   ```

**Test criteria:**
- ✅ Go keywords renamed
- ✅ Collisions detected
- ✅ Comments preserved
- ✅ Errors actionable

**Deliverable:** Robust, production-ready generator

**Files modified:**
- `tools/cgocopy-generate/` various files (~150 LOC)

**Risk:** LOW - Quality improvements

---

### Phase 7: Documentation & Examples (Week 4)

**Goal:** Comprehensive docs and examples

**Tasks:**
1. Update `pkg/cgocopy/README.md`:
   - Document Go generation workflow
   - Show before/after examples
   - Performance comparison table

2. Create examples:
   - `examples/auto_bridge/` - Simple auto-generated bridge
   - `examples/audio_engine/` - Real-world C library

3. Write guides:
   - Migration guide (v1.0 → v1.1)
   - Best practices
   - Troubleshooting

4. Update specs:
   - Mark GO_CODE_GENERATION.md as "Implemented"
   - Add "Lessons Learned" section

**Deliverable:** Complete documentation

**Files created:**
- `examples/auto_bridge/` (new example)
- `docs/guides/go_generation.md` (guide)

**Risk:** ZERO - Docs only

---

### Phase 8: Testing & Validation (Week 4)

**Goal:** Comprehensive test coverage

**Tasks:**
1. Unit tests:
   - Type mapping: all C types → Go types
   - Name conversion: all edge cases
   - Dependency sorting: complex graphs

2. Integration tests:
   - Generate code from real C headers
   - Compile and run generated code
   - Validate Copy works correctly

3. Performance tests:
   - Regression tests vs v1.0.0
   - String conversion overhead
   - Dynamic array overhead

4. Platform tests:
   - Linux, macOS, Windows
   - x86_64, ARM64

**Deliverable:** 90%+ coverage, all tests green

**Risk:** LOW - Testing only

---

### Phase 9: Release v1.1.0 (Week 4)

**Goal:** Public release of Go generation feature

**Tasks:**
1. Final code review
2. Update version: v1.0.0 → v1.1.0
3. Write release notes:
   - Highlight Go code generation
   - Performance benchmarks
   - Migration guide
4. Tag release: `git tag v1.1.0`
5. Update pkg.go.dev
6. Announce:
   - GitHub release
   - Reddit r/golang
   - Twitter/X

**Deliverable:** Public v1.1.0 release

---

## Risk Mitigation

### Risk 1: Breaking Existing Code

**Mitigation:**
- All new features are opt-in
- Existing API unchanged
- Backward compatibility tests
- Release as minor version (1.1.0, not 2.0.0)

### Risk 2: Performance Regression

**Mitigation:**
- Benchmark suite covers all code paths
- CI runs benchmarks on every commit
- String conversion is opt-in (users can keep []byte if needed)

### Risk 3: Tool Complexity

**Mitigation:**
- Start simple (primitives only)
- Add features incrementally
- Extensive testing at each phase
- Can abandon dynamic arrays if too complex (defer to v1.2)

### Risk 4: User Confusion

**Mitigation:**
- Clear documentation
- Migration guide
- Examples showing old vs new approach
- Deprecation warnings (not hard breaks)

---

## Timeline Summary

| Phase | Duration | Focus | Risk | Dependencies |
|-------|----------|-------|------|--------------|
| 0. Preparation | Done | Spec & benchmarks | ZERO | None |
| 1. Go structs | Week 1 | Tool: Basic generation | LOW | Phase 0 |
| 2. Init function | Week 1-2 | Tool: Registration | LOW | Phase 1 |
| 3. String support | Week 2 | Runtime: Strings | LOW | Phase 2 |
| 4. Arrays/nested | Week 2-3 | Tool: Composites | LOW | Phase 3 |
| 5. Dynamic arrays | Week 3 | Tool+Runtime: Optional | MED | Phase 4 |
| 6. Edge cases | Week 3-4 | Tool: Polish | LOW | Phase 4 |
| 7. Documentation | Week 4 | Docs & examples | ZERO | Phase 6 |
| 8. Testing | Week 4 | Full test suite | LOW | Phase 7 |
| 9. Release | Week 4 | v1.1.0 release | LOW | Phase 8 |

**Total duration:** 4 weeks  
**Critical path:** Phases 1-4 (core features)  
**Optional:** Phase 5 (can defer to v1.2)

---

## Success Metrics

### Functional
- ✅ Generate valid Go code from 100% of supported C structs
- ✅ Zero runtime errors in generated code
- ✅ 90%+ test coverage

### Performance
- ✅ No regression vs v1.0.0 for existing code
- ✅ String conversion: <35ns overhead
- ✅ Dynamic arrays: <1ns overhead

### User Experience
- ✅ One-command workflow: `go generate`
- ✅ Zero manual registration required
- ✅ Documentation covers 95% of use cases
- ✅ <5 GitHub issues in first month

---

## Final Recommendation

**✅ EVOLVE IN-PLACE (pkg/cgocopy) - NOT a separate package**

**Reasons:**
1. 90% of work is tool-only (isolated changes)
2. Runtime changes are minimal (<100 LOC) and non-breaking
3. Zero user migration required
4. Can ship incrementally (tool first, runtime later)
5. Single codebase = easier maintenance
6. Users get automatic benefits on upgrade

**Release strategy:**
- v1.0.0 (current) → v1.1.0 (Go generation)
- Backward compatible minor version bump
- Optional feature flag (-go)
- Can defer complex features (dynamic arrays) to v1.2

**Start with:** Phase 1 (tool extension) - lowest risk, highest value!
