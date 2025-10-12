package cgocopy2

import (
	"errors"
	"fmt"
	"reflect"
)

// Common errors returned by cgocopy2.
var (
	// ErrNotRegistered is returned when attempting to copy a type
	// that has not been precompiled.
	ErrNotRegistered = errors.New("type not registered: call Precompile[T]() first")
	
	// ErrNilPointer is returned when a nil C pointer is passed to Copy.
	ErrNilPointer = errors.New("nil C pointer")
	
	// ErrInvalidType is returned when a type cannot be used with cgocopy2.
	ErrInvalidType = errors.New("invalid type for cgocopy")
	
	// ErrUnsupportedType is returned when a field type is not supported.
	ErrUnsupportedType = errors.New("unsupported field type")
	
	// ErrFieldMismatch is returned when C and Go struct fields don't match.
	ErrFieldMismatch = errors.New("field count or types don't match")
	
	// ErrMetadataNotFound is returned when C metadata is not available.
	ErrMetadataNotFound = errors.New("C metadata not found for type")
	
	// ErrAlreadyRegistered is returned when trying to register a type twice.
	ErrAlreadyRegistered = errors.New("type already registered")
	
	// ErrTagFormat is returned when a cgocopy struct tag has invalid format.
	ErrTagFormat = errors.New("invalid cgocopy tag format")
)

// ValidationError represents a validation failure for a specific field.
type ValidationError struct {
	TypeName  string
	FieldName string
	GoType    reflect.Type
	CType     string
	Reason    string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.FieldName != "" {
		return fmt.Sprintf("validation failed for %s.%s: %s (Go type: %v, C type: %s)",
			e.TypeName, e.FieldName, e.Reason, e.GoType, e.CType)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.TypeName, e.Reason)
}

// Is allows ValidationError to be used with errors.Is.
func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

// RegistrationError represents an error during type registration.
type RegistrationError struct {
	Type   reflect.Type
	Reason string
	Cause  error
}

// Error implements the error interface.
func (e *RegistrationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to register type %v: %s: %v", e.Type, e.Reason, e.Cause)
	}
	return fmt.Sprintf("failed to register type %v: %s", e.Type, e.Reason)
}

// Unwrap allows RegistrationError to be used with errors.Unwrap.
func (e *RegistrationError) Unwrap() error {
	return e.Cause
}

// CopyError represents an error during copying.
type CopyError struct {
	Type      reflect.Type
	FieldName string
	Reason    string
	Cause     error
}

// Error implements the error interface.
func (e *CopyError) Error() string {
	typeName := e.Type.Name()
	if typeName == "" {
		typeName = e.Type.String()
	}
	
	if e.FieldName != "" {
		if e.Cause != nil {
			return fmt.Sprintf("copy failed for %s.%s: %s: %v", typeName, e.FieldName, e.Reason, e.Cause)
		}
		return fmt.Sprintf("copy failed for %s.%s: %s", typeName, e.FieldName, e.Reason)
	}
	
	if e.Cause != nil {
		return fmt.Sprintf("copy failed for %s: %s: %v", typeName, e.Reason, e.Cause)
	}
	return fmt.Sprintf("copy failed for %s: %s", typeName, e.Reason)
}

// Unwrap allows CopyError to be used with errors.Unwrap.
func (e *CopyError) Unwrap() error {
	return e.Cause
}

// newValidationError creates a new ValidationError.
func newValidationError(typeName, fieldName string, goType reflect.Type, cType, reason string) *ValidationError {
	return &ValidationError{
		TypeName:  typeName,
		FieldName: fieldName,
		GoType:    goType,
		CType:     cType,
		Reason:    reason,
	}
}

// newRegistrationError creates a new RegistrationError.
func newRegistrationError(typ reflect.Type, reason string, cause error) *RegistrationError {
	return &RegistrationError{
		Type:   typ,
		Reason: reason,
		Cause:  cause,
	}
}

// newCopyError creates a new CopyError.
func newCopyError(typ reflect.Type, fieldName, reason string, cause error) *CopyError {
	return &CopyError{
		Type:      typ,
		FieldName: fieldName,
		Reason:    reason,
		Cause:     cause,
	}
}
