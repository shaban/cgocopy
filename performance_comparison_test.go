package cgocopy

import (
	"reflect"
	"testing"
)

// CGO struct with StringPtr
type PluginInfoCGO struct {
	ID           uint32
	Name         StringPtr
	Manufacturer StringPtr
	Category     StringPtr
	Version      float32
}

// Pure Go struct
type PluginInfoGo struct {
	ID           uint32
	Name         string
	Manufacturer string
	Category     string
	Version      float32
}

// Benchmark: Direct + lazy StringPtr.String()
func BenchmarkDirectArrayPlusStrings(b *testing.B) {
	cPlugin := createPluginInfo()
	defer freePluginInfo(cPlugin)

	var cgoCopy PluginInfoCGO

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fast copy
		Direct(&cgoCopy, cPlugin)

		// Access all strings (simulating real usage)
		_ = cgoCopy.Name.String()
		_ = cgoCopy.Manufacturer.String()
		_ = cgoCopy.Category.String()
	}
}

// Benchmark: Registry.Copy() with automatic string conversion
func BenchmarkRegistryCopy(b *testing.B) {
	cPlugin := createPluginInfo()
	defer freePluginInfo(cPlugin)

	registry := NewRegistry()
	idOff, nameOff, mfgOff, catOff, verOff := pluginInfoOffsets()
	layout := []FieldInfo{
		{Offset: idOff, TypeName: "uint32_t"},
		{Offset: nameOff, TypeName: "char*"},
		{Offset: mfgOff, TypeName: "char*"},
		{Offset: catOff, TypeName: "char*"},
		{Offset: verOff, TypeName: "float"},
	}

	cSize := pluginInfoSize()
	err := registry.Register(reflect.TypeOf(PluginInfoGo{}), cSize, layout, PluginStringConverter{})
	if err != nil {
		b.Fatal(err)
	}

	var goCopy PluginInfoGo

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy with automatic string conversion
		if err := registry.Copy(&goCopy, cPlugin); err != nil {
			b.Fatal(err)
		}

		// Strings already converted - just access fields
		_ = goCopy.Name
		_ = goCopy.Manufacturer
		_ = goCopy.Category
	}
}

// Benchmark: Direct + convert all at once (fair comparison)
func BenchmarkDirectArrayPlusConvertAll(b *testing.B) {
	cPlugin := createPluginInfo()
	defer freePluginInfo(cPlugin)

	var cgoCopy PluginInfoCGO

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fast copy
		Direct(&cgoCopy, cPlugin)

		// Convert all strings immediately
		goCopy := PluginInfoGo{
			ID:           cgoCopy.ID,
			Name:         cgoCopy.Name.String(),
			Manufacturer: cgoCopy.Manufacturer.String(),
			Category:     cgoCopy.Category.String(),
			Version:      cgoCopy.Version,
		}
		_ = goCopy
	}
}

// Benchmark: Just the struct copy (no strings accessed)
func BenchmarkDirectNoStrings(b *testing.B) {
	cPlugin := createPluginInfo()
	defer freePluginInfo(cPlugin)

	var cgoCopy PluginInfoCGO

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Direct(&cgoCopy, cPlugin)
		// Don't access strings - just the struct copy
		_ = cgoCopy.ID
		_ = cgoCopy.Version
	}
}

// Test correctness
func TestPerformanceComparisonCorrectness(t *testing.T) {
	cPlugin := createPluginInfo()
	defer freePluginInfo(cPlugin)

	// Test Direct + StringPtr
	var cgoCopy PluginInfoCGO
	Direct(&cgoCopy, cPlugin)

	if cgoCopy.ID != 42 {
		t.Errorf("Direct: ID = %d, want 42", cgoCopy.ID)
	}
	if cgoCopy.Name.String() != "SuperDelay" {
		t.Errorf("Direct: Name = %s, want SuperDelay", cgoCopy.Name.String())
	}

	// Test Registry.Copy
	registry := NewRegistry()
	idOff, nameOff, mfgOff, catOff, verOff := pluginInfoOffsets()
	layout := []FieldInfo{
		{Offset: idOff, TypeName: "uint32_t"},
		{Offset: nameOff, TypeName: "char*"},
		{Offset: mfgOff, TypeName: "char*"},
		{Offset: catOff, TypeName: "char*"},
		{Offset: verOff, TypeName: "float"},
	}

	cSize := pluginInfoSize()
	err := registry.Register(reflect.TypeOf(PluginInfoGo{}), cSize, layout, PluginStringConverter{})
	if err != nil {
		t.Fatal(err)
	}

	var goCopy PluginInfoGo
	err = registry.Copy(&goCopy, cPlugin)
	if err != nil {
		t.Fatal(err)
	}

	if goCopy.ID != 42 {
		t.Errorf("Registry: ID = %d, want 42", goCopy.ID)
	}
	if goCopy.Name != "SuperDelay" {
		t.Errorf("Registry: Name = %s, want SuperDelay", goCopy.Name)
	}
	if goCopy.Manufacturer != "Waves" {
		t.Errorf("Registry: Manufacturer = %s, want Waves", goCopy.Manufacturer)
	}
}
