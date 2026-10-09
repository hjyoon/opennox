package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// 004E17B0 does not require a unit attacker. Retain its cached player/NPC
// marker and equipment prefix for world-fired arrows/fireballs and DeathBall
// CRUSH, without interpreting the world owner's nonexistent unit update.
func playerDamageWorldProjectile4E17B0(
	target, source, weapon *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if target.Flags().HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if r.Frame == nil {
			return reject("missing world projectile invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing world projectile invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	player := target.Class().Has(object.ClassPlayer)
	if player && r.CoopMode != nil && r.CoopMode() && source.FindOwnerChainPlayer() == target {
		return true, false
	}
	if target.UpdateData == nil || (!player && (!target.Class().Has(object.ClassMonster) || uint32(target.SubClass())&0x10 == 0)) {
		return reject("unsupported world projectile target update")
	}
	var absorption float32
	var cached playerDamageGreatSwordContext4E17B0
	cached.prefixed = true
	if player {
		ud := target.UpdateDataPlayer()
		if ud.Player == nil {
			return reject("missing world projectile player info")
		}
		if ud.Player.Field3680&1 != 0 {
			return true, false
		}
		absorption = math.Float32frombits(ud.Field57)
		cached.weaponFlags, cached.armorFlags = ud.Player.WeaponEquip, ud.Player.ArmorEquip
		cached.state, cached.marker, cached.markerType = &ud.State, &ud.Field76, &ud.Field75
		r.playerPrefix = &playerDamagePrefix4E17B0{update: ud, armorFlags: cached.armorFlags, weaponFlags: cached.weaponFlags}
		*cached.marker = 0
		if ud.Player.ObserveTarget() != nil {
			if r.ObserveClear == nil {
				return reject("missing world projectile observe service")
			}
			r.ObserveClear(target)
		}
	} else {
		ud := target.UpdateDataMonster()
		absorption = math.Float32frombits(ud.Field518)
		cached.weaponFlags, cached.armorFlags = ud.WeaponEquipFlags, ud.ArmorEquipFlags
		cached.marker, cached.markerType = &ud.Field547, &ud.Field546
		*cached.marker = 0
	}
	liveRecord := func() bool {
		return target.UpdateData != nil && player == target.Class().Has(object.ClassPlayer) &&
			(player || target.Class().Has(object.ClassMonster))
	}
	if !liveRecord() {
		return reject("unsupported live world projectile prefix record")
	}
	attack := weapon
	if attack == nil {
		attack = source
	}
	// Reflect precedes attribution, and reloads projectile class/subclass
	// after every reflect/ownership callback (004E193C/1956/195C).
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && attack.Class().Has(object.ClassMissile) {
		if r.BlockDirection == nil {
			return reject("missing world projectile Reflect direction")
		}
		if r.BlockDirection(target, attack.PosVec) {
			if r.ProjectileReflect == nil {
				return reject("missing world projectile Reflect effect")
			}
			r.ProjectileReflect(attack, target)
			if uint32(attack.SubClass())&0x40 == 0 {
				if r.ClearOwner == nil || r.SetOwner == nil {
					return reject("missing world projectile Reflect ownership")
				}
				r.ClearOwner(attack)
				r.SetOwner(target, attack)
			}
			if attack.Class().Has(object.ClassMissile) && uint32(attack.SubClass())&2 != 0 {
				if r.ChangeOwner == nil {
					return reject("missing world projectile Reflect ChangeOwner")
				}
				r.ChangeOwner(attack, target)
			}
			if r.Audio == nil {
				return reject("missing world projectile Reflect audio")
			}
			r.Audio(122, target)
			return true, false
		}
	}
	pos := attack.PrevPos // Snapshot BEFORE exclusion, as 004E1A49/004E1ABE.
	exclusion := r.BlockSourceExcluded
	if weapon == nil {
		exclusion = r.BlockSourceOnlyExcluded
	}
	if exclusion == nil {
		return reject("missing world projectile exclusion")
	}
	excluded := exclusion(attack)
	if !liveRecord() {
		return reject("unsupported live world projectile attribution record")
	}
	if (weapon != nil && source != weapon) || (weapon == nil && typ == object.DamageCrush) {
		*cached.marker, *cached.markerType = 1, uint32(attack.TypeInd)
	}
	front := false
	if !excluded {
		if r.BlockDirection == nil {
			return reject("missing world projectile block direction")
		}
		front = r.BlockDirection(target, pos)
	}
	if !liveRecord() {
		return reject("unsupported live world projectile block record")
	}
	if front {
		shield := (player && *cached.state == PlayerState16) || (!player && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK)
		if shield && cached.armorFlags&0x3000000 != 0 {
			return playerDamageWorldProjectileShield4E17B0(target, source, weapon, attack, damage, typ, r)
		}
		// Only the original single facing/exclusion pair is externally called.
		r.BlockSourceExcluded = func(*Object) bool { return excluded }
		r.BlockSourceOnlyExcluded = r.BlockSourceExcluded
		r.BlockDirection = func(*Object, types.Pointf) bool { return front }
		if applicable, h, result := playerDamageGreatSwordMissileBlock4E17B0(target, source, weapon, cached, damage, typ, r); applicable {
			return h, result
		}
		if player && cached.weaponFlags&0x400 == 0 && *cached.state == PlayerState1 && cached.armorFlags&0x3000000 != 0 {
			if r.BerserkShieldBlock == nil {
				return reject("missing world projectile berserker shield service")
			}
			if r.BerserkShieldBlock(target) {
				return playerDamageWorldProjectileShield4E17B0(target, source, weapon, attack, damage, typ, r)
			}
		}
	}
	if !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported world projectile armor durability")
	}
	if !liveRecord() {
		return reject("unsupported live world projectile carry record")
	}
	effective, wear := damage, damage
	if typ != object.DamageFlame {
		// Case 3 uses full armor; case 2 uses HALF armor. FLAME (case 1)
		// wears armor but bypasses HP absorption and the fractional carry.
		scale := float64(absorption)
		if typ == object.DamageCrush {
			scale *= 0.5
		}
		var carry *uint32
		if player {
			carry = &target.UpdateDataPlayer().Field21
		} else {
			carry = &target.UpdateDataMonster().Field1
		}
		accumulated := float32((1.0-scale)*float64(damage)) + math.Float32frombits(*carry)
		effective = playerDamageRound4E17B0(accumulated)
		*carry = math.Float32bits(accumulated - float32(effective))
		wear = damage - effective
	}
	playerDamageApplyArmor4E17B0(target, source, weapon, wear, typ, r)
	if target.Class().HasAny(object.MaskUnits) {
		if player != target.Class().Has(object.ClassPlayer) {
			return reject("unsupported live world projectile cached marker layout")
		}
		if *cached.marker == 0 {
			*cached.marker, *cached.markerType = 2, uint32(typ)
		}
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing live world projectile Quest scale")
		}
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing world projectile default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}

// Shield audio precedes LIVE projectile tests; the balance callback precedes
// LIVE inventory selection. The NPC block action has its own break handling.
func playerDamageWorldProjectileShield4E17B0(target, source, weapon, attack *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0) (bool, bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if r.Audio == nil {
		return reject("missing world projectile shield audio")
	}
	r.Audio(878, target)
	if attack.Class().Has(object.ClassMissile) && uint32(attack.SubClass())&0x70 == 0 {
		if r.ProjectileReflect == nil {
			return reject("missing world projectile shield reflection")
		}
		r.ProjectileReflect(attack, target)
		if attack.Class().Has(object.ClassMissile) && uint32(attack.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing world projectile shield ownership")
			}
			r.ClearOwner(attack)
			r.SetOwner(target, attack)
		}
	}
	if r.BlockDamagePercent == nil {
		return reject("missing world projectile shield balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	shield := playerDamageShieldItem4E17B0(target)
	if r.DamageBlockItem == nil || (shield != nil && (r.CanDamageBlockItem == nil || !r.CanDamageBlockItem(shield))) {
		return reject("unsupported live world projectile shield durability")
	}
	if !r.DamageBlockItem(shield, target, source, weapon, amount, typ) && r.Unsupported != nil {
		r.Unsupported("world projectile shield durability failed", target, source, weapon, damage, typ)
	}
	if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
		if target.Class().Has(object.ClassPlayer) {
			if r.PlayerSetState == nil || target.UpdateData == nil {
				return reject("missing live world projectile broken shield state")
			}
			r.PlayerSetState(target, PlayerState13)
		} else {
			if r.Melee.MonsterPopBlockAction == nil || !target.Class().Has(object.ClassMonster) || target.UpdateData == nil {
				return reject("missing live world projectile broken shield action")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, false
}
