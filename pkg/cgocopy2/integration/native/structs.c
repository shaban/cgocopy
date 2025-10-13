#include <stdlib.h>
#include <string.h>
#include "../../native2/cgocopy_macros.h"
#include "structs.h"
#include "metadata_api.h"

// Register SimplePerson with cgocopy metadata
CGOCOPY_STRUCT(SimplePerson,
    CGOCOPY_FIELD(SimplePerson, id),
    CGOCOPY_FIELD(SimplePerson, score),
    CGOCOPY_FIELD(SimplePerson, active)
)

SimplePerson* create_simple_person(int id, double score, _Bool active) {
    SimplePerson* p = (SimplePerson*)malloc(sizeof(SimplePerson));
    p->id = id;
    p->score = score;
    p->active = active;
    return p;
}

// Register User with cgocopy metadata
CGOCOPY_STRUCT(User,
    CGOCOPY_FIELD(User, user_id),
    CGOCOPY_FIELD(User, username),
    CGOCOPY_FIELD(User, email)
)

User* create_user(int id, const char* username, const char* email) {
    User* u = (User*)malloc(sizeof(User));
    u->user_id = id;
    u->username = strdup(username);
    u->email = strdup(email);
    return u;
}

void free_user(User* u) {
    if (u) {
        free(u->username);
        free(u->email);
        free(u);
    }
}

// Register Student with cgocopy metadata
CGOCOPY_STRUCT(Student,
    CGOCOPY_FIELD(Student, student_id),
    CGOCOPY_FIELD(Student, name),
    CGOCOPY_ARRAY_FIELD(Student, grades, int),
    CGOCOPY_ARRAY_FIELD(Student, scores, float)
)

Student* create_student(int id, const char* name) {
    Student* s = (Student*)malloc(sizeof(Student));
    s->student_id = id;
    s->name = strdup(name);
    s->grades[0] = 85;
    s->grades[1] = 90;
    s->grades[2] = 78;
    s->grades[3] = 92;
    s->grades[4] = 88;
    s->scores[0] = 85.5f;
    s->scores[1] = 90.2f;
    s->scores[2] = 88.7f;
    return s;
}

void free_student(Student* s) {
    if (s) {
        free(s->name);
        free(s);
    }
}

// Register Point3D with cgocopy metadata
CGOCOPY_STRUCT(Point3D,
    CGOCOPY_FIELD(Point3D, x),
    CGOCOPY_FIELD(Point3D, y),
    CGOCOPY_FIELD(Point3D, z)
)

// Register GameObject with cgocopy metadata
CGOCOPY_STRUCT(GameObject,
    CGOCOPY_FIELD(GameObject, name),
    CGOCOPY_FIELD(GameObject, position),
    CGOCOPY_FIELD(GameObject, velocity)
)

GameObject* create_game_object(const char* name, double px, double py, double pz, double vx, double vy, double vz) {
    GameObject* obj = (GameObject*)malloc(sizeof(GameObject));
    obj->name = strdup(name);
    obj->position.x = px;
    obj->position.y = py;
    obj->position.z = pz;
    obj->velocity.x = vx;
    obj->velocity.y = vy;
    obj->velocity.z = vz;
    return obj;
}

void free_game_object(GameObject* obj) {
    if (obj) {
        free(obj->name);
        free(obj);
    }
}

// Register AllTypes with cgocopy metadata
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

AllTypes* create_all_types(void) {
    AllTypes* t = (AllTypes*)malloc(sizeof(AllTypes));
    t->i8 = -42;
    t->u8 = 200;
    t->i16 = -1000;
    t->u16 = 50000;
    t->i32 = -100000;
    t->u32 = 3000000;
    t->i64 = -9000000000LL;
    t->u64 = 18000000000ULL;
    t->f32 = 3.14159f;
    t->f64 = 2.718281828;
    t->flag = 1;
    return t;
}

// ============================================================================
// Metadata API Implementation
// ============================================================================
// These getter functions provide access to the static metadata structs.
// Static symbols are not visible to Go via CGO, so we need these non-static
// getter functions as a bridge.

const cgocopy_struct_info* get_SimplePerson_metadata(void) {
    return &cgocopy_metadata_SimplePerson;
}

const cgocopy_struct_info* get_User_metadata(void) {
    return &cgocopy_metadata_User;
}

const cgocopy_struct_info* get_Student_metadata(void) {
    return &cgocopy_metadata_Student;
}

const cgocopy_struct_info* get_Point3D_metadata(void) {
    return &cgocopy_metadata_Point3D;
}

const cgocopy_struct_info* get_GameObject_metadata(void) {
    return &cgocopy_metadata_GameObject;
}

const cgocopy_struct_info* get_AllTypes_metadata(void) {
    return &cgocopy_metadata_AllTypes;
}
