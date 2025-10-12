package cgocopy2

import (
	"errors"
	"reflect"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		contains string
	}{
		{
			name: "with field name",
			err: &ValidationError{
				TypeName:  "User",
				FieldName: "Name",
				GoType:    reflect.TypeOf(""),
				CType:     "char*",
				Reason:    "type mismatch",
			},
			contains: "User.Name",
		},
		{
			name: "without field name",
			err: &ValidationError{
				TypeName: "Product",
				Reason:   "missing metadata",
			},
			contains: "Product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Error("Error() returned empty string")
			}
			if tt.contains != "" {
				if len(msg) < len(tt.contains) || !contains(msg, tt.contains) {
					t.Errorf("Error() = %q, should contain %q", msg, tt.contains)
				}
			}
		})
	}
}

func TestValidationError_Is(t *testing.T) {
	err := &ValidationError{TypeName: "Test", Reason: "test"}
	
	if !errors.Is(err, &ValidationError{}) {
		t.Error("ValidationError should match ValidationError type")
	}
	
	if errors.Is(err, ErrInvalidType) {
		t.Error("ValidationError should not match ErrInvalidType")
	}
}

func TestRegistrationError_Error(t *testing.T) {
	type TestType struct{}
	goType := reflect.TypeOf(TestType{})
	
	tests := []struct {
		name     string
		err      *RegistrationError
		contains string
	}{
		{
			name: "with cause",
			err: &RegistrationError{
				Type:   goType,
				Reason: "invalid field",
				Cause:  ErrUnsupportedType,
			},
			contains: "invalid field",
		},
		{
			name: "without cause",
			err: &RegistrationError{
				Type:   goType,
				Reason: "already registered",
			},
			contains: "already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Error("Error() returned empty string")
			}
			if !contains(msg, tt.contains) {
				t.Errorf("Error() = %q, should contain %q", msg, tt.contains)
			}
		})
	}
}

func TestRegistrationError_Unwrap(t *testing.T) {
	cause := ErrUnsupportedType
	err := &RegistrationError{
		Type:   reflect.TypeOf(0),
		Reason: "test",
		Cause:  cause,
	}
	
	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestCopyError_Error(t *testing.T) {
	type TestType struct{}
	goType := reflect.TypeOf(TestType{})
	
	tests := []struct {
		name     string
		err      *CopyError
		contains string
	}{
		{
			name: "with field and cause",
			err: &CopyError{
				Type:      goType,
				FieldName: "Name",
				Reason:    "nil pointer",
				Cause:     ErrNilPointer,
			},
			contains: "Name",
		},
		{
			name: "without field",
			err: &CopyError{
				Type:   goType,
				Reason: "not registered",
			},
			contains: "not registered",
		},
		{
			name: "with unnamed type",
			err: &CopyError{
				Type:   reflect.TypeOf(0),
				Reason: "test error",
			},
			contains: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Error("Error() returned empty string")
			}
			if !contains(msg, tt.contains) {
				t.Errorf("Error() = %q, should contain %q", msg, tt.contains)
			}
		})
	}
}

func TestCopyError_Unwrap(t *testing.T) {
	cause := ErrNotRegistered
	err := &CopyError{
		Type:   reflect.TypeOf(0),
		Reason: "test",
		Cause:  cause,
	}
	
	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error constants are defined and not nil
	errorConstants := []error{
		ErrNotRegistered,
		ErrNilPointer,
		ErrInvalidType,
		ErrUnsupportedType,
		ErrFieldMismatch,
		ErrMetadataNotFound,
		ErrAlreadyRegistered,
		ErrTagFormat,
	}
	
	for i, err := range errorConstants {
		if err == nil {
			t.Errorf("error constant at index %d is nil", i)
		}
		if err.Error() == "" {
			t.Errorf("error constant at index %d has empty message", i)
		}
	}
}

func TestNewValidationError(t *testing.T) {
	err := newValidationError("User", "Name", reflect.TypeOf(""), "char*", "test reason")
	
	if err.TypeName != "User" {
		t.Errorf("TypeName = %q, want %q", err.TypeName, "User")
	}
	if err.FieldName != "Name" {
		t.Errorf("FieldName = %q, want %q", err.FieldName, "Name")
	}
	if err.CType != "char*" {
		t.Errorf("CType = %q, want %q", err.CType, "char*")
	}
	if err.Reason != "test reason" {
		t.Errorf("Reason = %q, want %q", err.Reason, "test reason")
	}
}

func TestNewRegistrationError(t *testing.T) {
	cause := ErrUnsupportedType
	err := newRegistrationError(reflect.TypeOf(0), "test reason", cause)
	
	if err.Type != reflect.TypeOf(0) {
		t.Errorf("Type = %v, want %v", err.Type, reflect.TypeOf(0))
	}
	if err.Reason != "test reason" {
		t.Errorf("Reason = %q, want %q", err.Reason, "test reason")
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewCopyError(t *testing.T) {
	cause := ErrNilPointer
	err := newCopyError(reflect.TypeOf(0), "Field", "test reason", cause)
	
	if err.Type != reflect.TypeOf(0) {
		t.Errorf("Type = %v, want %v", err.Type, reflect.TypeOf(0))
	}
	if err.FieldName != "Field" {
		t.Errorf("FieldName = %q, want %q", err.FieldName, "Field")
	}
	if err.Reason != "test reason" {
		t.Errorf("Reason = %q, want %q", err.Reason, "test reason")
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
