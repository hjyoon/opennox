package legacy

/*
#include <stdlib.h>
#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "client__gui__guiquit.h"

extern obj_5D4594_754088_t *ptr_5D4594_754088, *ptr_5D4594_754092;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

//export sub_42EB90
func sub_42EB90(value C.int) C.int {
	*memmap.PtrInt32(0x5D4594, 754052) = int32(value)
	return value
}

//export sub_42EBA0
func sub_42EBA0() C.int {
	return C.int(memmap.Int32(0x5D4594, 754052))
}

//export sub_42EDC0
func sub_42EDC0() {
	C.free(unsafe.Pointer(C.ptr_5D4594_754088))
	C.ptr_5D4594_754088 = nil
	C.free(unsafe.Pointer(C.ptr_5D4594_754092))
	C.ptr_5D4594_754092 = nil
}

//export sub_436550
func sub_436550() C.int {
	blocked := sub_459DA0() != 0 || nox_gui_xxx_check_446360() != 0 || sub_49CB40() != 0 ||
		sub_49C810() != 0 || sub_446950() != 0 || sub_4706A0() != 0 || nox_gui_console_flagXxx_451410() != 0
	frame := gameFrame()
	if !blocked && gameFrame() != 2 {
		return C.int(bool2int(gameFrame()-memmap.Uint32(0x5D4594, 811920) == 1))
	}
	*memmap.PtrUint32(0x5D4594, 811920) = frame
	return 1
}

//export sub_437100
func sub_437100() {
	show := nox_client_getRenderGUI()
	state := memmap.PtrInt32(0x5D4594, 811064)
	if *state == show || noxflags.HasEngine(noxflags.EngineNoRendering) {
		return
	}
	*state = show
	C.sub_4721A0(C.int(show))
	sub_460EA0(C.int(show))
	C.nox_window_set_visible_unk5(C.int(show))
	sub_45D500(C.int(show))
	sub_455A00(C.int(show))
	sub_455F10(C.int(show))
	sub_4706C0(show)
	if show == 0 {
		sub_478000()
	}
}

//export nox_xxx_playerAnimCheck_4372B0
func nox_xxx_playerAnimCheck_4372B0() C.int {
	player := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	if player == nil {
		return 1
	}
	return C.int(bool2int(player.field_69 == 1 || player.field_69 == 2 || player.field_69 == 51))
}
