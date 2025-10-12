package cgocopy

import (
	"reflect"
	"testing"
	"unsafe"
)

type GoMacroPrimitives struct {
	I8    int8
	U8    uint8
	I16   int16
	U16   uint16
	I32   int32
	U32   uint32
	I64   int64
	U64   uint64
	F32   float32
	F64   float64
	Flag  bool
	Data  unsafe.Pointer
	Name  string
	Label [12]byte
}

type GoMacroInner struct {
	Value int32
	Ratio float64
}

type GoMacroComposite struct {
	Base   GoMacroPrimitives
	Points [2]GoMacroInner
}

type GoMacroStrings struct {
	Normal    string
	Empty     string
	NullValue string
	Unicode   string
}

func TestMacroPrimitivesCopy(t *testing.T) {
	registry := NewRegistry()

	meta := macroPrimitivesMetadata()
	if meta.Name == "" {
		t.Fatal("expected metadata name for MacroPrimitives")
	}
	if len(meta.Fields) != 14 {
		t.Fatalf("expected 14 fields, got %d", len(meta.Fields))
	}
	if meta.Fields[len(meta.Fields)-1].Kind != FieldArray || meta.Fields[len(meta.Fields)-1].ElemCount != 12 {
		t.Fatalf("label field should be fixed array of 12 bytes: %+v", meta.Fields[len(meta.Fields)-1])
	}

	if err := registry.Register(reflect.TypeOf(GoMacroPrimitives{}), meta.Size, meta.Fields, TestConverter{}); err != nil {
		t.Fatalf("failed to register MacroPrimitives: %v", err)
	}

	cPtr := createMacroPrimitives()
	defer freeMacroPrimitives(cPtr)

	var out GoMacroPrimitives
	if err := registry.Copy(&out, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if out.I8 != -8 || out.U8 != 8 {
		t.Errorf("unexpected int8/uint8 values: %d/%d", out.I8, out.U8)
	}
	if out.I16 != -160 || out.U16 != 160 {
		t.Errorf("unexpected int16/uint16 values: %d/%d", out.I16, out.U16)
	}
	if out.I32 != -3200 || out.U32 != 3200 {
		t.Errorf("unexpected int32/uint32 values: %d/%d", out.I32, out.U32)
	}
	if out.I64 != -64000 || out.U64 != 64000 {
		t.Errorf("unexpected int64/uint64 values: %d/%d", out.I64, out.U64)
	}
	if out.F32 < 3.49 || out.F32 > 3.51 {
		t.Errorf("unexpected float32 value: %f", out.F32)
	}
	if out.F64 < 6.74 || out.F64 > 6.76 {
		t.Errorf("unexpected float64 value: %f", out.F64)
	}
	if !out.Flag {
		t.Error("expected flag to be true")
	}
	if out.Data == nil {
		t.Error("expected data pointer to be non-nil")
	}
	if out.Name != "Macro Primitives" {
		t.Errorf("unexpected string: %q", out.Name)
	}
	labelText := string(out.Label[:5])
	if labelText != "macro" {
		t.Errorf("unexpected label prefix: %q", labelText)
	}
}

func TestMacroCompositeCopy(t *testing.T) {
	registry := NewRegistry()

	innerMeta := macroInnerMetadata()
	if err := registry.Register(reflect.TypeOf(GoMacroInner{}), innerMeta.Size, innerMeta.Fields); err != nil {
		t.Fatalf("register inner failed: %v", err)
	}

	primMeta := macroPrimitivesMetadata()
	if err := registry.Register(reflect.TypeOf(GoMacroPrimitives{}), primMeta.Size, primMeta.Fields, TestConverter{}); err != nil {
		t.Fatalf("register primitives failed: %v", err)
	}

	compositeMeta := macroCompositeMetadata()
	if len(compositeMeta.Fields) != 2 {
		t.Fatalf("expected 2 fields in composite metadata, got %d", len(compositeMeta.Fields))
	}
	if compositeMeta.Fields[1].Kind != FieldArray || compositeMeta.Fields[1].ElemCount != 2 {
		t.Fatalf("expected points to be array of length 2: %+v", compositeMeta.Fields[1])
	}

	if err := registry.Register(reflect.TypeOf(GoMacroComposite{}), compositeMeta.Size, compositeMeta.Fields); err != nil {
		t.Fatalf("register composite failed: %v", err)
	}

	cPtr := createMacroComposite()
	defer freeMacroComposite(cPtr)

	var out GoMacroComposite
	if err := registry.Copy(&out, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if out.Base.Name != "Macro Primitives" {
		t.Errorf("unexpected base name: %q", out.Base.Name)
	}
	if out.Base.Data == nil {
		t.Error("expected base data pointer to be non-nil")
	}
	expectedPoints := []struct {
		value int32
		ratio float64
	}{{10, 1.5}, {20, 2.5}}

	for i, exp := range expectedPoints {
		if out.Points[i].Value != exp.value {
			t.Errorf("point %d value: got %d want %d", i, out.Points[i].Value, exp.value)
		}
		if out.Points[i].Ratio < exp.ratio-0.0001 || out.Points[i].Ratio > exp.ratio+0.0001 {
			t.Errorf("point %d ratio: got %f want %f", i, out.Points[i].Ratio, exp.ratio)
		}
	}

	// Ensure nested struct preserved primitive label data
	labelText := string(out.Base.Label[:5])
	if labelText != "macro" {
		t.Errorf("unexpected base label: %q", labelText)
	}
}

func TestMacroStringsCopy(t *testing.T) {
	registry := NewRegistry()

	meta := macroStringsMetadata()
	if meta.Name == "" {
		t.Fatal("expected metadata name for MacroStrings")
	}
	if len(meta.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(meta.Fields))
	}
	for i, field := range meta.Fields {
		if field.Kind != FieldString {
			t.Fatalf("field %d expected string kind, got %v", i, field.Kind)
		}
	}

	if err := registry.Register(reflect.TypeOf(GoMacroStrings{}), meta.Size, meta.Fields, TestConverter{}); err != nil {
		t.Fatalf("failed to register MacroStrings: %v", err)
	}

	cPtr := createMacroStrings()
	defer freeMacroStrings(cPtr)

	var out GoMacroStrings
	if err := registry.Copy(&out, cPtr); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	if out.Normal != "Macro String" {
		t.Errorf("unexpected normal string: %q", out.Normal)
	}
	if out.Empty != "" {
		t.Errorf("expected empty string to be empty, got %q", out.Empty)
	}
	if out.NullValue != "" {
		t.Errorf("expected null_value to map to empty string, got %q", out.NullValue)
	}
	if out.Unicode != "Macro UTF-8 🚀" {
		t.Errorf("unexpected unicode string: %q", out.Unicode)
	}
}
