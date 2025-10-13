// Test helper functions for creating and freeing C structs
// These are NOT auto-generated - they're test-specific helpers

#include <stdlib.h>
#include <string.h>
#include "structs.h"

SimplePerson* create_simple_person(int id, double score, _Bool active) {
    SimplePerson* p = (SimplePerson*)malloc(sizeof(SimplePerson));
    p->id = id;
    p->score = score;
    p->active = active;
    return p;
}

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
