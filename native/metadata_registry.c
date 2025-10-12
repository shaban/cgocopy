#include "cgocopy_metadata.h"

static cgocopy_struct_registry_node *cgocopy_registry_head = NULL;

void cgocopy_registry_add(cgocopy_struct_registry_node *node) {
    // TODO: add synchronization once metadata registration runs concurrently.
    if (node == NULL) {
        return;
    }
    node->next = cgocopy_registry_head;
    cgocopy_registry_head = node;
}

const cgocopy_struct_info *cgocopy_lookup_struct_info(const char *name) {
    if (name == NULL) {
        return NULL;
    }
    cgocopy_struct_registry_node *cursor = cgocopy_registry_head;
    while (cursor != NULL) {
        if (cursor->info != NULL && cursor->info->name != NULL && strcmp(cursor->info->name, name) == 0) {
            return cursor->info;
        }
        cursor = cursor->next;
    }
    return NULL;
}
