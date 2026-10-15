package opennox

import "github.com/opennox/opennox/v1/legacy"

// The graphical host's 00457500 path selects slot0, refreshes it from the
// running server, copies exactly58 bytes to slot1 and selects slot1. Dedicated
// MainLoop skips that window, but operator commands still edit slot1.
// Keep the existing packed C copy rather than assigning the padded Go struct.
func initDedicatedHostSettings(refresh func()) {
	sub_4165D0(0)
	refresh()
	legacy.Sub_4165F0(0, 1)
	sub_4165D0(1)
}
