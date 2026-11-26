package legacy

/*
#include "client__gui__servopts__guiserv.h"
#include "client__gui__window.h"
#include "noxstring.h"
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

var serverLadderLabelFormat = internWStr("%s %s")

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
		formatFixedWide(text, serverLadderLabelFormat, formatValue{text: unsafe.Pointer(ladder)}, formatValue{text: unsafe.Pointer(kind)})
	case noxflags.HasGame(serverOptionsClanLadder):
		kind := g2ServerLoad("Servopts.wnd:Clan", serverOptionsStringsFile, 385)
		ladder := g2ServerLoad("Servopts.wnd:Ladder", serverOptionsStringsFile, 384)
		formatFixedWide(text, serverLadderLabelFormat, formatValue{text: unsafe.Pointer(ladder)}, formatValue{text: unsafe.Pointer(kind)})
	default:
		formatFixedWide(text, g2ServerLoad("WindowDir:Empty", serverOptionsStringsFile, 389))
	}
	return putServerOptionsLabel(10210, text)
}

//export sub_4580E0
func sub_4580E0(name *C.char) C.int {
	settings := serverOptionsText(1045968)
	formatFixedWide(settings, g2ServerLoad("SettingsMsg", serverOptionsStringsFile, 736), formatValue{text: unsafe.Pointer(name)})
	mode := serverOptionsText(1046096)
	if noxflags.HasGame(noxflags.GameModeChat) {
		formatFixedWide(mode, g2ServerLoad("GameType", serverOptionsStringsFile, 739))
	} else {
		title := nox_xxx_guiServerOptionsGetGametypeName_4573C0(C.short(noxflags.GetGame()))
		file := GoString((*C.char)(g2ServerMemOff(0x587000, 131072)))
		formatFixedWide(mode, g2ServerLoad("GameTypeIs", file, 743), formatValue{text: unsafe.Pointer(title)})
	}
	message := serverOptionsText(1046224)
	if noxflags.HasGame(noxflags.GameHost) {
		formatFixedWide(message, g2ServerLoad("GoMessage", serverOptionsStringsFile, 748))
	} else {
		formatFixedWide(message, g2ServerLoad("OptsMessage", serverOptionsStringsFile, 752))
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
	formatFixedWide(text, g2ServerLoad("NumTeamsMsg", serverOptionsStringsFile, 2071), formatValue{number: int32(count)})
	return putServerOptionsLabel(10110, text)
}
