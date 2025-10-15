package cgocopy2

/*
#include <stdlib.h>
#include <string.h>

// Simple person struct for benchmarking
typedef struct {
    int id;
    char name[64];
    double balance;
    int active;
} BenchPerson;

// Create test data
static BenchPerson* create_bench_person(int id, const char* name, double balance, int active) {
    BenchPerson* p = (BenchPerson*)malloc(sizeof(BenchPerson));
    p->id = id;
    strncpy(p->name, name, 63);
    p->name[63] = '\0';
    p->balance = balance;
    p->active = active;
    return p;
}

// Complex struct with nested data
typedef struct {
    double x;
    double y;
    double z;
} Point3D;

typedef struct {
    int id;
    char name[64];
    Point3D position;
    Point3D velocity;
    float health;
    int level;
} GameObject;

static GameObject* create_game_object(int id, const char* name, float health, int level) {
    GameObject* obj = (GameObject*)malloc(sizeof(GameObject));
    obj->id = id;
    strncpy(obj->name, name, 63);
    obj->name[63] = '\0';
    obj->position = (Point3D){100.5, 200.7, 300.9};
    obj->velocity = (Point3D){10.1, 20.2, 30.3};
    obj->health = health;
    obj->level = level;
    return obj;
}
*/
import "C"
import "unsafe"

// Go equivalents for benchmarking
type BenchPerson struct {
	ID      int32
	Name    [64]byte
	Balance float64
	Active  int32
}

type Point3DBench struct {
	X float64
	Y float64
	Z float64
}

type GameObjectBench struct {
	ID       int32
	Name     [64]byte
	Position Point3DBench
	Velocity Point3DBench
	Health   float32
	Level    int32
}

// Helper functions to create C structs (callable from tests)
func CreateBenchPerson(id int32, name string, balance float64, active int32) unsafe.Pointer {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return unsafe.Pointer(C.create_bench_person(C.int(id), cName, C.double(balance), C.int(active)))
}

func CreateGameObject(id int32, name string, health float32, level int32) unsafe.Pointer {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return unsafe.Pointer(C.create_game_object(C.int(id), cName, C.float(health), C.int(level)))
}

func FreeCPointer(ptr unsafe.Pointer) {
	C.free(ptr)
}

// Register types for cgocopy
func init() {
	PrecompileWithC[BenchPerson](CStructInfo{
		Name: "BenchPerson",
		Size: unsafe.Sizeof(BenchPerson{}),
		Fields: []CFieldInfo{
			{Name: "id", Type: "int32", Offset: 0, Size: 4},
			{Name: "name", Type: "int8", Offset: 4, Size: 1, IsArray: true, ArrayLen: 64},
			{Name: "balance", Type: "float64", Offset: 72, Size: 8},
			{Name: "active", Type: "int32", Offset: 80, Size: 4},
		},
	})

	PrecompileWithC[Point3DBench](CStructInfo{
		Name: "Point3D",
		Size: unsafe.Sizeof(Point3DBench{}),
		Fields: []CFieldInfo{
			{Name: "x", Type: "float64", Offset: 0, Size: 8},
			{Name: "y", Type: "float64", Offset: 8, Size: 8},
			{Name: "z", Type: "float64", Offset: 16, Size: 8},
		},
	})

	PrecompileWithC[GameObjectBench](CStructInfo{
		Name: "GameObject",
		Size: unsafe.Sizeof(GameObjectBench{}),
		Fields: []CFieldInfo{
			{Name: "id", Type: "int32", Offset: 0, Size: 4},
			{Name: "name", Type: "int8", Offset: 4, Size: 1, IsArray: true, ArrayLen: 64},
			{Name: "position", Type: "struct", Offset: 72, Size: 24},
			{Name: "velocity", Type: "struct", Offset: 96, Size: 24},
			{Name: "health", Type: "float32", Offset: 120, Size: 4},
			{Name: "level", Type: "int32", Offset: 124, Size: 4},
		},
	})
}
