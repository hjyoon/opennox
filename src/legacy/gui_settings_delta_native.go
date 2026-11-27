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
	sub_4DF020()
}

//export sub_4DF020
func sub_4DF020() C.int {
	// The GUI applies only nonempty edit fields. Start from the current
	// settings so clearing an input retains its configured value.
	settings := *(*server.Settings2)(cliGamedataNative416590(1))
	status := int32(game2ServerApply459AA0(unsafe.Pointer(&settings)))
	return C.int(reportGUISettingsIfChanged4DF020(&settings, status))
}

func reportGUISettingsIfChanged4DF020(settings *server.Settings2, status int32) int32 {
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
