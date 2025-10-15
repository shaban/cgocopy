package cgocopy2

import (
	"encoding/json"
	"testing"
)

// JSON-friendly versions (for fair comparison)
type BenchPersonJSON struct {
	ID      int32   `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
	Active  bool    `json:"active"`
}

type Point3DJSON struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type GameObjectJSON struct {
	ID       int32       `json:"id"`
	Name     string      `json:"name"`
	Position Point3DJSON `json:"position"`
	Velocity Point3DJSON `json:"velocity"`
	Health   float32     `json:"health"`
	Level    int32       `json:"level"`
}

// ============================================================================
// Benchmark 1: Simple Struct (4 fields)
// ============================================================================

func BenchmarkCgocopy_SimplePerson(b *testing.B) {
	cPerson := CreateBenchPerson(42, "John Doe", 1234.56, 1)
	defer FreeCPointer(cPerson)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Copy[BenchPerson](cPerson)
	}
}

func BenchmarkJSON_SimplePerson(b *testing.B) {
	person := BenchPersonJSON{
		ID:      42,
		Name:    "John Doe",
		Balance: 1234.56,
		Active:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Serialize
		jsonBytes, _ := json.Marshal(person)
		// Deserialize
		var result BenchPersonJSON
		_ = json.Unmarshal(jsonBytes, &result)
	}
}

// ============================================================================
// Benchmark 2: Complex Struct (6 fields + nested structs)
// ============================================================================

func BenchmarkCgocopy_GameObject(b *testing.B) {
	cObj := CreateGameObject(123, "Player", 100.0, 50)
	defer FreeCPointer(cObj)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Copy[GameObjectBench](cObj)
	}
}

func BenchmarkJSON_GameObject(b *testing.B) {
	obj := GameObjectJSON{
		ID:   123,
		Name: "Player",
		Position: Point3DJSON{
			X: 100.5,
			Y: 200.7,
			Z: 300.9,
		},
		Velocity: Point3DJSON{
			X: 10.1,
			Y: 20.2,
			Z: 30.3,
		},
		Health: 100.0,
		Level:  50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Serialize
		jsonBytes, _ := json.Marshal(obj)
		// Deserialize
		var result GameObjectJSON
		_ = json.Unmarshal(jsonBytes, &result)
	}
}

// ============================================================================
// Performance Summary (Apple M1 Pro):
// ============================================================================
// SimplePerson (4 fields):
//   - cgocopy:  44.11ns (1 alloc)  <-- 20.5x faster than JSON
//   - JSON:    904.2ns  (7 allocs)
//
// GameObject (6 fields + nested):
//   - cgocopy:  49.23ns (1 alloc)  <-- 52.9x faster than JSON
//   - JSON:     2606ns  (9 allocs)
//
// Conclusion: cgocopy is 20-50x faster than JSON for real-world structs!
