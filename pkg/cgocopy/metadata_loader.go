package cgocopy

/*
#include "../../native/cgocopy_metadata.h"
*/
import "C"

import (
	"unsafe"
)

// StructMetadata represents C struct metadata captured via cgocopy_metadata.h macros.
type StructMetadata struct {
	Name      string
	Size      uintptr
	Alignment uintptr
	Fields    []FieldInfo
}

func loadStructMetadata(info *C.cgocopy_struct_info) StructMetadata {
	if info == nil {
		return StructMetadata{}
	}

	metadata := StructMetadata{
		Name:      goStringFromCString(info.name),
		Size:      uintptr(info.size),
		Alignment: uintptr(info.alignment),
	}

	count := int(info.field_count)
	if count > 0 && info.fields != nil {
		cFields := unsafe.Slice(info.fields, count)
		metadata.Fields = make([]FieldInfo, count)
		for i, cf := range cFields {
			field := FieldInfo{
				Offset:    uintptr(cf.offset),
				Size:      uintptr(cf.size),
				TypeName:  goStringFromCString(cf.type_name),
				Kind:      fieldKindFromC(cf.kind),
				ElemType:  goStringFromCString(cf.elem_type),
				ElemCount: uintptr(cf.elem_count),
				IsString:  bool(cf.is_string),
			}

			// Ensure Kind/IsString align with inferred metadata for backward compatibility
			field.Kind = resolveFieldKind(field)
			field.IsString = field.IsString || field.Kind == FieldString

			metadata.Fields[i] = field
		}
	}

	return metadata
}

func fieldKindFromC(kind C.cgocopy_field_kind) FieldKind {
	switch kind {
	case C.CGOCOPY_FIELD_POINTER_KIND:
		return FieldPointer
	case C.CGOCOPY_FIELD_STRING_KIND:
		return FieldString
	case C.CGOCOPY_FIELD_ARRAY_KIND:
		return FieldArray
	case C.CGOCOPY_FIELD_STRUCT_KIND:
		return FieldStruct
	case C.CGOCOPY_FIELD_PRIMITIVE_KIND:
		fallthrough
	default:
		return FieldPrimitive
	}
}

func goStringFromCString(str *C.char) string {
	if str == nil {
		return ""
	}
	return C.GoString(str)
}
