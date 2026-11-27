package legacy

/*
#include "GAME2.h"
#include "GAME2_1.h"
#include "client__gui__guibook.h"

*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const clientAbilityTableOffset = 1047764

type clientAbilityState struct {
	id, field4        uint32
	available, active uint32
	reward, frame     uint32
}

func clientAbilityStates() []clientAbilityState {
	return memmap.PtrT[[6]clientAbilityState](0x5D4594, clientAbilityTableOffset)[1:]
}

//export sub_460EB0
func sub_460EB0(index C.int, flag C.char) {
	if index >= 0 && index < 140 {
		*memmap.PtrUint8(0x5D4594, 1049544+uintptr(index)) = uint8(flag)
	}
}

//export sub_461090
func sub_461090(id, availability C.int) *C.char {
	frame := gameFrame()
	if availability != 0 {
		frame = 0
	}
	states := clientAbilityStates()
	for i := range states {
		state := &states[i]
		if state.id == uint32(id) {
			state.available = uint32(availability)
			state.frame = frame
		}
	}
	return (*C.char)(memmap.PtrOff(0x5D4594, clientAbilityTableOffset+6*24+20))
}

//export sub_4610D0
func sub_4610D0(index C.uchar) *C.char {
	if index != 6 {
		id := memmap.Uint32(0x5D4594, clientAbilityTableOffset+24*uintptr(index))
		return sub_461090(C.int(id), 1)
	}
	var result *C.char
	for _, state := range clientAbilityStates() {
		result = sub_461090(C.int(state.id), 1)
	}
	return result
}

//export sub_461120
func sub_461120(id, active C.int) *C.char {
	mask := uint32(1) << (uint32(id) & 31)
	states := clientAbilityStates()
	for i := range states {
		state := &states[i]
		if state.id != uint32(id) {
			continue
		}
		if active != 0 {
			state.active |= mask
		} else {
			state.active &^= mask
		}
	}
	return (*C.char)(memmap.PtrOff(0x5D4594, clientAbilityTableOffset+6*24+12))
}

//export sub_461160
func sub_461160(id C.int) C.int {
	mask := uint32(1) << (uint32(id) & 31)
	for _, state := range clientAbilityStates() {
		if state.id == uint32(id) {
			return C.int(bool2int(state.active&mask != 0))
		}
	}
	return 0
}

//export sub_4611B0
func sub_4611B0() C.int {
	id := C.int(quickbarControls.target_ability_id)
	if id == 0 {
		return 0
	}
	result := nox_xxx_clientSendAbil_45DAF0(id)
	quickbarControls.target_ability_id, quickbarControls.targeting_ability = 0, 0
	return result
}

//export nox_xxx_netAbilityRewardCli_4611E0
func nox_xxx_netAbilityRewardCli_4611E0(id, value C.int, source *C.char) {
	clientAbilityReward(id, value, source != nil)
}

func clientAbilityReward(id, value C.int, notify bool) {
	if id < 1 || id >= 6 {
		return
	}
	states := clientAbilityStates()
	for i := range states {
		state := &states[i]
		if state.id != uint32(id) || state.reward == uint32(value) {
			continue
		}
		if noxflags.HasGame(noxflags.GameClient) {
			if p := Get_dword_8531A0_2576(); p != nil {
				p.SpellLvl[id] = uint32(value)
			}
		}
		state.reward = uint32(value)
		if value != 0 {
			bit := C.int(bool2int(notify))
			nox_xxx_abilityReward_45D290(id, C.uintptr_t(bit), bit)
		}
	}
}
