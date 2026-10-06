package opennox

import "github.com/opennox/opennox/v1/server"

// aiDependencyEnemyDistance546E8E reads the live CurrentEnemy through
// 00546A83's entry-cached update, not the unit's replacement update or an
// object argument slot. GAME.EXE keeps 004E6C00's retained surface distance
// until FCOMP and reads the radius only after that call has completed.
func aiDependencyEnemyDistance546E8E(unit *server.Object, update *server.MonsterUpdateData, slot *server.AIStackItem, closer bool) bool {
	enemy := update.CurrentEnemy
	if enemy == nil {
		return !closer // 00546B40 fails close; 00546EA1 passes far.
	}
	distance := aiDependencyObjectSurfaceDistance546B13(unit, enemy)
	radius := float64(slot.ArgF32(0))
	if closer {
		return !(distance > radius) // 00546B55: less/equal/unordered pass.
	}
	return distance >= radius // 00546EB6 tests C0 only: unordered fails.
}
