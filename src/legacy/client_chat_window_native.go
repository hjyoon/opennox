package legacy

/*
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "client__gui__window.h"

extern nox_window* dword_5d4594_1064856;
extern nox_window* dword_5d4594_1064860;
extern void* dword_5d4594_1064864;
extern unsigned int dword_5d4594_1064868, nox_player_netCode_85319C;
extern int nox_win_width, nox_win_height;
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const clientChatTeamOffset = 1064872

func clientChatEntryData() *gui.EntryFieldData {
	return (*gui.EntryFieldData)(C.dword_5d4594_1064864)
}

//export nox_client_chatStart_46A430
func nox_client_chatStart_46A430(team C.int) {
	if noxflags.HasGame(noxflags.GameModeCoop) || C.dword_5d4594_1064868 != 0 {
		return
	}
	data := clientChatEntryData()
	if data == nil {
		return
	}
	data.Text[0] = 0
	data.Field_1052 &^= 0xffff
	root := C.dword_5d4594_1064856
	nox_xxx_wndShowModalMB_46A8C0(root)
	sub_46C690(root)
	nox_xxx_windowFocus_46B500(C.dword_5d4594_1064860)
	C.dword_5d4594_1064868 = 1
	*memmap.PtrInt32(0x5D4594, clientChatTeamOffset) = int32(team)
}

func clientChatMessage(text []uint16, narrow, team bool, code, x, y uint16) *noxnet.MsgText {
	for len(text) != 0 && text[0] == ' ' {
		text = text[1:]
	}
	if len(text) == 0 {
		return nil
	}
	const utf16Flag = noxnet.TextFlags(4)
	flag, stride := utf16Flag, 2
	if narrow {
		flag, stride = noxnet.TextUTF8, 1
	}
	if team {
		flag |= noxnet.TextTeam
	}
	size := byte(len(text) + 1)
	data := make([]byte, stride*int(size))
	for i, unit := range text[:min(len(text), int(size))] {
		if narrow {
			data[i] = byte(unit)
		} else {
			binary.LittleEndian.PutUint16(data[2*i:], unit)
		}
	}
	return &noxnet.MsgText{NetCode: code, Flags: flag, PosX: x, PosY: y, Size: size, Data: data}
}

//export nox_xxx_cmdSayDo_46A4B0
func nox_xxx_cmdSayDo_46A4B0(text *C.wchar2_t, team C.int) C.size_t {
	code := C.nox_player_netCode_85319C
	drawable := nox_xxx_netSpriteByCodeDynamic_45A6F0(int32(code))
	var x, y uint16 = 0xffff, 0xffff
	if drawable != nil {
		x, y = uint16(drawable.pos.x), uint16(drawable.pos.y)
	}
	ptr := (*uint16)(unsafe.Pointer(text))
	n := alloc.StrLen(ptr)
	if n == 0 {
		return 0
	}
	units := unsafe.Slice(ptr, n)
	message := clientChatMessage(units, cliCanTalkNative4100F0(ptr) != 0, team != 0, uint16(code), x, y)
	if message == nil {
		return C.size_t(n)
	}
	packet, err := netmsg.Append(nil, message)
	if err != nil {
		return 0
	}
	return C.size_t(nox_netlist_addToMsgListCli_40EBC0(31, 0, (*C.uchar)(unsafe.Pointer(&packet[0])), int32(len(packet))))
}

//export sub_46A5D0
func sub_46A5D0(win *C.nox_window, draw *C.nox_window_data) C.int {
	root, data := C.dword_5d4594_1064856, clientChatEntryData()
	if root == nil || data == nil {
		return 0
	}
	nox_xxx_wndShowModalMB_46A8C0(root)
	nox_xxx_windowFocus_46B500(C.dword_5d4594_1064860)
	var textWidth, prefixWidth C.int
	nox_xxx_drawGetStringSize_43F840(nil, (*wchar2_t)(unsafe.Pointer(&data.Text[0])), &textWidth, nil, 0)
	nox_xxx_drawGetStringSize_43F840(nil, (*wchar2_t)(unsafe.Pointer(&data.Text[256])), &prefixWidth, nil, 0)
	width := min(320, max(100, textWidth+prefixWidth+10))
	nox_window_setPos_46A9B0(root, int32((C.nox_win_width-width)/2), int32(root.off_y))
	C.sub_46AB20(win, width, 20)
	return nox_xxx_wndEditDrawNoImage_488160(win, draw)
}

//export sub_46A6A0
func sub_46A6A0() C.int {
	root := C.dword_5d4594_1064856
	if wndIsShown_nox_xxx_wndIsShown_46ACC0(root) != 0 {
		return 0
	}
	entry := C.dword_5d4594_1064860
	if nox_xxx_wndGetFocus_46B4F0() == entry {
		nox_xxx_windowFocus_46B500(nil)
	}
	nox_xxx_wnd_46C6E0(root)
	nox_window_set_hidden(root, 1)
	root.flags &^= 8
	if entry != nil {
		entry.flags &^= 8
	}
	set_dword_5d4594_3799468(1)
	C.dword_5d4594_1064868 = 0
	return 1
}

//export sub_46A730
func sub_46A730() *C.nox_window {
	x, y := int32(C.nox_win_width/2), int32(2*C.nox_win_height/3)
	*memmap.PtrInt32(0x5D4594, 1064876) = x
	*memmap.PtrInt32(0x5D4594, 1064880) = y
	root := nox_new_window_from_file(internCStr("GuiChat.wnd"), C.sub_46A820)
	C.dword_5d4594_1064856 = root
	if root == nil {
		return nil
	}
	nox_window_setPos_46A9B0(root, x, y)
	entry := nox_xxx_wndGetChildByID_46B0C0(root, 9201)
	C.dword_5d4594_1064860 = entry
	if entry == nil {
		return nil
	}
	nox_xxx_wndSetWindowProc_46B300_go(entry, C.sub_46A7E0)
	C.dword_5d4594_1064864 = entry.widget_data
	return root
}

//export sub_46A7E0
func sub_46A7E0(win *C.nox_window, event C.int, key, pressed C.uintptr_t) C.int {
	if event != 21 || key != 1 {
		return nox_xxx_wndEditProc_487D70(win, event, key, pressed)
	}
	if pressed == 2 {
		nox_xxx_consoleEsc_49B7A0()
	}
	return 1
}

//export sub_46A820
func sub_46A820(win *C.nox_window, event, arg1, arg2 C.int) C.int {
	if event == 16415 {
		if data := clientChatEntryData(); data != nil && uint16(data.Field_1052) != 0 {
			nox_xxx_cmdSayDo_46A4B0((*C.wchar2_t)(unsafe.Pointer(&data.Text[0])), C.int(memmap.Int32(0x5D4594, clientChatTeamOffset)))
		}
		sub_46A6A0()
	}
	return 0
}

//export sub_46A860
func sub_46A860() C.int {
	var result C.int
	if root := C.dword_5d4594_1064856; root != nil {
		result = C.int(nox_xxx_windowDestroyMB_46C4E0(root))
		C.dword_5d4594_1064856 = nil
	}
	C.dword_5d4594_1064860, C.dword_5d4594_1064864 = nil, nil
	C.dword_5d4594_1064868 = 0
	*memmap.PtrInt32(0x5D4594, clientChatTeamOffset) = 0
	return result
}

//export nox_xxx_wndRetNULL_46A8A0
func nox_xxx_wndRetNULL_46A8A0() C.int { return 0 }

//export nox_xxx_wndRetNULL_0_46A8B0
func nox_xxx_wndRetNULL_0_46A8B0() C.int { return 0 }
