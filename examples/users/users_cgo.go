package users

//go:generate ../../tools/cgocopy-generate/cgocopy-generate -input=native/structs.h -output=native/structs_meta.c -api=native/metadata_api.h

/*
#cgo CFLAGS: -I${SRCDIR}/../../pkg/cgocopy
#include <stddef.h>
#include "native/metadata_api.h"
#include "native/structs_meta.c"
#include "native/helpers.c"
*/
import "C"

import (
	"unsafe"

	cgocopy "github.com/shaban/cgocopy/pkg/cgocopy"
)

func extractCMetadata(cStructInfoPtr *C.cgocopy_struct_info) cgocopy.CStructInfo {
	if cStructInfoPtr == nil {
		panic("nil C struct info pointer")
	}

	fieldCount := int(cStructInfoPtr.field_count)
	fields := make([]cgocopy.CFieldInfo, fieldCount)

	cFieldsSlice := (*[1 << 30]C.cgocopy_field_info)(unsafe.Pointer(cStructInfoPtr.fields))[:fieldCount:fieldCount]

	for i := 0; i < fieldCount; i++ {
		cField := &cFieldsSlice[i]
		fields[i] = cgocopy.CFieldInfo{
			Name:      C.GoString(cField.name),
			Type:      C.GoString(cField._type),
			Offset:    uintptr(cField.offset),
			Size:      uintptr(cField.size),
			IsPointer: cField.is_pointer != 0,
			IsArray:   cField.is_array != 0,
			ArrayLen:  int(cField.array_len),
		}
	}

	return cgocopy.CStructInfo{
		Name:   C.GoString(cStructInfoPtr.name),
		Size:   uintptr(cStructInfoPtr.size),
		Fields: fields,
	}
}

func init() {
	if err := cgocopy.PrecompileWithC[UserRole](
		extractCMetadata(C.get_UserRole_metadata()),
	); err != nil {
		panic(err)
	}

	if err := cgocopy.PrecompileWithC[UserDetails](
		extractCMetadata(C.get_UserDetails_metadata()),
	); err != nil {
		panic(err)
	}

	if err := cgocopy.PrecompileWithC[User](
		extractCMetadata(C.get_User_metadata()),
	); err != nil {
		panic(err)
	}
}

func createSampleUsers() (unsafe.Pointer, int) {
	var count C.size_t
	ptr := C.createUsers(&count)
	return unsafe.Pointer(ptr), int(count)
}

func freeSampleUsers(ptr unsafe.Pointer, count int) {
	C.freeUsers((*C.User)(ptr), C.size_t(count))
}
