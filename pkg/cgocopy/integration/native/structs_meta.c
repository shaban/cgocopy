// GENERATED CODE - DO NOT EDIT
// Generated from: native/structs.h

#include <stdlib.h>
#include "../../native/cgocopy_macros.h"
#include "structs.h"


// Metadata for SimplePerson
CGOCOPY_STRUCT(SimplePerson,
    CGOCOPY_FIELD(SimplePerson, id),
    CGOCOPY_FIELD(SimplePerson, score),
    CGOCOPY_FIELD(SimplePerson, active)
)

const cgocopy_struct_info* get_SimplePerson_metadata(void) {
    return &cgocopy_metadata_SimplePerson;
}


// Metadata for User
CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, user_id),
    CGOCOPY_FIELD(User, username),
    CGOCOPY_FIELD(User, email)
)

const cgocopy_struct_info* get_User_metadata(void) {
    return &cgocopy_metadata_User;
}


// Metadata for Student
CGOCOPY_STRUCT(Student,
    CGOCOPY_FIELD(Student, student_id),
    CGOCOPY_FIELD(Student, name),
    CGOCOPY_FIELD(Student, grades),
    CGOCOPY_FIELD(Student, scores)
)

const cgocopy_struct_info* get_Student_metadata(void) {
    return &cgocopy_metadata_Student;
}


// Metadata for Point3D
CGOCOPY_STRUCT(Point3D,
    CGOCOPY_FIELD(Point3D, x),
    CGOCOPY_FIELD(Point3D, y),
    CGOCOPY_FIELD(Point3D, z)
)

const cgocopy_struct_info* get_Point3D_metadata(void) {
    return &cgocopy_metadata_Point3D;
}


// Metadata for GameObject
CGOCOPY_STRUCT(GameObject,
    CGOCOPY_FIELD(GameObject, name),
    CGOCOPY_FIELD(GameObject, position),
    CGOCOPY_FIELD(GameObject, velocity)
)

const cgocopy_struct_info* get_GameObject_metadata(void) {
    return &cgocopy_metadata_GameObject;
}


// Metadata for AllTypes
CGOCOPY_STRUCT(AllTypes,
    CGOCOPY_FIELD(AllTypes, i8),
    CGOCOPY_FIELD(AllTypes, u8),
    CGOCOPY_FIELD(AllTypes, i16),
    CGOCOPY_FIELD(AllTypes, u16),
    CGOCOPY_FIELD(AllTypes, i32),
    CGOCOPY_FIELD(AllTypes, u32),
    CGOCOPY_FIELD(AllTypes, i64),
    CGOCOPY_FIELD(AllTypes, u64),
    CGOCOPY_FIELD(AllTypes, f32),
    CGOCOPY_FIELD(AllTypes, f64),
    CGOCOPY_FIELD(AllTypes, flag)
)

const cgocopy_struct_info* get_AllTypes_metadata(void) {
    return &cgocopy_metadata_AllTypes;
}

