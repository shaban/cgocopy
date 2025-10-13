#ifndef INTEGRATION_STRUCTS_H
#define INTEGRATION_STRUCTS_H

#include <stdint.h>

// Test struct 1: Simple primitives
typedef struct {
    int id;
    double score;
    _Bool active;
} SimplePerson;

// Test struct 2: With strings
typedef struct {
    int user_id;
    char* username;
    char* email;
} User;

// Test struct 3: With arrays
typedef struct {
    int student_id;
    char* name;
    int grades[5];
    float scores[3];
} Student;

// Test struct 4: Nested structs
typedef struct {
    double x;
    double y;
    double z;
} Point3D;

typedef struct {
    char* name;
    Point3D position;
    Point3D velocity;
} GameObject;

// Test struct 5: All primitive types
typedef struct {
    int8_t   i8;
    uint8_t  u8;
    int16_t  i16;
    uint16_t u16;
    int32_t  i32;
    uint32_t u32;
    int64_t  i64;
    uint64_t u64;
    float    f32;
    double   f64;
    _Bool    flag;
} AllTypes;

// Constructor/destructor functions
SimplePerson* create_simple_person(int id, double score, _Bool active);
User* create_user(int id, const char* username, const char* email);
void free_user(User* u);
Student* create_student(int id, const char* name);
void free_student(Student* s);
GameObject* create_game_object(const char* name, double px, double py, double pz, double vx, double vy, double vz);
void free_game_object(GameObject* obj);
AllTypes* create_all_types(void);

#endif // INTEGRATION_STRUCTS_H
