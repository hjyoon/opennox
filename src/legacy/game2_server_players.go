package legacy

/*
#include "game2_server_bridge.h"
#include "map_find_player_start_4f7ab0.h"
*/
import "C"
import "unsafe"

func g2ServerFirst(head *C.nox_list_item_t) unsafe.Pointer {
	return unsafe.Pointer(accessListNativeNext416720((*accessListNode416720)(unsafe.Pointer(head))))
}
func g2ServerNext(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(accessListNativeNext416720((*accessListNode416720)(p)))
}
func g2ServerClearList(head *C.nox_list_item_t) {
	for p := g2ServerFirst(head); p != nil; {
		next := g2ServerNext(p)
		accessListNativeRemove416720((*accessListNode416720)(p))
		C.free(p)
		p = next
	}
	accessListNativeClear416720((*accessListNode416720)(unsafe.Pointer(head)))
}

//export sub_456500
func sub_456500() C.int {
	root := C.dword_5d4594_1045684
	if root == nil {
		return 0
	}
	players, teams := g2ServerChild(root, 10501), g2ServerChild(root, 10502)
	if players == nil || teams == nil {
		return 0
	}
	g2ServerEvent(players, 16399, 0, 0)
	g2ServerEvent(teams, 16399, 0, 0)
	g2ServerClearList(&C.nox_gui_server_players.players)
	g2ServerClearList(&C.nox_gui_server_players.teams)
	g2ServerEnabled(C.dword_5d4594_1045688, false)
	g2ServerEnabled(C.dword_5d4594_1045692, false)
	for team := C.nox_server_teamFirst_418B10(); team != nil; team = C.nox_server_teamNext_418B60(team) {
		C.sub_457230(&team.name[0])
	}
	for player := C.nox_common_playerInfoGetFirst_416EA0(); player != nil; player = C.nox_common_playerInfoGetNext_416EE0(player) {
		if player.playerInd == 31 && g2ServerEngine(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) {
			continue
		}
		sub_457140(C.int(player.netCode), &player.name_final[0])
		value := C.nox_xxx_objGetTeamByNetCode_418C80(C.int(player.netCode))
		if value != nil && C.nox_xxx_servObjectHasTeam_419130(value) != 0 {
			sub_4571A0(C.int(player.netCode), C.int(value.id))
		}
	}
	visible := !g2ServerGame(4096)
	g2ServerEnabled(players, visible)
	return g2ServerEnabled(teams, visible)
}

//export sub_456640
func sub_456640(w *C.nox_window, d *C.nox_window_data) C.int {
	root := C.dword_5d4594_1045684
	if w == nil || d == nil || root == nil {
		return 0
	}
	g2ServerDrawBackground(w, d, true)
	if g2ServerGame(1) && !g2ServerGame(0x8000) && C.sub_40A740() == 0 {
		teams, assign := g2ServerChild(root, 10502), g2ServerChild(root, 10503)
		if teams == nil || assign == nil {
			return 1
		}
		if int32(g2ServerEvent(teams, 16404, 0, 0)) < 0 {
			g2ServerEnabled(assign, false)
		} else {
			players := g2ServerChild(root, 10501)
			var selections *int32
			if players != nil {
				selections = (*int32)(unsafe.Pointer(g2ServerEvent(players, 16404, 0, 0)))
			}
			g2ServerEnabled(assign, selections != nil && *selections >= 0)
		}
	}
	teams := g2ServerChild(root, 10502)
	if teams != nil && int32(g2ServerEvent(teams, 16404, 0, 0)) < 0 {
		g2ServerEnabled(C.dword_5d4594_1045688, false)
		g2ServerEnabled(C.dword_5d4594_1045692, false)
	}
	return 1
}

