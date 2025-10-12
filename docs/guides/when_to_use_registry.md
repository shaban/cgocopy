# When to Use the Registry

A pragmatic checklist for choosing between the zero-cost `Direct` helpers and the safer `Registry.Copy` pipeline.

## TL;DR

```
Does the struct contain char*?
│
├─ YES → Use Registry.Copy with UTF8Converter (or your converter)
│
└─ NO  → Use Direct/DirectArray for raw speed
```

## Direct (and DirectArray)

Pick Direct when all of the following hold:

- Every field is a primitive or fixed-length array.
- You control both the C and Go definitions, so layouts stay in sync.
- You can keep the C memory alive for the full lifetime of any copied pointers.

**Benefits**

- ~0.3ns per copy, zero allocations.
- Inlined code paths—no dispatch overhead.
- Ideal for hot loops and real-time workloads.

**Trade-offs**

- No validation. A layout mismatch silently corrupts memory.
- Strings/pointers are copied verbatim—you own the lifetime guarantees.

## Registry.Copy (+ UTF8Converter)

Reach for the registry when any of these apply:

- The struct (or a nested struct) contains `char*` pointers.
- You want Go `string` values with clear ownership semantics.
- The layout is defined by a third party and may change.
- You prefer a one-time validation step instead of a leap of faith.

**Workflow**

1. Describe the C layout with `FieldInfo` (or load it from the metadata helpers under `native/`).
2. Register the Go struct—ideally during `init` with `MustRegister`.
3. Pass `DefaultCStringConverter` (or your custom converter) so `char*` fields become Go strings.
4. Call `Copy` as needed; the C memory can be freed immediately afterwards.

**Performance notes**

- ~110ns for primitive-only structs, ~170ns when strings are present.
- Cost scales with the number/length of strings (the converter walks each byte once).
- Validation happens at registration, not during each copy.

## Custom converters

Implement the `CStringConverter` interface when you need:

- Non-UTF8 encodings.
- Arena/pooled allocations instead of `string`.
- Lazy evaluation (store the pointer and convert later).

The registry never imposes policy—your converter decides how C memory is handled.

## Real-world pattern

- Register nested metadata in dependency order (e.g., `SensorReading`, then `LargeSensorBlock`).
- Reuse a single `Registry` instance across the application.
- Mix approaches: keep using `Direct` for plain buffers, use the registry where strings or validation matter.

## Summary

| Scenario | Recommendation |
|----------|----------------|
| Hot loop, primitives only | `Direct` / `DirectArray` |
| Needs Go strings | `Registry.Copy` + `DefaultCStringConverter` |
| Third-party layout | `Registry.Copy` for validation |
| Custom encoding | `Registry.Copy` + custom converter |
