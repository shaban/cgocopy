// Package cgocopy2 provides improved type-safe copying between C and Go structures
// with simplified macros, thread-safe registry, and tag support.
package cgocopy2

import (
	"reflect"
	"sync"
)

// FieldType represents the type of a struct field.
type FieldType int

const (
	// FieldTypeInvalid represents an uninitialized or invalid field type.
	FieldTypeInvalid FieldType = iota

	// FieldTypePrimitive represents numeric and boolean types:
	// int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
	// float32, float64, bool, byte, rune.
	FieldTypePrimitive

	// FieldTypeString represents Go string type.
	FieldTypeString

	// FieldTypeStruct represents nested struct types.
	FieldTypeStruct

	// FieldTypeArray represents fixed-size array types.
	FieldTypeArray

	// FieldTypeSlice represents dynamic slice types.
	FieldTypeSlice

	// FieldTypePointer represents pointer types.
	FieldTypePointer
)

// String returns the string representation of a FieldType.
func (ft FieldType) String() string {
	switch ft {
	case FieldTypePrimitive:
		return "Primitive"
	case FieldTypeString:
		return "String"
	case FieldTypeStruct:
		return "Struct"
	case FieldTypeArray:
		return "Array"
	case FieldTypeSlice:
		return "Slice"
	case FieldTypePointer:
		return "Pointer"
	default:
		return "Invalid"
	}
}

// FieldInfo contains metadata about a single struct field.
type FieldInfo struct {
	// Name is the field name in the Go struct.
	Name string

	// CName is the field name in the C struct (may differ due to tags).
	CName string

	// Type is the category of the field (primitive, string, struct, etc).
	Type FieldType

	// Offset is the byte offset of this field in the Go struct.
	Offset uintptr

	// Size is the size in bytes of this field.
	Size uintptr

	// Skip indicates whether this field should be skipped during copying
	// (set via `cgocopy:"-"` tag).
	Skip bool

	// Index is the field index in the struct (for reflect.StructField access).
	Index int

	// ReflectType is the reflect.Type of the field for type checking.
	ReflectType reflect.Type

	// ArrayLen is the length for array types (0 for non-arrays).
	ArrayLen int

	// ElemType is the element type for arrays, slices, and pointers.
	ElemType reflect.Type
}

// StructMetadata contains all metadata needed to copy a struct type.
type StructMetadata struct {
	// TypeName is the fully qualified Go type name.
	TypeName string

	// CTypeName is the C struct name (without "struct" prefix).
	CTypeName string

	// Fields contains metadata for each field in the struct.
	Fields []FieldInfo

	// GoType is the reflect.Type of the Go struct.
	GoType reflect.Type

	// Size is the total size of the struct in bytes.
	Size uintptr

	// HasNestedStructs indicates if any fields are nested structs.
	HasNestedStructs bool

	// IsPrimitive indicates if this is a simple primitive type
	// (used for FastCopy optimization).
	IsPrimitive bool
}

// Registry is the thread-safe registry for struct metadata.
type Registry struct {
	mu       sync.RWMutex
	metadata map[reflect.Type]*StructMetadata

	// cTypeMap maps C type names to reflect.Type for faster lookup.
	cTypeMap map[string]reflect.Type
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		metadata: make(map[reflect.Type]*StructMetadata),
		cTypeMap: make(map[string]reflect.Type),
	}
}

// Register adds struct metadata to the registry.
// This is called internally by Precompile.
func (r *Registry) Register(goType reflect.Type, metadata *StructMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metadata[goType] = metadata
	r.cTypeMap[metadata.CTypeName] = goType
}

// Get retrieves struct metadata from the registry.
// Returns nil if the type has not been registered.
func (r *Registry) Get(goType reflect.Type) *StructMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.metadata[goType]
}

// GetByCName retrieves struct metadata by C type name.
// Returns nil if the type has not been registered.
func (r *Registry) GetByCName(cTypeName string) *StructMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	goType, ok := r.cTypeMap[cTypeName]
	if !ok {
		return nil
	}
	return r.metadata[goType]
}

// IsRegistered checks if a type has been precompiled.
func (r *Registry) IsRegistered(goType reflect.Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.metadata[goType]
	return ok
}

// Count returns the number of registered types.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.metadata)
}

// Clear removes all registered types (primarily for testing).
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metadata = make(map[reflect.Type]*StructMetadata)
	r.cTypeMap = make(map[string]reflect.Type)
}

// globalRegistry is the package-level registry instance.
var globalRegistry = NewRegistry()

// CFieldInfo represents metadata for a C struct field extracted from C macros.
type CFieldInfo struct {
	// Name is the field name in the C struct.
	Name string

	// Type is the C type string (e.g., "int32", "string", "float64").
	Type string

	// Offset is the byte offset in the C struct.
	Offset uintptr

	// Size is the size in bytes.
	Size uintptr

	// IsPointer indicates if this is a pointer field.
	IsPointer bool

	// IsArray indicates if this is an array field.
	IsArray bool

	// ArrayLen is the array length (0 if not an array).
	ArrayLen int
}

// CStructInfo represents metadata for a complete C struct extracted from C macros.
type CStructInfo struct {
	// Name is the C struct name.
	Name string

	// Size is the total size of the C struct.
	Size uintptr

	// Fields contains metadata for all fields.
	Fields []CFieldInfo
}
