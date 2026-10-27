package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// ToxicCloud and SmallToxicCloud are SIMPLE|DANGEROUS, not missiles. Their
// immediate call can carry a nil, self, imaginary, unit or owner-chain source.
// Periodic weapon-less poison and player victims retain their separate paths.
func playerDamageCloudPoisonShape4E17B0(weapon *Object, typ object.DamageType) bool {
	return typ == object.DamagePoison && weapon != nil &&
		weapon.Class().Has(object.ClassSimple) && weapon.Class().Has(object.ClassDangerous) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}

func playerDamageMonsterCloudPoison4E17B0(
	target, source, cloud *Object,
	update *MonsterUpdateData,
	equipment playerDamageGreatSwordContext4E17B0,
	damage int32,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	const typ = object.DamagePoison
	// 004E18D1 clears the entry NPC update before Reflect and exclusions.
	// Nonmissile case 5 does not enter Reflect, even when buff 27 is present.
	update.Field547 = 0
	if source != nil {
		// 004E1A49 snapshots PrevPos before exclusion/type callbacks. Marker
		// attribution reads the live cloud type after that callback instead.
		attackPos := cloud.PrevPos
		if runtime.BlockSourceExcluded == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON exclusion", target, source, cloud, damage, typ)
		}
		excluded := runtime.BlockSourceExcluded(cloud)
		if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
			return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC cloud POISON marker record", target, source, cloud, damage, typ)
		}
		if source != cloud {
			update.Field547, update.Field546 = 1, uint32(cloud.TypeInd)
		}
		if !excluded {
			if runtime.BlockDirection == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON direction", target, source, cloud, damage, typ)
			}
			front := runtime.BlockDirection(target, attackPos)
			if !target.Class().Has(object.ClassMonster) || target.Class().Has(object.ClassPlayer) || target.UpdateData == nil {
				return playerDamageUnsupported4E17B0(runtime, "unsupported live NPC cloud POISON block record", target, source, cloud, damage, typ)
			}
			if front && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK && equipment.armorFlags&0x3000000 != 0 {
				return playerDamageMonsterCloudPoisonBlock4E17B0(target, source, cloud, damage, runtime)
			}
			// Equipment stays cached, while the facing callback may have
			// changed this admitted cloud into a live missile. Readiness and
			// the GreatSword effects follow without repeating the prefix.
			if front && equipment.weaponFlags&0x400 != 0 && cloud.Class().Has(object.ClassMissile) {
				if blocked, h, result := playerDamageNPCCloudGreatSword4E17B0(target, source, cloud, damage, runtime); blocked {
					return h, result
				}
			}
		}
	}
	// Case 5 -> 004E1E83: signed raw damage, without armor absorption,
	// fractional carry, armor wear or the Shield/electric reduction paths.
	if update.Field547 == 0 {
		update.Field547, update.Field546 = 2, uint32(typ)
	}
	// 004E2025 queries the global flag before the live player-class test.
	// GodMode ordinarily protects players only, not this NPC recipient.
	if runtime.GodMode != nil && runtime.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	effective := damage
	if runtime.QuestMode != nil && runtime.QuestMode() {
		if runtime.QuestDamageScale == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON quest scale", target, source, cloud, damage, typ)
		}
		scale := runtime.QuestDamageScale()
		scaled := experienceSpill32TowardZero(experienceMul53TowardZero(float64(scale), float64(effective)))
		// 004E2063 uses signed FISTP under the startup control word.
		effective = netFXTruncateSignedDWORD419A70(scaled)
		if damage > 0 && effective < 1 {
			effective = 1
		}
	}
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON default service", target, source, cloud, damage, typ)
	}
	return true, runtime.DefaultDamage(target, source, cloud, effective, typ)
}

