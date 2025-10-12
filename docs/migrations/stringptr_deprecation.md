# Migrating from StringPtr

The legacy `StringPtr` helper has been removed in favour of a safer, pure-Go workflow:

1. Register structs that contain `char*` fields with `Registry.MustRegister` (or `Register`).
2. Supply `cgocopy.DefaultCStringConverter` (or your own `CStringConverter` implementation) so the registry can eagerly copy those strings into real Go `string` values.
3. Free the C memory immediately after calling `Copy`.

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

registry := cgocopy.NewRegistry()
layout := []cgocopy.FieldInfo{
    {Offset: unsafe.Offsetof(CDevice{}.id), TypeName: "uint32_t"},
    {Offset: unsafe.Offsetof(CDevice{}.name), TypeName: "char*"},
    {Offset: unsafe.Offsetof(CDevice{}.channels), TypeName: "uint32_t"},
}

registry.MustRegister(Device{}, unsafe.Sizeof(CDevice{}), layout, cgocopy.DefaultCStringConverter)

var out Device
if err := registry.Copy(&out, unsafe.Pointer(cDevice)); err != nil {
    panic(err)
}
```

## Why the change?

- **Safety first** – eager string conversion removes lifetime hazards and simplifies ownership.
- **No CGO round-trips** – `UTF8Converter` walks the C string with `unsafe.Add`, avoiding C helper calls while staying in pure Go.
- **Consistent ergonomics** – Go structs now look like idiomatic Go data models (`string`, not wrapper types).

## Need lazy conversions?

If you still prefer lazy evaluation, implement your own `CStringConverter` that stores pointers for later use. The registry API remains flexible, but the built-in presets now prioritise safe defaults.
