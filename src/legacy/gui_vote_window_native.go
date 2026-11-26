package legacy

/*
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "client__gui__guivote.h"
#include "noxstring.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type voteNames [32][28]uint16

var voteWindow struct {
	root, players, topics *nox_window
}

func voteNameIn(name *wchar2_t, names [][28]uint16) bool {
	for i := range names {
		if C.nox_wcscmp(name, (*wchar2_t)(unsafe.Pointer(&names[i][0]))) == 0 {
			return true
		}
	}
	return false
}

func voteNameCount(count C.uint32_t) int {
	return min(len(voteNames{}), max(0, int(int32(count))))
}

//export nox_xxx_showObserverWindow_48CA70
func nox_xxx_showObserverWindow_48CA70(hidden C.int) C.int {
	return C.int(nox_window_set_hidden(observerStatusIcon, int32(hidden)))
}

//export nox_xxx_cliToggleObsWindow_4357A0
func nox_xxx_cliToggleObsWindow_4357A0() C.int {
	player := Get_dword_8531A0_2576()
	return nox_xxx_showObserverWindow_48CA70(C.int(bool2int(player == nil || player.Field3680&1 == 0)))
}

//export sub_48CAD0
func sub_48CAD0() C.int {
	window := voteWindow.root
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(window) != 0 {
		return 0
	}
	nox_xxx_wnd_46C6E0(window)
	g2ServerHidden(window, true)
	return 1
}

//export sub_48D000_initGuiKick
func sub_48D000_initGuiKick() C.int {
	root := nox_new_window_from_file(internCStr("GuiKick.wnd"), C.nox_xxx_guiKick_48D0A0)
	storeWindowReference(&voteWindow.root, root)
	if root == nil {
		return 0
	}
	storeWindowReference(&voteWindow.players, g2ServerChild(root, 4320))
	storeWindowReference(&voteWindow.topics, g2ServerChild(root, 4321))
	nox_window_setPos_46A9B0(root, (int32(clientDisplayStorage.width)-int32(root.width))/2, int32(root.off_y))
	g2ServerHidden(root, true)
	legacyWords.dword_5d4594_1197324, legacyWords.dword_5d4594_1197328 = 0, 0
	sub_48D4A0()
	return 1
}

//export nox_xxx_guiKick_48D0A0
func nox_xxx_guiKick_48D0A0(win *nox_window, event C.int, control *nox_window, eventArg C.uintptr_t) C.int {
	if event != 16391 {
		return 0
	}
	switch nox_xxx_wndGetID_46B0A0(control) {
	case 4311:
		switch legacyWords.dword_5d4594_1197308 {
		case 0, 1, 3:
			sub_48D120()
			sub_48CAD0()
		case 2:
			sub_48D340()
			sub_48CAD0()
		case 4:
			sub_48D410()
		}
	case 4312:
		sub_48CAD0()
	}
	nox_xxx_clientPlaySoundSpecial_452D80(921, 100)
	return 1
}

func voteCurrentNames() *voteNames {
	return memmap.PtrT[voteNames](0x5D4594, 1193720)
}

//export sub_48D120
func sub_48D120() C.int {
	live := voteCurrentNames()
	previous := memmap.PtrT[voteNames](0x5D4594, 1195512)
	legacyWords.dword_5d4594_1197328 = legacyWords.dword_5d4594_1197324
	*previous = *live
	legacyWords.dword_5d4594_1197324 = 0
	win := voteWindow.players
	selected := g2ServerEvent(win, 16404, 0, 0)
	if data := g2ServerList(win); selected != 0 && data != nil && data.Field_4 != 0 {
		indices := unsafe.Slice((*uint32)(unsafe.Pointer(selected)), int(data.Count))
		for _, index := range indices {
			if index == ^uint32(0) || legacyWords.dword_5d4594_1197324 == C.uint32_t(len(live)) {
				break
			}
			if name := listBoxRowText(win, uintptr(index)); name != nil {
				entry := &live[legacyWords.dword_5d4594_1197324]
				alloc.StrCopyZero16P(entry[:], (*uint16)(unsafe.Pointer(name)))
				legacyWords.dword_5d4594_1197324++
			}
		}
	}

	before := previous[:voteNameCount(legacyWords.dword_5d4594_1197328)]
	after := live[:voteNameCount(legacyWords.dword_5d4594_1197324)]
	for i := range before {
		name := (*wchar2_t)(unsafe.Pointer(&before[i][0]))
		if !voteNameIn(name, after) {
			nox_xxx_voteSend_48D260(name)
		}
	}
	for i := range after {
		name := (*wchar2_t)(unsafe.Pointer(&after[i][0]))
		if !voteNameIn(name, before) {
			nox_xxx_netSendRenameMb_48D2D0(name)
		}
	}
	return C.int(legacyWords.dword_5d4594_1197324)
}

func sendPlayerVote(name *wchar2_t, command uint8) C.int {
	players := &GetServer().S().Players
	for p := players.First(); p != nil; p = players.Next(p) {
		final := (*wchar2_t)(unsafe.Pointer(&p.NameFinal[0]))
		if C.nox_wcscmp(final, name) != 0 {
			continue
		}
		message := [26]uint16{uint16(netmsg.MSG_VOTE) | uint16(command)<<8}
		alloc.StrCopy16P(message[1:], (*uint16)(unsafe.Pointer(name)))
		return C.int(nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(unsafe.Pointer(&message[0])), int32(unsafe.Sizeof(message))))
	}
	return 0
}

//export nox_xxx_voteSend_48D260
func nox_xxx_voteSend_48D260(name *wchar2_t) C.int { return sendPlayerVote(name, 2) }

//export nox_xxx_netSendRenameMb_48D2D0
func nox_xxx_netSendRenameMb_48D2D0(name *wchar2_t) C.int { return sendPlayerVote(name, 0) }

func sendQuestResetVote(command uint8) C.int {
	message := [2]byte{byte(netmsg.MSG_VOTE), command}
	return nox_xxx_netClientSend2_4E53C0(31, C.nox_important_payload_t(unsafe.Pointer(&message[0])), 2, nil, 1)
}

//export sub_48D3B0
func sub_48D3B0() C.int { return sendQuestResetVote(4) }

//export nox_xxx_clientVote_48D3E0
func nox_xxx_clientVote_48D3E0() C.int { return sendQuestResetVote(5) }

//export sub_48D340
func sub_48D340() C.int {
	vote := g2ServerEvent(voteWindow.topics, 16404, 0, 0) == 0
	next := C.uint32_t(bool2int(vote))
	prior := legacyWords.dword_5d4594_1197336
	legacyWords.dword_5d4594_1197332 = next
	result := C.int(next)
	if vote && prior == 0 {
		result = sub_48D3B0()
	} else if !vote && prior == 1 {
		result = nox_xxx_clientVote_48D3E0()
	}
	legacyWords.dword_5d4594_1197336 = legacyWords.dword_5d4594_1197332
	return result
}

//export sub_48D410
func sub_48D410() C.uintptr_t {
	selection := g2ServerEvent(voteWindow.topics, 16404, 0, 0)
	switch selection {
	case 0:
		return sub_48CB10(2)
	case 1:
		return sub_48CB10(3)
	default:
		return C.uintptr_t(selection)
	}
}

//export sub_48D450
func sub_48D450() C.int {
	root := voteWindow.root
	nox_xxx_wnd_46C6E0(root)
	nox_xxx_wndClearCaptureMain_46ADE0(root)
	nox_xxx_windowDestroyMB_46C4E0(root)
	voteWindow.root, voteWindow.players, voteWindow.topics = nil, nil, nil
	legacyWords.dword_5d4594_1197324, legacyWords.dword_5d4594_1197328 = 0, 0
	return sub_48D4A0()
}

//export sub_48D4A0
func sub_48D4A0() C.int {
	legacyWords.dword_5d4594_1197332, legacyWords.dword_5d4594_1197336 = 0, 0
	return 0
}

//export sub_48D4B0
func sub_48D4B0(mode C.int) C.int {
	*g2ServerWord(1197304) = uint32(mode)
	return sub_4C3460(C.int(bool2int(mode != 1)))
}

const voteWindowSource = `C:\NoxPost\src\client\Gui\GUIVote.c`

func voteText(id string, line int32) *wchar2_t {
	return nox_strman_loadString_40F1D0(internCStr(id), nil, internCStr(voteWindowSource), line)
}

func voteTopic(labelLine, topicLine int32, topic string) *wchar2_t {
	return internWStr(GoWString(voteText("VoteTopicLabel", labelLine)) + " " + GoWString(voteText(topic, topicLine)))
}

func addVotePlayers(win *nox_window) {
	color := uintptr(4)
	var team *nox_team_t
	if nox_xxx_getTeamCounter_417DD0() != 0 {
		membership := nox_xxx_objGetTeamByNetCode_418C80(C.int(ClientPlayerNetCode()))
		if membership != nil {
			team = nox_xxx_getTeamByID_418AB0(int32(membership.id))
		}
		if team == nil {
			return
		}
		color = uintptr(memmap.Uint32(0x587000, 156400+8*uintptr(team.field_57%10)))
	}
	currentPlayer := Get_dword_8531A0_2576()
	players := &GetServer().S().Players
	names := voteCurrentNames()[:voteNameCount(legacyWords.dword_5d4594_1197324)]
	index := uintptr(0)
	for p := players.First(); p != nil; p = players.Next(p) {
		if p == currentPlayer {
			continue
		}
		if team != nil {
			membership := nox_xxx_objGetTeamByNetCode_418C80(C.int(p.NetCode()))
			if membership == nil || nox_xxx_teamCompare2_419180(membership, team.field_57) == 0 {
				continue
			}
		}
		name := (*wchar2_t)(unsafe.Pointer(&p.NameFinal[0]))
		g2ServerText(win, 16397, unsafe.Pointer(name), color)
		if voteNameIn(name, names) {
			g2ServerEvent(win, 16405, index, 0)
		}
		index++
	}
}

//export sub_48CB10
func sub_48CB10(mode C.int) C.uintptr_t {
	root := voteWindow.root
	players, topics := voteWindow.players, voteWindow.topics
	g2ServerEvent(players, 16399, 0, 0)
	g2ServerEvent(topics, 16399, 0, 0)
	legacyWords.dword_5d4594_1197308 = C.uint32_t(mode)
	switch mode {
	case 4:
		if p := Get_dword_8531A0_2576(); p != nil && p.Field4792 == 0 {
			message := voteText("GUIVote.c:NotAllowedVote", 452)
			title := voteText("guiquit.c:Vote", 451)
			return C.uintptr_t(uintptr(nox_xxx_dialogMsgBoxCreate_449A10(nil, title, message, 33, nil, nil)))
		}
		g2ServerHidden(players, true)
		g2ServerHidden(topics, false)
		C.sub_46AEE0(g2ServerChild(root, 4301), voteText("SelectVoteTopic", 465))
		g2ServerText(topics, 16397, unsafe.Pointer(voteTopic(468, 470, "VoteResetServer")), 4)
		g2ServerText(topics, 16397, unsafe.Pointer(voteTopic(474, 476, "VoteKickPlayer")), 4)
	case 2:
		g2ServerHidden(players, true)
		g2ServerHidden(topics, false)
		C.sub_46AEE0(g2ServerChild(root, 4301), voteText("Vote:ResetQuest", 491))
		g2ServerText(topics, 16397, unsafe.Pointer(voteText("WindowDir:Yes", 495)), 4)
		g2ServerText(topics, 16397, unsafe.Pointer(voteText("WindowDir:No", 498)), 4)
		g2ServerEvent(topics, 16403, uintptr(bool2int(legacyWords.dword_5d4594_1197332 != 1)), 0)
	case 0, 1, 3:
		g2ServerHidden(players, false)
		g2ServerHidden(topics, true)
		C.sub_46AEE0(g2ServerChild(root, 4301), voteText("VoteKickPlayer", 520))
		addVotePlayers(players)
	default:
		return C.uintptr_t(mode)
	}
	return C.uintptr_t(nox_xxx_wndShowModalMB_46A8C0(root))
}
