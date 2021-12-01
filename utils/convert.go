package utils

import (
	"reflect"
	"unsafe"
)

// AsString returns a string that refers to the data backing the slice s.
func AsString(b []byte) string {
	p := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data)

	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(p)
	hdr.Len = len(b)

	return s
}

// Uint64 returns a pointer value for the uint64 value passed in.
func Uint64(v uint64) *uint64 {
	return &v
}
