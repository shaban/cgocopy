package cgocopy2

import (
	"reflect"
	"strings"
	"unsafe"
)

// Precompile analyzes a Go struct type and registers its metadata for efficient copying.
// This function must be called at initialization time for all types that will be copied.
// It uses reflection to analyze the struct layout and extract field information.
//
// Example:
//
//	type User struct {
//	    ID   int    `cgocopy:"id"`
//	    Name string `cgocopy:"name"`
//	}
//
//	func init() {
//	    cgocopy2.Precompile[User]()
//	}
func Precompile[T any]() error {
	var zero T
	goType := reflect.TypeOf(zero)

	// Dereference pointer types
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	// Check if already registered
	if globalRegistry.IsRegistered(goType) {
		return newRegistrationError(goType, "type already registered", ErrAlreadyRegistered)
	}

	// Validate that it's a struct
	if goType.Kind() != reflect.Struct {
		return newRegistrationError(goType, "only struct types can be precompiled", ErrInvalidType)
	}

	// Analyze the struct
	metadata, err := analyzeStruct(goType)
	if err != nil {
		return newRegistrationError(goType, "failed to analyze struct", err)
	}

	// Register the metadata
	globalRegistry.Register(goType, metadata)

	return nil
}

// PrecompileWithC analyzes a Go struct type and registers it with C struct metadata.
// This function should be used when copying from C structs that may have different
// memory layouts than their Go counterparts (different padding/alignment).
//
// The C metadata should be extracted from C11 CGOCOPY_STRUCT macros which provide
// accurate field offsets for the C struct layout.
//
// Example:
//
//	type User struct {
//	    ID   int32  `cgocopy:"user_id"`
//	    Name string `cgocopy:"username"`
//	}
//
//	func init() {
//	    cMetadata := cgocopy2.CStructInfo{
//	        Name: "User",
//	        Size: uintptr(C.sizeof_User),
//	        Fields: extractCFields(C.CGOCOPY_GET_METADATA(C.User)),
//	    }
//	    cgocopy2.PrecompileWithC[User](cMetadata)
//	}
func PrecompileWithC[T any](cInfo CStructInfo) error {
	var zero T
	goType := reflect.TypeOf(zero)

	// Dereference pointer types
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	// Check if already registered
	if globalRegistry.IsRegistered(goType) {
		return newRegistrationError(goType, "type already registered", ErrAlreadyRegistered)
	}

	// Validate that it's a struct
	if goType.Kind() != reflect.Struct {
		return newRegistrationError(goType, "only struct types can be precompiled", ErrInvalidType)
	}

	// Analyze the struct using C metadata
	metadata, err := analyzeStructWithC(goType, cInfo)
	if err != nil {
		return newRegistrationError(goType, "failed to analyze struct with C metadata", err)
	}

	// Register the metadata
	globalRegistry.Register(goType, metadata)

	return nil
}

// analyzeStruct uses reflection to extract field metadata from a struct type.
func analyzeStruct(goType reflect.Type) (*StructMetadata, error) {
	typeName := goType.Name()
	if typeName == "" {
		typeName = goType.String()
	}

	metadata := &StructMetadata{
		TypeName:         typeName,
		CTypeName:        typeName, // Default to same name, can be overridden
		GoType:           goType,
		Size:             goType.Size(),
		Fields:           make([]FieldInfo, 0, goType.NumField()),
		HasNestedStructs: false,
		IsPrimitive:      false,
	}

	// Analyze each field
	for i := 0; i < goType.NumField(); i++ {
		field := goType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse struct tags
		cName, skip := parseTag(field.Tag.Get("cgocopy"))
		if skip {
			continue
		}
		if cName == "" {
			cName = field.Name
		}

		// Determine field type category
		fieldType, err := categorizeFieldType(field.Type)
		if err != nil {
			return nil, newValidationError(typeName, field.Name, field.Type, "", err.Error())
		}

		// Track nested structs
		if fieldType == FieldTypeStruct {
			metadata.HasNestedStructs = true
		}

		// Get array length and element type for compound types
		arrayLen := 0
		var elemType reflect.Type
		if fieldType == FieldTypeArray {
			arrayLen = field.Type.Len()
			elemType = field.Type.Elem()
		} else if fieldType == FieldTypeSlice || fieldType == FieldTypePointer {
			elemType = field.Type.Elem()
		}

		fieldInfo := FieldInfo{
			Name:        field.Name,
			CName:       cName,
			Type:        fieldType,
			Offset:      field.Offset,
			Size:        field.Type.Size(),
			Skip:        false,
			Index:       i,
			ReflectType: field.Type,
			ArrayLen:    arrayLen,
			ElemType:    elemType,
		}

		metadata.Fields = append(metadata.Fields, fieldInfo)
	}

	// Check if this is effectively a primitive (single primitive field)
	if len(metadata.Fields) == 1 && metadata.Fields[0].Type == FieldTypePrimitive {
		metadata.IsPrimitive = true
	}

	return metadata, nil
}

