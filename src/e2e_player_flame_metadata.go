package opennox

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

// GAME.EXE 004E1E3B..004E1E49 copies the raw damage-type DWORD into
// the player marker. An unowned Flame is its own terminal attacker, so
// DefaultDamage snapshots the Flame's previous position.
func e2ePlayerFlameMetadata(player, flame *server.Object) bool {
	if player == nil || flame == nil || !player.Class().Has(object.ClassPlayer) || player.UpdateData == nil {
		return false
	}
	update := player.UpdateDataPlayer()
	return update.Field76 == 2 && update.Field75 == uint32(object.DamageFlame) &&
		player.Obj130 == flame && player.Field131 == uint32(object.DamageFlame) && player.Pos132 == flame.PrevPos
}
