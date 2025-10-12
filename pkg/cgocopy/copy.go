package cgocopy

import (
	"fmt"
	"reflect"
	"unsafe"
)

type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// Copy performs a metadata-guided copy from the C source pointer into the
// provided destination pointer. The struct type must have been registered and
// Finalize() must have been called before invoking Copy.
func Copy[T any](dst *T, cPtr unsafe.Pointer) error {
	return copyWithRegistry(defaultRegistry, dst, cPtr)
}

func copyWithRegistry[T any](reg *Registry, dst *T, cPtr unsafe.Pointer) error {
	if dst == nil {
		return ErrNilDestination
	}
	if cPtr == nil {
		return ErrNilSourcePointer
	}

	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return ErrDestinationNotStructPointer
	}
	elemsType := dstValue.Type().Elem()
	if elemsType.Kind() != reflect.Struct {
		return ErrDestinationNotStructPointer
	}

	if reg == nil {
		return ErrNilRegistry
	}

	if !reg.IsFinalized() {
		return ErrRegistryNotFinalized
	}

	mapping, ok := reg.GetMapping(elemsType)
	if !ok {
		return ErrStructNotRegistered
	}

	structVal := dstValue.Elem()
	if mapping.CanFastPath {
		dstPtr := unsafe.Pointer(structVal.UnsafeAddr())
		copyBytes(dstPtr, cPtr, mapping.GoSize)
		return nil
	}
	return copyStructWithMapping(reg, mapping, structVal, cPtr)
}

func copyStructWithMapping(reg *Registry, mapping *StructMapping, dstVal reflect.Value, cPtr unsafe.Pointer) error {
	if mapping == nil {
		return fmt.Errorf("cgocopy: missing struct mapping")
	}

	for idx, fieldMapping := range mapping.Fields {
		field := dstVal.Field(idx)
		fieldSrc := unsafe.Add(cPtr, fieldMapping.COffset)

		switch {
		case fieldMapping.IsString:
			if err := setStringField(mapping, field, fieldSrc); err != nil {
				return err
			}
		case fieldMapping.IsArray:
			if err := copyArrayField(reg, field, fieldMapping, fieldSrc); err != nil {
				return err
			}
		case fieldMapping.IsNested:
			nestedMapping, ok := reg.GetMapping(field.Type())
			if !ok {
				return fmt.Errorf("%w: %s", ErrStructNotRegistered, field.Type())
			}
			if err := copyStructWithMapping(reg, nestedMapping, field, fieldSrc); err != nil {
				return err
			}
		default:
			dstAddr := unsafe.Pointer(field.UnsafeAddr())
			copyBytes(dstAddr, fieldSrc, fieldMapping.Size)
		}
	}

	return nil
}

func setStringField(mapping *StructMapping, field reflect.Value, cFieldPtr unsafe.Pointer) error {
	converter := mapping.StringConverter
	if converter == nil {
		return fmt.Errorf("cgocopy: missing string converter for struct %s", mapping.GoTypeName)
	}

	cStringPtr := *(*unsafe.Pointer)(cFieldPtr)
	var goStr string
	if cStringPtr != nil {
		goStr = converter.CStringToGo(cStringPtr)
	} else {
		goStr = ""
	}
	field.SetString(goStr)
	return nil
}

func copyArrayField(reg *Registry, field reflect.Value, mappingField FieldMapping, cFieldPtr unsafe.Pointer) error {
	if field.Kind() == reflect.Slice {
		return copySliceField(reg, field, mappingField, cFieldPtr)
	}

	if mappingField.ArrayElemIsNested {
		elemType := field.Type().Elem()
		nestedMapping, ok := reg.GetMapping(elemType)
		if !ok {
			return fmt.Errorf("%w: %s", ErrStructNotRegistered, elemType)
		}

		elemSize := mappingField.ArrayElemSize
		elemCount := int(mappingField.ArrayLen)
		for i := 0; i < elemCount; i++ {
			elemValue := field.Index(i)
			elemSrc := unsafe.Add(cFieldPtr, uintptr(i)*elemSize)
			if err := copyStructWithMapping(reg, nestedMapping, elemValue, elemSrc); err != nil {
				return err
			}
		}
		return nil
	}

	if !field.CanAddr() {
		return fmt.Errorf("cgocopy: array field of type %s is not addressable", field.Type())
	}
	dstAddr := unsafe.Pointer(field.UnsafeAddr())
	copyBytes(dstAddr, cFieldPtr, mappingField.Size)
	return nil
}

