package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// The original movable gate precedes the first origin-coordinate read.
// Keep the lazy pointer boundary in the C export as well as the common path.
func applyForceExport52DF80(unit *server.Object, origin unsafe.Pointer, force float32) {
	if !unit.IsMovable() {
		return
	}
	GetServer().ApplyForce(unit, AsPointf(origin), float64(force))
}
