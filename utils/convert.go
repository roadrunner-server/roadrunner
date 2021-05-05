package utils

import (
	"reflect"
	"unsafe"
)

// AsBytes returns a slice that refers to the data backing the string s.
func AsBytes(s string) []byte {
	// get the pointer to the data of the string
	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)

	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Data = uintptr(p)
	// we need to set the cap and len for the string to byte convert
	// because string is shorter than []bytes
	hdr.Cap = len(s)
	hdr.Len = len(s)

	return b
}

// AsString returns a string that refers to the data backing the slice s.
func AsString(b []byte) string {
	p := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data)

	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(p)
	hdr.Len = len(b)

	return s
}
