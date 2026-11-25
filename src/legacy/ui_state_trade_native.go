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
extern unsigned int nox_game_createOrJoin_815048;
extern unsigned int dword_5d4594_814548;
extern uint32_t dword_5d4594_1047932;
extern uint32_t dword_5d4594_1062552;
extern uint32_t dword_5d4594_1049864;
#include "client_ui_words.h"
extern uint32_t dword_5d4594_1096252;
extern uint32_t dword_5d4594_1098624;
extern uint32_t dword_5d4594_1098628;
extern uint32_t dword_5d4594_1123520;
static void* ui_state_native_word_address(unsigned i) { switch(i) {
case 0: return &dword_5d4594_527988;
case 1: return &dword_5d4594_805820;
case 2: return &nox_xxx_useAudio_587000_80772;
case 3: return &dword_5d4594_815052;
case 4: return &nox_game_createOrJoin_815048;
case 5: return &dword_5d4594_814548;
case 6: return &dword_5d4594_1047932;
case 7: return &dword_5d4594_1062552;
case 8: return &dword_5d4594_1049864;
case 9: return &dword_5d4594_1064868;
case 10: return &dword_5d4594_1096252;
case 11: return &dword_5d4594_1098624;
case 12: return &dword_5d4594_1098628;
case 13: return &dword_5d4594_1123520;
default: return 0; }}
*/
import "C"

import (
	"github.com/opennox/opennox/v1/server"
	"unsafe"
)

// C supplies addresses only; the authoritative32bit load stays in Go.
func uiStateSharedWord(index int) uint32 {
	return *(*uint32)(C.ui_state_native_word_address(C.uint(index)))
}

// Each getter keeps the original signed32 C ABI and reads the existing shared word.
//
//export sub_41E2F0
func sub_41E2F0() C.int { return C.int(uiStateSharedWord(0)) }

//export nox_client_mousePriKey_430AF0
func nox_client_mousePriKey_430AF0() C.int { return C.int(uiStateSharedWord(1)) }

//export nox_xxx_cursor_430B00
func nox_xxx_cursor_430B00() C.int { return C.int(uiStateSharedWord(2)) }

//export sub_43AF30
func sub_43AF30() C.int { return C.int(uiStateSharedWord(3)) }

//export sub_43AF40
func sub_43AF40() C.int { return C.int(uiStateSharedWord(4)) }

//export sub_43AF80
func sub_43AF80() C.int { return C.int(uiStateSharedWord(5)) }

//export sub_4611A0
func sub_4611A0() C.int { return C.int(uiStateSharedWord(6)) }

//export sub_4674A0
func sub_4674A0() C.int { return C.int(uiStateSharedWord(7)) }

func Nox_client_gold_4674A0() uint32 { return uint32(sub_4674A0()) }

//export sub_4675B0
func sub_4675B0() C.int { return C.int(uiStateSharedWord(8)) }

//export sub_46A4A0
func sub_46A4A0() C.int { return C.int(uiStateSharedWord(9)) }

//export nox_xxx_playerGet_470A90
func nox_xxx_playerGet_470A90() C.int { return C.int(uiStateSharedWord(10)) }

//export sub_478030
func sub_478030() C.int { return C.int(uiStateSharedWord(11)) }

//export sub_479590
func sub_479590() C.int { return C.int(uiStateSharedWord(12)) }

//export sub_47A260
func sub_47A260() C.int { return C.int(uiStateSharedWord(13)) }

//export sub_50FD20
func sub_50FD20(ptr unsafe.Pointer, owner *C.nox_object_t) C.int {
	return C.int(game4TradeOfferValue50FD20((*server.TradeSession)(ptr), asObjectS((*nox_object_t)(owner))))
}
