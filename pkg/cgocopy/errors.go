package cgocopy

import "errors"

var (
	ErrNilRegistry                 = errors.New("cgocopy: registry must not be nil")
	ErrNotAStructType              = errors.New("cgocopy: type must be a struct")
	ErrAnonymousStruct             = errors.New("cgocopy: anonymous struct types are not supported")
	ErrMetadataMissing             = errors.New("cgocopy: metadata not found for struct")
	ErrNilDestination              = errors.New("cgocopy: destination pointer must not be nil")
	ErrNilSourcePointer            = errors.New("cgocopy: source pointer must not be nil")
	ErrDestinationNotStructPointer = errors.New("cgocopy: destination must be a pointer to struct")
	ErrRegistryFinalized           = errors.New("cgocopy: registry has been finalized")
	ErrRegistryNotFinalized        = errors.New("cgocopy: registry must be finalized before copying")
	ErrStructNotRegistered         = errors.New("cgocopy: struct type is not registered")
)
