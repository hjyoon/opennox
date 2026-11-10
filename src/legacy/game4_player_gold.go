package legacy

/*
#include <stdint.h>
#include "GAME4.h"
#include "GAME5_2.h"
// Convert the mixed CRC/pointer result before it crosses into Go. A small
// opaque token must never become a pointer-shaped Go GC root.
static uintptr_t game4_protection_delta_word(uint32_t token, int32_t delta) {
	return (uintptr_t)sub_56F920((int32_t)token, delta);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

// These return uintptr because protection may return either an opaque CRC word
// or an actual native allocation. Only the C ABI adapter casts it to uint32_t*.
//
//export nox_game4_player_gold_add_go
func nox_game4_player_gold_add_go(unit *C.nox_object_t, delta C.int) C.uintptr_t {
	var result C.uintptr_t
	_ = playerInventoryAddGoldNative41AC30(asObjectS(unit), uint32(delta), func(token uint32, amount int32) {
		result = C.game4_protection_delta_word(C.uint32_t(token), C.int32_t(amount))
	})
	return result
}

//export nox_game4_player_gold_sub_go
func nox_game4_player_gold_sub_go(unit *C.nox_object_t, amount C.uint) C.uintptr_t {
	var result C.uintptr_t
	_ = playerInventorySubGoldNative41AC30(asObjectS(unit), uint32(amount), func(token uint32, delta int32) {
		result = C.game4_protection_delta_word(C.uint32_t(token), C.int32_t(delta))
	})
	return result
}

//export nox_object_setGold_4FA620
func nox_object_setGold_4FA620(unit *C.nox_object_t, delta C.int) {
	Nox_object_setGold_4FA620(asObjectS(unit), int(int32(delta)))
}

//export nox_xxx_playerGetGold_4FA6B0
func nox_xxx_playerGetGold_4FA6B0(unit *C.nox_object_t) C.int {
	o := asObjectS(unit)
	if o == nil || o.UpdateData == nil {
		return 0
	}
	u := (*server.PlayerUpdateData)(o.UpdateData)
	if u.Player == nil {
		return 0
	}
	return C.int(int32(u.Player.GoldVal))
}

//export nox_object_getGold_4FA6D0
func nox_object_getGold_4FA6D0(unit *C.nox_object_t) C.int {
	return C.int(int32(objectGetGoldNative4FA6D0(asObjectS(unit))))
}
