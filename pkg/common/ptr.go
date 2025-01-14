package common

import "reflect"

// Ptr returns the pointer to the value passed in.
func Ptr[T any](v T) *T {
	return &v
}

// Val returns the value pointed to by the pointer passed in.
// If the pointer is nil, it returns the zero value of the type.
func Val[T any](v *T) T {
	if v == nil {
		return reflect.Zero(reflect.TypeOf(v).Elem()).Interface().(T)
	}
	return *v
}
