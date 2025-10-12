# Usage Guide

How to choose between `Direct` and `Registry.Copy` after the StringPtr removal.

```
Do you have `char*` fields?
│
├─ YES → Use Registry.Copy with UTF8Converter (strings copied eagerly)
│
└─ NO  → Use Direct/DirectArray (primitives and fixed arrays only)
```

## Direct

Best for:

- Plain structs comprised of integers, floats, enums, or fixed-size arrays
- Hot paths where zero allocations and raw speed matter

Characteristics:

- `Direct` and `DirectArray` are compiler-inlined
- No safety checks; Go struct layout must exactly match the C struct
- Strings/pointers are copied verbatim – keep the C memory alive yourself

```go
type CSensor struct {
    uint32_t id;
    float    value;
};

type Sensor struct {
    ID    uint32
    Value float32
};

var s Sensor
cgocopy.Direct(&s, unsafe.Pointer(cSensor))
```

## Registry.Copy + UTF8Converter

Use when:

- The C struct contains `char*` fields that need to become Go `string`
- You want validation of offsets, sizes, and array metadata
- Nested structs or arrays are involved

How it works:

1. Describe the C layout with `FieldInfo` metadata (or call helper generators).
2. Register the Go struct once (typically in `init`).
3. Provide `DefaultCStringConverter` (or bespoke converter) so `char*` → `string` copies happen automatically.
4. Call `Copy` whenever you need a Go value; the C memory can be freed immediately afterwards.

```go
type CDevice struct {
    uint32_t id;
    char*    name;
    uint32_t channels;
};

type Device struct {
    ID       uint32
    Name     string
    Channels uint32
}

layout := []cgocopy.FieldInfo{
    {Offset: unsafe.Offsetof(CDevice{}.id), TypeName: "uint32_t"},
    {Offset: unsafe.Offsetof(CDevice{}.name), TypeName: "char*"},
    {Offset: unsafe.Offsetof(CDevice{}.channels), TypeName: "uint32_t"},
}

registry := cgocopy.NewRegistry()
registry.MustRegister(Device{}, unsafe.Sizeof(CDevice{}), layout, cgocopy.DefaultCStringConverter)
```

Performance profile:

| Aspect                | Direct            | Registry.Copy + UTF8Converter |
|-----------------------|-------------------|-------------------------------|
| String handling       | Manual            | Automatic                     |
| Memory ownership      | Caller-managed    | Go owns copies immediately    |
| Validation            | None              | Offset/size checked at register time |
| Typical per-copy cost | 0.31–0.36ns (primitives) | 140–170ns (strings + metadata setup) |

## Custom converters

`UTF8Converter` covers standard UTF-8 `char*`. If you need lazy conversion, custom allocators, or non-UTF8 encodings, implement the `CStringConverter` interface and plug it into `Register`/`MustRegister`.
