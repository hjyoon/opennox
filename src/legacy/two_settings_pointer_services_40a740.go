package legacy

/*
#include "GAME1.h"
#include "GAME3_2.h"
#include <stdint.h>
_Static_assert(sizeof(int)==4, "settings pointer public int width");
typedef char* (*two_settings_pointer_info_fn_40a740)(void);
typedef nox_object_t* (*two_settings_pointer_first_unit_fn_40a740)(void);
typedef nox_object_t* (*two_settings_pointer_next_unit_fn_40a740)(const nox_object_t*);
_Static_assert(_Generic(&sub_4165B0,two_settings_pointer_info_fn_40a740:1,default:0), "Info native char pointer ABI");
_Static_assert(_Generic(&nox_xxx_getFirstPlayerUnit_4DA7C0,two_settings_pointer_first_unit_fn_40a740:1,default:0), "FirstUnit native pointer ABI");
_Static_assert(_Generic(&nox_xxx_getNextPlayerUnit_4DA7F0,two_settings_pointer_next_unit_fn_40a740:1,default:0), "NextUnit native pointer ABI");
static void* two_settings_pointer_info_40a740(void) { return sub_4165B0(); }
static void* two_settings_pointer_first_unit_40a740(void) { return nox_xxx_getFirstPlayerUnit_4DA7C0(); }
static void* two_settings_pointer_next_unit_40a740(void* unit) { return nox_xxx_getNextPlayerUnit_4DA7F0((const nox_object_t*)unit); }
static void* two_settings_pointer_foreign_pointer_40a740(uintptr_t bits) { return (void*)bits; }
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Only full words read from misaligned foreign pointer slots use this
// conversion. An actual Go pointer is not hidden in an integer for cgo.
func twoSettingsPointerNativeForeignPointer40A740(bits uintptr) unsafe.Pointer {
	return C.two_settings_pointer_foreign_pointer_40a740(C.uintptr_t(bits))
}

// The observer intercepts before this function and before any typed helper
// conversion. These are existing production services, not original helper
// bodies reconstructed from test input. No pointer is carried in a DWORD.
func twoSettingsPointerNativeServiceDefault40A740(site uint32, arguments []twoSettingsPointerNativeArg40A740) twoSettingsPointerNativeReply40A740 {
	defer func() { runtime.KeepAlive(arguments) }()
	switch site {
	case 0x40A741:
		return twoSettingsPointerNativeReply40A740{Pointer: C.two_settings_pointer_info_40a740()}
	case 0x40A74D:
		return twoSettingsPointerNativeReply40A740{Word: rawDefaultDamageGameFlagsNative4E0B30(arguments[0].Word)}
	case 0x40A778:
		return twoSettingsPointerNativeReply40A740{Word: rawDefaultDamageGameplayFlagsNative4E0B30(int32(arguments[0].Word))}
	case 0x40A784:
		return twoSettingsPointerNativeReply40A740{Pointer: unsafe.Pointer(nox_server_teamFirst_418B10())}
	case 0x40A79A:
		return twoSettingsPointerNativeReply40A740{Pointer: C.two_settings_pointer_first_unit_40a740()}
	case 0x40A7C9:
		return twoSettingsPointerNativeReply40A740{Pointer: C.two_settings_pointer_next_unit_40a740(arguments[0].Pointer)}
	case 0x40A7D9:
		return twoSettingsPointerNativeReply40A740{Pointer: unsafe.Pointer(nox_server_teamNext_418B60((*nox_team_t)(arguments[0].Pointer)))}
	case 0x40A7EE:
		return twoSettingsPointerNativeReply40A740{Pointer: unsafe.Pointer(nox_common_playerInfoGetFirst_416EA0())}
	case 0x40A819:
		return twoSettingsPointerNativeReply40A740{Pointer: unsafe.Pointer(nox_common_playerInfoGetNext_416EE0((*nox_playerInfo)(arguments[0].Pointer)))}
	default:
		panic("unknown two settings pointer service site")
	}
}
