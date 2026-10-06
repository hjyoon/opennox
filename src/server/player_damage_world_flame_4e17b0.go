package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// Stock Flame and red FlameCleanse are world fires, not spell missiles. Their
// colliders can supply a nil, self, imaginary or terminal unit parent. Only
// case 1 is admitted here; weapon/wand/unit/missile damage keeps its own path.
func playerDamageWorldFlameShape4E17B0(weapon *Object, typ object.DamageType) bool {
	const worldFire = object.ClassFire | object.ClassSimple | object.ClassDangerous
	return typ == object.DamageFlame && weapon != nil &&
		weapon.Class()&worldFire == worldFire &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWeapon|object.ClassWand|object.ClassMissile)
}

// GAME.EXE 004E17B0 case 1 wears armor using the full signed incoming
// damage. It does not consume HP absorption/carry. Entry equipment, state
// and markers keep their cached update base; armor and shield wear reload
// their own live records after callbacks.
func playerDamageWorldFlame4E17B0(
	target, source, flame *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, flame, damage, typ)
	}
	if target.Flags().HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if r.Frame == nil {
			return reject("missing world FLAME invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing world FLAME invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source != nil &&
		source.FindOwnerChainPlayer() == target && target.Class().Has(object.ClassPlayer) {
		return true, false
	}
	player := target.Class().Has(object.ClassPlayer)
	if !player && (!target.Class().Has(object.ClassMonster) || uint32(target.SubClass())&0x10 == 0) {
		return true, false
	}
	if target.UpdateData == nil {
		return reject("missing world FLAME entry update")
	}
	var cached playerDamageGreatSwordContext4E17B0
	var update *PlayerUpdateData
	if player {
		update = target.UpdateDataPlayer()
		if update.Player == nil {
			return reject("missing world FLAME player info")
		}
		cached = playerDamageGreatSwordContext4E17B0{
			weaponFlags: update.Player.WeaponEquip, armorFlags: update.Player.ArmorEquip,
			state: &update.State, marker: &update.Field76, markerType: &update.Field75,
		}
		if update.Player.Field3680&1 != 0 {
			return true, false
		}
	} else {
		ud := target.UpdateDataMonster()
		cached = playerDamageGreatSwordContext4E17B0{
			weaponFlags: ud.WeaponEquipFlags, armorFlags: ud.ArmorEquipFlags,
			marker: &ud.Field547, markerType: &ud.Field546,
		}
	}
	*cached.marker = 0
	if player && update.Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing world FLAME observe service")
		}
		r.ObserveClear(target)
	}
	// Ordinary FIRE does not enter buff 27. A callback introducing MISSILE
	// needs its separate live Reflect prefix, not a PE32 fallback.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && flame.Class().Has(object.ClassMissile) {
		return reject("unsupported live world FLAME Reflect missile")
	}
	liveLayout := func() bool {
		return target.UpdateData != nil && target.Class().Has(object.ClassPlayer) == player &&
			(player || target.Class().Has(object.ClassMonster))
	}
	if source != nil {
		// 004E1A49 precedes exclusions; attribution instead reads the live
		// source type at 004E1AA8. A nil source skips this entire prefix.
		pos := flame.PrevPos
		if r.BlockSourceExcluded == nil {
			return reject("missing world FLAME exclusion service")
		}
		excluded := r.BlockSourceExcluded(flame)
		if !liveLayout() {
			return reject("unsupported live world FLAME cached marker layout")
		}
		if source != flame {
			*cached.marker, *cached.markerType = 1, uint32(flame.TypeInd)
		}
		if !excluded {
			if r.BlockDirection == nil {
				return reject("missing world FLAME direction service")
			}
			front := r.BlockDirection(target, pos)
			if !liveLayout() {
				return reject("unsupported live world FLAME block layout")
			}
			if front {
				shield := (player && *cached.state == PlayerState16) ||
					(!player && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK)
				if shield && cached.armorFlags&0x3000000 != 0 {
					return playerDamageWorldFlameShield4E17B0(target, source, flame, damage, typ, r)
				}
				// Pure case 1 cannot be intercepted by a GreatSword/staff.
				if cached.weaponFlags&0x400 != 0 && flame.Class().Has(object.ClassMissile) {
					return reject("unsupported live world FLAME GreatSword missile")
				}
				// Keep the existing optional gameex berserker-shield rule.
				if player && cached.weaponFlags&0x400 == 0 && *cached.state == PlayerState1 && cached.armorFlags&0x3000000 != 0 {
					if r.BerserkShieldBlock == nil {
						return reject("missing world FLAME berserker shield service")
					}
					if r.BerserkShieldBlock(target) {
						return playerDamageWorldFlameShield4E17B0(target, source, flame, damage, typ, r)
					}
				}
			}
		}
	}
	if !liveLayout() || !playerDamageArmorReady4E17B0(target, r) {
		return reject("unsupported live world FLAME armor record")
	}
	playerDamageApplyArmor4E17B0(target, source, flame, damage, typ, r)
	if !liveLayout() {
		return reject("unsupported live world FLAME armor marker layout")
	}
	if *cached.marker == 0 {
		*cached.marker, *cached.markerType = 2, uint32(typ)
	}
	// 004E2025 queries the global flag before testing live player class;
	// armor wear and the cached marker have already been applied.
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	effective := damage
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing world FLAME Quest scale")
		}
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if damage > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing world FLAME default damage")
	}
	return true, r.DefaultDamage(target, source, flame, effective, typ)
}

// 004E1B98..004E1BFD: audio precedes live projectile reflection and owner
// transfer. Balance precedes the live inventory lookup, including nil shields.
func playerDamageWorldFlameShield4E17B0(
	target, source, flame *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, flame, damage, typ)
	}
	if r.Audio == nil {
		return reject("missing world FLAME shield audio")
	}
	r.Audio(878, target)
	if flame.Class().Has(object.ClassMissile) && uint32(flame.SubClass())&0x70 == 0 {
		if r.ProjectileReflect == nil {
			return reject("missing world FLAME shield reflection")
		}
		r.ProjectileReflect(flame, target)
		if flame.Class().Has(object.ClassMissile) && uint32(flame.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing world FLAME shield ownership")
			}
			r.ClearOwner(flame)
			r.SetOwner(target, flame)
		}
	}
	if r.BlockDamagePercent == nil {
		return reject("missing world FLAME shield balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	shield := playerDamageShieldItem4E17B0(target)
	if r.DamageBlockItem == nil || (shield != nil && (r.CanDamageBlockItem == nil || !r.CanDamageBlockItem(shield))) {
		return reject("unsupported world FLAME live shield wear")
	}
	if !r.DamageBlockItem(shield, target, source, flame, amount, typ) && r.Unsupported != nil {
		r.Unsupported("world FLAME shield wear failed", target, source, flame, damage, typ)
	}
	if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
		if target.Class().Has(object.ClassPlayer) {
			if r.PlayerSetState == nil || target.UpdateData == nil {
				return reject("unsupported live world FLAME broken-shield player")
			}
			r.PlayerSetState(target, PlayerState13)
		} else if target.Class().Has(object.ClassMonster) {
			if r.Melee.MonsterPopBlockAction == nil {
				return reject("missing world FLAME broken-shield NPC action")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, false
}
