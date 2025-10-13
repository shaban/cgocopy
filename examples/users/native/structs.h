#ifndef USERS_STRUCTS_H
#define USERS_STRUCTS_H

#include <stdint.h>

// User role - simple string wrapper  
typedef struct {
    char* name;
} UserRole;

// User details - contains profile information
typedef struct {
    char* full_name;
    uint32_t level;
    char* department;
} UserDetails;

// User - main user structure with nested details
typedef struct {
    uint32_t id;
    char* email;
    UserDetails details;
    double account_balance;
} User;

#endif // USERS_STRUCTS_H
