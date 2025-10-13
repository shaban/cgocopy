package cgocopy

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// CStringConverter converts C strings (char*) to Go strings without relying on cgo helpers.
type CStringConverter interface {
	CStringToGo(ptr unsafe.Pointer) string
}

// FieldKind classifies how a field should be copied between C and Go structs.
type FieldKind uint8

const (
	FieldPrimitive FieldKind = iota
	FieldPointer
	FieldString
	FieldArray
	FieldStruct
)

// FieldInfo describes an individual C field discovered via metadata.
type FieldInfo struct {
	Offset    uintptr
	Size      uintptr
	TypeName  string
	Kind      FieldKind
	ElemType  string
	ElemCount uintptr
	IsString  bool
}

// FieldMapping represents a validated mapping between a C field and its Go counterpart.
type FieldMapping struct {
	COffset           uintptr
	GoOffset          uintptr
	Size              uintptr
	CType             string
	GoType            reflect.Type
	Kind              FieldKind
	IsNested          bool
	IsString          bool
	IsArray           bool
	IsSlice           bool
	ArrayLen          uintptr
	ArrayElemKind     FieldKind
	ArrayElemType     string
	ArrayElemSize     uintptr
	ArrayElemGoType   reflect.Type
	ArrayElemIsNested bool
	ArrayElemIsString bool
	NestedMapping     *StructMapping
	ArrayElemMapping  *StructMapping
}

// StructMapping stores all field mappings for a registered struct pair.
type StructMapping struct {
	CSize           uintptr
	GoSize          uintptr
	Fields          []FieldMapping
	CTypeName       string
	GoTypeName      string
	StringConverter CStringConverter
	CanFastPath     bool
}

// Registry maintains struct mappings and lifecycle information.
type Registry struct {
	mu        sync.Mutex
	mappings  map[reflect.Type]*StructMapping
	finalized atomic.Bool
}

var defaultRegistry = newRegistry()

// newRegistry constructs a fresh registry instance.
func newRegistry() *Registry {
	return &Registry{mappings: make(map[reflect.Type]*StructMapping)}
}

// Reset reinitializes the package-level registry. Intended for tests and examples.
func Reset() {
	defaultRegistry = newRegistry()
}

// Finalize marks the package-level registry immutable. Call this once registrations complete.
func Finalize() {
	defaultRegistry.Finalize()
}

// IsFinalized reports whether the package-level registry has been finalized.
func IsFinalized() bool {
	return defaultRegistry.IsFinalized()
}

// GetMapping returns the mapping registered for the provided struct type using the
// package-level registry.
func GetMapping[T any]() (*StructMapping, bool) {
	var zero T
	goType := reflect.TypeOf(zero)
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}
	if goType.Kind() != reflect.Struct {
		return nil, false
	}
	return defaultRegistry.GetMapping(goType)
}

