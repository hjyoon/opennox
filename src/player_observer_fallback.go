package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

// playerObserverFallback_4E6150 preserves the original candidate order. The
// GameBall type cache is initialized before the player is read, and exhausting
// an initial player scan falls back through GameBall before scanning from the
// first player unit again.
func playerObserverFallback_4E6150(
	pl *server.Player,
	ensureGameBallID func(),
	hasFlagBall func() bool,
	findGameBall func() *server.Object,
	firstPlayerUnit func() *server.Object,
	nextPlayerUnit func(*server.Object) *server.Object,
	playerByNetCode func(uint32) *server.Player,
) *server.Object {
	ensureGameBallID()
	camera := pl.CameraFollowObj

	var candidate *server.Object
	switch {
	case camera != nil && camera.ObjClass.Has(object.ClassPlayer):
		candidate = nextPlayerUnit(camera)
	case camera == nil && hasFlagBall():
		candidate = findGameBall()
		if candidate == nil {
			candidate = firstPlayerUnit()
		}
	default:
		candidate = firstPlayerUnit()
	}

	if candidate != nil {
		if !candidate.ObjClass.Has(object.ClassPlayer) {
			return candidate
		}
		if candidate = playerObserverUsableUnit_4E6150(candidate, nextPlayerUnit, playerByNetCode); candidate != nil {
			return candidate
		}
	}
	if ball := findGameBall(); ball != nil {
		return ball
	}
	return playerObserverUsableUnit_4E6150(firstPlayerUnit(), nextPlayerUnit, playerByNetCode)
}

func playerObserverUsableUnit_4E6150(
	unit *server.Object,
	nextPlayerUnit func(*server.Object) *server.Object,
	playerByNetCode func(uint32) *server.Player,
) *server.Object {
	for unit != nil {
		pl := playerByNetCode(unit.NetCode)
		if unit.ObjFlags&object.FlagDead == 0 && pl.Field3680&1 == 0 {
			return unit
		}
		unit = nextPlayerUnit(unit)
	}
	return nil
}
