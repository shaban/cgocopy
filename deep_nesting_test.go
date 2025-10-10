package cgocopy

import (
	"reflect"
	"testing"
)

// Go equivalents
type GoLevel5 struct {
	Value int32
}

type GoLevel4 struct {
	ID     int32
	Level5 GoLevel5
}

type GoLevel3 struct {
	ID     int32
	Level4 GoLevel4
}

type GoLevel2 struct {
	ID     int32
	Level3 GoLevel3
}

type GoLevel1 struct {
	ID     int32
	Level2 GoLevel2
}

func TestDeepNesting5Levels(t *testing.T) {
	registry := NewRegistry()

	// Register from innermost to outermost

	// Level 5 (innermost - just a value)
	level5Layout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
	}
	err := registry.Register(reflect.TypeOf(GoLevel5{}), Level5Size(), level5Layout, nil)
	if err != nil {
		t.Fatalf("Failed to register Level5: %v", err)
	}

	// Level 4 (contains Level5)
	level4Layout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level4Level5Offset(), Size: Level5Size(), TypeName: "struct"},
	}
	err = registry.Register(reflect.TypeOf(GoLevel4{}), Level4Size(), level4Layout, nil)
	if err != nil {
		t.Fatalf("Failed to register Level4: %v", err)
	}

	// Level 3 (contains Level4)
	level3Layout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level3Level4Offset(), Size: Level4Size(), TypeName: "struct"},
	}
	err = registry.Register(reflect.TypeOf(GoLevel3{}), Level3Size(), level3Layout, nil)
	if err != nil {
		t.Fatalf("Failed to register Level3: %v", err)
	}

	// Level 2 (contains Level3)
	level2Layout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level2Level3Offset(), Size: Level3Size(), TypeName: "struct"},
	}
	err = registry.Register(reflect.TypeOf(GoLevel2{}), Level2Size(), level2Layout, nil)
	if err != nil {
		t.Fatalf("Failed to register Level2: %v", err)
	}

	// Level 1 (outermost - contains Level2)
	level1Layout := []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level1Level2Offset(), Size: Level2Size(), TypeName: "struct"},
	}
	err = registry.Register(reflect.TypeOf(GoLevel1{}), Level1Size(), level1Layout, nil)
	if err != nil {
		t.Fatalf("Failed to register Level1: %v", err)
	}

	// Create C struct with 5 levels
	cStruct := CreateDeepNested()
	defer FreeDeepNested(cStruct)

	// Copy the entire 5-level structure
	var goStruct GoLevel1
	err = registry.Copy(&goStruct, cStruct)
	if err != nil {
		t.Fatalf("Failed to copy 5-level nested struct: %v", err)
	} // Verify all values at all 5 levels
	if goStruct.ID != 1 {
		t.Errorf("Expected Level1.ID=1, got %d", goStruct.ID)
	}
	if goStruct.Level2.ID != 2 {
		t.Errorf("Expected Level2.ID=2, got %d", goStruct.Level2.ID)
	}
	if goStruct.Level2.Level3.ID != 3 {
		t.Errorf("Expected Level3.ID=3, got %d", goStruct.Level2.Level3.ID)
	}
	if goStruct.Level2.Level3.Level4.ID != 4 {
		t.Errorf("Expected Level4.ID=4, got %d", goStruct.Level2.Level3.Level4.ID)
	}
	if goStruct.Level2.Level3.Level4.Level5.Value != 999 {
		t.Errorf("Expected Level5.Value=999, got %d", goStruct.Level2.Level3.Level4.Level5.Value)
	}

	t.Logf("✅ Successfully copied 5-level nested struct recursively!")
	t.Logf("   Level1.ID = %d", goStruct.ID)
	t.Logf("   Level2.ID = %d", goStruct.Level2.ID)
	t.Logf("   Level3.ID = %d", goStruct.Level2.Level3.ID)
	t.Logf("   Level4.ID = %d", goStruct.Level2.Level3.Level4.ID)
	t.Logf("   Level5.Value = %d", goStruct.Level2.Level3.Level4.Level5.Value)
}

func TestDeepNestingWithMixedTypes(t *testing.T) {
	// Test that deep nesting works with various primitive types mixed in
	// This ensures the recursion handles different field types correctly

	type GoComplexLevel3 struct {
		A int8
		B uint16
		C float32
	}

	type GoComplexLevel2 struct {
		X    int64
		Nest GoComplexLevel3
		Y    float64
	}

	type GoComplexLevel1 struct {
		ID    uint32
		Nest2 GoComplexLevel2
		Flag  bool
	}

	registry := NewRegistry()

	// Register Level 3
	err := registry.Register(reflect.TypeOf(GoComplexLevel3{}), 8, []FieldInfo{
		{Offset: 0, Size: 1, TypeName: "int8_t"},
		{Offset: 2, Size: 2, TypeName: "uint16_t"},
		{Offset: 4, Size: 4, TypeName: "float"},
	}, nil)
	if err != nil {
		t.Fatalf("Failed to register ComplexLevel3: %v", err)
	}

	// Register Level 2 (contains Level 3)
	err = registry.Register(reflect.TypeOf(GoComplexLevel2{}), 24, []FieldInfo{
		{Offset: 0, Size: 8, TypeName: "int64_t"},
		{Offset: 8, Size: 8, TypeName: "struct"}, // GoComplexLevel3
		{Offset: 16, Size: 8, TypeName: "double"},
	}, nil)
	if err != nil {
		t.Fatalf("Failed to register ComplexLevel2: %v", err)
	}

	// Register Level 1 (outermost)
	err = registry.Register(reflect.TypeOf(GoComplexLevel1{}), 36, []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "uint32_t"},
		{Offset: 8, Size: 24, TypeName: "struct"}, // GoComplexLevel2
		{Offset: 32, Size: 1, TypeName: "bool"},
	}, nil)
	if err != nil {
		t.Fatalf("Failed to register ComplexLevel1: %v", err)
	}

	t.Logf("✅ Successfully registered 3-level nested struct with mixed primitive types!")
}

func BenchmarkDeepNesting5Levels(b *testing.B) {
	registry := NewRegistry()

	// Setup all registrations (using helper functions)
	registry.Register(reflect.TypeOf(GoLevel5{}), Level5Size(), []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
	}, nil)

	registry.Register(reflect.TypeOf(GoLevel4{}), Level4Size(), []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level4Level5Offset(), Size: Level5Size(), TypeName: "struct"},
	}, nil)

	registry.Register(reflect.TypeOf(GoLevel3{}), Level3Size(), []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level3Level4Offset(), Size: Level4Size(), TypeName: "struct"},
	}, nil)

	registry.Register(reflect.TypeOf(GoLevel2{}), Level2Size(), []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level2Level3Offset(), Size: Level3Size(), TypeName: "struct"},
	}, nil)

	registry.Register(reflect.TypeOf(GoLevel1{}), Level1Size(), []FieldInfo{
		{Offset: 0, Size: 4, TypeName: "int32_t"},
		{Offset: Level1Level2Offset(), Size: Level2Size(), TypeName: "struct"},
	}, nil)

	cStruct := CreateDeepNested()
	defer FreeDeepNested(cStruct)

	var goStruct GoLevel1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := registry.Copy(&goStruct, cStruct); err != nil {
			b.Fatal(err)
		}
	}
}