// Register validates and stores a mapping between a Go struct and its C metadata.
func (r *Registry) Register(goType reflect.Type, cSize uintptr, cLayout []FieldInfo, converter ...CStringConverter) error {
	if r == nil {
		return ErrNilRegistry
	}
	if goType == nil {
		return fmt.Errorf("cgocopy: goType must not be nil")
	}
	if goType.Kind() != reflect.Struct {
		return fmt.Errorf("cgocopy: goType must be a struct, got %v", goType.Kind())
	}
	if r.finalized.Load() {
		return ErrRegistryFinalized
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.finalized.Load() {
		return ErrRegistryFinalized
	}

	var stringConverter CStringConverter
	if len(converter) > 0 {
		stringConverter = converter[0]
	}

	mapping, err := r.buildMapping(goType, cSize, cLayout, stringConverter)
	if err != nil {
		return err
	}

	r.mappings[goType] = mapping
	return nil
}

// buildMapping validates metadata and constructs a StructMapping.
func (r *Registry) buildMapping(goType reflect.Type, cSize uintptr, cLayout []FieldInfo, converter CStringConverter) (*StructMapping, error) {
	numGoFields := goType.NumField()
	if len(cLayout) != numGoFields {
		return nil, fmt.Errorf("cgocopy: field count mismatch: C has %d fields, Go has %d fields", len(cLayout), numGoFields)
	}

	mapping := &StructMapping{
		CSize:           cSize,
		GoSize:          goType.Size(),
		Fields:          make([]FieldMapping, 0, numGoFields),
		CTypeName:       fmt.Sprintf("C_%s", goType.Name()),
		GoTypeName:      goType.Name(),
		StringConverter: converter,
		CanFastPath:     true,
	}

	archInfo := GetArchInfo()
	hasStrings := false

	for i := 0; i < numGoFields; i++ {
		goField := goType.Field(i)
		cField := cLayout[i]

		fieldKind := resolveFieldKind(cField)
		isString := fieldKind == FieldString || cField.TypeName == "char*" || cField.IsString
		if isString {
			hasStrings = true
			if goField.Type.Kind() != reflect.String {
				return nil, fmt.Errorf("cgocopy: field %d (%s) expected Go string but found %v", i, goField.Name, goField.Type.Kind())
			}
			fieldKind = FieldString
		}

		if cField.Size == 0 {
			if fieldKind == FieldArray {
				// defer to element handling below
			} else {
				size := getSizeFromTypeName(cField.TypeName, archInfo)
				if size == 0 && fieldKind != FieldStruct {
					return nil, fmt.Errorf("cgocopy: field %d (%s) has Size=0 and unknown TypeName '%s'", i, goField.Name, cField.TypeName)
				}
				cField.Size = size
			}
		}

		goFieldKind := goField.Type.Kind()
		goFieldSize := goField.Type.Size()

		switch fieldKind {
		case FieldString:
			mapping.CanFastPath = false
		case FieldStruct:
			mapping.CanFastPath = false
			if goField.Type.Kind() != reflect.Struct {
				return nil, fmt.Errorf("cgocopy: field %d (%s) declared as struct but Go kind is %v", i, goField.Name, goField.Type.Kind())
			}
			nestedMapping, ok := r.mappings[goField.Type]
			if !ok {
				return nil, fmt.Errorf("cgocopy: field %d (%s) nested struct %v must be registered first", i, goField.Name, goField.Type)
			}
			// Cache nested mapping to avoid lookups during copy.
			fieldMapping := FieldMapping{
				COffset:       cField.Offset,
				GoOffset:      goField.Offset,
				Size:          goFieldSize,
				CType:         cField.TypeName,
				GoType:        goField.Type,
				Kind:          fieldKind,
				IsString:      isString,
				IsNested:      true,
				NestedMapping: nestedMapping,
			}

			mapping.Fields = append(mapping.Fields, fieldMapping)
			continue
		case FieldArray:
			mapping.CanFastPath = false
			if goFieldKind != reflect.Array && goFieldKind != reflect.Slice {
				return nil, fmt.Errorf("cgocopy: field %d (%s) expected array or slice, got %v", i, goField.Name, goFieldKind)
			}

			isSlice := goFieldKind == reflect.Slice

			if !isSlice {
				if uintptr(goField.Type.Len()) != cField.ElemCount {
					return nil, fmt.Errorf("cgocopy: field %d (%s) array length mismatch: C=%d Go=%d", i, goField.Name, cField.ElemCount, goField.Type.Len())
				}
			}

			elemTypeName := cField.ElemType
			if elemTypeName == "" {
				elemTypeName = cField.TypeName
			}

			elemGoType := goField.Type.Elem()
			arrayElemKind := FieldPrimitive
			arrayElemIsString := false

			goElemSize := elemGoType.Size()
			var cElemSize uintptr
			if cField.Size != 0 {
				if cField.ElemCount == 0 {
					return nil, fmt.Errorf("cgocopy: field %d (%s) array ElemCount is 0", i, goField.Name)
				}
				if cField.Size%cField.ElemCount != 0 {
					return nil, fmt.Errorf("cgocopy: field %d (%s) array size %d not divisible by element count %d", i, goField.Name, cField.Size, cField.ElemCount)
				}
				cElemSize = cField.Size / cField.ElemCount
			}

			if cElemSize == 0 {
				cElemSize = goElemSize
			}

			if cField.Size == 0 {
				cField.Size = cElemSize * cField.ElemCount
			}

			if !isSlice && cField.Size != goFieldSize {
				return nil, fmt.Errorf("cgocopy: field %d (%s) array size mismatch: C=%d bytes, Go=%d bytes", i, goField.Name, cField.Size, goFieldSize)
			}

			switch elemGoType.Kind() {
			case reflect.Struct:
				arrayElemKind = FieldStruct
				if _, ok := r.mappings[elemGoType]; !ok {
					return nil, fmt.Errorf("cgocopy: field %d (%s) array element struct %v must be registered first", i, goField.Name, elemGoType)
				}
			case reflect.String:
				arrayElemKind = FieldString
				arrayElemIsString = true
			case reflect.Ptr, reflect.UnsafePointer:
				arrayElemKind = FieldPointer
			default:
				arrayElemKind = FieldPrimitive
			}

			switch arrayElemKind {
			case FieldPrimitive, FieldPointer:
				if err := validateTypeCompatibility(elemGoType, elemTypeName); err != nil {
					return nil, fmt.Errorf("cgocopy: field %d (%s) array element type incompatible: %w", i, goField.Name, err)
				}
			case FieldString:
				return nil, fmt.Errorf("cgocopy: field %d (%s) array of Go strings is not supported", i, goField.Name)
			}

			fieldMapping := FieldMapping{
				COffset:           cField.Offset,
				GoOffset:          goField.Offset,
				Size:              goFieldSize,
				CType:             cField.TypeName,
				GoType:            goField.Type,
				Kind:              fieldKind,
				IsString:          isString,
				IsArray:           true,
				IsSlice:           isSlice,
				ArrayLen:          cField.ElemCount,
				ArrayElemKind:     arrayElemKind,
				ArrayElemType:     elemTypeName,
				ArrayElemSize:     cElemSize,
				ArrayElemGoType:   elemGoType,
				ArrayElemIsNested: arrayElemKind == FieldStruct,
				ArrayElemIsString: arrayElemIsString,
			}
			if fieldMapping.ArrayElemIsNested {
				fieldMapping.IsNested = true
				if elemMapping, ok := r.mappings[elemGoType]; ok {
					fieldMapping.ArrayElemMapping = elemMapping
				}
			}

			mapping.Fields = append(mapping.Fields, fieldMapping)
			continue

		case FieldPointer:
			if err := validateTypeCompatibility(goField.Type, cField.TypeName); err != nil {
				return nil, fmt.Errorf("cgocopy: field %d (%s) type incompatible: %w", i, goField.Name, err)
			}
			mapping.CanFastPath = false
		default:
			if cField.Size != goFieldSize {
				return nil, fmt.Errorf("cgocopy: field %d (%s) size mismatch: C=%d bytes, Go=%d bytes", i, goField.Name, cField.Size, goFieldSize)
			}
			if err := validateTypeCompatibility(goField.Type, cField.TypeName); err != nil {
				return nil, fmt.Errorf("cgocopy: field %d (%s) type incompatible: %w", i, goField.Name, err)
			}
		}

		fieldMapping := FieldMapping{
			COffset:  cField.Offset,
			GoOffset: goField.Offset,
			Size:     goFieldSize,
			CType:    cField.TypeName,
			GoType:   goField.Type,
			Kind:     fieldKind,
			IsString: isString,
		}

		mapping.Fields = append(mapping.Fields, fieldMapping)
	}

	if hasStrings && converter == nil {
		return nil, fmt.Errorf("cgocopy: struct has string fields but no CStringConverter provided")
	}

	return mapping, nil
}

// Finalize marks the registry immutable. Further registration attempts fail.
func (r *Registry) Finalize() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.finalized.Store(true)
	r.mu.Unlock()
}

