package main

import (
	"strings"
	"testing"
)

func TestParseStructs_Simple(t *testing.T) {
	input := `
typedef struct {
    int id;
    double value;
} SimplePerson;
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structs))
	}

	s := structs[0]
	if s.Name != "SimplePerson" {
		t.Errorf("expected name 'SimplePerson', got '%s'", s.Name)
	}

	if len(s.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(s.Fields))
	}

	if s.Fields[0].Name != "id" || s.Fields[0].Type != "int" {
		t.Errorf("field 0: got %+v", s.Fields[0])
	}

	if s.Fields[1].Name != "value" || s.Fields[1].Type != "double" {
		t.Errorf("field 1: got %+v", s.Fields[1])
	}
}

func TestParseStructs_WithPointers(t *testing.T) {
	input := `
typedef struct {
    uint32_t user_id;
    char* username;
    char* email;
} User;
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structs))
	}

	s := structs[0]
	if len(s.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(s.Fields))
	}

	// Check pointer fields
	if s.Fields[1].Type != "char*" {
		t.Errorf("expected 'char*', got '%s'", s.Fields[1].Type)
	}
}

func TestParseStructs_WithArrays(t *testing.T) {
	input := `
struct Student {
    int id;
    char name[32];
    int grades[5];
};
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structs))
	}

	s := structs[0]
	if s.Name != "Student" {
		t.Errorf("expected name 'Student', got '%s'", s.Name)
	}

	if len(s.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(s.Fields))
	}

	// Check array fields
	nameField := s.Fields[1]
	if nameField.Name != "name" {
		t.Errorf("expected field name 'name', got '%s'", nameField.Name)
	}
	if nameField.ArraySize != "32" {
		t.Errorf("expected array size '32', got '%s'", nameField.ArraySize)
	}

	gradesField := s.Fields[2]
	if gradesField.ArraySize != "5" {
		t.Errorf("expected array size '5', got '%s'", gradesField.ArraySize)
	}
}

func TestParseStructs_NestedStructs(t *testing.T) {
	input := `
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
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 2 {
		t.Fatalf("expected 2 structs, got %d", len(structs))
	}

	// Check first struct
	if structs[0].Name != "Point3D" {
		t.Errorf("expected first struct 'Point3D', got '%s'", structs[0].Name)
	}

	// Check second struct with nested fields
	gameObject := structs[1]
	if gameObject.Name != "GameObject" {
		t.Errorf("expected second struct 'GameObject', got '%s'", gameObject.Name)
	}

	if len(gameObject.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(gameObject.Fields))
	}

	// Check nested struct fields
	if gameObject.Fields[1].Type != "Point3D" {
		t.Errorf("expected type 'Point3D', got '%s'", gameObject.Fields[1].Type)
	}
}

func TestParseStructs_MultipleStructs(t *testing.T) {
	input := `
typedef struct {
    int id;
} Simple;

typedef struct {
    float x;
    float y;
} Point;
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 2 {
		t.Fatalf("expected 2 structs, got %d", len(structs))
	}

	if structs[0].Name != "Simple" {
		t.Errorf("expected first struct 'Simple', got '%s'", structs[0].Name)
	}

	if structs[1].Name != "Point" {
		t.Errorf("expected second struct 'Point', got '%s'", structs[1].Name)
	}
}

func TestRemoveComments_LineComments(t *testing.T) {
	input := `
// This is a comment
struct Test { // inline comment
    int x;
};
`

	result := removeComments(input)

	if strings.Contains(result, "//") {
		t.Error("line comments should be removed")
	}

	if !strings.Contains(result, "struct Test") {
		t.Error("struct definition should remain")
	}
}

func TestRemoveComments_BlockComments(t *testing.T) {
	input := `
/* This is a block comment */
struct Test {
    int x; /* inline block */
};
/* Multi-line
   comment
   here */
`

	result := removeComments(input)

	if strings.Contains(result, "/*") || strings.Contains(result, "*/") {
		t.Error("block comments should be removed")
	}

	if !strings.Contains(result, "struct Test") {
		t.Error("struct definition should remain")
	}
}

func TestParseStructs_WithComments(t *testing.T) {
	input := `
// User struct for authentication
typedef struct {
    int id;        // User ID
    char* name;    // User name
    /* email address */
    char* email;
} User;
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structs))
	}

	s := structs[0]
	if len(s.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(s.Fields))
	}
}

func TestParseStructs_EmptyInput(t *testing.T) {
	input := ``

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 0 {
		t.Errorf("expected 0 structs, got %d", len(structs))
	}
}

func TestParseStructs_NoStructs(t *testing.T) {
	input := `
// Just some comments
int x = 5;
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 0 {
		t.Errorf("expected 0 structs, got %d", len(structs))
	}
}

func TestParseStructs_AllPrimitiveTypes(t *testing.T) {
	input := `
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
`

	structs, err := parseStructs(input)
	if err != nil {
		t.Fatalf("parseStructs failed: %v", err)
	}

	if len(structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(structs))
	}

	s := structs[0]
	if s.Name != "AllTypes" {
		t.Errorf("expected name 'AllTypes', got '%s'", s.Name)
	}

	if len(s.Fields) != 11 {
		t.Errorf("expected 11 fields, got %d", len(s.Fields))
	}
}
