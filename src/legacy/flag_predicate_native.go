package legacy

import (
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/server"
)

func windowFlags46ADA0(win *gui.Window) int32 {
	if win == nil {
		return -2
	}
	// The C entry reads the flags even when the window is marked destroyed.
	return int32(win.Flags)
}

func waypointFlagMask579EE0(wp *server.Waypoint, mask byte) bool {
	return wp != nil && wp.HasFlag2Mask(mask)
}