// IsFinalized reports whether Finalize has been invoked.
func (r *Registry) IsFinalized() bool {
	if r == nil {
		return false
	}
	return r.finalized.Load()
}

// Copy copies data from C memory to the destination struct using registered mappings.
func (r *Registry) Copy(dst interface{}, cPtr unsafe.Pointer) error {
	if r == nil {
		return ErrNilRegistry
	}
	if !r.finalized.Load() {
		return ErrRegistryNotFinalized
	}
	if cPtr == nil {
		return ErrNilSourcePointer
	}

	dstVal := reflect.ValueOf(dst)
	if !dstVal.IsValid() || dstVal.IsNil() {
		return ErrNilDestination
	}
	if dstVal.Kind() != reflect.Ptr {
		return ErrDestinationNotStructPointer
	}
	elem := dstVal.Type().Elem()
	if elem.Kind() != reflect.Struct {
		return ErrDestinationNotStructPointer
	}

	mapping, ok := r.GetMapping(elem)
	if !ok {
		return ErrStructNotRegistered
	}

	return copyStructWithMapping(r, mapping, dstVal.Elem(), cPtr)
}

// CopyNoReflection mirrors Copy but routes through the no-reflection copy path.
func (r *Registry) CopyNoReflection(dst interface{}, cPtr unsafe.Pointer) error {
	if r == nil {
		return ErrNilRegistry
	}
	if !r.finalized.Load() {
		return ErrRegistryNotFinalized
	}
	if cPtr == nil {
		return ErrNilSourcePointer
	}

	dstVal := reflect.ValueOf(dst)
	if !dstVal.IsValid() || dstVal.IsNil() {
		return ErrNilDestination
	}
	if dstVal.Kind() != reflect.Ptr {
		return ErrDestinationNotStructPointer
	}

	elem := dstVal.Type().Elem()
	if elem.Kind() != reflect.Struct {
		return ErrDestinationNotStructPointer
	}

	dstPtr := unsafe.Pointer(dstVal.Pointer())
	return copyNoReflectionGeneric(r, dstPtr, elem, cPtr)
}

