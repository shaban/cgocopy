#ifndef CGOCOPY_MACROS_H
#define CGOCOPY_MACROS_H

/*
 * cgocopy2 - Simplified C Macros for Struct Metadata
 * 
 * Requires: C11 or later (for _Generic support)
 * 
 * Usage:
 *   CGOCOPY_STRUCT(MyStruct,
 *     CGOCOPY_FIELD(id),
 *     CGOCOPY_FIELD(name),
 *     CGOCOPY_FIELD(score)
 *   )
 * 
 * Features:
 *   - Automatic type detection via C11 _Generic
 *   - No manual type strings
 *   - Compile-time type safety
 *   - Works with nested structs, arrays, pointers
 */

#include <stddef.h>
#include <stdint.h>

// ============================================================================
// Type Detection via C11 _Generic
// ============================================================================

#define CGOCOPY_TYPE_NAME(x) _Generic((x), \
    _Bool: "bool", \
    char: "int8", \
    signed char: "int8", \
    unsigned char: "uint8", \
    short: "int16", \
    unsigned short: "uint16", \
    int: "int32", \
    unsigned int: "uint32", \
    long: "int64", \
    unsigned long: "uint64", \
    long long: "int64", \
    unsigned long long: "uint64", \
    float: "float32", \
    double: "float64", \
    char*: "string", \
    const char*: "string", \
    default: "struct" \
)

#define CGOCOPY_TYPE_SIZE(x) _Generic((x), \
    _Bool: sizeof(_Bool), \
    char: sizeof(char), \
    signed char: sizeof(signed char), \
    unsigned char: sizeof(unsigned char), \
    short: sizeof(short), \
    unsigned short: sizeof(unsigned short), \
    int: sizeof(int), \
    unsigned int: sizeof(unsigned int), \
    long: sizeof(long), \
    unsigned long: sizeof(unsigned long), \
    long long: sizeof(long long), \
    unsigned long long: sizeof(unsigned long long), \
    float: sizeof(float), \
    double: sizeof(double), \
    char*: sizeof(char*), \
    const char*: sizeof(const char*), \
    default: sizeof(x) \
)

// ============================================================================
// Field Metadata Structure
// ============================================================================

typedef struct {
    const char* name;        // Field name
    const char* type;        // Type name (auto-detected)
    size_t offset;           // Offset in struct (bytes)
    size_t size;             // Size of field (bytes)
    int is_pointer;          // 1 if pointer, 0 otherwise
    int is_array;            // 1 if array, 0 otherwise
    size_t array_len;        // Array length (0 if not array)
} cgocopy_field_info;

// ============================================================================
// Struct Metadata Structure
// ============================================================================

typedef struct {
    const char* name;              // Struct name
    size_t size;                   // Total struct size
    size_t field_count;            // Number of fields
    cgocopy_field_info* fields;    // Field metadata array
} cgocopy_struct_info;

// ============================================================================
// Helper Macros for Field Detection
// ============================================================================

// Check if a type is a pointer type via _Generic
#define CGOCOPY_IS_POINTER_TYPE(x) \
    _Generic((x), \
        char*: 1, \
        const char*: 1, \
        void*: 1, \
        default: 0 \
    )

// For arrays, we use a clever sizeof trick:
// - For arrays: sizeof(array) > sizeof(ptr) 
// - For scalars/pointers: sizeof(scalar/ptr) <= sizeof(ptr)
// But we need to be careful with strings which are char*

// Determine if a field is an array (non-pointer with size > pointer size for primitives)
// This is a heuristic: arrays will have sizeof > sizeof first element for multi-element arrays
#define CGOCOPY_IS_ARRAY_INTERNAL(type_sample) \
    (!CGOCOPY_IS_POINTER_TYPE(type_sample) && sizeof(type_sample) > sizeof(void*))

// Safer array length calculation with type check
#define CGOCOPY_SAFE_ARRAY_LEN(arr, elem_type) \
    (sizeof(arr) / sizeof(elem_type))

// ============================================================================
// Main Macro: CGOCOPY_FIELD
// ============================================================================

