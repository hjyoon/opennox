package legacy

/*
#include "GAME1.h"
#include "common__object__weaplook.h"
#include "common__object__armrlook.h"
#include <stdint.h>
typedef int (*tt_title_to_mask_415960)(wchar2_t*);
typedef wchar2_t* (*tt_mask_to_title_415960)(int);
typedef char** (*tt_load_415960)(void);
_Static_assert(sizeof(int)==4,"title masks remain DWORD-sized");
_Static_assert(_Generic(&sub_415960,tt_title_to_mask_415960:1,default:0),"weapon input native pointer");
_Static_assert(_Generic(&sub_4159F0,tt_mask_to_title_415960:1,default:0),"weapon output native pointer");
_Static_assert(_Generic(&sub_415DA0,tt_title_to_mask_415960:1,default:0),"armor input native pointer");
_Static_assert(_Generic(&sub_415E80,tt_mask_to_title_415960:1,default:0),"armor output native pointer");
_Static_assert(_Generic(&nox_xxx_loadModifyers_4158C0,tt_load_415960:1,default:0),"legacy weapon loader ABI unchanged");
_Static_assert(_Generic(&nox_xxx_loadLook_415D50,tt_load_415960:1,default:0),"legacy armor loader ABI unchanged");
static uint32_t tt_fixture_to_mask_415960(int armor,wchar2_t* p) {
    return (uint32_t)(armor?sub_415DA0(p):sub_415960(p));
}
static wchar2_t* tt_fixture_from_mask_415960(int armor,uint32_t mask) {
    return armor?sub_415E80((int)mask):sub_4159F0((int)mask);
}
static uintptr_t tt_fixture_load_415960(int armor) {
    return (uintptr_t)(armor?nox_xxx_loadLook_415D50():nox_xxx_loadModifyers_4158C0());
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func titleFixtureToMask415960(armor bool, query unsafe.Pointer) uint32 {
	a := C.int(0)
	if armor {
		a = 1
	}
	v := uint32(C.tt_fixture_to_mask_415960(a, (*C.wchar2_t)(query)))
	runtime.KeepAlive(query)
	return v
}

func titleFixtureFromMask415960(armor bool, mask uint32) unsafe.Pointer {
	a := C.int(0)
	if armor {
		a = 1
	}
	return unsafe.Pointer(C.tt_fixture_from_mask_415960(a, C.uint32_t(mask)))
}

func titleFixtureLoad415960(armor bool) uintptr {
	a := C.int(0)
	if armor {
		a = 1
	}
	return uintptr(C.tt_fixture_load_415960(a))
}
