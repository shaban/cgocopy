package users

/*
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "../../native/cgocopy_metadata.h"

typedef struct {
    char* name;
} UserRole;

typedef struct {
    char* full_name;
    UserRole roles[4];
    uint32_t level;
} UserDetails;

typedef struct {
    uint32_t id;
    char* email;
    UserDetails details;
    double account_balance;
} User;

static const char* role_samples[2][4] = {
    {"author", "client", "developer", "admin"},
    {"admin", "moderator", "support", "analyst"}
};

static const char* name_samples[2] = {
    "Ada Lovelace",
    "Alan Turing"
};

static const char* email_samples[2] = {
    "ada@example.com",
    "alan@example.net"
};

User* createUsers(size_t* count) {
    size_t c = 2;
    User* users = (User*)calloc(c, sizeof(User));
    for (size_t i = 0; i < c; ++i) {
        users[i].id = (uint32_t)(1000 + i);
        users[i].email = strdup(email_samples[i]);
        users[i].details.full_name = strdup(name_samples[i]);
        users[i].details.level = (uint32_t)(i + 1);
        for (size_t r = 0; r < 4; ++r) {
            users[i].details.roles[r].name = strdup(role_samples[i][r % 4]);
        }
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
        for (size_t r = 0; r < 4; ++r) {
            free(users[i].details.roles[r].name);
        }
    }
    free(users);
}

CGOCOPY_STRUCT_BEGIN(UserRole)
    CGOCOPY_FIELD_STRING(UserRole, name),
CGOCOPY_STRUCT_END(UserRole)

CGOCOPY_STRUCT_BEGIN(UserDetails)
    CGOCOPY_FIELD_STRING(UserDetails, full_name),
    CGOCOPY_FIELD_ARRAY_STRUCT(UserDetails, roles, UserRole, 4),
    CGOCOPY_FIELD_PRIMITIVE(UserDetails, level, uint32_t)
CGOCOPY_STRUCT_END(UserDetails)

CGOCOPY_STRUCT_BEGIN(User)
    CGOCOPY_FIELD_PRIMITIVE(User, id, uint32_t),
    CGOCOPY_FIELD_STRING(User, email),
    CGOCOPY_FIELD_STRUCT(User, details, UserDetails),
    CGOCOPY_FIELD_PRIMITIVE(User, account_balance, double)
CGOCOPY_STRUCT_END(User)
*/
import "C"

import (
	"unsafe"

	cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

func cUserRoleMetadata() cgocopy.StructMetadata {
	return cgocopy.StructMetadataFromC(unsafe.Pointer(C.cgocopy_get_UserRole_info()))
}

func cUserDetailsMetadata() cgocopy.StructMetadata {
	return cgocopy.StructMetadataFromC(unsafe.Pointer(C.cgocopy_get_UserDetails_info()))
}

func cUserMetadata() cgocopy.StructMetadata {
	return cgocopy.StructMetadataFromC(unsafe.Pointer(C.cgocopy_get_User_info()))
}

func newSampleUsers() (unsafe.Pointer, int) {
	var count C.size_t
	ptr := C.createUsers(&count)
	return unsafe.Pointer(ptr), int(count)
}

func freeSampleUsers(ptr unsafe.Pointer, count int) {
	C.freeUsers((*C.User)(ptr), C.size_t(count))
}