/*
 * Define a field's metadata using automatic type detection
 * 
 * Usage: CGOCOPY_FIELD(StructType, field_name)
 * 
 * This macro:
 * 1. Extracts the field name as a string
 * 2. Detects the field type via _Generic
 * 3. Calculates the offset using offsetof()
 * 4. Determines if it's a pointer or array
 * 5. For arrays, manually populate array_len (see CGOCOPY_ARRAY_FIELD)
 * 
 * Note: For non-array fields, array_len will be 0
 */
#define CGOCOPY_FIELD(structtype, field) \
    { \
        .name = #field, \
        .type = CGOCOPY_TYPE_NAME(((structtype){0}).field), \
        .offset = offsetof(structtype, field), \
        .size = sizeof(((structtype){0}).field), \
        .is_pointer = CGOCOPY_IS_POINTER_TYPE(((structtype){0}).field), \
        .is_array = 0, \
        .array_len = 0 \
    }

/*
 * Special macro for array fields - requires element type
 * 
 * Usage: CGOCOPY_ARRAY_FIELD(StructType, field_name, element_type)
 * Example: CGOCOPY_ARRAY_FIELD(Student, grades, int)
 */
#define CGOCOPY_ARRAY_FIELD(structtype, field, elemtype) \
    { \
        .name = #field, \
        .type = #elemtype "[]", \
        .offset = offsetof(structtype, field), \
        .size = sizeof(((structtype){0}).field), \
        .is_pointer = 0, \
        .is_array = 1, \
        .array_len = sizeof(((structtype){0}).field) / sizeof(elemtype) \
    }

// ============================================================================
// Main Macro: CGOCOPY_STRUCT
// ============================================================================

/*
 * Register a struct with its field metadata
 * 
 * Usage:
 *   CGOCOPY_STRUCT(MyStruct,
 *     CGOCOPY_FIELD(MyStruct, id),
 *     CGOCOPY_FIELD(MyStruct, name),
 *     CGOCOPY_FIELD(MyStruct, values)
 *   )
 * 
 * This generates a global cgocopy_struct_info variable named:
 *   cgocopy_metadata_<structname>
 */
#define CGOCOPY_STRUCT(structtype, ...) \
    static cgocopy_field_info cgocopy_fields_##structtype[] = { \
        __VA_ARGS__ \
    }; \
    static cgocopy_struct_info cgocopy_metadata_##structtype = { \
        .name = #structtype, \
        .size = sizeof(structtype), \
        .field_count = sizeof(cgocopy_fields_##structtype) / \
                       sizeof(cgocopy_field_info), \
        .fields = cgocopy_fields_##structtype \
    };

// ============================================================================
// Helper: Get Metadata for a Struct Type
// ============================================================================

/*
 * Get the metadata for a registered struct
 * 
 * Usage: cgocopy_struct_info* info = CGOCOPY_GET_METADATA(MyStruct);
 */
#define CGOCOPY_GET_METADATA(structtype) \
    (&cgocopy_metadata_##structtype)

// ============================================================================
// Example Usage (commented out - for reference)
// ============================================================================

#if 0
// Example C struct
typedef struct {
    int id;
    char* name;
    float scores[3];
    double average;
} Student;

// Register with cgocopy2 (generates metadata automatically)
CGOCOPY_STRUCT(Student,
    CGOCOPY_FIELD(Student, id),
    CGOCOPY_FIELD(Student, name),
    CGOCOPY_FIELD(Student, scores),
    CGOCOPY_FIELD(Student, average)
)

// Later, access the metadata
void print_student_metadata(void) {
    cgocopy_struct_info* info = CGOCOPY_GET_METADATA(Student);
    printf("Struct: %s (size: %zu bytes, %zu fields)\n",
           info->name, info->size, info->field_count);
    
    for (size_t i = 0; i < info->field_count; i++) {
        cgocopy_field_info* field = &info->fields[i];
        printf("  Field: %s (type: %s, offset: %zu, size: %zu",
               field->name, field->type, field->offset, field->size);
        if (field->is_pointer) printf(", pointer");
        if (field->is_array) printf(", array[%zu]", field->array_len);
        printf(")\n");
    }
}
#endif

#endif // CGOCOPY_MACROS_H
