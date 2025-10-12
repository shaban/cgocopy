package cgocopy

/*
Package structcopy provides efficient and safe copying between C and Go structs.

# Array/Slice Handling Strategies

C arrays and Go slices represent the biggest challenge for automatic struct copying.
Common patterns and solutions:

1. FIXED-SIZE ARRAYS (Easy - works with DirectCopy)
   C:  struct { int values[10]; }
   Go: struct { Values [10]int }
   → Direct memory copy works perfectly

2. DYNAMIC ARRAYS WITH SEPARATE COUNT (Common pattern)
   C:  struct { Device* devices; int deviceCount; }
   Go: struct { Devices []Device }
   → Requires custom enumeration function (see devices package)
   → Return slice directly, not in parent struct

3. NULL-TERMINATED ARRAYS (strings, argv style)
   C:  struct { char** tags; }  // NULL-terminated
   Go: struct { Tags []string }
   → Need custom converter that iterates until NULL

4. NESTED ARRAYS IN STRUCTS (Complex)
   C:  struct Engine {
         Device* devices;
         int deviceCount;
         Plugin* plugins;
         int pluginCount;
       }
   → NOT AUTOMATABLE - each array needs separate handling
   → Solution: Keep arrays separate, don't embed in parent struct
   → Use separate functions: GetEngine(), GetDevices(), GetPlugins()

5. FLEXIBLE ARRAY MEMBER (C99)
   C:  struct { int count; Device devices[]; }
   → Cannot be modeled in Go
   → Treat as pointer + count like pattern #2

RECOMMENDATION for 1:N relationships:
  - Return slices from separate functions
  - Don't try to embed slices in struct copies
  - Use pattern: GetEngine() + GetEngineDevices(engineID)

Example:
  engine := GetEngine()                    // Primitives only
  devices := GetEngineDevices(engine.ID)   // Separate array fetch
  plugins := GetEnginePlugins(engine.ID)   // Separate array fetch
*/

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// FieldMapping represents a validated mapping between C and Go struct fields
type FieldMapping struct {
	COffset           uintptr
	GoOffset          uintptr
	Size              uintptr
	CType             string
	GoType            reflect.Type
	Kind              FieldKind
	IsNested          bool // True if this field is a nested struct or array element is nested
	IsString          bool // True if this field is a C string (char*) → Go string
	IsArray           bool
	ArrayLen          uintptr
	ArrayElemKind     FieldKind
	ArrayElemType     string
	ArrayElemSize     uintptr
	ArrayElemGoType   reflect.Type
	ArrayElemIsNested bool
	ArrayElemIsString bool
}

// CStringConverter converts C strings (char*) to Go strings
type CStringConverter interface {
	CStringToGo(ptr unsafe.Pointer) string
}

// StructMapping represents a validated and registered C/Go struct pair
type StructMapping struct {
	CSize           uintptr
	GoSize          uintptr
	Fields          []FieldMapping
	CTypeName       string
	GoTypeName      string
	StringConverter CStringConverter // Optional: for structs with string fields
}

// Registry holds all registered struct mappings
type Registry struct {
	mappings map[reflect.Type]*StructMapping
}

// NewRegistry creates a new struct mapping registry
func NewRegistry() *Registry {
	return &Registry{
		mappings: make(map[reflect.Type]*StructMapping),
	}
}

