package legacy

/*
#include "client__gui__guiggovr.h"
#include "client__gui__window.h"

extern int nox_win_width, nox_win_height;

static void nox_quest_over_format_count(wchar2_t* out, const wchar2_t* format, int count) {
	nox_swprintf(out, format, count);
}
static void nox_quest_over_format_text(wchar2_t* out, const wchar2_t* format) {
	nox_swprintf(out, format);
}
*/
import "C"

import (
	"encoding/binary"
	"unsafe"
)

//export sub_49B4B0
func sub_49B4B0(raw *C.ushort) C.int {
	g2ServerHidden(questGameOverWindow, false)
	g2ServerEnabled(questGameOverWindow, true)
	nox_xxx_clientPlaySoundSpecial_452D80(1007, 100)
	var width, height C.int
	nox_window_get_size(questGameOverWindow, &width, &height)
	nox_window_setPos_46A9B0(questGameOverWindow, int32(C.nox_win_width/2-width/2), int32(C.nox_win_height/2-height/2))
	words := unsafe.Slice((*byte)(unsafe.Pointer(raw)), 8)
	const source = `C:\NoxPost\src\client\Gui\GUIGGOvr.c`
	counts := [...]struct {
		key    string
		line   int
		offset uintptr
	}{
		{"GGOver.wnd:GeneratorsDestroyed", 178, 1302172},
		{"GGOver.wnd:NumSecretsFound", 181, 1301916},
		{"GGOver.wnd:Kills", 183, 1302428},
	}
	for i, count := range counts {
		value := binary.LittleEndian.Uint16(words[2+2*i : 4+2*i])
		text := g2ServerLoad(count.key, source, count.line)
		C.nox_quest_over_format_count((*C.wchar2_t)(g2ServerMem(count.offset)), text, C.int(value))
	}
	C.nox_quest_over_format_text((*C.wchar2_t)(g2ServerMem(1303196)), (*C.wchar2_t)(g2ServerMem(1303460)))
	for _, field := range [...]struct {
		id     int
		offset uintptr
	}{{10710, 1302172}, {10705, 1302940}, {10706, 1302684}, {10707, 1301916}, {10708, 1302428}, {10711, 1303196}} {
		g2ServerText(g2ServerChild(questGameOverWindow, field.id), 16385, g2ServerMem(field.offset), 0)
	}
	result := C.int(gameFrame())
	*g2ServerWord(1303456) = gameFrame()
	return result
}

//export sub_49B6E0
func sub_49B6E0() C.int { return nox_client_questGameOverTimer_native_49B6E0() }
