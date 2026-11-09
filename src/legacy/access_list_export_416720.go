package legacy

/*
#include "GAME1.h"
*/
import "C"
import "unsafe"

//export sub_416720
func sub_416720() { sub416720Native() }

//export nox_native_access_add_banned_416770
func nox_native_access_add_banned_416770(minutes C.int, name *C.wchar2_t, serial unsafe.Pointer) *C.int {
	return (*C.int)(sub416770Native(int32(minutes), unsafe.Pointer(name), serial))
}

//export sub_416820
func sub_416820(index C.int) { sub416820Native(int32(index)) }

//export sub_416860
func sub_416860(index C.int) *C.int { return (*C.int)(sub416860Native(int32(index))) }

//export sub_4168A0
func sub_4168A0(name *C.wchar2_t) *C.int { return (*C.int)(sub4168A0Native(unsafe.Pointer(name))) }

//export sub_4168E0
func sub_4168E0() *C.nox_access_allowed_entry_t {
	return (*C.nox_access_allowed_entry_t)(sub4168E0Native())
}

//export sub_4168F0
func sub_4168F0(p *C.nox_access_allowed_entry_t) *C.nox_access_allowed_entry_t {
	return (*C.nox_access_allowed_entry_t)(sub4168F0Native(unsafe.Pointer(p)))
}

//export sub_416900
func sub_416900() *C.nox_access_banned_entry_t {
	return (*C.nox_access_banned_entry_t)(sub416900Native())
}

//export sub_416910
func sub_416910(p *C.nox_access_banned_entry_t) *C.nox_access_banned_entry_t {
	return (*C.nox_access_banned_entry_t)(sub416910Native(unsafe.Pointer(p)))
}

//export sub_416920
func sub_416920() C.int { return C.int(sub416920Native()) }

//export sub_416950
func sub_416950() *C.int { return (*C.int)(sub416950Native()) }
