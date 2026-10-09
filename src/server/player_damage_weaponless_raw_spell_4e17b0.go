package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// Plasma, Death Ray and Mana Bomb supply a terminal unit owner, nil, or
// the class-zero ImaginaryCaster, and no weapon. The sealed PE type switch
// maps 14/15/16 to the same signed raw HP tail at 004E1E83, not electric
// armor/protection cases 9/17 or SentryGlobe's nonnil-weapon ray prefix.
func playerDamageWeaponlessRawSpellShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	return weapon == nil && (typ == object.DamagePlasma || typ == object.DamageManaBomb || typ == object.DamageZapRay) &&
		(source == nil || source.Class() == 0 || (source.Class().HasAny(object.MaskUnits) &&
			!source.Class().HasAny(object.ClassWeapon|object.ClassWand|object.ClassMissile)))
}

// GAME.EXE 004E17B0: cached entry equipment/update, live observer/Reflect,
// and source-only Fist/Meteor exclusions precede case 15/16's signed raw HP
// tail at 004E1E83. No armor absorption, carry, wear or electric scale runs.
func playerDamageWeaponlessRawSpell4E17B0(
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
			return reject("missing raw spell invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing raw spell invulnerability audio")
			}
			r.Audio(71, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source != nil && source.FindOwnerChainPlayer() == target &&
		target.Class().Has(object.ClassPlayer) && typ != object.DamageManaBomb {
		return true, false
	}
	player := target.Class().Has(object.ClassPlayer)
	if target.UpdateData == nil {
		return reject("missing raw spell entry update")
	}
	cached := target.UpdateData
	var armorFlags uint32
	if player {
		ud := target.UpdateDataPlayer()
		if ud.Player == nil {
			return reject("missing raw spell player info")
		}
		armorFlags = ud.Player.ArmorEquip
		if ud.Player.Field3680&1 != 0 {
			return true, false
		}
	} else {
		armorFlags = target.UpdateDataMonster().ArmorEquipFlags
	}
	marker, _ := playerDamageCasterImpactMarker4E17B0(target, cached)
	*marker = 0
	if player && target.UpdateDataPlayer().Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing raw spell observe clear")
		}
		r.ObserveClear(target)
	}
	// A replaced update of the same semantic layout is valid. A changed
	// Player/Monster layout cannot reuse the PE32 offset through a native
	// cached pointer, and is reported at use instead of corrupting memory.
	liveLayout := func() bool {
		return target.UpdateData != nil && target.Class().Has(object.ClassPlayer) == player &&
			(player || target.Class().Has(object.ClassMonster))
	}
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && source != nil {
		missile := source.Class().Has(object.ClassMissile)
		if missile || typ == object.DamageZapRay {
			if r.BlockDirection == nil {
				return reject("missing raw spell Reflect direction")
			}
			if r.BlockDirection(target, source.PosVec) {
				if missile {
					if r.ProjectileReflect == nil {
						return reject("missing raw spell Reflect service")
					}
					r.ProjectileReflect(source, target)
					// These subclass/class reads occur AFTER reflection, not
					// when the initial missile branch was selected.
					if uint32(source.SubClass())&0x40 == 0 {
						if r.ClearOwner == nil || r.SetOwner == nil {
							return reject("missing raw spell Reflect ownership")
						}
						r.ClearOwner(source)
						r.SetOwner(target, source)
					}
					if source.Class().Has(object.ClassMissile) && uint32(source.SubClass())&2 != 0 {
						if r.ChangeOwner == nil {
							return reject("missing raw spell Reflect change owner")
						}
						r.ChangeOwner(source, target)
					}
				}
				if typ == object.DamageZapRay {
					if r.PointFX == nil {
						return reject("missing raw spell Reflect FX")
					}
					r.PointFX(132, target.PosVec)
				}
				if r.Audio == nil {
					return reject("missing raw spell Reflect audio")
				}
				r.Audio(122, target)
				return true, false
			}
		}
	}
	if source != nil {
		pos := source.PrevPos
		if r.BlockSourceOnlyExcluded == nil {
			return reject("missing raw spell source-only exclusions")
		}
		excluded := r.BlockSourceOnlyExcluded(source)
		// Case 15 jumps directly to the raw switch after exclusions, before
		// physical direction/shield/GreatSword/staff/berserker predicates.
		if typ != object.DamageManaBomb && !excluded {
			if r.BlockDirection == nil {
				return reject("missing raw spell block direction")
			}
			front := r.BlockDirection(target, pos)
			if !liveLayout() {
				return reject("unsupported raw spell block layout")
			}
			if front {
				shield := (player && (*PlayerUpdateData)(cached).State == PlayerState16) ||
					(!player && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK)
				if shield && armorFlags&0x3000000 != 0 {
					return playerDamageWeaponlessRawSpellShield4E17B0(target, source, damage, typ, r)
				}
				// Type 16 with an ordinary nonmissile source does not enter
				// GreatSword/staff/berserker block. Those types stay separate.
				if source.Class().Has(object.ClassMissile) {
					return reject("unsupported live raw spell alternate missile block")
				}
			}
		}
	}
	if !liveLayout() {
		return reject("unsupported raw spell cached marker layout")
	}
	if marker, kind := playerDamageCasterImpactMarker4E17B0(target, cached); *marker == 0 {
		*marker, *kind = 2, uint32(typ)
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	effective := damage
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing raw spell Quest scale")
		}
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if damage > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing raw spell default damage")
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}

// 004E1B98..004E1BFD: shield sound, live projectile reflection/owner,
// binary32 block wear, live inventory lookup and broken-shield recovery.
func playerDamageWeaponlessRawSpellShield4E17B0(
	target, source *Object, damage int32, typ object.DamageType, r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	reject := func(reason string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, reason, target, source, nil, damage, typ)
	}
	if r.Audio == nil {
		return reject("missing raw spell shield audio")
	}
	r.Audio(878, target)
	if source.Class().Has(object.ClassMissile) && uint32(source.SubClass())&0x70 == 0 {
		if r.ProjectileReflect == nil {
			return reject("missing raw spell shield reflection")
		}
		r.ProjectileReflect(source, target)
		if source.Class().Has(object.ClassMissile) && uint32(source.SubClass())&2 == 0 {
			if r.ClearOwner == nil || r.SetOwner == nil {
				return reject("missing raw spell shield ownership")
			}
			r.ClearOwner(source)
			r.SetOwner(target, source)
		}
	}
	if r.BlockDamagePercent == nil {
		return reject("missing raw spell shield balance")
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	shield := playerDamageShieldItem4E17B0(target)
	if r.DamageBlockItem == nil || (shield != nil && (r.CanDamageBlockItem == nil || !r.CanDamageBlockItem(shield))) {
		return reject("unsupported raw spell live shield wear")
	}
	if !r.DamageBlockItem(shield, target, source, nil, amount, typ) && r.Unsupported != nil {
		r.Unsupported("raw spell shield wear failed", target, source, nil, damage, typ)
	}
	if shield != nil && shield.Flags().Has(object.FlagDestroyed) {
		if target.Class().Has(object.ClassPlayer) {
			if r.PlayerSetState == nil || target.UpdateData == nil {
				return reject("unsupported live raw spell broken-shield player")
			}
			r.PlayerSetState(target, PlayerState13)
		} else if target.Class().Has(object.ClassMonster) {
			if r.Melee.MonsterPopBlockAction == nil {
				return reject("missing raw spell broken-shield NPC action")
			}
			r.Melee.MonsterPopBlockAction(target)
		}
	}
	return true, false
}
