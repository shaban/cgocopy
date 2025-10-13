package integration

import (
	"testing"

	"github.com/shaban/cgocopy/pkg/cgocopy2"
)

// Test 1: Simple struct with primitives
func TestIntegration_SimplePerson(t *testing.T) {
	cPerson := CreateSimplePerson(42, 95.5, true)
	defer FreePointer(cPerson)

	person, err := cgocopy2.Copy[SimplePerson](cPerson)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if person.ID != 42 {
		t.Errorf("ID mismatch: got %d, want 42", person.ID)
	}
	if person.Score != 95.5 {
		t.Errorf("Score mismatch: got %f, want 95.5", person.Score)
	}
	if !person.Active {
		t.Errorf("Active mismatch: got %v, want true", person.Active)
	}
}

// Test 2: Struct with strings
func TestIntegration_User(t *testing.T) {
	cUser := CreateUser(123, "john_doe", "john@example.com")
	defer FreeUser(cUser)

	user, err := cgocopy2.Copy[User](cUser)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if user.UserID != 123 {
		t.Errorf("UserID mismatch: got %d, want 123", user.UserID)
	}
	if user.Username != "john_doe" {
		t.Errorf("Username mismatch: got %q, want %q", user.Username, "john_doe")
	}
	if user.Email != "john@example.com" {
		t.Errorf("Email mismatch: got %q, want %q", user.Email, "john@example.com")
	}
}

// Test 3: Struct with arrays
func TestIntegration_Student(t *testing.T) {
	cStudent := CreateStudent(456, "Alice")
	defer FreeStudent(cStudent)

	student, err := cgocopy2.Copy[Student](cStudent)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if student.StudentID != 456 {
		t.Errorf("StudentID mismatch: got %d, want 456", student.StudentID)
	}
	if student.Name != "Alice" {
		t.Errorf("Name mismatch: got %q, want %q", student.Name, "Alice")
	}

	expectedGrades := [5]int32{85, 90, 78, 92, 88}
	if student.Grades != expectedGrades {
		t.Errorf("Grades mismatch: got %v, want %v", student.Grades, expectedGrades)
	}

	expectedScores := [3]float32{85.5, 90.2, 88.7}
	for i, score := range student.Scores {
		if score < expectedScores[i]-0.01 || score > expectedScores[i]+0.01 {
			t.Errorf("Scores[%d] mismatch: got %f, want %f", i, score, expectedScores[i])
		}
	}
}

// Test 4: Nested structs
func TestIntegration_GameObject(t *testing.T) {
	cObj := CreateGameObject("Player", 100.0, 200.0, 300.0, 10.0, 20.0, 30.0)
	defer FreeGameObject(cObj)

	obj, err := cgocopy2.Copy[GameObject](cObj)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if obj.Name != "Player" {
		t.Errorf("Name mismatch: got %q, want %q", obj.Name, "Player")
	}
	if obj.Position.X != 100.0 || obj.Position.Y != 200.0 || obj.Position.Z != 300.0 {
		t.Errorf("Position mismatch: got (%f, %f, %f), want (100, 200, 300)",
			obj.Position.X, obj.Position.Y, obj.Position.Z)
	}
	if obj.Velocity.X != 10.0 || obj.Velocity.Y != 20.0 || obj.Velocity.Z != 30.0 {
		t.Errorf("Velocity mismatch: got (%f, %f, %f), want (10, 20, 30)",
			obj.Velocity.X, obj.Velocity.Y, obj.Velocity.Z)
	}
}

// Test 5: All primitive types
func TestIntegration_AllTypes(t *testing.T) {
	cTypes := CreateAllTypes()
	defer FreePointer(cTypes)

	types, err := cgocopy2.Copy[AllTypes](cTypes)
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if types.I8 != -42 {
		t.Errorf("I8 mismatch: got %d, want -42", types.I8)
	}
	if types.U8 != 200 {
		t.Errorf("U8 mismatch: got %d, want 200", types.U8)
	}
	if types.I16 != -1000 {
		t.Errorf("I16 mismatch: got %d, want -1000", types.I16)
	}
	if types.U16 != 50000 {
		t.Errorf("U16 mismatch: got %d, want 50000", types.U16)
	}
	if types.I32 != -100000 {
		t.Errorf("I32 mismatch: got %d, want -100000", types.I32)
	}
	if types.U32 != 3000000 {
		t.Errorf("U32 mismatch: got %d, want 3000000", types.U32)
	}
	if types.I64 != -9000000000 {
		t.Errorf("I64 mismatch: got %d, want -9000000000", types.I64)
	}
	if types.U64 != 18000000000 {
		t.Errorf("U64 mismatch: got %d, want 18000000000", types.U64)
	}
	if types.F32 < 3.14-0.01 || types.F32 > 3.15 {
		t.Errorf("F32 mismatch: got %f, want ~3.14159", types.F32)
	}
	if types.F64 < 2.71-0.01 || types.F64 > 2.72 {
		t.Errorf("F64 mismatch: got %f, want ~2.718281828", types.F64)
	}
	if !types.Flag {
		t.Errorf("Flag mismatch: got %v, want true", types.Flag)
	}
}

// Test 6: FastCopy with primitives
func TestIntegration_FastCopy_Int32(t *testing.T) {
	cInt := CreateInt32(12345)
	result := cgocopy2.FastCopy[int32](cInt)

	if result != 12345 {
		t.Errorf("FastCopy[int32] failed: got %d, want 12345", result)
	}
}

func TestIntegration_FastCopy_Float64(t *testing.T) {
	cDouble := CreateFloat64(3.14159)
	result := cgocopy2.FastCopy[float64](cDouble)

	if result < 3.14 || result > 3.15 {
		t.Errorf("FastCopy[float64] failed: got %f, want ~3.14159", result)
	}
}

// Test 7: Validation
func TestIntegration_Validation(t *testing.T) {
	if err := cgocopy2.ValidateStruct[SimplePerson](); err != nil {
		t.Errorf("SimplePerson validation failed: %v", err)
	}
	if err := cgocopy2.ValidateStruct[User](); err != nil {
		t.Errorf("User validation failed: %v", err)
	}
	if err := cgocopy2.ValidateStruct[Student](); err != nil {
		t.Errorf("Student validation failed: %v", err)
	}
	if err := cgocopy2.ValidateStruct[GameObject](); err != nil {
		t.Errorf("GameObject validation failed: %v", err)
	}
	if err := cgocopy2.ValidateStruct[AllTypes](); err != nil {
		t.Errorf("AllTypes validation failed: %v", err)
	}

	errors := cgocopy2.ValidateAll()
	if len(errors) > 0 {
		t.Errorf("ValidateAll found errors: %v", errors)
	}
}

// Test 8: GetRegisteredTypes
func TestIntegration_GetRegisteredTypes(t *testing.T) {
	types := cgocopy2.GetRegisteredTypes()

	expectedCount := 6
	if len(types) < expectedCount {
		t.Errorf("Expected at least %d registered types, got %d: %v", expectedCount, len(types), types)
	}

	hasUser := false
	hasGameObject := false
	for _, typeName := range types {
		if typeName == "User" {
			hasUser = true
		}
		if typeName == "GameObject" {
			hasGameObject = true
		}
	}

	if !hasUser {
		t.Errorf("User type not found in registered types: %v", types)
	}
	if !hasGameObject {
		t.Errorf("GameObject type not found in registered types: %v", types)
	}
}
