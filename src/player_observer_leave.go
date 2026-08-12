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

// playerLeaveMonsterObserver_4E60E0 keeps the original reload points for the
// player's unit. Selection callbacks may change PlayerUnit, and the unlock and
// follow calls must receive the value current at their respective call sites.
func playerLeaveMonsterObserver_4E60E0(
	pl *server.Player,
	getPossess func(*server.Object) *server.Object,
	findGoodSlave func(*server.Player) *server.Object,
	clearObserve func(*server.Object),
	findFallback func(*server.Player) *server.Object,
	unlock func(*server.Object),
	follow func(*server.Object, *server.Object),
) {
	unit := pl.PlayerUnit
	if unit == nil {
		return
	}
	var target *server.Object
	if getPossess(unit) != nil {
		target = findGoodSlave(pl)
		if target == nil {
			clearObserve(pl.PlayerUnit)
			return
		}
	} else {
		target = findFallback(pl)
	}
	unlock(pl.PlayerUnit)
	follow(pl.PlayerUnit, target)
}
