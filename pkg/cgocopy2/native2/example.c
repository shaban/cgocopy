/*
 * Example C file demonstrating cgocopy2 macro usage
 * 
 * Compile with: gcc -std=c11 -c example.c
 */

#include "cgocopy_macros.h"
#include <stdio.h>

// ============================================================================
// Example 1: Simple struct with primitives
// ============================================================================

typedef struct {
    int id;
    double score;
    _Bool active;
} SimpleStruct;

CGOCOPY_STRUCT(SimpleStruct,
    CGOCOPY_FIELD(SimpleStruct, id),
    CGOCOPY_FIELD(SimpleStruct, score),
    CGOCOPY_FIELD(SimpleStruct, active)
)

// ============================================================================
// Example 2: Struct with string
// ============================================================================

typedef struct {
    int user_id;
    char* username;
    char* email;
} User;

CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, user_id),
    CGOCOPY_FIELD(User, username),
    CGOCOPY_FIELD(User, email)
)

// ============================================================================
// Example 3: Struct with arrays
// ============================================================================

typedef struct {
    int student_id;
    char* name;
    int grades[5];
    float scores[3];
} Student;

CGOCOPY_STRUCT(Student,
    CGOCOPY_FIELD(Student, student_id),
    CGOCOPY_FIELD(Student, name),
    CGOCOPY_ARRAY_FIELD(Student, grades, int),
    CGOCOPY_ARRAY_FIELD(Student, scores, float)
)

// ============================================================================
// Example 4: Nested struct
// ============================================================================

typedef struct {
    double x;
    double y;
    double z;
} Point3D;

CGOCOPY_STRUCT(Point3D,
    CGOCOPY_FIELD(Point3D, x),
    CGOCOPY_FIELD(Point3D, y),
    CGOCOPY_FIELD(Point3D, z)
)

typedef struct {
    char* name;
    Point3D position;
    Point3D velocity;
} GameObject;

CGOCOPY_STRUCT(GameObject,
    CGOCOPY_FIELD(GameObject, name),
    CGOCOPY_FIELD(GameObject, position),
    CGOCOPY_FIELD(GameObject, velocity)
)

// ============================================================================
// Example 5: Struct with all types
// ============================================================================

typedef struct {
    // Integers
    int8_t   i8;
    uint8_t  u8;
    int16_t  i16;
    uint16_t u16;
    int32_t  i32;
    uint32_t u32;
    int64_t  i64;
    uint64_t u64;
    
    // Floats
    float  f32;
    double f64;
    
    // Bool
    _Bool flag;
    
    // String
    char* text;
    
    // Array
    int numbers[10];
} ComprehensiveStruct;

CGOCOPY_STRUCT(ComprehensiveStruct,
    CGOCOPY_FIELD(ComprehensiveStruct, i8),
    CGOCOPY_FIELD(ComprehensiveStruct, u8),
    CGOCOPY_FIELD(ComprehensiveStruct, i16),
    CGOCOPY_FIELD(ComprehensiveStruct, u16),
    CGOCOPY_FIELD(ComprehensiveStruct, i32),
    CGOCOPY_FIELD(ComprehensiveStruct, u32),
    CGOCOPY_FIELD(ComprehensiveStruct, i64),
    CGOCOPY_FIELD(ComprehensiveStruct, u64),
    CGOCOPY_FIELD(ComprehensiveStruct, f32),
    CGOCOPY_FIELD(ComprehensiveStruct, f64),
    CGOCOPY_FIELD(ComprehensiveStruct, flag),
    CGOCOPY_FIELD(ComprehensiveStruct, text),
    CGOCOPY_ARRAY_FIELD(ComprehensiveStruct, numbers, int)
)

// ============================================================================
// Helper function to print metadata
// ============================================================================

void print_struct_metadata(cgocopy_struct_info* info) {
    printf("Struct: %s\n", info->name);
    printf("  Size: %zu bytes\n", info->size);
    printf("  Field count: %zu\n", info->field_count);
    printf("  Fields:\n");
    
    for (size_t i = 0; i < info->field_count; i++) {
        cgocopy_field_info* field = &info->fields[i];
        printf("    [%zu] %s: %s (offset=%zu, size=%zu",
               i, field->name, field->type, field->offset, field->size);
        
        if (field->is_pointer) {
            printf(", pointer");
        }
        if (field->is_array) {
            printf(", array[%zu]", field->array_len);
        }
        printf(")\n");
    }
    printf("\n");
}

// ============================================================================
// Main (for testing)
// ============================================================================

#ifdef STANDALONE_TEST
int main(void) {
    printf("=== cgocopy2 Macro Examples ===\n\n");
    
    print_struct_metadata(CGOCOPY_GET_METADATA(SimpleStruct));
    print_struct_metadata(CGOCOPY_GET_METADATA(User));
    print_struct_metadata(CGOCOPY_GET_METADATA(Student));
    print_struct_metadata(CGOCOPY_GET_METADATA(Point3D));
    print_struct_metadata(CGOCOPY_GET_METADATA(GameObject));
    print_struct_metadata(CGOCOPY_GET_METADATA(ComprehensiveStruct));
    
    return 0;
}
#endif
