package legacy

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
		return settingEligibilityNativeReply40AA70{Pointer: sub416640Native()}
	case 0x40AA8A, 0x40AACF, 0x40AB02, 0x40AB10, 0x40AB49, 0x40AB77, 0x40AB88:
		return settingEligibilityNativeReply40AA70{Word: rawDefaultDamageGameFlagsNative4E0B30(arguments[0].Word)}
	case 0x40AA96:
		return settingEligibilityNativeReply40AA70{Word: uint32(sub40A770Native())}
	case 0x40AAC1:
		return settingEligibilityNativeReply40AA70{Word: uint32(sub40A740Native())}
	case 0x40AAED:
		return settingEligibilityNativeReply40AA70{Pointer: unsafe.Pointer(nox_server_teamByXxx_418AE0(int32(arguments[0].Word)))}
	case 0x40AB1E:
		return settingEligibilityNativeReply40AA70{Word: rawDefaultDamageGameplayFlagsNative4E0B30(int32(arguments[0].Word))}
	case 0x40AB34, 0x40AB56:
		return settingEligibilityNativeReply40AA70{Word: uint32(sub417DE0Native())}
	case 0x40AB63:
		return settingEligibilityNativeReply40AA70{Word: uint32(sub417DC0Native())}
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
