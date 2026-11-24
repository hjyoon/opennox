package legacy

/*
#include "GAME3_2.h"
*/
import "C"

import (
	"bytes"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func Sub_4DF020() {
	readAndReportGUISettings4DF020()
}

//export nox_gui_settings_delta_4DF020
func nox_gui_settings_delta_4DF020() C.intptr_t {
	return C.intptr_t(readAndReportGUISettings4DF020())
}

func readAndReportGUISettings4DF020() int {
	// The GUI applies only nonempty edit fields. Start from the current
	// settings so clearing an input retains its configured value.
	settings := *(*server.Settings2)(cliGamedataNative416590(1))
	status := int(nox_game2_server_apply_459AA0(unsafe.Pointer(&settings)))
	return reportGUISettingsIfChanged4DF020(&settings, status)
}

func reportGUISettingsIfChanged4DF020(settings *server.Settings2, status int) int {
	const wireSize = 58
	wire := unsafe.Slice((*byte)(unsafe.Pointer(settings)), wireSize)
	last := memmap.Slice(0x5D4594, 1563214)[:wireSize]
	if bytes.Equal(wire, last) {
		return status
	}
	world := GetServer().S()
	for player := world.Players.First(); player != nil; player = world.Players.Next(player) {
		if player.Index() != server.HostPlayerIndex {
			Nox_xxx_netGuiGameSettings_4DD9B0(1, settings, player.Index())
		}
	}
	copy(last, wire)
	return 0
}
