package legacy

/*
#include "cgo_const.h"
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "noxstring.h"

extern nox_window* dword_5d4594_1046492;
extern nox_window* dword_5d4594_1046496;
extern nox_window* dword_5d4594_1046500;
extern nox_window* dword_5d4594_1046504;
extern nox_window* dword_5d4594_1046516;
extern nox_window* dword_5d4594_1046520;

static void format_server_option_map_row(wchar2_t* out, const wchar2_t* format, const char* map, int min, int max) {
	nox_swprintf(out, format, map, min, max);
}

static void narrow_server_option_map_row(char* out, const wchar2_t* text) {
	nox_sprintf(out, "%S", text);
}
*/
import "C"

import (
	"strings"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func serverOptionMapRowColor(name string, mode uint16) uintptr {
	var standard server.Settings2
	alloc.StrCopy(standard.Field0[:], name)
	user := standard
	readMapRuleDefaults57A1E0(&standard, nil, nil, ruleSourceMap, mode)
	readMapRuleDefaults57A1E0(&user, internCStr("user.rul"), nil, ruleSourceMap|ruleSourceUser, mode)
	if standard.Field24 != user.Field24 || standard.Field44 != user.Field44 || standard.Field48 != user.Field48 {
		return 6
	}
	return ^uintptr(0)
}

func serverOptionMapNameFromRow(text *C.wchar2_t) string {
	if text == nil {
		return ""
	}
	var narrow [200]C.char
	C.narrow_server_option_map_row(&narrow[0], text)
	name, _, _ := strings.Cut(strings.TrimLeft(GoString(&narrow[0]), "\t"), "\t")
	return name
}

//export nox_client_guiserv_updateMapList_458230
func nox_client_guiserv_updateMapList_458230(mode C.int, current *C.char, apply C.bool) {
	list := C.dword_5d4594_1046496
	g2ServerEvent(list, 16399, 0, 0)
	*g2ServerWord(1046552) = uint32(mode)
	matched := false
	row := uintptr(0)
	var title [100]C.wchar2_t
	for mapInfo := nox_common_maplist_first_4D09B0(); mapInfo != nil; mapInfo = nox_common_maplist_next_4D09C0(mapInfo) {
		if mapInfo.field_6 == 0 || sub_4CFFC0(mapInfo)&mode == 0 {
			continue
		}
		name := &mapInfo.name[0]
		color := serverOptionMapRowColor(GoString(name), uint16(mode))
		format := g2ServerLoad("RecPlayers", serverOptionsStringsFile, 823)
		C.format_server_option_map_row(&title[0], format, name, C.int(mapInfo.field_8_0), C.int(mapInfo.field_8_1))
		g2ServerText(list, 16397, unsafe.Pointer(&title[0]), color)
		if C.nox_strcmpi(current, name) == 0 {
			matched = true
			g2ServerEvent(list, 16403, row, 0)
			g2ServerEvent(list, 16412, row, 0)
		}
		row++
	}
	if !matched {
		g2ServerEvent(list, 16403, 0, 0)
		g2ServerEvent(list, 16412, 0, 0)
	}
	if !bool(apply) {
		return
	}

	record := sub_4165B0()
	settings := (*server.Settings2)(unsafe.Pointer(record))
	name := ""
	selected := g2ServerEvent(list, 16404, 0, 0)
	if int(selected) >= 0 {
		name = serverOptionMapNameFromRow(listBoxRowText(list, selected))
	}
	alloc.StrCopy(settings.Field0[:9], name)
	readMapRuleDefaults57A1E0(settings, internCStr("user.rul"), nil, ruleSourceMap|ruleSourceUser|ruleSourceInternet, uint16(mode))
	sub_459880(record)
}

//export sub_459880
func sub_459880(data *C.char) C.int {
	mode := noxflags.GameFlag(g2ServerMode(data))
	unlocked := *(*byte)(g2ServerAt(data, 57)) == 0
	labelKey, labelLine := "Servopts.wnd:KillLimit", 1926
	switch {
	case mode.Has(noxflags.GameModeCTF):
		labelKey, labelLine = "Servopts.wnd:CaptureLimit", 1900
	case mode.Has(noxflags.GameModeElimination):
		labelKey, labelLine = "Servopts.wnd:DeathLimit", 1912
	}
	if unlocked && g2ServerGame(uint32(noxflags.GameHost)) &&
		(mode&(noxflags.GameModeCTF|noxflags.GameModeElimination) != 0 ||
			!g2ServerGame(uint32(serverOptionsIndividualLadder|serverOptionsClanLadder))) {
		g2ServerEnabled(C.dword_5d4594_1046516, true)
		g2ServerEnabled(C.dword_5d4594_1046520, true)
	}
	title := g2ServerLoad(labelKey, serverOptionsStringsFile, labelLine)
	C.nox_wcscpy(&C.dword_5d4594_1046516.draw_data.text[0], title)
	sub_4580E0(data)
	sub_459A40((*C.char)(g2ServerAt(data, 9)))
	if g2ServerGame(uint32(noxflags.GameHost)) && !g2ServerGame(uint32(serverOptionsIndividualLadder|serverOptionsClanLadder)) {
		g2ServerEnabled(C.dword_5d4594_1046500, *(*byte)(g2ServerAt(data, 57)) == 0)
		g2ServerEnabled(C.dword_5d4594_1046504, *(*byte)(g2ServerAt(data, 57)) == 0)
	}
	toggle := nox_xxx_wndGetChildByID_46B0C0(C.dword_5d4594_1046492, 10122)
	if *(*byte)(g2ServerAt(data, 57)) != 0 {
		toggle.draw_data.field_0 |= 4
	} else {
		toggle.draw_data.field_0 &^= 4
	}
	sub_453F70((*C.nox_cgo_const_void)(g2ServerAt(data, 24)))
	sub_4535E0((*C.int)(g2ServerAt(data, 44)))
	return sub_4535F0(*(*C.int)(g2ServerAt(data, 48)))
}
