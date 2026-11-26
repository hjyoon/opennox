package legacy

/*
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "noxstring.h"


static void format_server_option_text(wchar2_t* out, const wchar2_t* text) {
	nox_swprintf(out, text);
}

static void format_server_option_ladder(wchar2_t* out, const wchar2_t* ladder, const wchar2_t* kind) {
	nox_swprintf(out, L"%s %s", ladder, kind);
}

static void format_server_option_name(wchar2_t* out, const wchar2_t* format, const char* name) {
	nox_swprintf(out, format, name);
}

static void format_server_option_mode(wchar2_t* out, const wchar2_t* format, const wchar2_t* mode) {
	nox_swprintf(out, format, mode);
}

static void format_server_option_count(wchar2_t* out, const wchar2_t* format, int teams) {
	nox_swprintf(out, format, teams);
}
*/
import "C"

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const (
	serverOptionsIndividualLadder = noxflags.GameFlag(0x4000)
	serverOptionsClanLadder       = noxflags.GameFlag(0x8000)
)

func serverOptionsText(offset uintptr) *C.wchar2_t {
	return (*C.wchar2_t)(g2ServerMem(offset))
}

func putServerOptionsLabel(id int32, text *C.wchar2_t) C.int {
	window := nox_xxx_wndGetChildByID_46B0C0(serverSharedWindows.options, id)
	return C.int(g2ServerText(window, 16385, unsafe.Pointer(text), ^uintptr(0)))
}

//export sub_457FE0
func sub_457FE0() C.int {
	text := serverOptionsText(1045700)
	switch {
	case noxflags.HasGame(serverOptionsIndividualLadder):
		kind := g2ServerLoad("Servopts.wnd:Individual", serverOptionsStringsFile, 380)
		ladder := g2ServerLoad("Servopts.wnd:Ladder", serverOptionsStringsFile, 379)
		C.format_server_option_ladder(text, ladder, kind)
	case noxflags.HasGame(serverOptionsClanLadder):
		kind := g2ServerLoad("Servopts.wnd:Clan", serverOptionsStringsFile, 385)
		ladder := g2ServerLoad("Servopts.wnd:Ladder", serverOptionsStringsFile, 384)
		C.format_server_option_ladder(text, ladder, kind)
	default:
		C.format_server_option_text(text, g2ServerLoad("WindowDir:Empty", serverOptionsStringsFile, 389))
	}
	return putServerOptionsLabel(10210, text)
}

//export sub_4580E0
func sub_4580E0(name *C.char) C.int {
	settings := serverOptionsText(1045968)
	C.format_server_option_name(settings, g2ServerLoad("SettingsMsg", serverOptionsStringsFile, 736), name)
	mode := serverOptionsText(1046096)
	if noxflags.HasGame(noxflags.GameModeChat) {
		C.format_server_option_text(mode, g2ServerLoad("GameType", serverOptionsStringsFile, 739))
	} else {
		title := nox_xxx_guiServerOptionsGetGametypeName_4573C0(C.short(noxflags.GetGame()))
		file := GoString((*C.char)(g2ServerMemOff(0x587000, 131072)))
		C.format_server_option_mode(mode, g2ServerLoad("GameTypeIs", file, 743), title)
	}
	message := serverOptionsText(1046224)
	if noxflags.HasGame(noxflags.GameHost) {
		C.format_server_option_text(message, g2ServerLoad("GoMessage", serverOptionsStringsFile, 748))
	} else {
		C.format_server_option_text(message, g2ServerLoad("OptsMessage", serverOptionsStringsFile, 752))
	}
	putServerOptionsLabel(10121, settings)
	putServerOptionsLabel(10118, mode)
	return putServerOptionsLabel(10117, message)
}

//export sub_459CD0
func sub_459CD0() C.int {
	if serverSharedWindows.options == nil {
		return 0
	}
	var count uint8
	if sub_40A740() != 0 || noxflags.HasGame(serverOptionsClanLadder) {
		count = uint8(sub_417DE0())
	} else {
		count = uint8(nox_xxx_getTeamCounter_417DD0())
	}
	text := serverOptionsText(1046364)
	C.format_server_option_count(text, g2ServerLoad("NumTeamsMsg", serverOptionsStringsFile, 2071), C.int(count))
	return putServerOptionsLabel(10110, text)
}
