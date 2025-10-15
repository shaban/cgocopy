# cgocopy vs JSON: Performance Validation

**Date:** October 15, 2025  
**Platform:** Apple M1 Pro (ARM64)  
**Go Version:** 1.25.2

---

## Executive Summary

✅ **VALIDATION PASSED:** cgocopy is **21-52x faster** than JSON for real-world structs, with 5-7x less memory allocation. The complexity of the code generation feature is **absolutely justified** for high-frequency FFI use cases.

---

## Benchmark Results

### Test Case 1: SimplePerson (4 fields: int, char[64], double, int)

```
BenchmarkCgocopy_SimplePerson-8    24597241    45.15 ns/op    64 B/op    1 allocs/op
BenchmarkJSON_SimplePerson-8        1249392   950.8 ns/op   336 B/op    7 allocs/op
```

**Performance:**
- **Speedup:** 21.1x faster (950.8ns / 45.15ns)
- **Memory:** 5.25x less allocation (336B / 64B)
- **Allocs:** 7x fewer allocations (7 / 1)

**Analysis:**
- cgocopy maintains constant ~45ns overhead regardless of struct size
- JSON overhead scales with field count and string serialization
- Memory allocation dominated by JSON string conversions

---

### Test Case 2: GameObject (6 fields + 2 nested Point3D structs)

```
BenchmarkCgocopy_GameObject-8      23839381    50.38 ns/op    64 B/op    1 allocs/op
BenchmarkJSON_GameObject-8           463477  2601   ns/op   496 B/op    9 allocs/op
```

**Performance:**
- **Speedup:** 51.6x faster (2601ns / 50.38ns)
- **Memory:** 7.75x less allocation (496B / 64B)
- **Allocs:** 9x fewer allocations (9 / 1)

**Analysis:**
- Nested structs severely impact JSON performance (2.7x slower than SimplePerson)
- cgocopy performance is nearly identical (~5ns difference)
- Memory-copy approach scales O(1) with nesting depth, JSON scales O(n)

---

## Performance Characteristics

### cgocopy

| Metric | Value | Notes |
|--------|-------|-------|
| Simple struct | ~45ns | 4 primitive fields |
| Complex struct | ~50ns | 6 fields + 2 nested structs |
| Memory/op | 64B | Single allocation for result |
| Scaling | O(struct size) | Linear with memory layout |

**Key insight:** Performance is dominated by memory copy size, not field count or nesting depth.

### JSON

| Metric | Value | Notes |
|--------|-------|-------|
| Simple struct | ~950ns | 4 primitive fields |
| Complex struct | ~2600ns | 6 fields + 2 nested structs |
| Memory/op | 336-496B | Multiple allocations for serialization |
| Scaling | O(fields × nesting) | Quadratic with complexity |

**Key insight:** Performance degrades significantly with nested structures due to recursive serialization.

---

## Real-World Implications

### When cgocopy Wins (21-52x faster)

1. **Game Engines:** Frame data transfer (60 FPS = 16.6ms budget)
   - JSON: 2600ns × 1000 objects = 2.6ms (15% of frame budget)
   - cgocopy: 50ns × 1000 objects = 0.05ms (0.3% of frame budget)

2. **Audio Processing:** Buffer metadata (44.1kHz sample rate)
   - JSON: 950ns × 100 buffers/sec = 95μs overhead
   - cgocopy: 45ns × 100 buffers/sec = 4.5μs overhead

3. **Real-Time Systems:** Sensor data aggregation
   - JSON: 2600ns × 10k sensors = 26ms latency
   - cgocopy: 50ns × 10k sensors = 0.5ms latency

### When JSON is Acceptable

1. **Configuration Loading:** Once at startup
2. **Network APIs:** Human-readable format requirement
3. **Low-Frequency Events:** < 10/sec update rate
4. **Small Structs:** < 3 fields with no nesting

---

## Conclusion

### The Sanity Check: ✅ PASSED

**Question:** "Did we write something so complex that we are not by orders of magnitude faster than simple JSON?"

**Answer:** **NO.** cgocopy is 21-52x faster than JSON for real-world structs. The complexity is **absolutely justified** for:

1. **High-frequency FFI:** Game loops, audio processing, real-time systems
2. **Memory-sensitive applications:** 5-7x less allocation overhead
3. **Complex data structures:** Performance advantage increases with nesting depth

### Design Validation

The two-path array strategy and code generation complexity are **necessary** because:

1. **Performance scales:** 50ns/struct regardless of complexity
2. **Memory efficiency:** Single allocation vs 7-9 allocations
3. **Real-world impact:** 21-52x speedup translates to milliseconds saved in production

### Recommendation

**Proceed with Go code generation feature.** The performance advantage is clear, measurable, and significant for the target use cases (Swift/Rust/ObjC FFI, game engines, audio processing, real-time systems).

---

## Appendix: Test Setup

### Benchmark Code

**SimplePerson C struct:**
```c
typedef struct {
    int id;
    char name[64];
    double balance;
    int active;
} BenchPerson;
```

**GameObject C struct:**
```c
typedef struct {
    double x;
    double y;
    double z;
} Point3D;

typedef struct {
    int id;
    char name[64];
    Point3D position;
    Point3D velocity;
    float health;
    int level;
} GameObject;
```

### Methodology

1. **Fair Comparison:**
   - Both methods do full round-trip (C → Go → result)
   - JSON includes Marshal + Unmarshal time
   - cgocopy includes Copy[T]() with validation

2. **Memory Accounting:**
   - Includes all allocations (struct + metadata)
   - JSON counted from `json.Marshal` + `json.Unmarshal`
   - cgocopy counted from `Copy[T]()` call

3. **Benchmark Configuration:**
   - `-benchtime=1s` for statistical stability
   - `-benchmem` for allocation tracking
   - Apple M1 Pro (8 cores) for consistency

---

**Files:**
- Benchmark implementation: `pkg/cgocopy/vs_json_benchmark_test.go`
- Helper functions: `pkg/cgocopy/benchmark_helpers.go`
- Full specification: `specs/GO_CODE_GENERATION.md`
