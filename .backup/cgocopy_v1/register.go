package cgocopy

import (
	"fmt"
	"reflect"
)

// RegisterStruct validates and registers the mapping for the provided Go struct
// type using the package-level registry. Call Finalize() once all registrations
// are complete.
func RegisterStruct[T any](conv CStringConverter) error {
	var zero T
	return defaultRegistry.registerStruct(zero, conv)
}

// registerStruct registers the provided struct type on a specific registry instance.
func (r *Registry) registerStruct(goStruct any, conv CStringConverter) error {
	if r == nil {
		return ErrNilRegistry
	}
	if r.IsFinalized() {
		return ErrRegistryFinalized
	}

	goType := reflect.TypeOf(goStruct)
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}
	if goType.Kind() != reflect.Struct {
		return ErrNotAStructType
	}
	if goType.Name() == "" {
		return ErrAnonymousStruct
	}

	if _, ok := r.GetMapping(goType); ok {
		return nil
	}

	visited := make(map[reflect.Type]bool)
	return registerStructType(r, goType, conv, visited)
}

func registerStructType(r *Registry, goType reflect.Type, conv CStringConverter, visited map[reflect.Type]bool) error {
	if _, ok := r.GetMapping(goType); ok {
		return nil
	}
	if r.IsFinalized() {
		return ErrRegistryFinalized
	}
	if visited[goType] {
		return nil
	}
	visited[goType] = true

	metadata, err := lookupStructMetadata(goType.Name())
	if err != nil {
		return err
	}

	if len(metadata.Fields) != goType.NumField() {
		return fmt.Errorf("cgocopy: field count mismatch for %s (C=%d Go=%d)", goType.Name(), len(metadata.Fields), goType.NumField())
	}

	for i := 0; i < goType.NumField(); i++ {
		fieldInfo := metadata.Fields[i]
		goField := goType.Field(i)

		switch fieldInfo.Kind {
		case FieldStruct:
			if goField.Type.Kind() != reflect.Struct {
				return fmt.Errorf("cgocopy: field %q in %s expected struct, got %s", goField.Name, goType.Name(), goField.Type.Kind())
			}
			if err := registerStructType(r, goField.Type, conv, visited); err != nil {
				return err
			}
		case FieldArray:
			if goField.Type.Kind() != reflect.Array && goField.Type.Kind() != reflect.Slice {
				return fmt.Errorf("cgocopy: field %q in %s expected array or slice, got %s", goField.Name, goType.Name(), goField.Type.Kind())
			}
			if goField.Type.Kind() == reflect.Array {
				if uintptr(goField.Type.Len()) != fieldInfo.ElemCount {
					return fmt.Errorf("cgocopy: field %q array length mismatch (C=%d Go=%d)", goField.Name, fieldInfo.ElemCount, goField.Type.Len())
				}
			}
			if goField.Type.Elem().Kind() == reflect.Struct {
				if err := registerStructType(r, goField.Type.Elem(), conv, visited); err != nil {
					return err
				}
			}
		}
	}

	resolvedConverter := conv
	if resolvedConverter == nil && metadataHasString(metadata) {
		resolvedConverter = DefaultCStringConverter
	}

	if resolvedConverter != nil {
		return r.Register(goType, metadata.Size, metadata.Fields, resolvedConverter)
	}
	return r.Register(goType, metadata.Size, metadata.Fields)
}

func metadataHasString(md StructMetadata) bool {
	for _, field := range md.Fields {
		if field.Kind == FieldString {
			return true
		}
	}
	return false
}
