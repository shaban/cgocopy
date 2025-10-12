package cgocopy

import (
	"testing"
	"unsafe"

	base "github.com/shaban/cgocopy/pkg/cgocopy"
)

type (
	FieldMapping     = base.FieldMapping
	CStringConverter = base.CStringConverter
	StructMapping    = base.StructMapping
	Registry         = base.Registry
	FieldKind        = base.FieldKind
	FieldInfo        = base.FieldInfo
	StructMetadata   = base.StructMetadata
	ArchInfo         = base.ArchInfo
	OffsetPredictor  = base.OffsetPredictor
	UTF8Converter    = base.UTF8Converter
	TestConverter    = base.TestConverter
)

const (
	FieldPrimitive FieldKind = base.FieldPrimitive
	FieldPointer   FieldKind = base.FieldPointer
	FieldString    FieldKind = base.FieldString
	FieldArray     FieldKind = base.FieldArray
	FieldStruct    FieldKind = base.FieldStruct
)

var (
	DefaultCStringConverter = base.DefaultCStringConverter
)

func NewRegistry() *Registry { return base.NewRegistry() }

func CustomLayout(structName string, typeNames ...string) {
	base.CustomLayout(structName, typeNames...)
}

func AutoLayout(typeNames ...string) []FieldInfo {
	return base.AutoLayout(typeNames...)
}

func Direct[T any](dst *T, src unsafe.Pointer) {
	base.Direct(dst, src)
}

func DirectArray[T any](dst []T, src unsafe.Pointer, cElemSize uintptr) {
	base.DirectArray(dst, src, cElemSize)
}

func GetArchInfo() ArchInfo { return base.GetArchInfo() }

func NewOffsetPredictor(info *ArchInfo) *OffsetPredictor {
	return base.NewOffsetPredictor(info)
}

func CreateTestStruct() unsafe.Pointer  { return base.CreateTestStruct() }
func FreeTestStruct(ptr unsafe.Pointer) { base.FreeTestStruct(ptr) }
func TestStructSize() uintptr           { return base.TestStructSize() }

func GetAutoDeviceSize() uintptr           { return base.GetAutoDeviceSize() }
func GetComplexDeviceSize() uintptr        { return base.GetComplexDeviceSize() }
func CreateAutoDevice() unsafe.Pointer     { return base.CreateAutoDevice() }
func FreeAutoDevice(ptr unsafe.Pointer)    { base.FreeAutoDevice(ptr) }
func CreateComplexDevice() unsafe.Pointer  { return base.CreateComplexDevice() }
func FreeComplexDevice(ptr unsafe.Pointer) { base.FreeComplexDevice(ptr) }

func CreateInnerStruct() unsafe.Pointer  { return base.CreateInnerStruct() }
func CreateMiddleStruct() unsafe.Pointer { return base.CreateMiddleStruct() }
func CreateOuterStruct() unsafe.Pointer  { return base.CreateOuterStruct() }
func FreePtr(ptr unsafe.Pointer)         { base.FreePtr(ptr) }
func InnerStructSize() uintptr           { return base.InnerStructSize() }
func MiddleStructSize() uintptr          { return base.MiddleStructSize() }
func OuterStructSize() uintptr           { return base.OuterStructSize() }
func CreateDeepNested() unsafe.Pointer   { return base.CreateDeepNested() }
func FreeDeepNested(ptr unsafe.Pointer)  { base.FreeDeepNested(ptr) }
func Level1Size() uintptr                { return base.Level1Size() }
func Level2Size() uintptr                { return base.Level2Size() }
func Level3Size() uintptr                { return base.Level3Size() }
func Level4Size() uintptr                { return base.Level4Size() }
func Level5Size() uintptr                { return base.Level5Size() }
func Level1Level2Offset() uintptr        { return base.Level1Level2Offset() }
func Level2Level3Offset() uintptr        { return base.Level2Level3Offset() }
func Level3Level4Offset() uintptr        { return base.Level3Level4Offset() }
func Level4Level5Offset() uintptr        { return base.Level4Level5Offset() }

func BenchmarkOffsetPrediction(b *testing.B) { base.BenchmarkOffsetPrediction(b) }
func TestOffsetPrediction(t *testing.T)      { base.TestOffsetPrediction(t) }
