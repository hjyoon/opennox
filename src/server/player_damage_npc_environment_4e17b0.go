package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// Bounded NPC prefix for source-less world LAVA and non-unit DANGEROUS
// IMPALE. The caller has already applied the entry flags/invulnerability
// gates and cached the NPC update/equipment/absorption, as at 004E187E.
func playerDamageNPCEnvironment4E17B0(
	target, source, weapon *Object, cached playerDamageGreatSwordContext4E17B0,
	absorption float32, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	liveNPC := func() bool {
		return target.Class().Has(object.ClassMonster) && !target.Class().Has(object.ClassPlayer) && target.UpdateData != nil
	}
	*cached.marker = 0
	// 004E19CE skips the complete facing/defense prefix for nil sources.
	// A world spike skips Reflect, sword and staff, but a stance shield can
	// still block it. Snapshot PrevPos before the six source exclusions.
	if source != nil {
		position := weapon.PrevPos
		if r.BlockSourceExcluded == nil {
			return reject("missing NPC environment exclusion")
		}
		excluded := r.BlockSourceExcluded(weapon)
		if !liveNPC() {
			return reject("unsupported live NPC environment attribution record")
		}
		if source != weapon {
			*cached.marker, *cached.markerType = 1, uint32(weapon.TypeInd)
		}
		if !excluded {
			if r.BlockDirection == nil {
				return reject("missing NPC environment direction")
			}
			front := r.BlockDirection(target, position)
			if !liveNPC() {
				return reject("unsupported live NPC environment shield record")
			}
			if front && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK && cached.armorFlags&0x3000000 != 0 {
				// This is the same 004E1B98 audio/live-projectile/live-shield
				// sequence already restored for ordinary world FLAME.
				return playerDamageWorldFlameShield4E17B0(target, source, weapon, damage, typ, r)
			}
		}
	}
	if !liveNPC() || !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported NPC environment armor record")
	}
	effective, wear := damage, damage
	// The independently read 004E20A8 jump table sends case 3 to 004E1F84,
	// but case 12 to 004E1DC7. Lava wears armor using the raw signed damage
	// and never consumes absorption/carry. Impale reloads LIVE carry first.
	if typ == object.DamageImpale {
		carry := &target.UpdateDataMonster().Field1
		accumulated := float32((1-float64(absorption))*float64(damage)) + math.Float32frombits(*carry)
		effective = playerDamageRound4E17B0(accumulated)
		*carry = math.Float32bits(accumulated - float32(effective))
		wear = damage - effective
	}
	playerDamageApplyArmor4E17B0(target, source, weapon, wear, typ, r)
	if !liveNPC() {
		return reject("unsupported live NPC environment marker record")
	}
	if *cached.marker == 0 {
		*cached.marker, *cached.markerType = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	// The original queries GodMode after wear/marker, then exempts PLAYERS
	// only; an NPC must still reach the live Quest/default-damage tail.
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing NPC environment Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing NPC environment default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
