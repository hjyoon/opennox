package legacy

/*
#include "GAME1.h"
#include <stdlib.h>
static void* access_list_fixture_call_416720(int entry,int index,void* name,void* serial,int* word) {
    switch(entry) {
    case 0x416720:sub_416720();return 0;
    case 0x416770:return sub_416770(index,(wchar2_t*)name,(const char*)serial);
    case 0x416820:sub_416820(index);return 0;
    case 0x416860:return sub_416860(index);
    case 0x4168a0:return sub_4168A0((wchar2_t*)name);
    case 0x4168e0:return sub_4168E0();
    case 0x4168f0:return sub_4168F0((nox_access_allowed_entry_t*)name);
    case 0x416900:return sub_416900();
    case 0x416910:return sub_416910((nox_access_banned_entry_t*)name);
    case 0x416920:*word=sub_416920();return 0;
    case 0x416950:return sub_416950();
    default:abort();
    }
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func accessListFixtureCall416720(entry uint32, index int32, name, serial unsafe.Pointer) (unsafe.Pointer, uint32) {
	var word C.int
	p := C.access_list_fixture_call_416720(C.int(entry), C.int(index), name, serial, &word)
	runtime.KeepAlive(name)
	runtime.KeepAlive(serial)
	return p, uint32(word)
}
