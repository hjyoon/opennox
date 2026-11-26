package legacy

/*
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME4_1.h"
#include "client_ui_words.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

var clientUIWords = func() *C.nox_client_ui_words_t {
	state, _ := alloc.New(C.nox_client_ui_words_t{})
	state.audio_mask = 5
	state.render_bubbles = 1
	return state
}()

//export nox_client_ui_words_storage
func nox_client_ui_words_storage() *C.nox_client_ui_words_t { return clientUIWords }

// Each getter keeps the original signed 32-bit C ABI and reads the existing shared word.
//
//export sub_41E2F0
func sub_41E2F0() C.int { return C.int(clientUIWords.wol_status) }

//export nox_client_mousePriKey_430AF0
func nox_client_mousePriKey_430AF0() C.int { return C.int(clientUIWords.mouse_primary) }

//export nox_xxx_cursor_430B00
func nox_xxx_cursor_430B00() C.int { return C.int(clientUIWords.audio_mask) }

//export sub_43AF30
func sub_43AF30() C.int { return C.int(clientUIWords.wol_game_data_ready) }

//export sub_43AF40
func sub_43AF40() C.int { return C.int(clientUIWords.create_or_join) }

//export sub_43AF80
func sub_43AF80() C.int { return C.int(clientUIWords.network_state) }

//export sub_4611A0
func sub_4611A0() C.int { return C.int(quickbarControls.targeting_ability) }

//export sub_4674A0
func sub_4674A0() C.int { return C.int(clientUIWords.player_gold) }

func Nox_client_gold_4674A0() uint32 { return uint32(sub_4674A0()) }

//export sub_4675B0
func sub_4675B0() C.int { return C.int(inventoryState.mode) }

//export sub_46A4A0
func sub_46A4A0() C.int { return C.int(*clientChatActive) }

//export nox_xxx_playerGet_470A90
func nox_xxx_playerGet_470A90() C.int { return C.int(meterState.barsVisible) }

//export sub_478030
func sub_478030() C.int { return C.int(clientUIWords.shop_active) }

//export sub_479590
func sub_479590() C.int { return C.int(clientUIWords.shop_mode) }

//export sub_47A260
func sub_47A260() C.int { return C.int(clientUIWords.dialog_active) }

//export sub_50FD20
func sub_50FD20(ptr unsafe.Pointer, owner *C.nox_object_t) C.int {
	return C.int(game4TradeOfferValue50FD20((*server.TradeSession)(ptr), asObjectS((*nox_object_t)(owner))))
}