//export sub_456BB0
func sub_456BB0(team *C.nox_team_t) {
	local := (*C.nox_playerInfo)(unsafe.Pointer(uintptr(C.dword_8531A0_2576)))
	if team == nil || (local != nil && (local.field_4&1 != 0 || local.field_3680&1 != 0)) {
		return
	}
	server := g2ServerGame(1)
	value := C.nox_xxx_objGetTeamByNetCode_418C80(C.int(C.nox_player_netCode_85319C))
	if value == nil {
		return
	}
	if C.nox_xxx_servObjectHasTeam_419130(value) != 0 {
		if !server {
			C.sub_419960(team, value, C.short(C.nox_player_netCode_85319C))
			*g2ServerWord(1045696)++
			return
		}
		C.sub_4196D0(unsafe.Pointer(value), unsafe.Pointer(team), C.int(C.nox_player_netCode_85319C), 1)
		if !g2ServerGame(128) {
			p := C.nox_common_playerInfoGetByID_417040(C.int(C.nox_player_netCode_85319C))
			if p != nil && p.playerUnit != nil {
				var pos C.float2
				C.nox_xxx_mapFindPlayerStart_4F7AB0(&pos, p.playerUnit)
				C.nox_xxx_unitMove_4E7010(p.playerUnit, &pos)
			}
		}
	} else if server {
		C.nox_xxx_createAtImpl_4191D0(team.field_57, value, 1, C.int(C.nox_player_netCode_85319C), 1)
	} else {
		C.sub_419900(team, value, C.short(C.nox_player_netCode_85319C))
	}
	*g2ServerWord(1045696)++
}

//export sub_456D00
func sub_456D00(index C.int, out *C.wchar2_t) *C.wchar2_t {
	root := C.dword_5d4594_1045684
	if out == nil || root == nil || index < 0 {
		return nil
	}
	w := g2ServerChild(root, 10502)
	if w == nil {
		return nil
	}
	line := (*C.wchar2_t)(unsafe.Pointer(g2ServerEvent(w, 16406, uintptr(index), 0)))
	if line == nil {
		return nil
	}
	buf, free := g2ServerTemp(56)
	defer free()
	p := (*C.wchar2_t)(unsafe.Pointer(&buf[0]))
	C.nox_wcscpy(p, line)
	name := C.nox_wcstok(p, (*C.wchar2_t)(internWStr("\t\n\r")))
	if name == nil {
		return nil
	}
	return C.nox_wcscpy(out, name)
}

//export sub_456D60
func sub_456D60(destroy C.int) *C.int {
	if destroy != 0 && C.dword_5d4594_1045684 != nil {
		C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1045684)
	}
	g2ServerClearList(&C.nox_gui_server_players.players)
	g2ServerClearList(&C.nox_gui_server_players.teams)
	C.dword_5d4594_1045684 = nil
	C.dword_5d4594_1045688 = nil
	C.dword_5d4594_1045692 = nil
	return nil
}

//export sub_456DF0
func sub_456DF0(id C.int) C.int {
	root := C.dword_5d4594_1045684
	if root == nil {
		return 0
	}
	i := sub_456E40(id, 1)
	if i < 0 {
		return -1
	}
	w := g2ServerChild(root, 10501)
	if w == nil {
		return 0
	}
	return C.int(g2ServerEvent(w, 16398, uintptr(i), 0))
}

//export sub_456E40
func sub_456E40(id, remove C.int) C.int {
	i := 0
	for p := g2ServerFirst(&C.nox_gui_server_players.players); p != nil; p = g2ServerNext(p) {
		entry := (*C.nox_gui_server_player_entry)(p)
		if uint32(entry.net_code) == uint32(id) {
			if remove != 0 {
				accessListNativeRemove416720((*accessListNode416720)(p))
				C.free(p)
			}
			return C.int(i)
		}
		i++
	}
	return -1
}

//export sub_456EA0
func sub_456EA0(name *C.wchar2_t) C.int {
	root := C.dword_5d4594_1045684
	if root == nil {
		return 0
	}
	w := g2ServerChild(root, 10502)
	if w == nil {
		return 0
	}
	i := sub_456F10(name, 1)
	if i >= 0 {
		g2ServerEvent(w, 16404, 0, 0)
		g2ServerEvent(w, 16398, uintptr(i), 0)
	}
	return sub_456500()
}

//export sub_456F10
func sub_456F10(name *C.wchar2_t, remove C.int) C.int {
	if name == nil {
		return -1
	}
	i := 0
	for p := g2ServerFirst(&C.nox_gui_server_players.teams); p != nil; p = g2ServerNext(p) {
		entry := (*C.nox_gui_server_team_entry)(p)
		if C.game2_server_icmp(&entry.name[0], name) == 0 {
			if remove != 0 {
				accessListNativeRemove416720((*accessListNode416720)(p))
				C.free(p)
			}
			return C.int(i)
		}
		i++
	}
	return -1
}