func copySliceField(reg *Registry, field reflect.Value, mappingField FieldMapping, cFieldPtr unsafe.Pointer) error {
	length := int(mappingField.ArrayLen)

	if length == 0 {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}

	if field.IsNil() || field.Len() != length {
		field.Set(reflect.MakeSlice(field.Type(), length, length))
	} else if field.Cap() < length {
		field.Set(reflect.MakeSlice(field.Type(), length, length))
	} else {
		field.Set(field.Slice(0, length))
	}

	if mappingField.ArrayElemIsNested {
		elemType := field.Type().Elem()
		nestedMapping, ok := reg.GetMapping(elemType)
		if !ok {
			return fmt.Errorf("%w: %s", ErrStructNotRegistered, elemType)
		}

		elemSize := mappingField.ArrayElemSize
		for i := 0; i < length; i++ {
			elemValue := field.Index(i)
			elemSrc := unsafe.Add(cFieldPtr, uintptr(i)*elemSize)
			if err := copyStructWithMapping(reg, nestedMapping, elemValue, elemSrc); err != nil {
				return err
			}
		}
		return nil
	}

	if length == 0 || mappingField.ArrayElemSize == 0 {
		return nil
	}

	dstSlice := field.Slice(0, length)
	dstPtr := unsafe.Pointer(dstSlice.Index(0).Addr().Pointer())
	srcSlice := unsafe.Slice((*byte)(cFieldPtr), uintptr(length)*mappingField.ArrayElemSize)
	dstBytes := unsafe.Slice((*byte)(dstPtr), uintptr(length)*mappingField.ArrayElemSize)
	copy(dstBytes, srcSlice)
	return nil
}

func copyBytes(dst, src unsafe.Pointer, size uintptr) {
	if size == 0 {
		return
	}
	dstSlice := unsafe.Slice((*byte)(dst), size)
	srcSlice := unsafe.Slice((*byte)(src), size)
	copy(dstSlice, srcSlice)
}

// CopyNoReflection performs the same copy as Copy but avoids per-field reflection
// by relying on the precomputed layout stored in the registry mapping.
func CopyNoReflection[T any](dst *T, cPtr unsafe.Pointer) error {
	if dst == nil {
		return ErrNilDestination
	}
	var zero T
	goType := reflect.TypeOf(zero)
	if goType.Kind() != reflect.Struct {
		return ErrNotAStructType
	}
	return copyNoReflectionGeneric(defaultRegistry, unsafe.Pointer(dst), goType, cPtr)
}

// copyNoReflectionGeneric drives the no-reflection copy using the provided registry.
func copyNoReflectionGeneric(reg *Registry, dstPtr unsafe.Pointer, goType reflect.Type, cPtr unsafe.Pointer) error {
	if reg == nil {
		return ErrNilRegistry
	}
	if !reg.IsFinalized() {
		return ErrRegistryNotFinalized
	}
	if cPtr == nil {
		return ErrNilSourcePointer
	}
	if dstPtr == nil {
		return ErrNilDestination
	}
	if goType.Kind() != reflect.Struct {
		return ErrNotAStructType
	}

	mapping, ok := reg.GetMapping(goType)
	if !ok {
		return ErrStructNotRegistered
	}

	return copyStructNoReflection(reg, mapping, dstPtr, cPtr)
}

func copyStructNoReflection(reg *Registry, mapping *StructMapping, dstPtr, cPtr unsafe.Pointer) error {
	if mapping == nil {
		return fmt.Errorf("cgocopy: missing struct mapping")
	}

	if mapping.CanFastPath {
		copyBytes(dstPtr, cPtr, mapping.GoSize)
		return nil
	}

	for idx := range mapping.Fields {
		fieldMapping := &mapping.Fields[idx]
		fieldSrc := unsafe.Add(cPtr, fieldMapping.COffset)
		fieldDst := unsafe.Add(dstPtr, fieldMapping.GoOffset)

		switch {
		case fieldMapping.IsString:
			if err := setStringFieldNoReflection(mapping, fieldMapping, fieldDst, fieldSrc); err != nil {
				return err
			}
		case fieldMapping.IsArray:
			if err := copyArrayFieldNoReflection(reg, fieldMapping, fieldDst, fieldSrc); err != nil {
				return err
			}
		case fieldMapping.IsNested:
			if fieldMapping.NestedMapping == nil {
				return fmt.Errorf("%w: %s", ErrStructNotRegistered, fieldMapping.GoType)
			}
			if err := copyStructNoReflection(reg, fieldMapping.NestedMapping, fieldDst, fieldSrc); err != nil {
				return err
			}
		default:
			copyBytes(fieldDst, fieldSrc, fieldMapping.Size)
		}
	}

	return nil
}

