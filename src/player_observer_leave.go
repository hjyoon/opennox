package opennox

import "github.com/opennox/opennox/v1/server"

// playerLeaveObsByObserved_4E60A0 walks the player list in its original order.
// It reads each camera target on arrival and asks for the successor only after
// the matching player's leave callback has completed.
func playerLeaveObsByObserved_4E60A0(
	target *server.Object,
	first func() *server.Player,
	next func(*server.Player) *server.Player,
	leave func(*server.Player),
) {
	for pl := first(); pl != nil; pl = next(pl) {
		if pl.CameraFollowObj == target {
			leave(pl)
		}
	}
}
