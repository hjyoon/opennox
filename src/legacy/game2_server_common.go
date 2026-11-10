package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"runtime"
	"unsafe"
)

type g2ServerWindow = C.nox_window
type g2ServerDraw = C.nox_window_data

func g2ServerWord(off uintptr) *uint32                     { return memmap.PtrUint32(0x5D4594, off) }
func g2ServerByte(off uintptr) *byte                       { return memmap.PtrUint8(0x5D4594, off) }
func g2ServerMem(off uintptr) unsafe.Pointer               { return memmap.PtrOff(0x5D4594, off) }
func g2ServerMemOff(base, off uintptr) unsafe.Pointer      { return memmap.PtrOff(base, off) }
func g2ServerPointerSlot(base, off uintptr) unsafe.Pointer { return *memmap.PtrPtr(base, off) }
func g2ServerGame(flag uint32) bool                        { return bool(C.nox_common_gameFlags_check_40A5C0(C.uint(flag))) }
func g2ServerEngine(flag C.nox_engine_flag) bool           { return bool(C.nox_common_getEngineFlag(flag)) }
func g2ServerBool(v bool) C.int {
	if v {
		return 1
	}
	return 0
}
func g2ServerChild(w *g2ServerWindow, id int) *g2ServerWindow {
	return C.nox_xxx_wndGetChildByID_46B0C0(w, C.int(id))
}
func g2ServerEvent(w *g2ServerWindow, e int, a, b uintptr) uintptr {
	return uintptr(C.game2_server_event(w, C.int(e), C.uintptr_t(a), C.uintptr_t(b)))
}
func g2ServerText(w *g2ServerWindow, e int, p unsafe.Pointer, b uintptr) uintptr {
	v := g2ServerEvent(w, e, uintptr(p), b)
	runtime.KeepAlive(p)
	return v
}
func g2ServerEnabled(w *g2ServerWindow, v bool) C.int {
	return C.nox_xxx_wnd_46ABB0(w, g2ServerBool(v))
}
func g2ServerHidden(w *g2ServerWindow, v bool) C.int {
	return C.nox_window_set_hidden(w, g2ServerBool(v))
}
func g2ServerList(w *g2ServerWindow) *gui.ScrollListBoxData {
	if w == nil {
		return nil
	}
	return (*gui.ScrollListBoxData)(w.widget_data)
}
func g2ServerItems(d *gui.ScrollListBoxData) []gui.ScrollListBoxItem {
	if d == nil || d.Items == nil {
		return nil
	}
	return unsafe.Slice(d.Items, int(d.Count))
}
func g2ServerFirstVisible(d *gui.ScrollListBoxData) int {
	if d == nil || d.Items == nil || d.Field_11_0 == 0 {
		return 0
	}
	items := g2ServerItems(d)
	scroll := int(d.Field_13_1)
	for i := 0; i < int(d.Field_11_0); i++ {
		if int(int32(items[i].Field_0)) > scroll {
			return i
		}
	}
	return 0
}
func g2ServerName(p unsafe.Pointer) *C.wchar2_t { return (*C.wchar2_t)(p) }
func g2ServerCString(s string) *C.char          { return internCStr(s) }
func g2ServerTemp(units int) ([]uint16, func()) { return alloc.Make([]uint16{}, units) }
func g2ServerIntText(w *g2ServerWindow, value int32, sentinel uintptr) {
	b, free := g2ServerTemp(16)
	defer free()
	C.nox_itow(C.int(value), (*C.wchar2_t)(unsafe.Pointer(&b[0])), 10)
	g2ServerText(w, 16414, unsafe.Pointer(&b[0]), sentinel)
}
func g2ServerLoad(key, path string, line int) *C.wchar2_t {
	return C.nox_strman_loadString_40F1D0(internCStr(key), nil, internCStr(path), C.int(line))
}
func g2ServerPlayerByName(name *C.wchar2_t) *C.nox_playerInfo {
	if name == nil {
		return nil
	}
	for p := C.nox_common_playerInfoGetFirst_416EA0(); p != nil; p = C.nox_common_playerInfoGetNext_416EE0(p) {
		if C.game2_server_icmp(&p.name_final[0], name) == 0 {
			return p
		}
	}
	return nil
}
func g2ServerDrawBackground(w *C.nox_window, d *C.nox_window_data, requireImage bool) C.int {
	var x, y C.uint
	C.nox_client_wndGetPosition_46AA60(w, &x, &y)
	if int8(w.flags) >= 0 {
		if uint32(d.bg_color) != 0x80000000 {
			C.nox_client_drawRectFilledAlpha_49CF10(C.int(x), C.int(y), C.int(w.width), C.int(w.height))
		}
	} else if !requireImage || d.bg_image != nil {
		C.nox_client_drawImageAt_47D2C0((*C.nox_video_bag_image_t)(d.bg_image), C.int(x), C.int(y))
	}
	return 1
}
