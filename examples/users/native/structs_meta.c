// GENERATED CODE - DO NOT EDIT
// Generated from: native/structs.h

#include <stdlib.h>
#include "../../../pkg/cgocopy/native/cgocopy_macros.h"
#include "structs.h"


// Metadata for UserRole
CGOCOPY_STRUCT(UserRole,
    CGOCOPY_FIELD(UserRole, name)
)

const cgocopy_struct_info* get_UserRole_metadata(void) {
    return &cgocopy_metadata_UserRole;
}


// Metadata for UserDetails
CGOCOPY_STRUCT(UserDetails,
    CGOCOPY_FIELD(UserDetails, full_name),
    CGOCOPY_FIELD(UserDetails, level),
    CGOCOPY_FIELD(UserDetails, department)
)

const cgocopy_struct_info* get_UserDetails_metadata(void) {
    return &cgocopy_metadata_UserDetails;
}


// Metadata for User
CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, id),
    CGOCOPY_FIELD(User, email),
    CGOCOPY_FIELD(User, details),
    CGOCOPY_FIELD(User, account_balance)
)

const cgocopy_struct_info* get_User_metadata(void) {
    return &cgocopy_metadata_User;
}

