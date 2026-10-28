package opennox

import (
	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/server"
)

// Diagnostic copies distinguish the object aimed at by the input helper from
// the target actually accepted by the original game's pending Swap duration.
// No native pointer or input sink survives a read.
type e2eMonsterQuickbarSwapTarget struct {
	id               uint16
	created, trigger uint32
	wire             uint16
	player, target   e2eTargetObjectObservation
}

func e2eReadMonsterQuickbarSwapTarget(frame, castFrame uint32, unit *server.Object, head *server.DurSpell, world *server.Server) (e2eMonsterQuickbarSwapTarget, bool) {
	if world == nil || unit == nil || !unit.Class().Has(object.ClassPlayer) ||
		unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
		frame-castFrame > 120 {
		return e2eMonsterQuickbarSwapTarget{}, false
	}
	for record, reads := head, 0; record != nil && reads < 4096; record, reads = record.Next, reads+1 {
		if record.Spell != uint32(spell.SPELL_SWAP) || record.Caster16 != unit || record.Frame60-castFrame > frame-castFrame {
			continue
		}
		target := record.Target48
		if target == nil || !target.Class().Has(object.ClassMonster) || target.HealthData == nil || target.HealthData.Cur == 0 ||
			target.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			continue
		}
		wire := world.GetUnitNetCode(target)
		if wire <= 0 || wire > 0xffff {
			continue
		}
		return e2eMonsterQuickbarSwapTarget{id: record.ID, created: record.Frame60, trigger: record.Frame68 - 1,
			wire: uint16(wire), player: e2eReadTargetObject(unit), target: e2eReadTargetObject(target)}, true
	}
	return e2eMonsterQuickbarSwapTarget{}, false
}

func e2eObserveMonsterQuickbarSwapTarget(frame, castFrame uint32, playerID int32, input e2eMonsterInputIdentity, key keybind.Key) {
	if key != keybind.KeyF || noxServer == nil {
		return
	}
	unit := noxServer.Players.HostUnit()
	if unit == nil || unit.ScriptIDVal != playerID {
		return
	}
	world := noxServer.S()
	actual, found := e2eReadMonsterQuickbarSwapTarget(frame, castFrame, unit, world.Spells.Dur.List, world)
	if !found {
		return
	}
	e2eLog.Printf("MONSTER QUICKBAR SWAP TARGET: {\"frame\":%d,\"cast_frame\":%d,\"player_script_id\":%d,\"key\":%d,\"spell\":%d,\"duration_id\":%d,\"created_frame\":%d,\"trigger_frame\":%d,\"input_wire\":%d,\"input_extent\":%d,\"input_script_id\":%d,\"wire\":%d,\"extent\":%d,\"script_id\":%d,\"type_index\":%d,\"player\":[%g,%g],\"target\":[%g,%g]}",
		frame, castFrame, playerID, key, spell.SPELL_SWAP, actual.id, actual.created, actual.trigger,
		input.wire, input.extent, input.scriptID, actual.wire, actual.target.Extent, actual.target.ScriptID, actual.target.TypeIndex,
		actual.player.Position.X, actual.player.Position.Y, actual.target.Position.X, actual.target.Position.Y)
}