//export sub_456FA0
func sub_456FA0() C.int {
	if C.dword_5d4594_1045684 == nil {
		return 0
	}
	g2ServerClearList(&C.nox_gui_server_players.teams)
	g2ServerEnabled(C.dword_5d4594_1045688, false)
	g2ServerEnabled(C.dword_5d4594_1045692, false)
	return sub_456500()
}

//export sub_457120
func sub_457120(team *C.nox_team_t) C.uchar {
	if team == nil {
		return 3
	}
	return C.uchar(*(*byte)(g2ServerMemOff(0x587000, 128968+8*uintptr(team.def_ind%10))))
}

//export sub_457140
func sub_457140(id C.int, name *C.wchar2_t) C.int {
	root := C.dword_5d4594_1045684
	if root == nil || name == nil {
		return 0
	}
	w := g2ServerChild(root, 10501)
	if w == nil {
		return 0
	}
	p := C.calloc(1, C.sizeof_nox_gui_server_player_entry)
	if p == nil {
		return 0
	}
	entry := (*C.nox_gui_server_player_entry)(p)
	accessListNativeInit416720((*accessListNode416720)(p))
	accessListNativeAppend416720((*accessListNode416720)(unsafe.Pointer(&C.nox_gui_server_players.players)), (*accessListNode416720)(p))
	C.nox_wcscpy(&entry.name[0], name)
	entry.net_code = C.uint32_t(id)
	return C.int(g2ServerText(w, 16397, unsafe.Pointer(name), 3))
}
func g2ServerPlayerColor(index int) uint32 {
	switch index {
	case 0:
		return uint32(C.nox_color_black_2650656)
	case 1:
		return *(*uint32)(g2ServerMemOff(0x852978, 4))
	case 2:
		return *(*uint32)(g2ServerMemOff(0x85B3FC, 956))
	case 3:
		return *g2ServerWord(2597996)
	case 4:
		return uint32(C.nox_color_white_2523948)
	case 5:
		return uint32(C.nox_color_violet_2598268)
	case 6:
		return *(*uint32)(g2ServerMemOff(0x85B3FC, 940))
	case 7:
		return uint32(C.nox_color_red_2589776)
	case 8:
		return *(*uint32)(g2ServerMemOff(0x85B3FC, 984))
	case 9:
		return uint32(C.dword_8531A0_2572)
	case 10:
		return uint32(C.nox_color_green_2614268)
	case 11:
		return *(*uint32)(g2ServerMemOff(0x85B3FC, 944))
	case 12:
		return uint32(C.nox_color_cyan_2649820)
	case 13:
		return uint32(C.nox_color_blue_2650684)
	case 14:
		return uint32(C.nox_color_orange_2614256)
	case 15:
		return uint32(C.nox_color_yellow_2589772)
	case 16:
		return *(*uint32)(g2ServerMemOff(0x852978, 0))
	default:
		return *g2ServerWord(2597996)
	}
}

//export sub_4571A0
func sub_4571A0(playerID, teamID C.int) C.int {
	root := C.dword_5d4594_1045684
	if root == nil {
		return 0
	}
	w := g2ServerChild(root, 10501)
	d := g2ServerList(w)
	i := int(sub_456E40(playerID, 0))
	if d == nil || d.Items == nil || i < 0 || i >= int(d.Field_11_0) {
		return 0
	}
	color := 3
	if teamID != 0 {
		for p := g2ServerFirst(&C.nox_gui_server_players.teams); p != nil; p = g2ServerNext(p) {
			team := (*C.nox_gui_server_team_entry)(p)
			if uint32(team.team_id) == uint32(teamID) {
				color = int(team.color)
				break
			}
		}
	}
	g2ServerItems(d)[i].Field_129 = g2ServerPlayerColor(color)
	return 1
}

//export sub_457350
func sub_457350(oldID, newID C.int) {
	old := uint32(uint8(oldID))
	for p := g2ServerFirst(&C.nox_gui_server_players.teams); p != nil; p = g2ServerNext(p) {
		team := (*C.nox_gui_server_team_entry)(p)
		if uint32(team.team_id) == old {
			team.team_id = C.uint32_t(uint8(newID))
		}
	}
}

//export sub_4573A0
func sub_4573A0() C.int {
	if C.dword_5d4594_1045684 != nil {
		return sub_456500()
	}
	return 0
}

//export sub_4573B0
func sub_4573B0() { *g2ServerWord(1045696) = 0 }
