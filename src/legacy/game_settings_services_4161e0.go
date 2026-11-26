package legacy

/*
#include <stdint.h>
#include <limits.h>
#include "GAME1.h"
*/
import "C"

import "unsafe"

const gameSettingsPlainCharSigned4161E0 = C.CHAR_MIN < 0

func gameSettingsActualDirtyCell4161E0() unsafe.Pointer {
	return unsafe.Pointer(&legacyAddressableWords.dword_5d4594_371692)
}

func gameSettingsDefaultWord4161E0(service int, arg int32) uint32 {
	switch int(int32(service)) {
	case gameSettingsPlayerCount4161E0:
		return uint32(nox_common_playerInfoCount_416F40())
	case gameSettingsPlayerLimit4161E0:
		return uint32(nox_xxx_servGetPlrLimit_409FA0())
	case gameSettingsHeadless4161E0:
		return uint32(bool2int(bool(nox_common_getEngineFlag(C.nox_engine_flag(arg)))))
	case gameSettingsVideo4161E0:
		return uint32(sub_43BE50_get_video_mode_id())
	case gameSettingsFlags4161E0:
		return uint32(nox_common_gameFlags_getVal_40A5B0())
	case gameSettingsGameData4161E0:
		return uint32(int32(nox_xxx_servGamedataGet_40A020(C.short(arg))))
	case gameSettingsModeByte4161E0:
		return uint32(sub_40A180(C.short(arg)))
	case gameSettingsCheckFlag4161E0:
		return uint32(bool2int(bool(nox_common_gameFlags_check_40A5C0(C.uint(arg)))))
	case gameSettingsEquipmentMask4161E0:
		return uint32(sub_453710())
	default:
		return 0
	}
}

func gameSettingsDefaultString4161E0(service int) unsafe.Pointer {
	if int32(service) == gameSettingsServerName4161E0 {
		return unsafe.Pointer(nox_xxx_serverOptionsGetServername_40A4C0())
	}
	return unsafe.Pointer(nox_xxx_mapGetMapName_409B40())
}

func gameSettingsDefaultFill4161E0(service int, out []byte) {
	if int32(service) == gameSettingsSpellMask4161E0 {
		var words [5]C.uint32_t
		sub_454040(&words[0])
		copy(out, unsafe.Slice((*byte)(unsafe.Pointer(&words[0])), unsafe.Sizeof(words)))
		return
	}
	var word C.uint32_t
	sub_4536B0(&word)
	copy(out, unsafe.Slice((*byte)(unsafe.Pointer(&word)), unsafe.Sizeof(word)))
}

func gameSettingsCopy4161E0(dst, src unsafe.Pointer, n uintptr) {
	copy(unsafe.Slice((*byte)(dst), n), unsafe.Slice((*byte)(src), n))
}