// analyzeStructWithC uses reflection combined with C metadata to extract field information.
// This function uses C struct field offsets instead of Go struct field offsets to handle
// cases where C and Go have different struct layouts due to padding/alignment.
func analyzeStructWithC(goType reflect.Type, cInfo CStructInfo) (*StructMetadata, error) {
	typeName := goType.Name()
	if typeName == "" {
		typeName = goType.String()
	}

	metadata := &StructMetadata{
		TypeName:         typeName,
		CTypeName:        cInfo.Name,
		GoType:           goType,
		Size:             cInfo.Size,
		Fields:           make([]FieldInfo, 0, goType.NumField()),
		HasNestedStructs: false,
		IsPrimitive:      false,
	}

	// Create a map of C field names to C field metadata for name-based lookup
	cFieldMap := make(map[string]*CFieldInfo)
	for i := range cInfo.Fields {
		cFieldMap[cInfo.Fields[i].Name] = &cInfo.Fields[i]
	}

	// Analyze each Go field and match with C field
	cFieldIdx := 0 // Track position in C fields for positional matching
	for i := 0; i < goType.NumField(); i++ {
		field := goType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse struct tags
		cName, skip := parseTag(field.Tag.Get("cgocopy"))
		if skip {
			continue
		}

		var cField *CFieldInfo
		var ok bool

		if cName != "" {
			// Explicit tag: match by name
			cField, ok = cFieldMap[cName]
			if !ok {
				return nil, newValidationError(typeName, field.Name, field.Type, cName,
					"C field not found in metadata")
			}
		} else {
			// No tag: match by position (field index)
			if cFieldIdx >= len(cInfo.Fields) {
				return nil, newValidationError(typeName, field.Name, field.Type, "",
					"more Go fields than C fields")
			}
			cField = &cInfo.Fields[cFieldIdx]
			cName = cField.Name // Use C field name for reference
			cFieldIdx++
		}

		// Determine field type category
		fieldType, err := categorizeFieldType(field.Type)
		if err != nil {
			return nil, newValidationError(typeName, field.Name, field.Type, cName, err.Error())
		}

		// Track nested structs
		if fieldType == FieldTypeStruct {
			metadata.HasNestedStructs = true
		}

		// Get array length and element type for compound types
		arrayLen := 0
		var elemType reflect.Type
		if fieldType == FieldTypeArray {
			arrayLen = field.Type.Len()
			elemType = field.Type.Elem()
			// Validate array length matches C metadata
			if cField.IsArray && arrayLen != cField.ArrayLen {
				return nil, newValidationError(typeName, field.Name, field.Type, cName,
					"array length mismatch with C metadata")
			}
		} else if fieldType == FieldTypeSlice || fieldType == FieldTypePointer {
			elemType = field.Type.Elem()
		}

		// Use C field offset instead of Go field offset!
		fieldInfo := FieldInfo{
			Name:        field.Name,
			CName:       cName,
			Type:        fieldType,
			Offset:      cField.Offset, // <<< Key difference: using C offset
			Size:        cField.Size,   // <<< Using C size
			Skip:        false,
			Index:       i,
			ReflectType: field.Type,
			ArrayLen:    arrayLen,
			ElemType:    elemType,
		}

		metadata.Fields = append(metadata.Fields, fieldInfo)
	}

	// Check if this is effectively a primitive (single primitive field)
	if len(metadata.Fields) == 1 && metadata.Fields[0].Type == FieldTypePrimitive {
		metadata.IsPrimitive = true
	}

	return metadata, nil
}

// parseTag parses a cgocopy struct tag.
// Returns (cFieldName, skip).
//
// Supported formats:
//   - `cgocopy:"field_name"` - map to C field "field_name"
//   - `cgocopy:"-"` - skip this field
//   - `cgocopy:""` or no tag - use Go field name
func parseTag(tag string) (string, bool) {
	tag = strings.TrimSpace(tag)

	// Empty tag means use the field name as-is
	if tag == "" {
		return "", false
	}

	// Skip marker
	if tag == "-" {
		return "", true
	}

	// Otherwise, use the tag value as the C field name
	return tag, false
}

// categorizeFieldType determines the FieldType category for a reflect.Type.
func categorizeFieldType(t reflect.Type) (FieldType, error) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return FieldTypePrimitive, nil

	case reflect.String:
		return FieldTypeString, nil

	case reflect.Struct:
		return FieldTypeStruct, nil

	case reflect.Array:
		return FieldTypeArray, nil

	case reflect.Slice:
		return FieldTypeSlice, nil

	case reflect.Ptr:
		return FieldTypePointer, nil

	case reflect.Interface, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return FieldTypeInvalid, ErrUnsupportedType

	default:
		return FieldTypeInvalid, ErrUnsupportedType
	}
}

// IsRegistered returns true if the type T has been precompiled.
func IsRegistered[T any]() bool {
	var zero T
	goType := reflect.TypeOf(zero)
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}
	return globalRegistry.IsRegistered(goType)
}

// GetMetadata retrieves the precompiled metadata for type T.
// Returns nil if the type has not been precompiled.
func GetMetadata[T any]() *StructMetadata {
	var zero T
	goType := reflect.TypeOf(zero)
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}
	return globalRegistry.Get(goType)
}

// Reset clears all registered types from the global registry.
// This is primarily useful for testing.
func Reset() {
	globalRegistry.Clear()
}

// isPrimitiveKind checks if a reflect.Kind is a primitive type.
func isPrimitiveKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	default:
		return false
	}
}

// fieldPtr returns a pointer to a field in a struct value.
func fieldPtr(structPtr unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(structPtr) + offset)
}
