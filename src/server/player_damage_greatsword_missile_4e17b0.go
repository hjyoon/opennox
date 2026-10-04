package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// These are the update/equipment values cached by 004E184E..004E1898, not
// a PE32 record. State and damage markers retain the cached update base.
type playerDamageGreatSwordContext4E17B0 struct {
	weaponFlags, armorFlags uint32
	state                   *PlayerState
	marker, markerType      *uint32
}

func playerDamageGreatSwordItem4E22A0(target *Object) *Object {
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Flags().Has(object.FlagEquipped) && uint32(item.SubClass())&0x400 != 0 {
			return item
		}
	}
	return nil
}

// 004E1C0F..004E1D08 is the GreatSword (not a spell staff) defense. The
// earlier ordinary shield and Reflect Shield branches remain separate. This
// slice admits missiles; non-missile BLADE/IMPACT blocks belong to melee.
func playerDamageGreatSwordMissileBlock4E17B0(
	target, source, weapon *Object, cached playerDamageGreatSwordContext4E17B0,
	damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	attack := weapon
	if attack == nil {
		attack = source
	}
	if source == nil || attack == nil || !attack.Class().Has(object.ClassMissile) ||
		cached.weaponFlags&0x400 == 0 || typ == object.DamageManaBomb {
		return false, false, false
	}
	reject := func(reason string) (bool, bool, bool) {
		h, result := playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
		return true, h, result
	}
	player := target.Class().Has(object.ClassPlayer)
	if player && cached.state == nil {
		return reject("GreatSword live player lacks cached player update")
	}
	// A real shield posture takes the shield branch, including its electric
	// exclusions. It must not fall through into a GreatSword reflection.
	shieldPosture := (!player && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK) ||
		(player && *cached.state == PlayerState16)
	if cached.armorFlags&0x3000000 != 0 && shieldPosture {
		return false, false, false
	}
	excluded := r.BlockSourceExcluded
	if weapon == nil {
		excluded = r.BlockSourceOnlyExcluded
	}
	if excluded == nil || r.BlockDirection == nil {
		return reject("missing GreatSword missile block direction service")
	}
	// 004E1A42/004E1ABE copy the attack's previous position before testing
	// exclusions. A callback must not silently replace this facing input.
	pos := attack.PrevPos
	if excluded(attack) || !r.BlockDirection(target, pos) {
		return false, false, false
	}
	player = target.Class().Has(object.ClassPlayer)
	if player {
		if cached.state == nil || (r.playerPrefix == nil && (target.UpdateData == nil || target.UpdateDataPlayer().Player == nil)) {
			return reject("GreatSword live player lacks cached player update")
		}
		state := *cached.state
		if state != PlayerState13 && state != PlayerState18 && state != PlayerState19 && state != PlayerState20 {
			return false, false, false
		}
		if r.playerPrefix == nil && target.UpdateDataPlayer().Player.ObserveTarget() != nil {
			return reject("possessed player GreatSword missile block")
		}
	} else if !playerDamageMonsterBlockReady534340(target) {
		return false, false, false
	}
	item := playerDamageGreatSwordItem4E22A0(target)
	if item == nil {
		// The original helper dereferences nil after its weapon callback.
		// Keep that unsupported boundary safe before committing the prefix.
		return reject("missing equipped GreatSword block item")
	}
	if cached.marker == nil || cached.markerType == nil || r.ProjectileReflect == nil ||
		r.Audio == nil || r.BlockDamagePercent == nil ||
		r.Melee.CanDamageBlockWeapon == nil || !r.Melee.CanDamageBlockWeapon(item) ||
		r.Melee.DamageBlockWeapon == nil ||
		(player && (r.Melee.RandomInt == nil || r.PlayerSetState == nil)) ||
		(!player && (r.Melee.MonsterBlockAction == nil || r.Melee.MonsterPopBlockAction == nil)) {
		return reject("missing GreatSword missile block effect service")
	}
	if uint32(attack.SubClass())&2 == 0 && (r.ClearOwner == nil || r.SetOwner == nil) {
		return reject("missing GreatSword missile ownership service")
	}
	if r.playerPrefix == nil {
		*cached.marker = 0
		if (weapon != nil && source != weapon) || (weapon == nil && (typ == object.DamageClaw || typ == object.DamageCrush)) {
			*cached.marker, *cached.markerType = 1, uint32(attack.TypeInd)
		}
	}
	r.ProjectileReflect(attack, target)
	// 004E1C72/004E1C7C reload class and subclass after reflection. Unlike
	// Reflect Shield, subclass bit 2 retains the owner and calls no ChangeOwner.
	if attack.Class().Has(object.ClassMissile) && uint32(attack.SubClass())&2 == 0 {
		if r.ClearOwner == nil || r.SetOwner == nil {
			return reject("missing live GreatSword missile ownership service")
		}
		r.ClearOwner(attack)
		r.SetOwner(target, attack)
	}
	r.Audio(890, target)
	if target.Class().Has(object.ClassPlayer) {
		if r.Melee.RandomInt == nil || r.PlayerSetState == nil {
			return reject("missing live GreatSword player state service")
		}
		r.PlayerSetState(target, PlayerState(r.Melee.RandomInt(18, 20)))
	} else {
		if r.Melee.MonsterBlockAction == nil {
			return reject("missing live GreatSword monster action service")
		}
		r.Melee.MonsterBlockAction(target)
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	// 004E22A6 reloads the inventory head only after reflection, audio,
	// state/action and balance callbacks. Do not wear an eagerly cached item.
	item = playerDamageGreatSwordItem4E22A0(target)
	if item == nil || !r.Melee.CanDamageBlockWeapon(item) {
		return reject("unsupported live GreatSword block item")
	}
	if !r.Melee.DamageBlockWeapon(item, target, source, attack, amount, typ) && r.Unsupported != nil {
		r.Unsupported("GreatSword block durability failed", target, source, weapon, damage, typ)
	}
	if item.Flags().Has(object.FlagDestroyed) {
		if target.Class().Has(object.ClassPlayer) {
			if r.PlayerSetState == nil {
				return reject("missing live GreatSword break state service")
			}
			r.PlayerSetState(target, PlayerState13)
		} else {
			if r.Melee.MonsterPopBlockAction == nil {
				return reject("missing live GreatSword break action service")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, true, false
}
