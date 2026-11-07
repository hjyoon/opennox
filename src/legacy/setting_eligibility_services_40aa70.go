package legacy

/*
#include "GAME1.h"
#include <stdint.h>
_Static_assert(sizeof(int)==4, "setting eligibility public int width");
typedef int (*setting_eligibility_entry_fn_40aa70)(nox_playerInfo*);
typedef void* (*setting_eligibility_info_fn_40aa70)(void);
typedef char (*setting_eligibility_char_fn_40aa70)(void);
typedef int (*setting_eligibility_limit_fn_40aa70)(void);
_Static_assert(_Generic(&sub_40AA70,setting_eligibility_entry_fn_40aa70:1,default:0), "setting eligibility public typed pointer ABI");
_Static_assert(_Generic(&sub_416640,setting_eligibility_info_fn_40aa70:1,default:0), "Info416640 native void pointer ABI");
_Static_assert(_Generic(&sub_417DE0,setting_eligibility_char_fn_40aa70:1,default:0), "count417DE0 char ABI");
_Static_assert(_Generic(&sub_417DC0,setting_eligibility_limit_fn_40aa70:1,default:0), "limit417DC0 int ABI");
static void* setting_eligibility_info_40aa70(void) { return sub_416640(); }
// Only this actual product char fallback canonicalizes to low AL. Diagnostic
// raw replies bypass this helper and are masked by the root at each site.
static uint32_t setting_eligibility_char_AL_40aa70(void) { return (uint32_t)(uint8_t)(unsigned char)sub_417DE0(); }
static uint32_t setting_eligibility_limit_40aa70(void) { return (uint32_t)sub_417DC0(); }
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Defaults use the actual existing product services. Observation occurs before
// this function, before GetServer or any typed helper conversion/admission.
func settingEligibilityNativeServiceDefault40AA70(site uint32, arguments []settingEligibilityNativeArg40AA70) settingEligibilityNativeReply40AA70 {
	defer func() { runtime.KeepAlive(arguments) }()
	switch site {
	case 0x40AA72:
		return settingEligibilityNativeReply40AA70{Pointer: C.setting_eligibility_info_40aa70()}
	case 0x40AA8A, 0x40AACF, 0x40AB02, 0x40AB10, 0x40AB49, 0x40AB77, 0x40AB88:
		return settingEligibilityNativeReply40AA70{Word: rawDefaultDamageGameFlagsNative4E0B30(arguments[0].Word)}
	case 0x40AA96:
		return settingEligibilityNativeReply40AA70{Word: uint32(sub40A770Native())}
	case 0x40AAC1:
		return settingEligibilityNativeReply40AA70{Word: uint32(sub40A740Native())}
	case 0x40AAED:
		// Preserve the existing raw int32 ABI and its helper-domain limit.
		// Its high-bit key lookup issue is a separate OPEN helper port.
		return settingEligibilityNativeReply40AA70{Pointer: unsafe.Pointer(nox_server_teamByXxx_418AE0(int32(arguments[0].Word)))}
	case 0x40AB1E:
		return settingEligibilityNativeReply40AA70{Word: rawDefaultDamageGameplayFlagsNative4E0B30(int32(arguments[0].Word))}
	case 0x40AB34, 0x40AB56:
		return settingEligibilityNativeReply40AA70{Word: uint32(C.setting_eligibility_char_AL_40aa70())}
	case 0x40AB63:
		return settingEligibilityNativeReply40AA70{Word: uint32(C.setting_eligibility_limit_40aa70())}
	case 0x40AB94:
		return settingEligibilityNativeReply40AA70{Pointer: unsafe.Pointer(nox_common_playerInfoGetFirst_416EA0())}
	case 0x40ABA8:
		return settingEligibilityNativeReply40AA70{Pointer: unsafe.Pointer(nox_common_playerInfoGetNext_416EE0((*nox_playerInfo)(arguments[0].Pointer)))}
	case 0x40ABBC:
		return settingEligibilityNativeReply40AA70{Word: uint32(sub40AA00Native())}
	default:
		panic("unknown setting eligibility service site")
	}
}