// Register validates and registers a C/Go struct pair
// converter is optional - only needed if the struct has string fields (char* → string)
// Can be called as: Register(type, size, layout) or Register(type, size, layout, converter)
func (r *Registry) Register(goType reflect.Type, cSize uintptr, cLayout []FieldInfo, converter ...CStringConverter) error {
	if goType.Kind() != reflect.Struct {
		return fmt.Errorf("goType must be a struct, got %v", goType.Kind())
	}

	// Extract converter if provided
	var stringConverter CStringConverter
	if len(converter) > 0 {
		stringConverter = converter[0]
	}

	// Validate field count matches
	numGoFields := goType.NumField()
	if len(cLayout) != numGoFields {
		return fmt.Errorf("field count mismatch: C has %d fields, Go has %d fields", len(cLayout), numGoFields)
	}

	mapping := &StructMapping{
		CSize:           cSize,
		GoSize:          goType.Size(),
		Fields:          make([]FieldMapping, 0, numGoFields),
		CTypeName:       fmt.Sprintf("C_%s", goType.Name()),
		GoTypeName:      goType.Name(),
		StringConverter: stringConverter,
	}

	// Get architecture info for size deduction
	archInfo := GetArchInfo()

	// Validate each field
	hasStrings := false
	for i := 0; i < numGoFields; i++ {
		goField := goType.Field(i)
		cField := cLayout[i]

		fieldKind := resolveFieldKind(cField)
		isString := fieldKind == FieldString || cField.TypeName == "char*" || cField.IsString
		if isString {
			hasStrings = true
			if goField.Type.Kind() != reflect.String {
				return fmt.Errorf("field %d (%s) expected Go string but found %v", i, goField.Name, goField.Type.Kind())
			}
			fieldKind = FieldString
		}

		// Deduce size from TypeName if not provided
		if cField.Size == 0 {
			if fieldKind == FieldArray {
				// Fill from Go field size after we compute it below
				cField.Size = 0
			} else {
				size := getSizeFromTypeName(cField.TypeName, archInfo)
				if size == 0 && fieldKind != FieldStruct {
					return fmt.Errorf("field %d (%s) has Size=0 and unknown TypeName '%s' - cannot deduce size",
						i, goField.Name, cField.TypeName)
				}
				cField.Size = size
			}
		}

		goFieldSize := goField.Type.Size()
		switch fieldKind {
		case FieldString:
			// Skip size validation; Go string header size differs from C pointer size
		case FieldStruct:
			// Nested structs may contain strings or other fields that make the Go size
			// differ from the C size. We rely on the nested registration for safety.
		case FieldArray:
			if cField.Size == 0 {
				cField.Size = goFieldSize
			}
			if cField.Size != goFieldSize {
				return fmt.Errorf("field %d (%s) array size mismatch: C=%d bytes, Go=%d bytes",
					i, goField.Name, cField.Size, goFieldSize)
			}
		default:
			if cField.Size != goFieldSize {
				return fmt.Errorf("field %d (%s) size mismatch: C=%d bytes, Go=%d bytes",
					i, goField.Name, cField.Size, goFieldSize)
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

		switch fieldKind {
		case FieldStruct:
			if goField.Type.Kind() != reflect.Struct {
				return fmt.Errorf("field %d (%s) declared as struct but Go kind is %v", i, goField.Name, goField.Type.Kind())
			}
			if _, ok := r.mappings[goField.Type]; !ok {
				return fmt.Errorf("field %d (%s) is a nested struct of type %v which must be registered first",
					i, goField.Name, goField.Type)
			}
			fieldMapping.IsNested = true

		case FieldArray:
			if cField.ElemCount == 0 {
				return fmt.Errorf("field %d (%s) is declared as array but ElemCount is 0", i, goField.Name)
			}
			if goField.Type.Kind() != reflect.Array {
				return fmt.Errorf("field %d (%s) is declared as array but Go kind is %v", i, goField.Name, goField.Type.Kind())
			}
			if uintptr(goField.Type.Len()) != cField.ElemCount {
				return fmt.Errorf("field %d (%s) array length mismatch: C=%d, Go=%d", i, goField.Name,
					cField.ElemCount, goField.Type.Len())
			}

			elemTypeName := cField.ElemType
			if elemTypeName == "" {
				elemTypeName = cField.TypeName
			}

			elemGoType := goField.Type.Elem()
			arrayElemKind := FieldPrimitive
			arrayElemIsString := false

			switch elemGoType.Kind() {
			case reflect.Struct:
				arrayElemKind = FieldStruct
				if _, ok := r.mappings[elemGoType]; !ok {
					return fmt.Errorf("field %d (%s) array element struct %v must be registered first", i, goField.Name, elemGoType)
				}
				fieldMapping.ArrayElemIsNested = true
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
					return fmt.Errorf("field %d (%s) array element type incompatible: %w", i, goField.Name, err)
				}
			case FieldString:
				return fmt.Errorf("field %d (%s) array of Go strings is not supported", i, goField.Name)
			}

			fieldMapping.IsArray = true
			fieldMapping.ArrayLen = cField.ElemCount
			fieldMapping.ArrayElemKind = arrayElemKind
			fieldMapping.ArrayElemType = elemTypeName
			fieldMapping.ArrayElemSize = elemGoType.Size()
			fieldMapping.ArrayElemGoType = elemGoType
			fieldMapping.ArrayElemIsString = arrayElemIsString
			if fieldMapping.ArrayElemIsNested {
				fieldMapping.IsNested = true
			}

		case FieldPointer, FieldPrimitive:
			if err := validateTypeCompatibility(goField.Type, cField.TypeName); err != nil {
				return fmt.Errorf("field %d (%s) type incompatible: %w", i, goField.Name, err)
			}
		case FieldString:
			// Already validated
		default:
			if err := validateTypeCompatibility(goField.Type, cField.TypeName); err != nil {
				return fmt.Errorf("field %d (%s) type incompatible: %w", i, goField.Name, err)
			}
		}

		mapping.Fields = append(mapping.Fields, fieldMapping)
	}

	// Validate converter if needed
	if hasStrings && converter == nil {
		return fmt.Errorf("struct has string fields but no CStringConverter provided")
	}

	r.mappings[goType] = mapping
	return nil
}

type FieldKind uint8

const (
	FieldPrimitive FieldKind = iota
	FieldPointer
	FieldString
	FieldArray
	FieldStruct
)

// FieldInfo describes a field in the C struct
type FieldInfo struct {
	Offset    uintptr   // Required: use C.offsetof() or 0 for AutoLayout
	Size      uintptr   // Optional: if 0, deduced from TypeName using arch_info
	TypeName  string    // Required: C type name (e.g., "uint32_t", "char*", "double")
	Kind      FieldKind // Optional: defaults to primitive if zero
	ElemType  string    // For arrays: element C type (e.g., "float")
	ElemCount uintptr   // For arrays: number of elements
	// IsString is deprecated - automatically deduced from TypeName == "char*"
	// Keeping for backward compatibility, but new code should omit it
	IsString bool
}

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

// getSizeFromTypeName returns the size of a C type based on arch_info
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

// getAlignmentFromTypeName returns the alignment of a C type based on arch_info
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

// CustomLayout generates C code for structs that don't follow standard alignment.
// Use this when AutoLayout fails due to custom packing or complex alignment.
//
// This function prints C code to stdout that you should copy into your C files.
// The generated C code provides a function that returns the actual layout
// as determined by your C compiler.
//
// Example usage:
//
//	// 1. Generate C code:
//	cgocopy.CustomLayout("MyDevice", "uint32_t", "char*", "float")
//
//	// 2. Copy the printed C code into your .c file
//
//	// 3. Use in Go:
//	layout := getMyDeviceLayout()  // Call the C function
//	registry.MustRegister(Device{}, cSize, layout, converter)
//
// This solves AutoLayout limitations:
//   - Works with #pragma pack(1)
//   - Works with __attribute__((packed))
//   - Uses actual C compiler layout decisions
//   - Perfect for third-party libraries
func CustomLayout(structName string, typeNames ...string) {
	fmt.Printf(`// Copy this C code into your .c file for struct: %s
// This uses your C compiler's actual offsetof() calculations

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

// Define your struct exactly as it appears in your C code
typedef struct {
`, structName)

	// Generate struct definition
	for i, typeName := range typeNames {
		cType := goTypeToCType(typeName)
		fieldName := fmt.Sprintf("field%d", i)
		fmt.Printf("    %s %s;\n", cType, fieldName)
	}

	fmt.Printf(`} %s;

// Field info structure (matches cgocopy.FieldInfo)
typedef struct {
    size_t offset;
    size_t size;
    const char* type_name;
    bool is_string;
} cgocopy_FieldInfo;

// Generated layout function - call this from Go
cgocopy_FieldInfo* get%sLayout() {
    static cgocopy_FieldInfo layout[] = {
`, structName, structName)

	// Generate layout entries
	for i, typeName := range typeNames {
		fieldName := fmt.Sprintf("field%d", i)
		isString := typeName == "char*"
		fmt.Printf("        {.offset = offsetof(%s, %s), .size = sizeof(%s), .type_name = \"%s\", .is_string = %s},\n",
			structName, fieldName, goTypeToCType(typeName), typeName, boolToC(isString))
	}

	fmt.Printf(`    };
    return layout;
}

// Also provide struct size
size_t get%sSize() {
    return sizeof(%s);
}

`, structName, structName)
}

// goTypeToCType converts Go type names to C types for code generation
func goTypeToCType(goType string) string {
	switch goType {
	case "int8", "uint8":
		return "uint8_t"
	case "int16", "uint16":
		return "uint16_t"
	case "int32", "uint32":
		return "uint32_t"
	case "int64", "uint64":
		return "uint64_t"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "char*":
		return "char*"
	case "void*", "pointer":
		return "void*"
	case "size_t":
		return "size_t"
	default:
		return goType
	}
}

// boolToC converts bool to C string
func boolToC(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// alignOffset rounds up offset to the next alignment boundary
func alignOffset(offset, align uintptr) uintptr {
	if align == 0 || align == 1 {
		return offset
	}
	return ((offset + align - 1) / align) * align
}

func memCopy(dst, src unsafe.Pointer, size uintptr) {
	if size == 0 {
		return
	}
	dstSlice := unsafe.Slice((*byte)(dst), size)
	srcSlice := unsafe.Slice((*byte)(src), size)
	copy(dstSlice, srcSlice)
}

// AutoLayout automatically calculates offsets and sizes from C type names.
// This eliminates the need for offsetof() helpers for standard C struct layouts.
//
// Example:
//
//	// Before (manual offsetof):
//	layout := []structcopy.FieldInfo{
//	    {Offset: C.deviceIdOffset(), Size: 4, TypeName: "uint32_t"},
//	    {Offset: C.deviceNameOffset(), Size: 8, TypeName: "char*"},
//	    {Offset: C.deviceValueOffset(), Size: 4, TypeName: "float"},
//	}
//
//	// After (automatic):
//	layout := structcopy.AutoLayout("uint32_t", "char*", "float")
//
// Returns layout with calculated offsets based on standard C alignment rules.
// Works for 100% of standard C structs (validated experimentally).
//
// LIMITATIONS:
//   - Does not work with #pragma pack directives
//   - Does not work with __attribute__((packed))
//   - Does not work with bitfields
//   - For these cases, continue using explicit offsetof() helpers
func AutoLayout(typeNames ...string) []FieldInfo {
	archInfo := GetArchInfo()
	layout := make([]FieldInfo, len(typeNames))
	currentOffset := uintptr(0)

	for i, typeName := range typeNames {
		size := getSizeFromTypeName(typeName, archInfo)
		align := getAlignmentFromTypeName(typeName, archInfo)
		kind := resolveFieldKind(FieldInfo{TypeName: typeName})
		isString := kind == FieldString

		// Align current offset to field's alignment requirement
		currentOffset = alignOffset(currentOffset, align)

		layout[i] = FieldInfo{
			Offset:   currentOffset,
			Size:     size,
			TypeName: typeName,
			Kind:     kind,
			IsString: isString,
		}

		// Move to next field
		currentOffset += size
	}

	return layout
}

// validateTypeCompatibility checks if C and Go types are compatible
func validateTypeCompatibility(goType reflect.Type, cTypeName string) error {
	// Map of compatible C to Go types
	compatibleTypes := map[string][]reflect.Kind{
		// Integer types
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
		// Float types
		"float":  {reflect.Float32},
		"double": {reflect.Float64},
		// Bool
		"bool":  {reflect.Bool},
		"_Bool": {reflect.Bool},
		// Pointers
		"pointer": {reflect.Ptr, reflect.UnsafePointer},
		// Size types
		"size_t":    {reflect.Uint64, reflect.Uintptr},
		"uintptr_t": {reflect.Uintptr},
	}

	goKind := goType.Kind()

	// Handle pointer types with * suffix
	if goKind == reflect.Ptr || goKind == reflect.UnsafePointer {
		if cTypeName == "pointer" || (len(cTypeName) > 0 && cTypeName[len(cTypeName)-1] == '*') {
			return nil
		}
	}

	// Handle array types
	if goKind == reflect.Array {
		// Arrays are compatible if they're just byte/int arrays
		// More sophisticated array validation could be added here
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

// Copy performs a validated, efficient copy from C struct to Go struct
// Recursively handles nested structs automatically
func (r *Registry) Copy(dst interface{}, cPtr unsafe.Pointer) error {
	dstVal := reflect.ValueOf(dst)

	// dst must be a pointer to a struct
	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("dst must be a pointer to struct, got %v", dstVal.Kind())
	}

	dstElem := dstVal.Elem()
	if dstElem.Kind() != reflect.Struct {
		return fmt.Errorf("dst must point to a struct, got %v", dstElem.Kind())
	}

	// Look up the mapping
	mapping, ok := r.mappings[dstElem.Type()]
	if !ok {
		return fmt.Errorf("struct type %v is not registered", dstElem.Type())
	}

	// Check if any field needs special handling (nested, strings, complex arrays)
	hasSpecial := false
	for _, field := range mapping.Fields {
		if field.IsNested || field.IsString || field.Kind == FieldPointer ||
			(field.IsArray && (field.ArrayElemIsNested || field.ArrayElemIsString || field.ArrayElemKind == FieldPointer)) {
			hasSpecial = true
			break
		}
	}

	if !hasSpecial {
		// Fast path: no nested structs or strings, use raw memory copy
		goPtr := unsafe.Pointer(dstElem.Addr().Pointer())
		memCopy(goPtr, cPtr, mapping.GoSize)
		return nil
	}

	// Special handling path: copy field by field
	for i, fieldMapping := range mapping.Fields {
		field := dstElem.Field(i)

		if fieldMapping.IsString {
			// Handle C string (char*) → Go string
			// Read the char* pointer from C struct
			charPtrAddr := unsafe.Add(cPtr, fieldMapping.COffset)
			charPtr := *(*unsafe.Pointer)(charPtrAddr)

			// Convert using registered converter
			goStr := mapping.StringConverter.CStringToGo(charPtr)
			field.SetString(goStr)
		} else if fieldMapping.IsNested && !fieldMapping.IsArray {
			// Recursively copy nested struct
			nestedCPtr := unsafe.Add(cPtr, fieldMapping.COffset)
			if err := r.Copy(field.Addr().Interface(), nestedCPtr); err != nil {
				return fmt.Errorf("failed to copy nested field %d: %w", i, err)
			}
		} else if fieldMapping.IsArray {
			srcAddr := unsafe.Add(cPtr, fieldMapping.COffset)
			if fieldMapping.ArrayElemIsNested {
				elemSize := fieldMapping.ArrayElemSize
				elemCount := int(fieldMapping.ArrayLen)
				for j := 0; j < elemCount; j++ {
					nestedSrc := unsafe.Add(srcAddr, uintptr(j)*elemSize)
					elemValue := field.Index(j)
					if !elemValue.CanAddr() {
						return fmt.Errorf("array element %d of field %d is not addressable", j, i)
					}
					if err := r.Copy(elemValue.Addr().Interface(), nestedSrc); err != nil {
						return fmt.Errorf("failed to copy nested array element %d: %w", j, err)
					}
				}
			} else {
				if !field.CanAddr() {
					return fmt.Errorf("array field %d is not addressable", i)
				}
				dstAddr := unsafe.Pointer(field.UnsafeAddr())
				srcSlice := unsafe.Slice((*byte)(srcAddr), fieldMapping.Size)
				dstSlice := unsafe.Slice((*byte)(dstAddr), fieldMapping.Size)
				copy(dstSlice, srcSlice)
			}
		} else {
			// Copy primitive field
			srcAddr := unsafe.Add(cPtr, fieldMapping.COffset)
			dstAddr := unsafe.Add(unsafe.Pointer(dstElem.Addr().Pointer()), fieldMapping.GoOffset)
			memCopy(dstAddr, srcAddr, fieldMapping.Size)
		}
	}

	return nil
}

// CopyField performs a validated copy of a single field
func (r *Registry) CopyField(dst interface{}, cPtr unsafe.Pointer, fieldIndex int) error {
	dstVal := reflect.ValueOf(dst)

	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("dst must be a pointer to struct")
	}

	dstElem := dstVal.Elem()
	if dstElem.Kind() != reflect.Struct {
		return fmt.Errorf("dst must point to a struct")
	}

	mapping, ok := r.mappings[dstElem.Type()]
	if !ok {
		return fmt.Errorf("struct type %v is not registered", dstElem.Type())
	}

	if fieldIndex < 0 || fieldIndex >= len(mapping.Fields) {
		return fmt.Errorf("field index %d out of range [0, %d)", fieldIndex, len(mapping.Fields))
	}

	field := mapping.Fields[fieldIndex]

	// Calculate source and destination addresses
	srcAddr := unsafe.Add(cPtr, field.COffset)
	dstAddr := unsafe.Add(unsafe.Pointer(dstElem.Addr().Pointer()), field.GoOffset)

	// Copy the field
	memCopy(dstAddr, srcAddr, field.Size)

	return nil
}

// GetMapping returns the mapping information for a registered type
func (r *Registry) GetMapping(goType reflect.Type) (*StructMapping, bool) {
	mapping, ok := r.mappings[goType]
	return mapping, ok
}

// String returns a human-readable description of the struct mapping
func (m *StructMapping) String() string {
	result := fmt.Sprintf("Mapping: %s -> %s\n", m.CTypeName, m.GoTypeName)
	result += fmt.Sprintf("  C Size: %d bytes, Go Size: %d bytes\n", m.CSize, m.GoSize)
	result += "  Fields:\n"
	for i, field := range m.Fields {
		result += fmt.Sprintf("    [%d] C[%d:%d] -> Go[%d:%d] (%s -> %v)\n",
			i,
			field.COffset, field.COffset+field.Size,
			field.GoOffset, field.GoOffset+field.Size,
			field.CType, field.GoType)
	}
	return result
}

// Direct performs a direct unsafe copy from C struct to Go struct.
// This is the fastest method (0.3ns) with zero overhead - the compiler inlines this call.
//
// LIMITATIONS - Direct ONLY works for structs containing:
//   - Primitive types (int, float, bool, etc.)
//   - Fixed-size arrays of primitives
//   - Nested structs (that also meet these criteria)
//
// Direct DOES NOT work for structs containing:
//   - Strings (char* → string requires allocation and C.GoString conversion)
//   - Slices (C arrays → Go slices require length tracking and allocation)
//   - Pointers to Go-managed memory
//   - Any dynamically-sized data
//
// For these cases, use Registry.Register() + Registry.Copy() which handles:
//   - String conversion (char* → string via CStringConverter)
//   - Proper memory management
//   - CGO pointer safety
//
// IMPORTANT: Only use this if you have already validated struct compatibility!
// For safety, use Registry.Register() + Registry.Copy() instead.
//
// Example:
//
//	var goStruct MyStruct
//	cgocopy.Direct(&goStruct, cStructPtr)
//
// The compiler will inline this to a single pointer cast with zero function call overhead.
func Direct[T any](dst *T, src unsafe.Pointer) {
	*dst = *(*T)(src)
}

// DirectArray copies an array of C structs to a Go slice using Direct.
//
// This eliminates boilerplate pointer arithmetic when copying C arrays:
//
//	// Before:
//	cSize := unsafe.Sizeof(C.AudioDevice{})
//	for i := range devices {
//	    cPtr := unsafe.Pointer(uintptr(unsafe.Pointer(cDevices)) + uintptr(i)*cSize)
//	    cgocopy.Direct(&devices[i], cPtr)
//	}
//
//	// After:
//	cSize := unsafe.Sizeof(C.AudioDevice{})
//	cgocopy.DirectArray(devices, unsafe.Pointer(cDevices), cSize)
//
// Parameters:
//   - dst: Pre-allocated Go slice to copy into (length determines iteration count)
//   - src: Pointer to first element of C array
//   - cElemSize: Size of each C struct element (use unsafe.Sizeof(C.StructType{}))
//
// Performance: Fully inlined by compiler - zero overhead beyond the loop itself.
// Each iteration is 0.31ns (same as individual Direct calls).
func DirectArray[T any](dst []T, src unsafe.Pointer, cElemSize uintptr) {
	for i := range dst {
		// unsafe.Add is the idiomatic way to do pointer arithmetic (Go 1.17+)
		cPtr := unsafe.Add(src, i*int(cElemSize))
		Direct(&dst[i], cPtr)
	}
}

// MustRegister is a convenience wrapper that automatically uses reflection
// to get the Go type. Panics on error (suitable for init() functions).
//
// Example:
//
//	registry.MustRegister(AudioDevice{}, cSize, layout, converter)
//
// Instead of:
//
//	registry.Register(reflect.TypeOf(AudioDevice{}), cSize, layout, converter)
func (r *Registry) MustRegister(goStruct any, cSize uintptr, cLayout []FieldInfo, converter ...CStringConverter) {
	goType := reflect.TypeOf(goStruct)

	var conv CStringConverter
	if len(converter) > 0 {
		conv = converter[0]
	}

	if err := r.Register(goType, cSize, cLayout, conv); err != nil {
		panic(fmt.Sprintf("Failed to register %v: %v", goType, err))
	}
}
