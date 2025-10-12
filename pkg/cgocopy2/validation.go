package cgocopy2

import (
	"fmt"
	"reflect"
)

// ValidateStruct checks if a struct type is properly registered and can be copied.
// It performs comprehensive validation including:
//   - Checking if the type is registered
//   - Verifying all fields are supported types
//   - Ensuring nested structs are also registered
//   - Validating field metadata completeness
//
// This function is useful for debugging registration issues at initialization time.
//
// Example:
//
//	func init() {
//	    cgocopy2.Precompile[User]()
//	    if err := cgocopy2.ValidateStruct[User](); err != nil {
//	        log.Fatalf("User struct validation failed: %v", err)
//	    }
//	}
func ValidateStruct[T any]() error {
	var zero T
	goType := reflect.TypeOf(zero)

	// Dereference pointer types
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	// Check if it's a struct
	if goType.Kind() != reflect.Struct {
		return newValidationError(goType.String(), "", goType, "", "not a struct type")
	}

	// Check if registered
	metadata := globalRegistry.Get(goType)
	if metadata == nil {
		return newValidationError(goType.Name(), "", goType, "", 
			fmt.Sprintf("type not registered - call Precompile[%s]() first", goType.Name()))
	}

	// Validate metadata completeness
	if err := validateMetadata(metadata); err != nil {
		return err
	}

	// Validate nested structs are also registered
	if err := validateNestedStructs(metadata); err != nil {
		return err
	}

	return nil
}

// validateMetadata checks if struct metadata is complete and valid.
func validateMetadata(metadata *StructMetadata) error {
	if metadata.TypeName == "" {
		return newValidationError("", "", metadata.GoType, "", "metadata missing type name")
	}

	if metadata.GoType == nil {
		return newValidationError(metadata.TypeName, "", nil, "", "metadata missing Go type")
	}

	// Validate each field
	for i := range metadata.Fields {
		field := &metadata.Fields[i]
		if err := validateField(metadata.TypeName, field); err != nil {
			return err
		}
	}

	return nil
}

// validateField checks if a single field is valid.
func validateField(typeName string, field *FieldInfo) error {
	if field.Name == "" {
		return newValidationError(typeName, "", field.ReflectType, "", 
			"field has empty name")
	}

	if field.ReflectType == nil {
		return newValidationError(typeName, field.Name, nil, field.CName, 
			"field missing reflect type")
	}

	// Validate field type is supported
	switch field.Type {
	case FieldTypePrimitive:
		if !isPrimitiveKind(field.ReflectType.Kind()) {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				fmt.Sprintf("field marked as primitive but has kind %v", field.ReflectType.Kind()))
		}

	case FieldTypeString:
		if field.ReflectType.Kind() != reflect.String && 
		   field.ReflectType.Kind() != reflect.Ptr {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				fmt.Sprintf("field marked as string but has kind %v", field.ReflectType.Kind()))
		}

	case FieldTypeStruct:
		if field.ReflectType.Kind() != reflect.Struct {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				fmt.Sprintf("field marked as struct but has kind %v", field.ReflectType.Kind()))
		}

	case FieldTypeArray:
		if field.ReflectType.Kind() != reflect.Array {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				fmt.Sprintf("field marked as array but has kind %v", field.ReflectType.Kind()))
		}
		if field.ArrayLen <= 0 {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				"array field has invalid length")
		}
		if field.ElemType == nil {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				"array field missing element type")
		}

	case FieldTypeSlice:
		if field.ReflectType.Kind() != reflect.Slice {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				fmt.Sprintf("field marked as slice but has kind %v", field.ReflectType.Kind()))
		}
		if field.ElemType == nil {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				"slice field missing element type")
		}

	case FieldTypePointer:
		if field.ReflectType.Kind() != reflect.Ptr {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				fmt.Sprintf("field marked as pointer but has kind %v", field.ReflectType.Kind()))
		}
		if field.ElemType == nil {
			return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
				"pointer field missing element type")
		}

	default:
		return newValidationError(typeName, field.Name, field.ReflectType, field.CName,
			fmt.Sprintf("unsupported field type: %v", field.Type))
	}

	return nil
}

// validateNestedStructs ensures all nested struct types are registered.
func validateNestedStructs(metadata *StructMetadata) error {
	for i := range metadata.Fields {
		field := &metadata.Fields[i]

		if field.Skip {
			continue
		}

		// Check if field is a nested struct
		if field.Type == FieldTypeStruct {
			nestedMeta := globalRegistry.Get(field.ReflectType)
			if nestedMeta == nil {
				return newValidationError(metadata.TypeName, field.Name, field.ReflectType, field.CName,
					fmt.Sprintf("nested struct type %s not registered - call Precompile[%s]() first",
						field.ReflectType.Name(), field.ReflectType.Name()))
			}
		}

		// Check element type for arrays and slices
		if field.Type == FieldTypeArray || field.Type == FieldTypeSlice {
			if field.ElemType.Kind() == reflect.Struct {
				elemMeta := globalRegistry.Get(field.ElemType)
				if elemMeta == nil {
					return newValidationError(metadata.TypeName, field.Name, field.ElemType, field.CName,
						fmt.Sprintf("array/slice element type %s not registered - call Precompile[%s]() first",
							field.ElemType.Name(), field.ElemType.Name()))
				}
			}
		}

		// Check pointer element type
		if field.Type == FieldTypePointer {
			if field.ElemType.Kind() == reflect.Struct {
				elemMeta := globalRegistry.Get(field.ElemType)
				if elemMeta == nil {
					return newValidationError(metadata.TypeName, field.Name, field.ElemType, field.CName,
						fmt.Sprintf("pointer element type %s not registered - call Precompile[%s]() first",
							field.ElemType.Name(), field.ElemType.Name()))
				}
			}
		}
	}

	return nil
}

// ValidateAll validates all registered types.
// This is useful for checking the entire type registry at once.
func ValidateAll() []error {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	var errors []error
	for _, metadata := range globalRegistry.metadata {
		if err := validateMetadata(metadata); err != nil {
			errors = append(errors, err)
		}
		if err := validateNestedStructs(metadata); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// MustValidateStruct is like ValidateStruct but panics on error.
// This is useful for initialization-time validation where failures should be fatal.
func MustValidateStruct[T any]() {
	if err := ValidateStruct[T](); err != nil {
		panic(fmt.Sprintf("struct validation failed: %v", err))
	}
}

// GetRegisteredTypes returns a list of all registered type names.
// This is useful for debugging and introspection.
func GetRegisteredTypes() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	types := make([]string, 0, len(globalRegistry.metadata))
	for _, metadata := range globalRegistry.metadata {
		types = append(types, metadata.TypeName)
	}
	return types
}
