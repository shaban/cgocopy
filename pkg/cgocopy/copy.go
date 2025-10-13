package cgocopy2

import (
	"reflect"
	"unsafe"
)

// Copy creates a Go struct of type T by copying data from a C struct pointer.
// The type T must have been precompiled using Precompile[T]() before calling Copy.
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
//
//	func main() {
//	    cUser := getCUser() // Returns C.User pointer
//	    user := cgocopy2.Copy[User](unsafe.Pointer(cUser))
//	    fmt.Printf("User: %+v\n", user)
//	}
func Copy[T any](cPtr unsafe.Pointer) (T, error) {
	var zero T

	// Check for nil pointer
	if cPtr == nil {
		return zero, ErrNilPointer
	}

	// Get type and check registration
	goType := reflect.TypeOf(zero)
	if goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	metadata := globalRegistry.Get(goType)
	if metadata == nil {
		return zero, newCopyError(goType, "", "type not registered", ErrNotRegistered)
	}

	// Create a new instance
	result := reflect.New(goType).Elem()

	// Copy each field
	for i := range metadata.Fields {
		field := &metadata.Fields[i]
		if field.Skip {
			continue
		}

		// Get field value in result struct
		resultField := result.Field(field.Index)

		// Calculate C field pointer (assuming same layout for now)
		cFieldPtr := unsafe.Pointer(uintptr(cPtr) + field.Offset)

		// Copy based on field type
		if err := copyField(resultField, cFieldPtr, field); err != nil {
			return zero, newCopyError(goType, field.Name, "failed to copy field", err)
		}
	}

	return result.Interface().(T), nil
}

// copyField copies a single field from C to Go based on its type.
func copyField(goField reflect.Value, cPtr unsafe.Pointer, field *FieldInfo) error {
	if !goField.CanSet() {
		return ErrInvalidType
	}

	switch field.Type {
	case FieldTypePrimitive:
		return copyPrimitive(goField, cPtr, field.ReflectType)

	case FieldTypeString:
		return copyString(goField, cPtr)

	case FieldTypeStruct:
		return copyStruct(goField, cPtr, field.ReflectType)

	case FieldTypeArray:
		return copyArray(goField, cPtr, field)

	case FieldTypeSlice:
		return copySlice(goField, cPtr, field)

	case FieldTypePointer:
		return copyPointer(goField, cPtr, field)

	default:
		return ErrUnsupportedType
	}
}

// copyPrimitive copies a primitive field (int, float, bool, etc).
func copyPrimitive(goField reflect.Value, cPtr unsafe.Pointer, fieldType reflect.Type) error {
	switch fieldType.Kind() {
	case reflect.Int:
		goField.SetInt(int64(*(*int)(cPtr)))
	case reflect.Int8:
		goField.SetInt(int64(*(*int8)(cPtr)))
	case reflect.Int16:
		goField.SetInt(int64(*(*int16)(cPtr)))
	case reflect.Int32:
		goField.SetInt(int64(*(*int32)(cPtr)))
	case reflect.Int64:
		goField.SetInt(*(*int64)(cPtr))

	case reflect.Uint:
		goField.SetUint(uint64(*(*uint)(cPtr)))
	case reflect.Uint8:
		goField.SetUint(uint64(*(*uint8)(cPtr)))
	case reflect.Uint16:
		goField.SetUint(uint64(*(*uint16)(cPtr)))
	case reflect.Uint32:
		goField.SetUint(uint64(*(*uint32)(cPtr)))
	case reflect.Uint64:
		goField.SetUint(*(*uint64)(cPtr))

	case reflect.Float32:
		goField.SetFloat(float64(*(*float32)(cPtr)))
	case reflect.Float64:
		goField.SetFloat(*(*float64)(cPtr))

	case reflect.Bool:
		// Assume C bool is represented as uint8 (0 or 1)
		goField.SetBool(*(*uint8)(cPtr) != 0)

	default:
		return ErrUnsupportedType
	}

	return nil
}

// copyString copies a C string (char*) to a Go string.
// This assumes the C struct contains a char* pointer.
func copyString(goField reflect.Value, cPtr unsafe.Pointer) error {
	// Read the char* pointer from the C struct
	charPtr := *(*unsafe.Pointer)(cPtr)
	
	if charPtr == nil {
		goField.SetString("")
		return nil
	}

	// Convert C string to Go string
	// Walk the C string to find its length
	length := 0
	for {
		b := *(*byte)(unsafe.Pointer(uintptr(charPtr) + uintptr(length)))
		if b == 0 {
			break
		}
		length++
	}

	// Create a Go string from the C bytes
	if length == 0 {
		goField.SetString("")
		return nil
	}

	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = *(*byte)(unsafe.Pointer(uintptr(charPtr) + uintptr(i)))
	}

	goField.SetString(string(bytes))
	return nil
}