// GetMapping returns the struct mapping for the provided Go type if registered.
func (r *Registry) GetMapping(goType reflect.Type) (*StructMapping, bool) {
	if r == nil {
		return nil, false
	}
	mapping, ok := r.mappings[goType]
	return mapping, ok
}

// FastCopierFor returns a callable performing fast-path copies for the provided struct type.
func FastCopierFor[T any](r *Registry) (func(*T, unsafe.Pointer) error, error) {
	if r == nil {
		return nil, ErrNilRegistry
	}
	if !r.finalized.Load() {
		return nil, ErrRegistryNotFinalized
	}

	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return nil, ErrNotAStructType
	}

	mapping, ok := r.GetMapping(t)
	if !ok {
		return nil, ErrStructNotRegistered
	}
	if !mapping.CanFastPath {
		return nil, fmt.Errorf("cgocopy: fast path unavailable for %s", t.Name())
	}

	return func(dst *T, cPtr unsafe.Pointer) error {
		if dst == nil {
			return ErrNilDestination
		}
		if cPtr == nil {
			return ErrNilSourcePointer
		}
		*dst = *(*T)(cPtr)
		return nil
	}, nil
}

// FastCopier retrieves a fast-path copier from the default registry.
func FastCopier[T any]() (func(*T, unsafe.Pointer) error, error) {
	return FastCopierFor[T](defaultRegistry)
}

// MustRegister is a convenience helper mirroring Register but panicking on error.
func (r *Registry) MustRegister(goStruct any, cSize uintptr, cLayout []FieldInfo, converter ...CStringConverter) {
	if r == nil {
		panic(ErrNilRegistry)
	}
	goType := reflect.TypeOf(goStruct)
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}
	var conv CStringConverter
	if len(converter) > 0 {
		conv = converter[0]
	}
	if err := r.Register(goType, cSize, cLayout, conv); err != nil {
		panic(fmt.Sprintf("cgocopy: failed to register %v: %v", goType, err))
	}
}

