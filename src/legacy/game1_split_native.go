package legacy

/*
#include <stdlib.h>
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "client__gui__guiquit.h"
#include "client__gui__window.h"
#include "client__shell__selchar.h"
#include "common/fs/nox_fs.h"
#include "common__binfile.h"
#include "client__video__draw_common.h"
void nox_client_gui_set_flag_815132(int v);
char* nox_client_getChatMap_49FF40(short* pos);
void nox_xxx_setMouseBounds_430A70(int x1, int x2, int y1, int y2);
int nox_xxx_isQuest_4D6F50(void);
extern uint32_t dword_5d4594_815044;
extern uint32_t dword_5d4594_815748;
extern uint32_t dword_5d4594_816368, dword_5d4594_816372, dword_5d4594_816376;
extern uint32_t dword_587000_93156, dword_5d4594_825736;
extern nox_window* dword_5d4594_816412;
extern nox_window* dword_5d4594_825744;
extern nox_window* dword_5d4594_826028;
extern nox_window* dword_5d4594_826032;
extern nox_window* nox_wnd_quitMenu_825760;
extern void* dword_5d4594_826036;
extern uint32_t nox_color_black_2650656, nox_color_white_2523948, nox_color_orange_2614256;
extern int nox_win_width, nox_win_height;
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// The transfer owners are separate from the original adjacent DWORD lengths.
// All retained values are foreign pointers; the caller owns their lifetime.
var game1TransferBuffers [2]unsafe.Pointer

type game1SplitHooks struct {
	NewBackdrop   func() *nox_window
	DestroyWindow func(*nox_window) int32
	StringWidth   func(*wchar2_t) int32
}

var game1SplitTestHooks *game1SplitHooks

//export nox_xxx_xferSetBuffer_446520
func nox_xxx_xferSetBuffer_446520(index C.int, p unsafe.Pointer) {
	if uint32(index) < 2 {
		game1TransferBuffers[index] = p
	}
}

//export nox_xxx_xferGetBuffer_446520
func nox_xxx_xferGetBuffer_446520(index C.int) unsafe.Pointer {
	if uint32(index) < 2 {
		return game1TransferBuffers[index]
	}
	return nil
}

//export nox_client_xxx_switchChatMap_43B510
func nox_client_xxx_switchChatMap_43B510() {
	C.nox_client_gui_set_flag_815132(0)
	C.nox_xxx_setGameFlags_40A4D0(5)
	C.nox_xxx_setMouseBounds_430A70(0, C.nox_win_width-1, 0, C.nox_win_height-1)
	position := C.uint32_t(memmap.Uint32(0x5D4594, 814916))
	name := GoString(C.nox_client_getChatMap_49FF40((*C.short)(unsafe.Pointer(&position))))
	if strings.Contains(name, ".") {
		return
	}
	C.sub_409B50(internCStr(name))
	var extension [4]byte
	copy(extension[:], unsafe.Slice((*byte)(memmap.PtrOff(0x587000, 90856)), 4))
	name += string(extension[:])
	if !bool(C.nox_common_getEngineFlag(C.NOX_ENGINE_FLAG_REPLAY_READ)) {
		C.nox_xxx_gameSetMapPath_409D70(internCStr(name))
	}
	C.nox_common_gameFlags_unset_40A540(55280)
	C.nox_xxx_setGameFlags_40A4D0(128)
	if p := C.sub_4165D0(0); p != nil {
		flags := (*uint16)(unsafe.Add(unsafe.Pointer(p), 52))
		*flags = (*flags & 0x280F) | 0x80
	}
}

//export sub_43B6D0
func sub_43B6D0() C.int { return C.int(C.dword_5d4594_815044) }

//export sub_43BDB0
func sub_43BDB0() C.int { return C.int(memmap.Uint32(0x5D4594, 815092)) }

//export sub_43BC10
func sub_43BC10(text *wchar2_t, limit C.uchar) *C.ushort {
	if text == nil {
		return nil
	}
	n := int(C.nox_wcslen(text))
	for {
		var width C.int
		if h := game1SplitTestHooks; h != nil && h.StringWidth != nil {
			width = C.int(h.StringWidth(text))
		} else {
			C.nox_xxx_drawGetStringSize_43F840(nil, text, &width, nil, 0)
		}
		*(*uint16)(unsafe.Add(unsafe.Pointer(text), 2*n)) = 0
		// The original measures before removing the previous last character.
		// An empty string is the lower bound, including limits smaller than 5.
		if int32(width)+5 <= int32(limit) || n == 0 {
			return (*C.ushort)(unsafe.Pointer(text))
		}
		n--
	}
}

//export nox_sprintAddrPort_43BC80_go
func nox_sprintAddrPort_43BC80_go(addr *C.char, port C.ushort, dst *C.char) C.int {
	s := fmt.Sprintf("%s:%d", GoString(addr), uint16(port))
	copy(unsafe.Slice((*byte)(unsafe.Pointer(dst)), len(s)+1), s)
	*(*byte)(unsafe.Add(unsafe.Pointer(dst), len(s))) = 0
	return C.int(len(s))
}

//export sub_43C650
func sub_43C650() C.int {
	// The platform C ABI returns a DWORD even when its Go clock is wider.
	now := uint64(uint32(PlatformTicks()))
	previous := memmap.Uint64(0x5D4594, 815740)
	index := memmap.Uint32(0x5D4594, 815756)
	delta := now
	if previous != 0 {
		delta -= previous
	}
	*memmap.PtrUint64(0x5D4594, 815220+uintptr(index)*8) = delta
	*memmap.PtrUint64(0x5D4594, 815756) = (memmap.Uint64(0x5D4594, 815756) + 1) % 60
	*memmap.PtrUint64(0x5D4594, 815740) = now
	C.dword_5d4594_815748++
	return C.int(C.dword_5d4594_815748)
}

//export sub_43CEB0
func sub_43CEB0() {
	n := int32(C.dword_5d4594_815748)
	if n >= 60 {
		n = 60
	}
	if n <= 10 {
		*memmap.PtrUint64(0x587000, 91880) = 33
		return
	}
	var sum uint64
	for i := int32(0); i < n; i++ {
		sum += memmap.Uint64(0x5D4594, 815220+uintptr(i)*8)
	}
	*memmap.PtrUint64(0x587000, 91880) = sum / uint64(n)
}

func game1MusicFrame(index uint32) unsafe.Pointer {
	return memmap.PtrOff(0x5D4594, 815772+uintptr(16*(index+6*uint32(C.dword_5d4594_816372))))
}

//export sub_43DA80
func sub_43DA80() C.int {
	if int32(C.dword_5d4594_816368) < 6 {
		sub_43DD10(game1MusicFrame(uint32(C.dword_5d4594_816368)))
		C.dword_5d4594_816368++
		return 1
	}
	C.dword_5d4594_816368 = 6
	return 0
}

//export sub_43DAD0
func sub_43DAD0() {
	if C.dword_5d4594_816368 > 0 {
		C.dword_5d4594_816368--
		sub_43D9E0(game1MusicFrame(uint32(C.dword_5d4594_816368)))
	}
	C.dword_5d4594_816368 = 0
}

//export sub_43DB20
func sub_43DB20() C.int { return C.int(C.dword_5d4594_816368) }

//export sub_43DB30
func sub_43DB30(value C.int) C.int { C.dword_5d4594_816368 = C.uint32_t(value); return value }

//export sub_43DB40
func sub_43DB40(index C.int) *C.char { return (*C.char)(game1MusicFrame(uint32(index))) }

//export sub_43DB60
func sub_43DB60() C.int {
	if int32(C.dword_5d4594_816372) >= 3 {
		C.dword_5d4594_816372 = 3
		return 3
	}
	sub_43DA80()
	level := uint32(C.dword_5d4594_816372)
	*memmap.PtrUint32(0x5D4594, 816076+uintptr(4*level)) = uint32(C.dword_5d4594_816368)
	C.dword_5d4594_816368 = 0
	C.dword_5d4594_816372++
	return C.int(C.dword_5d4594_816372)
}

//export sub_43DBA0
func sub_43DBA0() {
	if C.dword_5d4594_816372 > 0 {
		C.dword_5d4594_816372--
		C.dword_5d4594_816368 = C.uint32_t(memmap.Uint32(0x5D4594, 816076+uintptr(4*uint32(C.dword_5d4594_816372))))
		sub_43DAD0()
	} else {
		C.dword_5d4594_816372 = 0
	}
}

//export sub_43DC00
func sub_43DC00() { C.dword_587000_93156 = 0 }

//export sub_43DC10
func sub_43DC10() C.int {
	result := C.dword_5d4594_816376
	if result != 0 {
		C.dword_587000_93156 = 1
	}
	return C.int(result)
}

//export sub_43DC30
func sub_43DC30() C.int { return C.int(C.dword_587000_93156) }

//export nox_xxx_gui_43E1A0
func nox_xxx_gui_43E1A0(show C.int) *C.uint32_t {
	if show != 0 {
		var w *nox_window
		if h := game1SplitTestHooks; h != nil && h.NewBackdrop != nil {
			w = h.NewBackdrop()
		} else {
			w = C.nox_window_new(nil, 552, 0, 0, C.nox_win_width, C.nox_win_height, nil)
		}
		C.dword_5d4594_816412 = w
		if w != nil {
			w.draw_data.bg_color = C.nox_color_black_2650656
		}
		return (*C.uint32_t)(unsafe.Pointer(w))
	}
	if w := C.dword_5d4594_816412; w != nil {
		if h := game1SplitTestHooks; h != nil && h.DestroyWindow != nil {
			h.DestroyWindow(w)
		} else {
			C.nox_xxx_windowDestroyMB_46C4E0(w)
		}
		C.dword_5d4594_816412 = nil
	}
	// All callers discard the old destroy-status-as-pointer return value.
	return nil
}

//export sub_43E8C0
func sub_43E8C0(value C.int) C.int { *memmap.PtrUint32(0x5D4594, 816408) = uint32(value); return value }

//export sub_43F0E0
func sub_43F0E0(raw *C.uint32_t) C.int {
	if raw == nil {
		return 0
	}
	fields := unsafe.Slice((*uint32)(unsafe.Pointer(raw)), 5)
	if fields[1] == 2 {
		if fields[3] == 2 {
			return 7
		}
		return 5
	}
	if fields[1] != 0 {
		return 0
	}
	result := C.int(0)
	if fields[4] == 2 {
		result++
	}
	if fields[3] == 2 {
		result += 2
	}
	return result
}

//export sub_4417E0_go
func sub_4417E0_go(dst *wchar2_t, name *C.char) {
	if dst == nil || name == nil {
		return
	}
	suffix := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, 103292)), 5)
	text := GoString(name) + string(suffix)
	if end := strings.IndexByte(text, 0); end >= 0 {
		text = text[:end]
	}
	// nox_vsnwprintf's %S widens bytes and ignores string width/precision.
	// Keep that format behavior, including bytes >= 0x80.
	wide, free := game1ByteWide(text + "\t\t")
	defer free()
	C.nox_wcscat(dst, (*wchar2_t)(unsafe.Pointer(&wide[0])))
}

func game1ByteWide(text string) ([]uint16, func()) {
	wide, free := alloc.Make([]uint16{}, len(text)+1)
	for i := range len(text) {
		wide[i] = uint16(text[i])
	}
	return wide, free
}

//export sub_445450
func sub_445450() *wchar2_t {
	var result *wchar2_t
	for i := uintptr(0); i < 3; i++ {
		result = C.nox_wcscpy((*wchar2_t)(memmap.PtrOff(0x5D4594, 823804+644*i)), (*wchar2_t)(memmap.PtrOff(0x5D4594, 825740)))
		*memmap.PtrUint32(0x5D4594, 824440+644*i) = 0
		*memmap.PtrUint8(0x5D4594, 824444+644*i) = 0
	}
	C.dword_5d4594_825736 = 0
	return result
}

//export nox_xxx_drawMessageLines_445530
func nox_xxx_drawMessageLines_445530() C.int {
	r := C.nox_draw_getViewport_437250()
	y := C.int(3*r.height/4 + r.y1 - 15)
	index := uint32(C.dword_5d4594_825736)
	result := C.int(0)
	for line := 0; line < 3; line++ {
		result = C.int(C.gameFrame())
		if memmap.Uint32(0x5D4594, 824440+uintptr(644*index)) < uint32(C.gameFrame()) {
			break
		}
		C.nox_xxx_drawSetTextColor_434390(C.int(C.nox_color_black_2650656))
		text := (*wchar2_t)(memmap.PtrOff(0x5D4594, 823804+uintptr(644*index)))
		var width C.int
		C.nox_xxx_drawGetStringSize_43F840(nil, text, &width, nil, 0)
		x := (C.nox_win_width - width) / 2
		for shadow := uintptr(0); shadow < 4; shadow++ {
			C.nox_xxx_drawString_43F6E0(nil, text, x+C.int(memmap.Uint32(0x587000, 107848+8*shadow)), y+C.int(memmap.Uint32(0x587000, 107852+8*shadow)))
		}
		color := C.nox_color_white_2523948
		if line != 0 {
			color = C.uint32_t(memmap.Uint32(0x5D4594, 2597996))
		}
		C.nox_xxx_drawSetTextColor_434390(C.int(color))
		C.nox_xxx_drawString_43F6E0(nil, text, x, y)
		y -= 4 + C.nox_xxx_guiFontHeightMB_43F320(nil)
		result = C.int(index)
		if index == 0 {
			index = 2
		} else {
			index--
		}
	}
	return result
}

//export nox_xxx_guiChatMode_4456E0
func nox_xxx_guiChatMode_4456E0(w *nox_window, unused unsafe.Pointer) C.int {
	if w == nil {
		return 0
	}
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(w, &x, &y)
	C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(w.draw_data.bg_image), C.int(x)+C.int(w.draw_data.img_px), C.int(y)+C.int(w.draw_data.img_py))
	return 1
}

//export nox_xxx_guiChatShowHide_445730
func nox_xxx_guiChatShowHide_445730(hidden C.int) C.int {
	return C.nox_window_set_hidden(C.dword_5d4594_825744, hidden)
}

//export sub_445770
func sub_445770() C.int {
	result := C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_825744)
	C.dword_5d4594_825744 = nil
	return result
}

//export nox_xxx_quitDialogYes_445B20
func nox_xxx_quitDialogYes_445B20() { C.nox_client_quit_4460C0(); C.sub_445C40() }

//export nox_xxx_quitDialogNo_445B30
func nox_xxx_quitDialogNo_445B30() C.int {
	return C.nox_xxx_wndSetCaptureMain_46ADC0(C.nox_wnd_quitMenu_825760)
}

//export sub_445BA0
func sub_445BA0() { C.sub_413A00(0) }

//export sub_445C20
func sub_445C20() {
	if C.nox_xxx_wndGetFlags_46ADA0(C.nox_wnd_quitMenu_825760)&0x10 == 0 {
		C.sub_445C40()
	}
}

//export sub_445FF0
func sub_445FF0() *C.uint32_t {
	if w := C.nox_wnd_quitMenu_825760; w != nil {
		w.draw_data.bg_color = C.nox_color_black_2650656
	}
	var result *nox_window
	for id := C.int(9001); id <= 9006; id++ {
		result = C.nox_xxx_wndGetChildByID_46B0C0(C.nox_wnd_quitMenu_825760, id)
		if result != nil {
			result.draw_data.text_color = C.nox_color_orange_2614256
		}
	}
	return (*C.uint32_t)(unsafe.Pointer(result))
}

//export nox_gui_xxx_check_446360
func nox_gui_xxx_check_446360() C.uint {
	if w := C.nox_wnd_quitMenu_825760; w != nil {
		return C.uint((^uint32(C.nox_xxx_wndGetFlags_46ADA0(w)) >> 4) & 1)
	}
	return 0
}

//export nox_motd_4463E0
func nox_motd_4463E0(index C.int) C.int {
	C.dword_5d4594_826036 = nil
	size := memmap.PtrUint32(0x5D4594, 826040+uintptr(4*uint32(index)))
	*size = 0
	f := C.nox_fs_open(internCStr("motd.txt"))
	if f == nil {
		return 0
	}
	defer C.nox_fs_close(f)
	length := C.nox_fs_fsize(f)
	if length < 0 || uint64(length) >= uint64(^uint32(0)) {
		return 0
	}
	*size = uint32(length)
	p := C.calloc(C.size_t(*size)+1, 1)
	if p == nil {
		return 0
	}
	C.dword_5d4594_826036 = p
	C.nox_binfile_fread_raw_40ADD0((*C.char)(p), C.size_t(*size), 1, f)
	*(*byte)(unsafe.Add(p, uintptr(*size))) = 0
	*size++
	return 0
}

//export sub_446490_go
func sub_446490_go(index C.int) C.uintptr_t {
	result := C.uintptr_t(0)
	if p := C.dword_5d4594_826036; p != nil {
		C.free(p)
		result = C.uintptr_t(C.intptr_t(index))
	}
	C.dword_5d4594_826036 = nil
	*memmap.PtrUint32(0x5D4594, 826040+uintptr(4*uint32(index))) = 0
	return result
}

//export nox_xxx_guiMotdLoad_4465C0
func nox_xxx_guiMotdLoad_4465C0() *nox_window {
	root := C.nox_new_window_from_file(internCStr("motd.wnd"), unsafe.Pointer(C.sub_4466C0))
	C.dword_5d4594_826028 = root
	if root == nil {
		return nil
	}
	w := C.nox_xxx_wndGetChildByID_46B0C0(root, 4203)
	C.dword_5d4594_826032 = w
	if w == nil {
		return nil
	}
	list := (*C.nox_scrollListBox_data)(w.widget_data)
	slider := C.nox_xxx_gLoadImg_42F970(internCStr("UISlider"))
	lit := C.nox_xxx_gLoadImg_42F970(internCStr("UISliderLit"))
	thumb := C.nox_xxx_wndGetChildByID_46B0C0(root, 4204)
	down := C.nox_xxx_wndGetChildByID_46B0C0(root, 4205)
	up := C.nox_xxx_wndGetChildByID_46B0C0(root, 4206)
	if list == nil || thumb == nil || down == nil || up == nil || thumb.field_100 == nil {
		return nil
	}
	thumb.field_100.width = 16
	thumb.field_100.height = 10
	C.sub_4B5700(thumb, nil, nil, unsafe.Pointer(slider), unsafe.Pointer(lit), unsafe.Pointer(lit))
	C.nox_xxx_wnd_46B280(thumb, w)
	C.nox_xxx_wnd_46B280(down, w)
	C.nox_xxx_wnd_46B280(up, w)
	list.field_9, list.field_7, list.field_8 = unsafe.Pointer(thumb), unsafe.Pointer(down), unsafe.Pointer(up)
	return root
}

//export sub_4466C0
func sub_4466C0(unused1 C.int, event C.int, unused2 C.int, unused3 C.int) C.int {
	if event == 16391 {
		C.nox_xxx_clientPlaySoundSpecial_452D80(766, 100)
		sub_446780()
	}
	return 0
}

//export sub_4466F0
func sub_4466F0(input *C.char, output *C.uint8_t) *C.char {
	if input == nil || output == nil {
		return nil
	}
	*output = 0
	src, dst := unsafe.Pointer(input), unsafe.Pointer(output)
	for {
		ch := *(*byte)(src)
		if ch == 0 {
			*(*byte)(dst) = 0
			return nil
		}
		if ch == '\n' || ch == '\r' {
			*(*byte)(dst) = 0
			src = unsafe.Add(src, 1)
			if ch == '\r' && *(*byte)(src) == '\n' {
				src = unsafe.Add(src, 1)
			}
			return (*C.char)(src)
		}
		*(*byte)(dst) = ch
		src, dst = unsafe.Add(src, 1), unsafe.Add(dst, 1)
	}
}

//export nox_xxx_motdAddSomeTextMB_446730_go
func nox_xxx_motdAddSomeTextMB_446730_go(input *C.uint8_t) C.uintptr_t {
	if input == nil || *input == 0 {
		return C.uintptr_t(uintptr(unsafe.Pointer(input)))
	}
	// The existing formatter widens bytes, rather than interpreting UTF-8.
	// The listbox copies synchronously; release this foreign owner afterwards.
	text, free := game1ByteWide(GoString((*C.char)(unsafe.Pointer(input))))
	defer free()
	return C.nox_window_call_field_94_fnc(C.dword_5d4594_826032, 16397, C.uintptr_t(uintptr(unsafe.Pointer(&text[0]))), ^C.uintptr_t(0), internCStr("game1_split_native.go"), 0)
}

//export sub_446780
func sub_446780() C.int {
	root, list := C.dword_5d4594_826028, C.dword_5d4594_826032
	if root == nil || list == nil || C.wndIsShown_nox_xxx_wndIsShown_46ACC0(root) != 0 {
		return 0
	}
	C.nox_xxx_windowFocus_46B500(nil)
	C.nox_window_set_hidden(root, 1)
	root.flags &= ^C.nox_window_flags(8)
	list.flags &= ^C.nox_window_flags(8)
	C.nox_window_call_field_94_fnc(list, 16399, 0, 0, internCStr("game1_split_native.go"), 0)
	return 1
}

//export nox_xxx_motd_4467F0
func nox_xxx_motd_4467F0() {
	if nox_gui_xxx_check_446360() != 0 || (C.nox_xxx_isQuest_4D6F50() != 0 && bool(C.nox_common_gameFlags_check_40A5C0(128))) {
		return
	}
	root, list := C.dword_5d4594_826028, C.dword_5d4594_826032
	if !bool(C.nox_common_getEngineFlag(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)) && root != nil && list != nil &&
		C.nox_xxx_wndGetFlags_46ADA0(root)&0x10 != 0 && C.sub_44A4A0() == 0 && C.sub_49C810() == 0 && C.sub_49CB40() == 0 {
		C.nox_xxx_wndShowModalMB_46A8C0(root)
		if other := C.nox_xxx_wndGetChildByID_46B0C0(nil, 4100); other != nil {
			C.nox_xxx_wndShowModalMB_46A8C0(other)
		}
		root.flags |= 8
		list.flags |= 8
		if p := nox_xxx_xferGetBuffer_446520(1); p != nil {
			// Split without the old fixed 256-byte stack destination. Final
			// unterminated lines and empty terminated lines retain their order.
			text := GoString((*C.char)(p))
			for len(text) != 0 {
				end := strings.IndexAny(text, "\r\n")
				line := text
				if end >= 0 {
					line = text[:end]
					n := end + 1
					if text[end] == '\r' && n < len(text) && text[n] == '\n' {
						n++
					}
					text = text[n:]
					if line == "" {
						line = " "
					}
				} else {
					text = ""
				}
				if line != "" {
					nox_xxx_motdAddSomeTextMB_446730_go((*C.uint8_t)(unsafe.Pointer(internCStr(line))))
				}
			}
		}
		C.nox_xxx_windowFocus_46B500(C.nox_xxx_wndGetChildByID_46B0C0(root, 4202))
	}
	*memmap.PtrUint32(0x5D4594, 826068) = 0
}

//export sub_446950
func sub_446950() C.int {
	if w := C.dword_5d4594_826028; w != nil && C.nox_xxx_wndGetFlags_46ADA0(w)&0x10 == 0 {
		return 1
	}
	return 0
}
