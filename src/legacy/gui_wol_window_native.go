package legacy

/*
#include "GAME1_2.h"
#include "GAME2_2.h"
#include "client__gui__window.h"
#include "client__shell__noxworld.h"

extern nox_gui_animation* nox_wnd_xxx_815040;
extern uint32_t dword_5d4594_815052;
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

type wolWorldWindows struct {
	root             *nox_window
	mapPanel         *nox_window
	regionPanel      *nox_window
	filter           *nox_window
	heading          *nox_window
	serverInfo       *nox_window
	serverDetails    *nox_window
	serverInfoButton *nox_window
	gameList         *nox_window
	gameColumns      [5]*nox_window
}

var wolWindows wolWorldWindows

var wolServerIcons [2][3][2]*C.nox_video_bag_image_t
var wolMapIcons [4]*C.nox_video_bag_image_t

//export nox_wol_server_icon
func nox_wol_server_icon(small, tier, lit C.int) *C.nox_video_bag_image_t {
	if small < 0 || small >= 2 || tier < 0 || tier >= 3 || lit < 0 || lit >= 2 {
		return nil
	}
	return wolServerIcons[small][tier][lit]
}

//export nox_wol_map_icon
func nox_wol_map_icon(quadrant C.int) *C.nox_video_bag_image_t {
	if quadrant < 0 || quadrant >= C.int(len(wolMapIcons)) {
		return nil
	}
	return wolMapIcons[quadrant]
}

func wolLoadLegacyImage(offset uintptr, name string) *nox_video_bag_image_t {
	img := nox_xxx_gLoadImg_42F970(internCStr(name))
	mirror := memmap.PtrUint32(0x5D4594, offset)
	*mirror = 0
	if unsafe.Sizeof(uintptr(0)) == 4 {
		*mirror = uint32(uintptr(unsafe.Pointer(img)))
	}
	return img
}

func wolLoadWorldIcons() {
	var offset uintptr = 814556
	for tier, color := range [...]string{"Green", "Yellow", "Red"} {
		for small, size := range [...]string{"Large", "Small"} {
			for lit, suffix := range [...]string{"", "Lit"} {
				wolServerIcons[small][tier][lit] = wolLoadLegacyImage(offset, "NWGameIcon"+size+color+suffix)
				offset += 4
			}
		}
	}
	for i, name := range [...]string{"NWMapULLg", "NWMapURLg", "NWMapLLLg", "NWMapLRLg"} {
		wolMapIcons[i] = wolLoadLegacyImage(814900+4*uintptr(i), name)
	}
}

func wolWorldTrigger(event int32, senderID int32) {
	root := wolWindows.root
	sender := nox_xxx_wndGetChildByID_46B0C0(root, senderID)
	nox_window_call_field_94_fnc(root, event, uintptr(unsafe.Pointer(sender)), 0, nil, 0)
}

func wolShowExistingWorld() {
	Set_dword_5d4594_815044(0)
	anim := C.nox_wnd_xxx_815040
	anim.state = C.NOX_GUI_ANIM_IN
	anim.fnc_done_out = (*[0]byte)(C.sub_438330)
	sub_43BE40(3)
	nox_xxx_clientPlaySoundSpecial_452D80(922, 100)
	nox_window_set_hidden(wolWindows.root, 0)
	nox_window_set_hidden(wolWindows.serverInfo, 0)
	nox_xxx_wnd_46ABB0_go(wolWindows.mapPanel, 1)
	if Get_nox_game_createOrJoin_815048() != 0 {
		sub_4375C0(0)
	}
	sub_46ACE0(wolWindows.root, 10047, 10051, C.int(bool2int(Get_dword_587000_87408() == 0)))
	if sub_4D6FA0() == 2 {
		wolWorldTrigger(16391, 10010)
		wolWorldTrigger(16391, 10010)
	}
}

func wolBindWorldControls(root *nox_window) {
	wolWindows.mapPanel = nox_xxx_wndGetChildByID_46B0C0(root, 10020)
	wolWindows.regionPanel = nox_xxx_wndGetChildByID_46B0C0(root, 10021)
	nox_xxx_wndSetWindowProc_46B300(wolWindows.mapPanel, C.sub_439D00)
	nox_xxx_wndSetWindowProc_46B300(wolWindows.regionPanel, C.sub_439D00)
	nox_xxx_wndSetProc_46B2C0(wolWindows.regionPanel, C.nox_xxx_windowMultiplayerSub_439E70)
	wolWindows.heading = nox_xxx_wndGetChildByID_46B0C0(root, 10011)
	wolWindows.gameList = nox_xxx_wndGetChildByID_46B0C0(root, 10037)
	wolWindows.serverDetails = nox_xxx_wndGetChildByID_46B0C0(root, 10034)
	wolWindows.serverInfo = nox_xxx_wndGetChildByID_46B0C0(root, 10033)
	wolWindows.filter = sub_489B80(root)
	wolWindows.serverInfoButton = nox_xxx_wndGetChildByID_46B0C0(root, 10001)
	for i := range wolWindows.gameColumns {
		wolWindows.gameColumns[i] = nox_xxx_wndGetChildByID_46B0C0(root, int32(10038+i))
	}
	host := nox_xxx_wndGetChildByID_46B0C0(root, 10002)
	if noxflags.HasGame(noxflags.GameFlag25) {
		nox_xxx_wnd_46ABB0_go(host, 0)
	}
	join := nox_xxx_wndGetChildByID_46B0C0(root, 10003)
	join.draw_data.field_0 &^= 4
	draw := nox_xxx_wndGetChildByID_46B0C0(root, 10046)
	nox_xxx_wndSetDrawFn_46B340(draw, C.sub_438C80)
	nox_xxx_wndSetProc_46B2C0(wolWindows.gameList, C.nox_xxx_windowMultiplayerSub_439E70)
	nox_window_set_hidden(wolWindows.gameList, 1)
	nox_window_set_hidden(wolWindows.serverInfo, 1)
	nox_xxx_wnd_46B280(wolWindows.serverInfoButton, root)
}

func wolBindListControls(parent, list *nox_window, sliderID, upID, downID int32) *nox_window {
	slider := nox_xxx_wndGetChildByID_46B0C0(parent, sliderID)
	up := nox_xxx_wndGetChildByID_46B0C0(parent, upID)
	down := nox_xxx_wndGetChildByID_46B0C0(parent, downID)
	data := game3ListData(list)
	nox_xxx_wnd_46B280(slider, list)
	nox_xxx_wnd_46B280(up, list)
	nox_xxx_wnd_46B280(down, list)
	data.Field_9, data.Field_7, data.Field_8 = unsafe.Pointer(slider), unsafe.Pointer(up), unsafe.Pointer(down)
	return slider.field_100
}

//export nox_game_showGameSel_4379F0
func nox_game_showGameSel_4379F0() C.int {
	nox_xxx_setQuest_4D6F60(0)
	sub_4D6F80(0)
	if !noxflags.HasGame(noxflags.GameFlag26) && sub_4D6F30() == 0 {
		Set_nox_game_createOrJoin_815048(0)
		Set_dword_587000_87412(-1)
	}
	clear(memmap.PtrT[[3]uint32](0x5D4594, 815076)[:])
	nox_game_addStateCode_43BDD0(10000)
	sub_4A24C0(1)
	sub_4A1BE0(0)
	if Get_dword_587000_87412() != -1 {
		sub_49FDB0(0)
	}
	if wolWindows.root != nil {
		wolShowExistingWorld()
		return 1
	}
	root := nox_new_window_from_file(internCStr("noxworld.wnd"), C.nox_xxx_windowMultiplayerSub_439E70)
	wolWindows.root = root
	if root == nil {
		return 0
	}
	sub_49FF20()
	nox_xxx_wndShowModalMB_46A8C0(root)
	nox_window_set_all_funcs(root, C.sub_439D00, nil, nil)
	anim := nox_gui_makeAnimation_43C5B0(root, 0, 0, 0, -480, 0, 20, 0, -40)
	C.nox_wnd_xxx_815040 = anim
	if anim == nil {
		return 0
	}
	anim.field_0 = 10000
	anim.field_12 = (*[0]byte)(C.sub_438370)
	anim.fnc_done_out = (*[0]byte)(C.sub_438330)
	wolBindWorldControls(root)
	Set_dword_587000_87412(0)
	nox_window_set_hidden(wolWindows.regionPanel, 1)
	wolSetHeading(wolWindows.heading, wolWindowString("JoinServer", 2517))
	sub_46ACE0(wolWindows.mapPanel, 10620, 10631, 1)
	sub_49FDB0(0)
	wolLoadWorldIcons()

	thumb := wolBindListControls(wolWindows.gameList, wolWindows.gameList, 10053, 10043, 10044)
	thumb.width, thumb.height = 16, 12
	nox_xxx_wndSetOffsetMB_46AE40(thumb, 0, -15)
	infoThumb := wolBindListControls(wolWindows.serverInfo, wolWindows.serverDetails, 10032, 10035, 10036)
	infoThumb.width, infoThumb.height = 16, 10
	sub_438480()
	sub_40A3C0(C.uint(nox_xxx_servGetPort_40A430()))
	Set_nox_game_createOrJoin_815048(0)
	Set_dword_5d4594_815044(0)
	C.dword_5d4594_815052 = 0
	sub_43AF90(0)
	Set_qword_5d4594_815068(uint64(nox_platform_get_ticks() + 1000))
	if Get_dword_587000_87408() == 1 {
		sub_4383A0()
	}
	nox_xxx_createSocketLocal_554B40(0)
	Set_nox_wol_server_result_cnt_815088(0)
	Nox_xxx_loadModifyers_4158C0()
	Nox_xxx_loadLook_415D50()
	sub_430C30_set_video_max(C.NOX_MAX_WIDTH, C.NOX_MAX_HEIGHT)
	C.nox_client_setCursorType_477610(0)
	if legacyAddressableWords.dword_5d4594_815096 != 0 {
		wolDisconnectedNotice("Kicked", 2616, &legacyAddressableWords.dword_5d4594_815096)
	} else if legacyAddressableWords.dword_5d4594_815100 != 0 {
		wolDisconnectedNotice("Timeout", 2623, &legacyAddressableWords.dword_5d4594_815100)
	}
	sub_43DE40(C.sub_438770)
	if sub_44A4A0() != 0 {
		sub_44A4B0()
	}
	nox_xxx_wndSetDrawFn_46B340(wolWindows.mapPanel, C.sub_438E30)
	for id := int32(10054); id <= 10057; id++ {
		nox_xxx_wndSetDrawFn_46B340(nox_xxx_wndGetChildByID_46B0C0(root, id), C.sub_438E30)
	}
	if Get_dword_587000_87408() == 0 {
		sub_46ACE0(root, 10047, 10051, 1)
	}
	if sub_4D6FA0() == 1 {
		wolWorldTrigger(16391, 10002)
		nox_xxx_setQuest_4D6F60(1)
		win := nox_xxx_wndGetChildByID_46B0C0(root, 10020)
		nox_window_call_field_93(win, 5, 15663512, 0)
	}
	return 1
}