// resolveFieldKind determines the field kind when metadata omits it.
func resolveFieldKind(info FieldInfo) FieldKind {
	if info.Kind != 0 {
		return info.Kind
	}
	if info.ElemCount > 0 {
		return FieldArray
	}
	if info.IsString || info.TypeName == "char*" {
		return FieldString
	}
	if info.TypeName == "struct" {
		return FieldStruct
	}
	if info.TypeName == "pointer" {
		return FieldPointer
	}
	if strings.HasSuffix(info.TypeName, "*") && info.TypeName != "char*" {
		return FieldPointer
	}
	return FieldPrimitive
}

// getSizeFromTypeName deduces the size of a C type using captured architecture information.
func getSizeFromTypeName(typeName string, archInfo ArchInfo) uintptr {
	switch typeName {
	case "int8_t", "uint8_t", "char":
		return archInfo.Int8Size
	case "int16_t", "uint16_t", "short":
		return archInfo.Int16Size
	case "int32_t", "uint32_t", "int":
		return archInfo.Int32Size
	case "int64_t", "uint64_t", "long", "long long":
		return archInfo.Int64Size
	case "float":
		return archInfo.Float32Size
	case "double":
		return archInfo.Float64Size
	case "char*", "void*", "pointer":
		return archInfo.PointerSize
	case "size_t":
		return archInfo.SizeTSize
	default:
		return 0
	}
}

// getAlignmentFromTypeName returns the expected alignment for a C type.
func getAlignmentFromTypeName(typeName string, archInfo ArchInfo) uintptr {
	switch typeName {
	case "int8_t", "uint8_t", "char":
		return archInfo.Int8Align
	case "int16_t", "uint16_t", "short":
		return archInfo.Int16Align
	case "int32_t", "uint32_t", "int":
		return archInfo.Int32Align
	case "int64_t", "uint64_t", "long", "long long":
		return archInfo.Int64Align
	case "float":
		return archInfo.Float32Align
	case "double":
		return archInfo.Float64Align
	case "char*", "void*", "pointer":
		return archInfo.PointerAlign
	case "size_t":
		return archInfo.SizeTAlign
	default:
		return 1
	}
}

// validateTypeCompatibility ensures the Go field kind can represent the C type.
func validateTypeCompatibility(goType reflect.Type, cTypeName string) error {
	compatibleTypes := map[string][]reflect.Kind{
		"int":       {reflect.Int, reflect.Int32},
		"int8_t":    {reflect.Int8},
		"int16_t":   {reflect.Int16},
		"int32_t":   {reflect.Int32},
		"int64_t":   {reflect.Int64},
		"uint8_t":   {reflect.Uint8},
		"uint16_t":  {reflect.Uint16},
		"uint32_t":  {reflect.Uint32},
		"uint64_t":  {reflect.Uint64},
		"char":      {reflect.Int8, reflect.Uint8},
		"short":     {reflect.Int16},
		"long":      {reflect.Int64},
		"long long": {reflect.Int64},
		"float":     {reflect.Float32},
		"double":    {reflect.Float64},
		"bool":      {reflect.Bool},
		"_Bool":     {reflect.Bool},
		"pointer":   {reflect.Ptr, reflect.UnsafePointer},
		"size_t":    {reflect.Uint64, reflect.Uintptr},
		"uintptr_t": {reflect.Uintptr},
	}

	goKind := goType.Kind()

	if goKind == reflect.Ptr || goKind == reflect.UnsafePointer {
		if cTypeName == "pointer" || (len(cTypeName) > 0 && cTypeName[len(cTypeName)-1] == '*') {
			return nil
		}
	}

	if goKind == reflect.Array {
		if strings.HasSuffix(cTypeName, "]") {
			return nil
		}
	}

	if validKinds, ok := compatibleTypes[cTypeName]; ok {
		for _, validKind := range validKinds {
			if goKind == validKind {
				return nil
			}
		}
	}

	return fmt.Errorf("incompatible types: C type '%s' cannot map to Go type '%v'", cTypeName, goType)
}
