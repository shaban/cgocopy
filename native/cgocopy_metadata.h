#ifndef CGOCOPY_METADATA_H
#define CGOCOPY_METADATA_H

#include <stddef.h>
#include <stdbool.h>
#include <string.h>

#ifndef _Alignof
#define _Alignof(type) __alignof__(type)
#endif

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    CGOCOPY_FIELD_PRIMITIVE_KIND = 0,
    CGOCOPY_FIELD_POINTER_KIND = 1,
    CGOCOPY_FIELD_STRING_KIND = 2,
    CGOCOPY_FIELD_ARRAY_KIND = 3,
    CGOCOPY_FIELD_STRUCT_KIND = 4,
} cgocopy_field_kind;

typedef struct {
    size_t offset;
    size_t size;
    const char *type_name;
    cgocopy_field_kind kind;
    const char *elem_type;
    size_t elem_count;
    bool is_string;
} cgocopy_field_info;

typedef struct {
    const char *name;
    size_t size;
    size_t alignment;
    size_t field_count;
    const cgocopy_field_info *fields;
} cgocopy_struct_info;

typedef struct cgocopy_struct_registry_node {
    const cgocopy_struct_info *info;
    struct cgocopy_struct_registry_node *next;
} cgocopy_struct_registry_node;

void cgocopy_registry_add(cgocopy_struct_registry_node *node);
const cgocopy_struct_info *cgocopy_lookup_struct_info(const char *name);

#define CGOCOPY_INTERNAL_STRINGIFY(x) CGOCOPY_INTERNAL_STRINGIFY_IMPL(x)
#define CGOCOPY_INTERNAL_STRINGIFY_IMPL(x) #x

#define CGOCOPY_INTERNAL_FIELD_INIT(kind_value, struct_type, field_name, type_literal, elem_literal, elem_count_value, string_flag) \
    {                                                                                                               \
        .offset = offsetof(struct_type, field_name),                                                                \
        .size = sizeof(((struct_type *)0)->field_name),                                                             \
        .type_name = (type_literal),                                                                                \
        .kind = (kind_value),                                                                                       \
        .elem_type = (elem_literal),                                                                                \
        .elem_count = (elem_count_value),                                                                           \
        .is_string = (string_flag)                                                                                  \
    }

#define CGOCOPY_STRUCT_BEGIN(struct_type)                                                                           \
    static const cgocopy_field_info cgocopy_fields_##struct_type[] = {

#define CGOCOPY_STRUCT_END(struct_type)                                                                             \
    };                                                                                                              \
    static const cgocopy_struct_info cgocopy_struct_info_##struct_type = {                                          \
        .name = CGOCOPY_INTERNAL_STRINGIFY(struct_type),                                                            \
        .size = sizeof(struct_type),                                                                                \
        .alignment = _Alignof(struct_type),                                                                         \
        .field_count = sizeof(cgocopy_fields_##struct_type) / sizeof(cgocopy_field_info),                           \
        .fields = cgocopy_fields_##struct_type,                                                                     \
    };                                                                                                              \
    static cgocopy_struct_registry_node cgocopy_registry_node_##struct_type;                                        \
    static void cgocopy_register_##struct_type(void) __attribute__((constructor));                                  \
    static void cgocopy_register_##struct_type(void) {                                                              \
        cgocopy_registry_node_##struct_type.info = &cgocopy_struct_info_##struct_type;                              \
        cgocopy_registry_node_##struct_type.next = NULL;                                                            \
        cgocopy_registry_add(&cgocopy_registry_node_##struct_type);                                                 \
    }                                                                                                               \
    static inline const cgocopy_struct_info *cgocopy_get_##struct_type##_info(void) {                               \
        return &cgocopy_struct_info_##struct_type;                                                                  \
    }

#define CGOCOPY_FIELD_PRIMITIVE(struct_type, field_name, c_type)                                                    \
    CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_PRIMITIVE_KIND, struct_type, field_name,                              \
                                CGOCOPY_INTERNAL_STRINGIFY(c_type), NULL, 0, false)

#define CGOCOPY_FIELD_POINTER(struct_type, field_name, c_type)                                                      \
    CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_POINTER_KIND, struct_type, field_name,                                \
                                CGOCOPY_INTERNAL_STRINGIFY(c_type), NULL, 0, false)

#define CGOCOPY_FIELD_STRING(struct_type, field_name)                                                               \
    CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_STRING_KIND, struct_type, field_name, "char*", NULL, 0, true)

#define CGOCOPY_FIELD_STRUCT(struct_type, field_name, nested_type)                                                  \
    CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_STRUCT_KIND, struct_type, field_name,                                 \
                                CGOCOPY_INTERNAL_STRINGIFY(nested_type), NULL, 0, false)

#define CGOCOPY_FIELD_ARRAY(struct_type, field_name, elem_type, elem_count)                                         \
    CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_ARRAY_KIND, struct_type, field_name,                                  \
                                CGOCOPY_INTERNAL_STRINGIFY(elem_type), CGOCOPY_INTERNAL_STRINGIFY(elem_type),       \
                                (elem_count), false)

#define CGOCOPY_FIELD_ARRAY_STRUCT(struct_type, field_name, nested_type, elem_count)                                \
    CGOCOPY_INTERNAL_FIELD_INIT(CGOCOPY_FIELD_ARRAY_KIND, struct_type, field_name,                                  \
                                CGOCOPY_INTERNAL_STRINGIFY(nested_type), CGOCOPY_INTERNAL_STRINGIFY(nested_type),   \
                                (elem_count), false)

#ifdef __cplusplus
}
#endif

#endif // CGOCOPY_METADATA_H