// The ordinary front-facing block is reachable for non-excluded custom cloud
// shapes. Stock ToxicCloud/SmallToxicCloud skip it. Keep its late live projectile
// checks from 004E1BA7..004E1BFD. The literal 2 is the shield subclass mask;
// EquipDamage still receives the incoming POISON type, not CRUSH.
func playerDamageMonsterCloudPoisonBlock4E17B0(
	target, source, cloud *Object,
	damage int32,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	const typ = object.DamagePoison
	if runtime.Audio == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON shield audio", target, source, cloud, damage, typ)
	}
	runtime.Audio(878, target)
	if cloud.Class().Has(object.ClassMissile) && uint32(cloud.SubClass())&0x70 == 0 {
		if runtime.ProjectileReflect == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON shield reflection", target, source, cloud, damage, typ)
		}
		runtime.ProjectileReflect(cloud, target)
		if cloud.Class().Has(object.ClassMissile) && uint32(cloud.SubClass())&2 == 0 {
			if runtime.ClearOwner == nil || runtime.SetOwner == nil {
				return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON shield owner", target, source, cloud, damage, typ)
			}
			runtime.ClearOwner(cloud)
			runtime.SetOwner(target, cloud)
		}
	}
	if runtime.BlockDamagePercent == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON shield percentage", target, source, cloud, damage, typ)
	}
	amount := playerDamageBlockAmount4E17B0(runtime.BlockDamagePercent(), damage)
	shield := playerDamageShieldItem4E17B0(target)
	if runtime.DamageBlockItem == nil || (shield != nil && (runtime.CanDamageBlockItem == nil || !runtime.CanDamageBlockItem(shield))) {
		return playerDamageUnsupported4E17B0(runtime, "unsupported NPC cloud POISON live shield wear", target, source, cloud, damage, typ)
	}
	if !runtime.DamageBlockItem(shield, target, source, cloud, amount, typ) && runtime.Unsupported != nil {
		runtime.Unsupported("NPC cloud POISON shield wear failed", target, source, cloud, damage, typ)
	}
	if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
		if runtime.Melee.MonsterPopBlockAction == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing NPC cloud POISON broken shield action", target, source, cloud, damage, typ)
		}
		runtime.Melee.MonsterPopBlockAction(target)
	}
	return true, false
}

// 004E1C35..004E1D08 after the cloud caller's cached equipment, marker,
// exclusion and facing prefix. A direction callback may have made the
// admitted nonmissile cloud a live missile; do not repeat that prefix.
func playerDamageNPCCloudGreatSword4E17B0(
	target, source, cloud *Object, damage int32, r PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	const typ = object.DamagePoison
	reject := func(reason string) (bool, bool, bool) {
		h, ok := playerDamageUnsupported4E17B0(r, reason, target, source, cloud, damage, typ)
		return true, h, ok
	}
	if !playerDamageMonsterBlockReady534340(target) {
		return false, false, false
	}
	if r.ProjectileReflect == nil {
		return reject("missing NPC cloud POISON GreatSword reflection")
	}
	r.ProjectileReflect(cloud, target)
	// The original reloads class and subclass after reflection. Subclass bit2
	// retains the owner; this branch never calls ChangeOwner.
	if cloud.Class().Has(object.ClassMissile) && uint32(cloud.SubClass())&2 == 0 {
		if r.ClearOwner == nil {
			return reject("missing NPC cloud POISON GreatSword clear-owner")
		}
		r.ClearOwner(cloud)
		if r.SetOwner == nil {
			return reject("missing NPC cloud POISON GreatSword set-owner")
		}
		r.SetOwner(target, cloud)
	}
	if r.Audio == nil {
		return reject("missing NPC cloud POISON GreatSword audio")
	}
	r.Audio(890, target)
	liveNPC := func() bool {
		return target.Class().Has(object.ClassMonster) && !target.Class().Has(object.ClassPlayer) && target.UpdateData != nil
	}
	if !liveNPC() {
		return reject("unsupported live NPC cloud POISON GreatSword action layout")
	}
	if r.Melee.MonsterBlockAction == nil {
		return reject("missing NPC cloud POISON GreatSword block action")
	}
	r.Melee.MonsterBlockAction(target)
	if r.BlockDamagePercent == nil {
		return reject("missing NPC cloud POISON GreatSword balance")
	}
	// GAME.EXE's startup PC53/toward-zero product and binary32 spill precede
	// 004E22A0's live inventory scan, including a balance callback replacement.
	amount := playerDamageBlockAmount4E17B0(r.BlockDamagePercent(), damage)
	item := playerDamageGreatSwordItem4E22A0(target)
	if r.Melee.DamageBlockWeapon == nil {
		return reject("missing NPC cloud POISON GreatSword durability")
	}
	if item != nil && item.HealthData != nil &&
		(r.Melee.CanDamageBlockWeapon == nil || !r.Melee.CanDamageBlockWeapon(item)) {
		return reject("unsupported live NPC cloud POISON GreatSword durability")
	}
	if !r.Melee.DamageBlockWeapon(item, target, source, cloud, amount, typ) && r.Unsupported != nil {
		r.Unsupported("NPC cloud POISON GreatSword durability failed", target, source, cloud, damage, typ)
	}
	if item == nil {
		// The original nil-item wear is a no-op, followed by a flags dereference
		// at 004E22EE. Keep preceding effects and report this original fault.
		return reject("original NPC cloud POISON GreatSword nil-item fault")
	}
	// 004E22EE tests the selected pointer after wear, even when wear relinks
	// the inventory. A replacement head is not the item that just broke.
	if item.Flags().Has(object.FlagDestroyed) {
		if !liveNPC() {
			return reject("unsupported live NPC cloud POISON GreatSword broken-item layout")
		}
		if r.Melee.MonsterPopBlockAction == nil {
			return reject("missing NPC cloud POISON GreatSword pop action")
		}
		r.Melee.MonsterPopBlockAction(target)
	}
	return true, true, false
}
