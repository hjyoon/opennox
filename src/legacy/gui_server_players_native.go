package legacy

/*
#include <stdlib.h>
#include "game2_server_bridge.h"
#include "client__gui__servopts__playrlst.h"
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

const serverPlayersStringsFile = "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\playrlst.c"

func attachServerListScrollbar(root, list *nox_window, thumbID, upID, downID int32, slider, lit unsafe.Pointer) {
	if root == nil || list == nil {
		return
	}
	data := g2ServerList(list)
	thumb := nox_xxx_wndGetChildByID_46B0C0(root, thumbID)
	up := nox_xxx_wndGetChildByID_46B0C0(root, upID)
	down := nox_xxx_wndGetChildByID_46B0C0(root, downID)
	if data == nil || thumb == nil || up == nil || down == nil {
		return
	}
	if thumb.field_100 != nil {
		thumb.field_100.width, thumb.field_100.height = 16, 10
	}
	sub_4B5700(thumb, nil, nil, slider, lit, lit)
	nox_xxx_wnd_46B280(thumb, list)
	nox_xxx_wnd_46B280(up, list)
	nox_xxx_wnd_46B280(down, list)
	data.Field_9, data.Field_7, data.Field_8 = unsafe.Pointer(thumb), unsafe.Pointer(up), unsafe.Pointer(down)
}

func serverPlayersWindowFile(language int32) *C.char {
	if language < 0 || language >= 9 {
		language = 0
	}
	return (*C.char)(*memmap.PtrPtr(0x587000, 129048+4*uintptr(language)))
}

//export nox_xxx_guiServerPlayersLoad_456270
func nox_xxx_guiServerPlayersLoad_456270(parent *C.nox_window) *C.nox_window {
	language := nox_strman_get_lang_code()
	if nox_xxx_guiFontHeightMB_43F320(nil) > 10 {
		language = 2
	}
	if C.dword_5d4594_1045684 != nil {
		return nil
	}
	root := nox_new_window_from_file(serverPlayersWindowFile(language), unsafe.Pointer(C.sub_4567C0))
	C.dword_5d4594_1045684 = root
	if root == nil {
		return nil
	}
	C.dword_5d4594_1045688 = nox_xxx_wndGetChildByID_46B0C0(root, 10507)
	C.dword_5d4594_1045692 = nox_xxx_wndGetChildByID_46B0C0(root, 10509)
	sub_46B120(root, parent)
	nox_xxx_wndSetDrawFn_46B340_go(root, unsafe.Pointer(C.sub_456640))
	nox_xxx_wndRetNULL_46A8A0()
	nox_common_list_clear_425760(&C.nox_gui_server_players.players)
	nox_common_list_clear_425760(&C.nox_gui_server_players.teams)

	slider := unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UISlider")))
	lit := unsafe.Pointer(nox_xxx_gLoadImg_42F970(internCStr("UISliderLit")))
	players := nox_xxx_wndGetChildByID_46B0C0(root, 10501)
	teams := nox_xxx_wndGetChildByID_46B0C0(root, 10502)
	attachServerListScrollbar(root, players, 10517, 10515, 10516, slider, lit)
	attachServerListScrollbar(root, teams, 10520, 10518, 10519, slider, lit)
	sub_456500()
	if title := nox_xxx_wndGetChildByID_46B0C0(root, 10504); title != nil && noxflags.HasGame(noxflags.GameModeChat) {
		text := g2ServerLoad("Title1", serverPlayersStringsFile, 631)
		g2ServerText(title, 16385, unsafe.Pointer(text), ^uintptr(0))
	}
	return C.dword_5d4594_1045684
}

func serverSelectedTeam(teams *nox_window, selected C.int) (*C.nox_team_t, bool) {
	if teams == nil {
		return nil, false
	}
	var name [56]C.wchar2_t
	if sub_456D00(selected, &name[0]) == nil {
		return nil, false
	}
	return sub_418A40(&name[0]), true
}

func serverRosterTeamButtons(teams *nox_window) {
	settings := sub_4165B0()
	selected := int32(g2ServerEvent(teams, 16404, 0, 0))
	remove := C.dword_5d4594_1045688
	rename := C.dword_5d4594_1045692
	if selected < 0 || noxflags.HasGame(serverOptionsClanLadder) || settings == nil || *(*int8)(g2ServerAt(settings, 53)) < 0 {
		g2ServerEnabled(remove, false)
		g2ServerEnabled(rename, false)
	} else {
		if noxflags.HasGame(noxflags.GameHost) {
			g2ServerEnabled(rename, true)
		}
		g2ServerEnabled(remove, noxflags.HasGame(noxflags.GameModeChat) || *g2ServerWord(1045696) == 0)
	}
	if noxflags.HasGame(noxflags.GameHost) && noxflags.HasEngine(noxflags.EngineNoRendering) {
		g2ServerEnabled(remove, false)
	}
}

func serverRosterAssignPlayers(root, teams *nox_window) {
	if teams == nil {
		return
	}
	selected := C.int(g2ServerEvent(teams, 16404, 0, 0))
	if selected >= 0 {
		team, ok := serverSelectedTeam(teams, selected)
		if !ok {
			return
		}
		players := nox_xxx_wndGetChildByID_46B0C0(root, 10501)
		if players == nil {
			return
		}
		selections := (*C.int)(unsafe.Pointer(g2ServerEvent(players, 16404, 0, 0)))
		if team != nil && selections != nil {
			for selection := selections; *selection >= 0; selection = (*C.int)(unsafe.Add(unsafe.Pointer(selection), unsafe.Sizeof(*selection))) {
				name := listBoxRowText(players, uintptr(*selection))
				player := g2ServerPlayerByName(name)
				if player == nil || player.field_3680&1 != 0 || player.field_4&1 != 0 {
					continue
				}
				member := nox_xxx_objGetTeamByNetCode_418C80(C.int(player.netCode))
				if member == nil {
					continue
				}
				if nox_xxx_servObjectHasTeam_419130(member) != 0 {
					sub_4196D0(unsafe.Pointer(member), unsafe.Pointer(team), C.int(player.netCode), 1)
				} else {
					nox_xxx_createAtImpl_4191D0(team.field_57, member, 1, C.int(player.netCode), 1)
				}
			}
		}
	}
	g2ServerEvent(teams, 16403, ^uintptr(0), 0)
	if players := nox_xxx_wndGetChildByID_46B0C0(root, 10501); players != nil {
		g2ServerEvent(players, 16403, ^uintptr(0), 0)
	}
}

//export sub_4567C0
func sub_4567C0(unusedWindow *C.nox_window, event C.int, window *C.nox_window, unusedArg C.uintptr_t) C.int {
	root := C.dword_5d4594_1045684
	if root == nil || window == nil {
		return 0
	}
	if event == 16400 && nox_xxx_wndGetID_46B0A0(window) == 10502 {
		serverRosterTeamButtons(window)
	}
	if event != 16391 && event != 16400 {
		return 0
	}
	id := nox_xxx_wndGetID_46B0A0(window)
	nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
	switch id {
	case 10509:
		prompt := g2ServerLoad("NewName", serverPlayersStringsFile, 504)
		title := g2ServerLoad("Rename", serverPlayersStringsFile, 504)
		nox_xxx_dialogMsgBoxCreate_449A10(root, title, prompt, 163, nil, nil)
	case 10507, 4001:
		teams := nox_xxx_wndGetChildByID_46B0C0(root, 10502)
		if teams == nil {
			return 0
		}
		selected := C.int(g2ServerEvent(teams, 16404, 0, 0))
		team, ok := serverSelectedTeam(teams, selected)
		if !ok {
			return 0
		}
		if id == 10507 {
			if team != nil {
				sub_456BB0(team)
			}
		} else {
			var dialogEntry C.uchar = 0xa8
			name := (*C.wchar2_t)(unsafe.Pointer(uintptr(sub_449E60(C.char(dialogEntry)))))
			if team != nil && name != nil && sub_4190F0(name) == 0 {
				C.nox_xxx_teamRenameMB_418CD0(team, name)
			}
		}
	case 10503:
		serverRosterAssignPlayers(root, nox_xxx_wndGetChildByID_46B0C0(root, 10502))
	}
	return 0
}

func serverRosterTeamTitle(line, name *C.wchar2_t, team *C.nox_team_t, settings *C.char, redLine int) {
	copyWideString(line, name)
	if !g2ServerGame(96) && (settings == nil || *(*byte)(g2ServerAt(settings, 52))&0x60 == 0) {
		return
	}
	if team.field_57 >= 3 {
		return
	}
	key := "RedFlag"
	if team.field_57 != 1 {
		key = "BlueFlag"
		redLine += 4
	}
	appendWideString(line, g2ServerLoad(key, serverPlayersStringsFile, redLine))
}

//export sub_457010
func sub_457010(team *C.nox_team_t, name *C.wchar2_t) C.int {
	root := C.dword_5d4594_1045684
	if root == nil || team == nil || name == nil {
		return 0
	}
	head := &C.nox_gui_server_players.teams
	for pointer := g2ServerFirst(head); pointer != nil; pointer = g2ServerNext(pointer) {
		entry := (*C.nox_gui_server_team_entry)(pointer)
		if entry.team_id == C.uint32_t(team.field_57) {
			copyWideString(&entry.name[0], name)
			break
		}
	}
	settings := sub_4165B0()
	teams := nox_xxx_wndGetChildByID_46B0C0(root, 10502)
	if teams == nil {
		return 0
	}
	selected := int32(g2ServerEvent(teams, 16404, 0, 0))
	if selected < 0 {
		return 0
	}
	var line [56]C.wchar2_t
	serverRosterTeamTitle(&line[0], name, team, settings, 778)
	row := uintptr(selected)
	g2ServerEvent(teams, 16398, row, 0)
	g2ServerEvent(teams, 16402, row, 0)
	g2ServerText(teams, 16397, unsafe.Pointer(&line[0]), uintptr(sub_457120(team)))
	return C.int(g2ServerEvent(teams, 16403, row, 0))
}

//export sub_457230
func sub_457230(name *C.wchar2_t) C.uintptr_t {
	root := C.dword_5d4594_1045684
	if root == nil || name == nil {
		return 0
	}
	teams := nox_xxx_wndGetChildByID_46B0C0(root, 10502)
	if teams == nil {
		return 0
	}
	team := sub_418A40(name)
	if team == nil {
		return 0
	}
	entry := (*C.nox_gui_server_team_entry)(C.calloc(1, C.size_t(unsafe.Sizeof(C.nox_gui_server_team_entry{}))))
	if entry == nil {
		return 0
	}
	entry.team_id = C.uint32_t(team.field_57)
	entry.def_ind = C.uint32_t(team.def_ind)
	entry.color = C.uint8_t(sub_457120(team))
	sub_425770(unsafe.Pointer(&entry.list))
	nox_common_list_append_4258E0(&C.nox_gui_server_players.teams, &entry.list)
	copyWideString(&entry.name[0], name)

	var line [56]C.wchar2_t
	serverRosterTeamTitle(&line[0], name, team, sub_4165B0(), 893)
	return C.uintptr_t(g2ServerText(teams, 16397, unsafe.Pointer(&line[0]), uintptr(entry.color)))
}
