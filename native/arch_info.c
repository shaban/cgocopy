#include <stddef.h>
#include <stdint.h>

// Test struct designed to FORCE padding and reveal alignment
// We intentionally put larger types after smaller ones to create gaps
typedef struct {
    int8_t    i8;      // 1 byte
    // padding here for i32?
    int32_t   i32;     // 4 bytes  
    // padding here for i64?
    int64_t   i64;     // 8 bytes
    int16_t   i16;     // 2 bytes (after i64 to force padding)
    // padding here for f64?
    double    f64;     // 8 bytes
    int8_t    i8_2;    // 1 byte (after double to force padding)
    // padding here for ptr?
    void*     ptr;     // 8 bytes (64-bit) or 4 bytes (32-bit)
    uint8_t   u8;      // 1 byte
    uint32_t  u32;     // 4 bytes
    uint16_t  u16;     // 2 bytes
    uint64_t  u64;     // 8 bytes
    float     f32;     // 4 bytes
    char*     charptr; // pointer size
    size_t    sizet;   // size_t size
} AlignmentTestStruct;

// Architecture information captured at compile-time
typedef struct {
    // Primitive sizes
    size_t int8_size;
    size_t int16_size;
    size_t int32_size;
    size_t int64_size;
    size_t uint8_size;
    size_t uint16_size;
    size_t uint32_size;
    size_t uint64_size;
    size_t float_size;
    size_t double_size;
    size_t pointer_size;
    size_t sizet_size;
    
    // Offsets in test struct (reveals natural alignment)
    size_t int8_offset;
    size_t int16_offset;
    size_t int32_offset;
    size_t int64_offset;
    size_t uint8_offset;
    size_t uint16_offset;
    size_t uint32_offset;
    size_t uint64_offset;
    size_t float_offset;
    size_t double_offset;
    size_t pointer_offset;
    size_t charptr_offset;
    size_t sizet_offset;
    
    // Total size of test struct (includes trailing padding)
    size_t test_struct_size;
    
    // Platform identifier
    int is_64bit;
    int is_little_endian;
} ArchitectureInfo;

// Function to get architecture info at runtime
ArchitectureInfo getArchitectureInfo() {
    ArchitectureInfo info;
    
    // Capture sizes
    info.int8_size = sizeof(int8_t);
    info.int16_size = sizeof(int16_t);
    info.int32_size = sizeof(int32_t);
    info.int64_size = sizeof(int64_t);
    info.uint8_size = sizeof(uint8_t);
    info.uint16_size = sizeof(uint16_t);
    info.uint32_size = sizeof(uint32_t);
    info.uint64_size = sizeof(uint64_t);
    info.float_size = sizeof(float);
    info.double_size = sizeof(double);
    info.pointer_size = sizeof(void*);
    info.sizet_size = sizeof(size_t);
    
    // Capture offsets (reveals alignment requirements)
    // NOTE: struct is designed to force padding
    info.int8_offset = offsetof(AlignmentTestStruct, i8);
    info.int32_offset = offsetof(AlignmentTestStruct, i32);
    info.int64_offset = offsetof(AlignmentTestStruct, i64);
    info.int16_offset = offsetof(AlignmentTestStruct, i16);
    info.double_offset = offsetof(AlignmentTestStruct, f64);
    info.uint8_offset = offsetof(AlignmentTestStruct, u8);
    info.pointer_offset = offsetof(AlignmentTestStruct, ptr);
    info.uint16_offset = offsetof(AlignmentTestStruct, u16);
    info.uint32_offset = offsetof(AlignmentTestStruct, u32);
    info.uint64_offset = offsetof(AlignmentTestStruct, u64);
    info.float_offset = offsetof(AlignmentTestStruct, f32);
    info.charptr_offset = offsetof(AlignmentTestStruct, charptr);
    info.sizet_offset = offsetof(AlignmentTestStruct, sizet);
    
    // Capture total size
    info.test_struct_size = sizeof(AlignmentTestStruct);
    
    // Platform detection
    info.is_64bit = (sizeof(void*) == 8) ? 1 : 0;
    
    // Endianness detection
    union {
        uint32_t i;
        uint8_t c[4];
    } test = {.i = 0x01020304};
    info.is_little_endian = (test.c[0] == 0x04) ? 1 : 0;
    
    return info;
}

// Helper to calculate alignment requirement from offset
size_t calculateAlignment(size_t prev_end, size_t current_offset) {
    if (current_offset == prev_end) {
        return 1; // No padding needed
    }
    size_t padding = current_offset - prev_end;
    // Alignment is the padding size (assuming single field)
    return current_offset - prev_end;
}
