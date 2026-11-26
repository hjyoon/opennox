package legacy

/*
#include "game2_server_bridge.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const serverAccessStrings = `C:\NoxPost\src\client\Gui\ServOpts\access.c`

var serverAccessClassLabels = [...]string{"WARRIOR", "WIZARD", "CONJURER"}

//export sub_454640
func sub_454640() C.int {
	list := serverAccessState.wnd_10123
	if list == nil {
		return 0
	}
	list.height = C.intptr_t(4*(nox_xxx_guiFontHeightMB_43F320(list.draw_data.font)+1) + 2)
	list.end_y = list.off_y + list.height
	lines := [...]int{88, 89, 94}
	var widest C.int
	for i, label := range serverAccessClassLabels {
		text := g2ServerLoad(label, serverAccessStrings, lines[i])
		var width C.int
		nox_xxx_drawGetStringSize_43F840(list.draw_data.font, text, &width, nil, 0)
		widest = max(widest, width)
	}
	list.width = C.intptr_t(widest) + 7
	list.end_x = list.off_x + list.width
	return C.int(list.end_x)
}

func setServerAccessLatency(control, checkbox *C.nox_window, value uint16) {
	if value != 0xffff {
		g2ServerEnabled(control, true)
		checkbox.draw_data.field_0 |= 4
		g2ServerIntText(control, int32(value), 0)
	}
}

//export sub_454740
func sub_454740() *C.int {
	s := serverAccessState
	settings := unsafe.Slice((*byte)(sub_416640()), 109)
	password := g2ServerChild(s.root, 10136)
	g2ServerText(password, 16414, unsafe.Pointer(nox_xxx_sysopGetPass_40A630()), 0)
	setServerAccessLatency(s.wnd_10130, s.wnd_10129, binary.LittleEndian.Uint16(settings[105:107]))
	setServerAccessLatency(s.wnd_10132, s.wnd_10131, binary.LittleEndian.Uint16(settings[107:109]))
	latency := g2ServerChild(s.root, 10124)
	if int8(settings[102]) < 0 {
		latency.draw_data.field_0 |= 4
	}
	if settings[100]&0x20 != 0 {
		g2ServerEnabled(s.wnd_10104, true)
		s.wnd_10103.draw_data.field_0 |= 4
	}
	g2ServerText(s.wnd_10104, 16414, unsafe.Pointer(&settings[78]), 0)
	if sub_4D6F30() != 0 {
		g2ServerEnabled(s.wnd_10102, false)
	} else {
		g2ServerEnabled(s.wnd_10102, true)
		if settings[100]&0x10 != 0 {
			s.wnd_10102.draw_data.field_0 = 4
		} else {
			g2ServerEnabled(g2ServerChild(s.root, 10206), false)
		}
	}
	g2ServerChild(s.root, 10207).draw_data.field_0 |= 4
	s.active_list = s.wnd_10105
	for i, label := range serverAccessClassLabels {
		text := g2ServerLoad(label, serverAccessStrings, 242+i)
		g2ServerText(s.wnd_10123, 16397, unsafe.Pointer(text), ^uintptr(0))
	}
	if settings[100]&0x10 != 0 {
		g2ServerHidden(s.wnd_10109, false)
		g2ServerHidden(s.wnd_10105, true)
	}
	if mask := settings[100]; mask != 0 {
		if data := g2ServerList(s.wnd_10123); data != nil && data.Field_12 != nil {
			selected := unsafe.Slice(data.Field_12, 4)
			var count int
			for i := range serverAccessClassLabels {
				if mask&(1<<i) != 0 {
					selected[count] = uint32(i)
					count++
					selected[count] = ^uint32(0)
				}
			}
		}
	}
	g2ServerIntText(s.wnd_10133, int32(settings[104]), 0)
	for player := nox_common_playerInfoGetFirst_416EA0(); player != nil; player = nox_common_playerInfoGetNext_416EE0(player) {
		if player.playerInd != 31 || !noxflags.HasEngine(noxflags.EngineNoRendering) {
			sub_455920(&player.name_final[0])
		}
	}
	return sub_455800()
}
