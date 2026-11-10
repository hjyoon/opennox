package legacy

/*
#include "game3_widgets_native.h"
*/
import "C"

import (
	"fmt"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"math"
	"unsafe"
)

// Records remain C-owned packed wire payloads; this private registry never
// stores native links or window pointers in their DWORD fields.
var game3NearbyServers [2500]*C.nox_gui_server_ent_t

func game3NearbyDistance(xy *uint32, x, y int16, radius float64) bool {
	p := unsafe.Slice(xy, 2)
	dx, dy := float64(int32(uint32(int32(x))-p[0])), float64(int32(uint32(int32(y))-p[1]))
	return math.Sqrt(dy*dy+dx*dx) <= radius
}

//export nox_game3_distance_go
func nox_game3_distance_go(xy *C.uint32_t, s *C.nox_gui_server_ent_t) C.int {
	if xy == nil || s == nil {
		return 0
	}
	// The record is packed. Copy the two scalar coordinates without an aligned
	// native short load; the record's fixed 169-byte format is not a native node.
	b := unsafe.Slice((*byte)(unsafe.Pointer(s)), 169)
	x, y := int16(uint16(b[44])|uint16(b[45])<<8), int16(uint16(b[46])|uint16(b[47])<<8)
	if game3NearbyDistance((*uint32)(unsafe.Pointer(xy)), x, y, *(*float64)(memmap.PtrOff(0x581450, 9720))) {
		return 1
	}
	return 0
}

//export sub_4A25C0
func sub_4A25C0(xy *C.uint32_t, head *C.nox_list_item_t) C.int {
	var count uint32
	for it := C.nox_common_list_getFirstSafe_425890(head); it != nil; it = C.nox_common_list_getNextSafe_4258A0(it) {
		if nox_game3_distance_go(xy, C.nox_game3_server_record(it)) != 0 {
			count++
		}
	}
	return C.int(int32(count))
}

//export sub_4A2610
func sub_4A2610(owner *C.nox_window, xy *C.uint32_t, head *C.nox_list_item_t) *C.nox_window {
	C.dword_5d4594_1307720 = 0
	for it := C.nox_common_list_getFirstSafe_425890(head); it != nil; it = C.nox_common_list_getNextSafe_4258A0(it) {
		s := C.nox_game3_server_record(it)
		if nox_game3_distance_go(xy, s) != 0 && int(C.dword_5d4594_1307720) < len(game3NearbyServers) {
			game3NearbyServers[int(C.dword_5d4594_1307720)] = s
			C.dword_5d4594_1307720++
		}
	}
	if C.dword_5d4594_1307720 == 0 {
		return C.dword_5d4594_1307716
	}
	popup := C.nox_new_window_from_file(internCStr("proxlist.wnd"), C.nox_game3_owner_callback(owner))
	C.dword_5d4594_1307716 = popup
	if popup == nil {
		return nil
	}
	if xy == nil {
		sub_4A2890()
		return nil
	}
	p := unsafe.Slice(xy, 2)
	var pos [2]C.uint32_t
	sub_4A2830(C.int(int32(uint32(p[0])+216)), C.int(int32(uint32(p[1])+27)), &pos[0])
	C.nox_window_setPos_46A9B0(popup, C.int(pos[0]), C.int(pos[1]))
	C.nox_xxx_wnd_46B280(popup, owner)
	child := func(id int) *C.nox_window { return C.nox_xxx_wndGetChildByID_46B0C0(popup, C.int(id)) }
	slider, up, down, list := child(10064), child(10062), child(10063), child(10061)
	if slider == nil || up == nil || down == nil || list == nil || list.widget_data == nil || slider.field_100 == nil {
		sub_4A2890()
		return nil
	}
	normal, lit := C.nox_xxx_gLoadImg_42F970(internCStr("UISlider")), C.nox_xxx_gLoadImg_42F970(internCStr("UISliderLit"))
	C.sub_4B5700(slider, nil, nil, unsafe.Pointer(normal), unsafe.Pointer(lit), unsafe.Pointer(lit))
	C.nox_xxx_wnd_46B280(slider, list)
	C.nox_xxx_wnd_46B280(up, list)
	C.nox_xxx_wnd_46B280(down, list)
	data := (*gui.ScrollListBoxData)(list.widget_data)
	data.Field_9 = unsafe.Pointer(slider)
	data.Field_7 = unsafe.Pointer(up)
	data.Field_8 = unsafe.Pointer(down)
	slider.field_100.width = 16
	slider.field_100.height = 10
	for i := 0; i < int(C.dword_5d4594_1307720); i++ {
		s := game3NearbyServers[i]
		raw := unsafe.Slice((*byte)(unsafe.Pointer(s)), 169)
		var name string
		if raw[120] != 0 {
			n := 0
			for n < 15 && raw[120+n] != 0 {
				n++
			}
			name = string(raw[120 : 120+n])
		} else {
			var buf [64]C.char
			C.nox_sprintAddrPort_43BC80((*C.char)(unsafe.Add(unsafe.Pointer(s), 12)), C.ushort(uint16(raw[109])|uint16(raw[110])<<8), &buf[0])
			name = C.GoString(&buf[0])
		}
		text, free := alloc.CString16(fmt.Sprintf("%s   %dms", name, int32(uint32(raw[96])|uint32(raw[97])<<8|uint32(raw[98])<<16|uint32(raw[99])<<24)))
		game3WidgetEvent(list, 16397, uintptr(unsafe.Pointer(text)), ^uintptr(0))
		free()
	}
	return popup
}

func game3NearbyPosition(x, y int32) (uint32, uint32) {
	a, b := uint32(x-100), uint32(y-20)
	if x-100+200 > 600 {
		a = 400
	}
	if y-20+200 > 451 {
		b = 251
	}
	if b < 27 {
		b = 27
	}
	if int32(a) < 216 {
		a = 216
	}
	return a, b
}

//export sub_4A2830
func sub_4A2830(x, y C.int, out *C.uint32_t) *C.uint32_t {
	if out != nil {
		a, b := game3NearbyPosition(int32(x), int32(y))
		p := unsafe.Slice(out, 2)
		p[0] = C.uint32_t(a)
		p[1] = C.uint32_t(b)
	}
	return out
}

//export sub_4A2890
func sub_4A2890() C.int {
	var result C.int
	if C.dword_5d4594_1307716 != nil {
		result = C.nox_xxx_windowDestroyMB_46C4E0(C.dword_5d4594_1307716)
		C.dword_5d4594_1307716 = nil
	}
	C.dword_5d4594_1307720 = 0
	clear(game3NearbyServers[:])
	return result
}

//export sub_4A28B0
func sub_4A28B0() C.int {
	if C.dword_5d4594_1307716 != nil {
		return 1
	}
	return 0
}

//export sub_4A28C0
func sub_4A28C0(i C.int) *C.nox_gui_server_ent_t {
	if i >= 0 && i < C.int(C.dword_5d4594_1307720) && int(i) < len(game3NearbyServers) {
		return game3NearbyServers[int(i)]
	}
	return nil
}
