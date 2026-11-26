package legacy

/*
#include <stdint.h>
#include "GAME4.h"
#include "GAME5_2.h"
*/
import "C"

import "github.com/opennox/opennox/v1/server"

// The legacy caller consumes a scalar protection result in a pointer-sized return slot.
//
//export nox_game4_player_gold_add_go
func nox_game4_player_gold_add_go(unit *C.nox_object_t, delta C.int) C.uintptr_t {
	return C.uintptr_t(game4GoldAdd(asObjectS(unit), uint32(delta)))
}

//export nox_game4_player_gold_sub_go
func nox_game4_player_gold_sub_go(unit *C.nox_object_t, amount C.uint) C.uintptr_t {
	return C.uintptr_t(game4GoldSub(asObjectS(unit), uint32(amount)))
}

func game4GoldAdd(unit *server.Object, amount uint32) uintptr {
	var result uintptr
	playerInventoryAddGoldNative41AC30(unit, amount, func(token uint32, delta int32) {
		result = uintptr(sub_56F920(C.int(token), C.int(delta)))
	})
	return result
}

func game4GoldSub(unit *server.Object, amount uint32) uintptr {
	var result uintptr
	playerInventorySubGoldNative41AC30(unit, amount, func(token uint32, delta int32) {
		result = uintptr(sub_56F920(C.int(token), C.int(delta)))
	})
	return result
}

//export nox_object_setGold_4FA620
func nox_object_setGold_4FA620(unit *C.nox_object_t, delta C.int) {
	Nox_object_setGold_4FA620(asObjectS(unit), int(int32(delta)))
}

//export nox_xxx_playerGetGold_4FA6B0
func nox_xxx_playerGetGold_4FA6B0(unit *C.nox_object_t) C.int {
	return C.int(game4GoldGet(asObjectS(unit)))
}

func game4GoldGet(unit *server.Object) uint32 {
	if unit == nil || unit.UpdateData == nil {
		return 0
	}
	u := (*server.PlayerUpdateData)(unit.UpdateData)
	if u.Player == nil {
		return 0
	}
	return u.Player.GoldVal
}

//export nox_object_getGold_4FA6D0
func nox_object_getGold_4FA6D0(unit *C.nox_object_t) C.int {
	return C.int(int32(objectGetGoldNative4FA6D0(asObjectS(unit))))
}
