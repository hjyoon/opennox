package server

import (
	"math"

	"github.com/opennox/libs/object"
)

// The NPC entry has already checked subclass, flags and invulnerability.
// Nil source/weapon skip reflection, facing and weapon defenses. Coop is
// still queried before caching the update, as in the original common prefix.
func playerDamageSourceLessNPC4E17B0(target *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0) (handled, result bool) {
	if r.CoopMode != nil {
		_ = r.CoopMode()
	}
	liveCarry := func() *uint32 {
		if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return nil
		}
		return &target.UpdateDataMonster().Field1
	}
	if liveCarry() == nil || r.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(r, "missing source-less NPC update/default service", target, nil, nil, damage, typ)
	}
	cached := target.UpdateDataMonster()
	absorption := math.Float32frombits(cached.Field518)
	cached.Field547 = 0
	return playerDamageSourceLessSwitch4E17B0(target, damage, typ, playerDamageSourceLessRecord4E17B0{
		absorption: absorption, marker: &cached.Field547, markerType: &cached.Field546, carry: liveCarry,
	}, r)
}
