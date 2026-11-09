package legacy

/*
#include "GAME1.h"
*/
import "C"
import "unsafe"

//export nox_access_native_head_416720
func nox_access_native_head_416720(banned C.int) unsafe.Pointer {
	return accessListNativeDefaultHead416720(banned != 0)
}

//export nox_access_allowed_name_4168E0
func nox_access_allowed_name_4168E0(entry *C.nox_access_allowed_entry_t) *C.wchar2_t {
	if entry == nil {
		return nil
	}
	return &entry.name[0]
}

//export nox_access_banned_name_416900
func nox_access_banned_name_416900(entry *C.nox_access_banned_entry_t) *C.wchar2_t {
	if entry == nil {
		return nil
	}
	return &entry.name[0]
}

//export nox_access_banned_serial_416900
func nox_access_banned_serial_416900(entry *C.nox_access_banned_entry_t) *C.char {
	if entry == nil {
		return nil
	}
	return &entry.serial[0]
}

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