// copyStruct copies a nested struct field.
func copyStruct(goField reflect.Value, cPtr unsafe.Pointer, fieldType reflect.Type) error {
	// Check if the nested struct is registered
	metadata := globalRegistry.Get(fieldType)
	if metadata == nil {
		return ErrNotRegistered
	}

	// Create a new instance of the nested struct
	nestedStruct := reflect.New(fieldType).Elem()

	// Copy each field of the nested struct
	for i := range metadata.Fields {
		field := &metadata.Fields[i]
		if field.Skip {
			continue
		}

		nestedField := nestedStruct.Field(field.Index)
		nestedCPtr := unsafe.Pointer(uintptr(cPtr) + field.Offset)

		if err := copyField(nestedField, nestedCPtr, field); err != nil {
			return err
		}
	}

	goField.Set(nestedStruct)
	return nil
}

// copyArray copies a fixed-size array field.
func copyArray(goField reflect.Value, cPtr unsafe.Pointer, field *FieldInfo) error {
	elemSize := field.ElemType.Size()
	elemKind := field.ElemType.Kind()

	// For primitive arrays, we can do bulk copy
	if isPrimitiveKind(elemKind) {
		for i := 0; i < field.ArrayLen; i++ {
			elemPtr := unsafe.Pointer(uintptr(cPtr) + uintptr(i)*elemSize)
			elemField := goField.Index(i)
			if err := copyPrimitive(elemField, elemPtr, field.ElemType); err != nil {
				return err
			}
		}
		return nil
	}

	// For string arrays
	if elemKind == reflect.String {
		for i := 0; i < field.ArrayLen; i++ {
			elemPtr := unsafe.Pointer(uintptr(cPtr) + uintptr(i)*elemSize)
			elemField := goField.Index(i)
			if err := copyString(elemField, elemPtr); err != nil {
				return err
			}
		}
		return nil
	}

	// For struct arrays
	if elemKind == reflect.Struct {
		for i := 0; i < field.ArrayLen; i++ {
			elemPtr := unsafe.Pointer(uintptr(cPtr) + uintptr(i)*elemSize)
			elemField := goField.Index(i)
			if err := copyStruct(elemField, elemPtr, field.ElemType); err != nil {
				return err
			}
		}
		return nil
	}

	return ErrUnsupportedType
}

// copySlice copies a slice field.
// Note: This assumes the C struct has a slice-like structure with pointer + length.
// The exact layout depends on how slices are represented in C.
func copySlice(goField reflect.Value, cPtr unsafe.Pointer, field *FieldInfo) error {
	// For now, we'll assume C slices are represented as:
	// struct { void* data; size_t len; }
	// This is a common pattern in C for representing Go slices.
	
	type cSlice struct {
		data unsafe.Pointer
		len  int
	}

	slice := (*cSlice)(cPtr)
	
	if slice.data == nil || slice.len == 0 {
		goField.Set(reflect.MakeSlice(field.ReflectType, 0, 0))
		return nil
	}

	// Create a new Go slice
	newSlice := reflect.MakeSlice(field.ReflectType, slice.len, slice.len)
	elemSize := field.ElemType.Size()
	elemKind := field.ElemType.Kind()

	// Copy elements based on type
	if isPrimitiveKind(elemKind) {
		for i := 0; i < slice.len; i++ {
			elemPtr := unsafe.Pointer(uintptr(slice.data) + uintptr(i)*elemSize)
			elemField := newSlice.Index(i)
			if err := copyPrimitive(elemField, elemPtr, field.ElemType); err != nil {
				return err
			}
		}
	} else if elemKind == reflect.String {
		for i := 0; i < slice.len; i++ {
			elemPtr := unsafe.Pointer(uintptr(slice.data) + uintptr(i)*elemSize)
			elemField := newSlice.Index(i)
			if err := copyString(elemField, elemPtr); err != nil {
				return err
			}
		}
	} else {
		return ErrUnsupportedType
	}

	goField.Set(newSlice)
	return nil
}

// copyPointer copies a pointer field.
func copyPointer(goField reflect.Value, cPtr unsafe.Pointer, field *FieldInfo) error {
	// Read the pointer value from C
	ptrValue := *(*unsafe.Pointer)(cPtr)
	
	if ptrValue == nil {
		goField.Set(reflect.Zero(field.ReflectType))
		return nil
	}

	// Create a new instance of the pointed-to type
	elemType := field.ElemType
	newElem := reflect.New(elemType)

	// If it's a primitive or string, copy directly
	if isPrimitiveKind(elemType.Kind()) {
		if err := copyPrimitive(newElem.Elem(), ptrValue, elemType); err != nil {
			return err
		}
	} else if elemType.Kind() == reflect.String {
		if err := copyString(newElem.Elem(), ptrValue); err != nil {
			return err
		}
	} else if elemType.Kind() == reflect.Struct {
		if err := copyStruct(newElem.Elem(), ptrValue, elemType); err != nil {
			return err
		}
	} else {
		return ErrUnsupportedType
	}

	goField.Set(newElem)
	return nil
}
