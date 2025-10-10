# CStringPtr: Lazy String Conversion for Maximum Performance

## The Problem

You want maximum copy performance (0.31ns DirectCopy), but your C struct has `char*` pointers:

```c
typedef struct {
    uint32_t id;
    char* name;      // Pointer to C string
    uint32_t channels;
} Device;
```

## Two Solutions

### Option 1: Eager Conversion with `string` (Current `devices` package)

```go
type Device struct {
    ID       uint32
    Name     string    // Go string (16 bytes)
    Channels uint32
}

// Must use Registry.Copy (handles string conversion)
func GetDevice() Device {
    cDevice := C.getDevice()
    defer C.freeDevice(cDevice)
    
    var device Device
    registry.Copy(&device, unsafe.Pointer(cDevice))  // ~50ns
    return device
}
```

**Pros:**
- ✅ Pure Go struct - safe to use anywhere
- ✅ Normal Go string type
- ✅ No manual memory management

**Cons:**
- ❌ Slower copy (~50ns)
- ❌ Allocates during copy
- ❌ Can't use DirectCopy

### Option 2: Lazy Pointer with `CStringPtr` (FASTEST!)

```go
type Device struct {
    ID       uint32
    Name     CStringPtr   // Just the pointer (8 bytes)
    Channels uint32
}

// Ultra-fast DirectCopy + lazy string conversion
func GetDevice() (Device, func()) {
    cDevice := C.getDevice()
    
    var device Device
    DirectCopy(&device, unsafe.Pointer(cDevice))  // 0.31ns
    
    cleanup := func() {
        C.freeDevice(cDevice)  // Free when done
    }
    
    return device, cleanup
}

// Usage:
device, cleanup := GetDevice()
defer cleanup()  // MUST call to free C memory

fmt.Println(device.Name.String())  // ~29ns - allocates only when called
```

**Pros:**
- ✅ **Ultra-fast copy (0.31ns)** - same as fixed buffer
- ✅ **Small Go struct** (8 bytes for pointer, not 256)
- ✅ **Works with `char*` pointers** - doesn't require `char[N]`
- ✅ **Lazy allocation** - only convert to string if/when needed
- ✅ Can use DirectCopy

**Cons:**
- ❌ **Must manage C memory lifetime** - caller responsible for cleanup
- ❌ **Cleanup function required** - must defer or remember to call
- ❌ **Not thread-safe without care** - C memory not GC-managed
- ❌ Allocates each time you call .String() (but cheaper: ~29ns vs ~50ns)

## Performance Comparison

| Method | Copy Time | String Access | Total (if accessed) | Allocations | Struct Size |
|--------|-----------|---------------|---------------------|-------------|-------------|
| `string` (Registry.Copy) | ~50ns | 0ns (already string) | ~50ns | 1 during copy | 40 bytes |
| `CStringPtr` (DirectCopy) | 0.31ns | ~29ns | ~29ns | 1 when calling .String() | 16 bytes |

**Winner:** `CStringPtr` is **42% faster** (~29ns vs 50ns) and **60% smaller** (16 vs 40 bytes)!

## When to Use Each

### Use `string` with Registry.Copy when:
- ✅ You want normal, idiomatic Go code
- ✅ String will be accessed multiple times
- ✅ Don't want to manage C memory lifetime
- ✅ 50ns overhead is acceptable
- ✅ Want safest, most idiomatic option

### Use `CStringPtr` with DirectCopy when:
- ✅ C has `char* name` (pointer - the common case)
- ✅ Want fastest possible performance (42% faster)
- ✅ String accessed rarely or conditionally
- ✅ Willing to manage C memory lifetime explicitly
- ✅ Want smallest struct size (60% smaller)

## Example: Conditional String Access

`CStringPtr` shines when strings are accessed conditionally:

```go
devices, cleanup := GetDevices()
defer cleanup()

// Filter by channels - never convert names!
outputDevices := make([]Device, 0)
for _, device := range devices {
    if device.Channels > 0 {
        outputDevices = append(outputDevices, device)
    }
}

// Only convert strings for matches
for _, device := range outputDevices {
    fmt.Println(device.Name.String())  // Lazy - only for output devices
}
```

**Savings:** If you filter 100 devices down to 3, you only pay for 3 string conversions (87ns) instead of 100 (5000ns)!

## Memory Management Pattern

```go
// Pattern 1: Single struct
device, cleanup := GetDevice()
defer cleanup()
// Use device...

// Pattern 2: Array of structs
devices, cleanup := GetDevices()
defer cleanup()
// Use devices...

// Pattern 3: Store for later (transfer ownership)
type DeviceCache struct {
    devices []Device
    cleanup func()
}

func (c *DeviceCache) Load() {
    c.devices, c.cleanup = GetDevices()
}

func (c *DeviceCache) Close() {
    if c.cleanup != nil {
        c.cleanup()
        c.cleanup = nil
    }
}
```

## Safety Note

**CRITICAL:** The C memory MUST remain valid for the entire lifetime of any struct containing `CStringPtr` fields. If you free the C memory too early:

```go
// ❌ DANGER - Use after free!
device, cleanup := GetDevice()
cleanup()  // Freed C memory
fmt.Println(device.Name.String())  // CRASH or garbage!

// ✅ CORRECT - Keep C memory alive
device, cleanup := GetDevice()
defer cleanup()  // Free at end of function
fmt.Println(device.Name.String())  // Safe
```

## Recommendation

- **Default:** Use `string` with Registry.Copy (safest, most idiomatic)
- **Performance critical:** Use `CStringPtr` with DirectCopy (42% faster, 60% smaller)

The `CStringPtr` approach gives you **maximum performance** (29ns total) while keeping structs **small** (8-byte pointers), but requires **explicit memory management**.
