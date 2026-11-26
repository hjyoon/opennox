package legacy

/*
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME4_1.h"
extern unsigned int dword_5d4594_527988;
extern uint32_t dword_5d4594_805820;
extern uint32_t nox_xxx_useAudio_587000_80772;
extern uint32_t dword_5d4594_815052;
extern uint32_t dword_5d4594_1047932;
extern uint32_t dword_5d4594_1062552;
#include "client_ui_words.h"
extern uint32_t dword_5d4594_1098624;
extern uint32_t dword_5d4594_1098628;
extern uint32_t dword_5d4594_1123520;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Each getter keeps the original signed 32-bit C ABI and reads the existing shared word.
//
//export sub_41E2F0
func sub_41E2F0() C.int { return C.int(C.dword_5d4594_527988) }

//export nox_client_mousePriKey_430AF0
func nox_client_mousePriKey_430AF0() C.int { return C.int(C.dword_5d4594_805820) }

//export nox_xxx_cursor_430B00
func nox_xxx_cursor_430B00() C.int { return C.int(C.nox_xxx_useAudio_587000_80772) }

//export sub_43AF30
func sub_43AF30() C.int { return C.int(C.dword_5d4594_815052) }

//export sub_43AF40
func sub_43AF40() C.int { return C.int(C.nox_game_createOrJoin_815048) }

//export sub_43AF80
func sub_43AF80() C.int { return C.int(C.dword_5d4594_814548) }

//export sub_4611A0
func sub_4611A0() C.int { return C.int(C.dword_5d4594_1047932) }

//export sub_4674A0
func sub_4674A0() C.int { return C.int(C.dword_5d4594_1062552) }

func Nox_client_gold_4674A0() uint32 { return uint32(sub_4674A0()) }

//export sub_4675B0
func sub_4675B0() C.int { return C.int(inventoryState.mode) }

//export sub_46A4A0
func sub_46A4A0() C.int { return C.int(*clientChatActive) }

//export nox_xxx_playerGet_470A90
func nox_xxx_playerGet_470A90() C.int { return C.int(meterState.barsVisible) }

//export sub_478030
func sub_478030() C.int { return C.int(C.dword_5d4594_1098624) }

//export sub_479590
func sub_479590() C.int { return C.int(C.dword_5d4594_1098628) }

//export sub_47A260
func sub_47A260() C.int { return C.int(C.dword_5d4594_1123520) }

//export sub_50FD20
func sub_50FD20(ptr unsafe.Pointer, owner *C.nox_object_t) C.int {
	return C.int(game4TradeOfferValue50FD20((*server.TradeSession)(ptr), asObjectS((*nox_object_t)(owner))))
}
