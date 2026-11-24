package legacy

/*
#include "GAME1_2.h"
#include "client__gui__window.h"
#include "client__shell__noxworld.h"

extern nox_window* nox_wol_wnd_gameList_815012;
extern nox_window* nox_wol_wnd_world_814980;
extern nox_window* dword_5d4594_815004;
extern nox_window *dword_5d4594_815016, *dword_5d4594_815020, *dword_5d4594_815024;
extern nox_window *dword_5d4594_815028, *dword_5d4594_815032;
extern nox_gui_animation* nox_wnd_xxx_815040;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

//export sub_437320
func sub_437320(record *C.nox_gui_server_ent_t) C.int {
	if record == nil {
		return 0
	}
	row := wolServerNode(unsafe.Pointer(record)).row
	if row == nil {
		return 0
	}
	tier := 0
	for tier < 2 && int32(record.ping) > memmap.Int32(0x587000, 87484+4*uintptr(tier)) {
		tier++
	}
	small := C.int(bool2int(Get_dword_587000_87412() == -1))
	row.draw_data.bg_image = unsafe.Pointer(C.nox_wol_server_icon(small, C.int(tier), 0))
	row.draw_data.hl_image = unsafe.Pointer(C.nox_wol_server_icon(small, C.int(tier), 1))
	row.draw_data.dis_image = row.draw_data.bg_image
	return 1
}

//export sub_4375C0
func sub_4375C0(enabled C.int) {
	if count := Get_nox_wol_server_result_cnt_815088(); count != 0 {
		C.sub_46AD20(C.nox_wol_wnd_world_814980, 10070, C.int(count+10069), enabled)
	}
}

//export sub_437860
func sub_437860(x, y C.int) C.int {
	bounds := memmap.PtrT[[4][4]int16](0x587000, 87528)
	for index, rect := range bounds {
		if C.int(rect[0]) < x && x < C.int(rect[2]) && C.int(rect[1]) < y && y < C.int(rect[3]) {
			return C.int(index)
		}
	}
	return 0
}

//export sub_4379C0
func sub_4379C0() {
	if Get_dword_587000_87408() == 1 {
		nox_window_call_field_94_fnc(C.nox_wol_wnd_gameList_815012, 16399, 0, 0, nil, 0)
	}
}

//export sub_438330
func sub_438330() C.int {
	anim := asGUIAnim(C.nox_wnd_xxx_815040)
	callback := anim.Func13Ptr
	anim.Free()
	if !noxflags.HasGame(noxflags.GameFlag29) {
		C.nox_client_guiXxx_43A9D0()
	}
	if callback != nil {
		ccall.CallIntVoid(callback)
	}
	return 1
}

//export sub_438370
func sub_438370() C.int {
	anim := asGUIAnim(C.nox_wnd_xxx_815040)
	if anim.State() == gui.AnimOutDone {
		return sub_438330()
	}
	anim.SetState(gui.AnimOut)
	sub_43BE40(int32(gui.AnimOut))
	nox_xxx_clientPlaySoundSpecial_452D80(923, 100)
	return 1
}

//export sub_438480
func sub_438480() C.int {
	list, info := C.nox_wol_wnd_gameList_815012, C.dword_5d4594_815004
	if list == nil || info == nil {
		return 0
	}
	gameData := (*gui.ScrollListBoxData)(asWindow(list).WidgetData)
	infoData := (*gui.ScrollListBoxData)(asWindow(info).WidgetData)
	if gameData == nil || infoData == nil {
		return 0
	}
	columns := [...]*nox_window{
		C.dword_5d4594_815016, C.dword_5d4594_815020, C.dword_5d4594_815024,
		C.dword_5d4594_815028, C.dword_5d4594_815032,
	}
	nox_xxx_wndSetProc_46B2C0_go(list, C.sub_439050)
	nox_xxx_wndSetWindowProc_46B300_go(list, C.sub_438EF0)
	for _, column := range columns {
		sub_46B120(column, list)
		nox_xxx_wndSetWindowProc_46B300_go(column, C.sub_438EF0)
	}
	nox_xxx_wndSetID_46B080((*nox_window)(infoData.Field_7), 10035)
	nox_xxx_wndSetID_46B080((*nox_window)(infoData.Field_8), 10036)
	nox_xxx_wndSetID_46B080((*nox_window)(infoData.Field_9), 10032)
	var result C.int
	for i, control := range [...]unsafe.Pointer{gameData.Field_7, gameData.Field_8, gameData.Field_9} {
		for _, column := range columns {
			result = C.int(nox_window_call_field_94_fnc(column, int32(16408+i), uintptr(control), 0, nil, 0))
		}
	}
	return result
}
