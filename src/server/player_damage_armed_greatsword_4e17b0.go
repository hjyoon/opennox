package server

import "github.com/opennox/libs/object"

// 004E1C35..004E1D08, after the armed EXPLOSION caller's cached equipment,
// marker, exclusion and facing prefix. Do not execute that prefix again.
// Pure non-missile type 7 cannot enter this branch. An unready cached player
// state or live NPC action falls through to the caller's armor/HP tail.
func playerDamageArmedExplosionGreatSword4E17B0(
	target, source, weapon *Object, damage int32, typ object.DamageType,
	cachedState *PlayerState, r PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	reject := func(reason string) (bool, bool, bool) {
		h, ok := playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
		return true, h, ok
	}
	player := target.Class().Has(object.ClassPlayer)
	if player {
		if cachedState == nil {
			return reject("missing armed EXPLOSION GreatSword cached state")
		}
		switch *cachedState {
		case PlayerState13, PlayerState18, PlayerState19, PlayerState20:
		default:
			return false, false, false
		}
	} else {
		if !target.Class().Has(object.ClassMonster) || target.UpdateData == nil {
			return reject("missing armed EXPLOSION GreatSword NPC action update")
		}
		if !playerDamageMonsterBlockReady534340(target) {
			return false, false, false
		}
	}
	if r.ProjectileReflect == nil {
		return reject("missing armed EXPLOSION GreatSword reflection")
	}
	r.ProjectileReflect(weapon, target)
	// Ownership uses the post-reflect class/subclass, unlike Reflect27's
	// unconditional subclass-0x40 rule. GreatSword never calls ChangeOwner.
	if weapon.Class().Has(object.ClassMissile) && uint32(weapon.SubClass())&2 == 0 {
		if r.ClearOwner == nil {
			return reject("missing armed EXPLOSION GreatSword clear-owner")
		}
		r.ClearOwner(weapon)
		if r.SetOwner == nil {
			return reject("missing armed EXPLOSION GreatSword set-owner")
		}
		r.SetOwner(target, weapon)
	}
	if r.Audio == nil {
		return reject("missing armed EXPLOSION GreatSword audio")
	}
	r.Audio(890, target)
	sameLayout := func() bool {
		return target.Class().Has(object.ClassPlayer) == player &&
			(player || target.Class().Has(object.ClassMonster)) && target.UpdateData != nil &&
			(!player || target.UpdateDataPlayer().Player != nil)
	}
	if !sameLayout() {
		return reject("unsupported live armed EXPLOSION GreatSword action layout")
	}
	if player {
		if r.Melee.RandomInt == nil {
			return reject("missing armed EXPLOSION GreatSword RNG")
		}
		state := PlayerState(r.Melee.RandomInt(18, 20))
		if !sameLayout() {
			return reject("unsupported live armed EXPLOSION GreatSword state layout")
		}
		if r.PlayerSetState == nil {
			return reject("missing armed EXPLOSION GreatSword player state")
		}
		r.PlayerSetState(target, state)
	} else {
		if r.Melee.MonsterBlockAction == nil {
			return reject("missing armed EXPLOSION GreatSword NPC block action")
		}
		r.Melee.MonsterBlockAction(target)
	}
	if r.BlockDamagePercent == nil {
		return reject("missing armed EXPLOSION GreatSword balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	// 004E22A0 scans live inventory after sound, state/action and balance.
	// The selected pointer, not a second inventory lookup, is tested for
	// destruction after the actual weapon-wear callback.
	item := playerDamageGreatSwordItem4E22A0(target)
	if r.Melee.DamageBlockWeapon == nil {
		return reject("missing armed EXPLOSION GreatSword durability")
	}
	if item != nil && item.HealthData != nil &&
		(r.Melee.CanDamageBlockWeapon == nil || !r.Melee.CanDamageBlockWeapon(item)) {
		return reject("unsupported live armed EXPLOSION GreatSword durability")
	}
	if !r.Melee.DamageBlockWeapon(item, target, source, weapon, amount, typ) && r.Unsupported != nil {
		r.Unsupported("armed EXPLOSION GreatSword durability failed", target, source, weapon, damage, typ)
	}
	if item == nil {
		// Original 004E1560 is a nil-item no-op; 004E22EE then faults at
		// [nil+0x10]. Retain all preceding effects and report that fault.
		return reject("original armed EXPLOSION GreatSword nil-item fault")
	}
	if item.Flags().Has(object.FlagDestroyed) {
		if !sameLayout() {
			return reject("unsupported live armed EXPLOSION GreatSword broken-item layout")
		}
		if player {
			if r.PlayerSetState == nil {
				return reject("missing armed EXPLOSION GreatSword idle state")
			}
			r.PlayerSetState(target, PlayerState13)
		} else {
			if r.Melee.MonsterPopBlockAction == nil {
				return reject("missing armed EXPLOSION GreatSword NPC pop")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, true, false
}
