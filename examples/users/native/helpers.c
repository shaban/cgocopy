// Helper functions for creating sample user data
// These are NOT auto-generated - they're example-specific helpers

#include <stdlib.h>
#include <string.h>
#include "structs.h"

static const char* name_samples[2] = {
    "Ada Lovelace",
    "Alan Turing"
};

static const char* email_samples[2] = {
    "ada@example.com",
    "alan@example.net"
};

static const char* dept_samples[2] = {
    "Mathematics",
    "Computer Science"
};

User* createUsers(size_t* count) {
    size_t c = 2;
    User* users = (User*)calloc(c, sizeof(User));
    
    for (size_t i = 0; i < c; ++i) {
        users[i].id = (uint32_t)(1000 + i);
        users[i].email = strdup(email_samples[i]);
        users[i].details.full_name = strdup(name_samples[i]);
        users[i].details.level = (uint32_t)(i + 1);
        users[i].details.department = strdup(dept_samples[i]);
        users[i].account_balance = 1234.56 + (double)i * 42.0;
    }
    
    *count = c;
    return users;
}

void freeUsers(User* users, size_t count) {
    if (!users) {
        return;
    }
    
    for (size_t i = 0; i < count; ++i) {
        free(users[i].email);
        free(users[i].details.full_name);
        free(users[i].details.department);
    }
    
    free(users);
}
