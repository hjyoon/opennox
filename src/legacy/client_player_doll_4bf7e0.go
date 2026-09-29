package legacy

/*
#include "GAME3_1.h"
*/
import "C"

import "unsafe"

type clientPlayerDollState4BF7E0 struct {
	armorMask     uint32
	weaponMask    uint32
	colorSkin     uint32
	colorHair     uint32
	colorMustache uint32
	colorGoatee   uint32
	colorBeard    uint32
	colorUnknown  uint32
	variant       byte
}

func clientPlayerDollStateNative4BF7E0() (clientPlayerDollState4BF7E0, bool) {
	var state C.nox_player_doll_state_t
	if C.nox_client_getPlayerDollState_4BF7E0(&state) == 0 {
		return clientPlayerDollState4BF7E0{}, false
	}
	return clientPlayerDollState4BF7E0{
		armorMask:     uint32(state.armor_mask),
		weaponMask:    uint32(state.weapon_mask),
		colorSkin:     uint32(state.color_skin),
		colorHair:     uint32(state.color_hair),
		colorMustache: uint32(state.color_mustache),
		colorGoatee:   uint32(state.color_goatee),
		colorBeard:    uint32(state.color_beard),
		colorUnknown:  uint32(state.color_unknown),
		variant:       byte(state.variant),
	}, true
}

func clientPlayerDollImageNative4BF9F0(tableOffset uintptr, layer int) unsafe.Pointer {
	return unsafe.Pointer(C.nox_client_getPlayerDollImage_4BF9F0(C.uintptr_t(tableOffset), C.int(layer)))
}