func setStringFieldNoReflection(mapping *StructMapping, fieldMapping *FieldMapping, fieldDst, cFieldPtr unsafe.Pointer) error {
	converter := mapping.StringConverter
	if converter == nil {
		return fmt.Errorf("cgocopy: missing string converter for struct %s", mapping.GoTypeName)
	}

	cStringPtr := *(*unsafe.Pointer)(cFieldPtr)
	target := (*string)(fieldDst)
	if cStringPtr != nil {
		*target = converter.CStringToGo(cStringPtr)
	} else {
		*target = ""
	}
	return nil
}

func copyArrayFieldNoReflection(reg *Registry, fieldMapping *FieldMapping, fieldDst, cFieldPtr unsafe.Pointer) error {
	if fieldMapping.IsSlice {
		return copySliceFieldNoReflection(reg, fieldMapping, fieldDst, cFieldPtr)
	}

	if fieldMapping.ArrayElemIsNested {
		if fieldMapping.ArrayElemMapping == nil {
			return fmt.Errorf("%w: %s", ErrStructNotRegistered, fieldMapping.ArrayElemGoType)
		}
		elemSize := fieldMapping.ArrayElemSize
		elemCount := int(fieldMapping.ArrayLen)
		for i := 0; i < elemCount; i++ {
			elemDst := unsafe.Add(fieldDst, uintptr(i)*elemSize)
			elemSrc := unsafe.Add(cFieldPtr, uintptr(i)*elemSize)
			if err := copyStructNoReflection(reg, fieldMapping.ArrayElemMapping, elemDst, elemSrc); err != nil {
				return err
			}
		}
		return nil
	}

	copyBytes(fieldDst, cFieldPtr, fieldMapping.Size)
	return nil
}

func copySliceFieldNoReflection(reg *Registry, fieldMapping *FieldMapping, fieldDst, cFieldPtr unsafe.Pointer) error {
	length := int(fieldMapping.ArrayLen)
	sliceValue := reflect.NewAt(fieldMapping.GoType, fieldDst).Elem()

	if length == 0 {
		if sliceValue.IsNil() {
			sliceValue.Set(reflect.MakeSlice(fieldMapping.GoType, 0, 0))
		} else {
			sliceValue.SetLen(0)
		}
		return nil
	}

	if sliceValue.IsNil() || sliceValue.Cap() < length {
		sliceValue.Set(reflect.MakeSlice(fieldMapping.GoType, length, length))
	} else if sliceValue.Len() != length {
		sliceValue.SetLen(length)
	}

	header := (*sliceHeader)(unsafe.Pointer(sliceValue.UnsafeAddr()))
	if fieldMapping.ArrayElemIsNested {
		if fieldMapping.ArrayElemMapping == nil {
			return fmt.Errorf("%w: %s", ErrStructNotRegistered, fieldMapping.ArrayElemGoType)
		}
		elemSize := fieldMapping.ArrayElemSize
		for i := 0; i < length; i++ {
			elemDst := unsafe.Add(header.Data, uintptr(i)*elemSize)
			elemSrc := unsafe.Add(cFieldPtr, uintptr(i)*elemSize)
			if err := copyStructNoReflection(reg, fieldMapping.ArrayElemMapping, elemDst, elemSrc); err != nil {
				return err
			}
		}
		return nil
	}

	total := uintptr(length) * fieldMapping.ArrayElemSize
	if total == 0 {
		return nil
	}

	dstBytes := unsafe.Slice((*byte)(header.Data), total)
	srcBytes := unsafe.Slice((*byte)(cFieldPtr), total)
	copy(dstBytes, srcBytes)
	return nil
}
