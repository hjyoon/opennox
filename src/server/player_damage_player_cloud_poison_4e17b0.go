package server

import "github.com/opennox/libs/object"

// Polyp contact creates an unowned ToxicCloud, so 0053D8C0 passes the cloud
// itself as source and weapon. Owned spell clouds also enter the same case 5.
// GAME.EXE retains the entry player marker/equipment through ObserveClear;
// poison skips armor absorption, wear, carry and elemental/Shield reduction.
func playerDamagePlayerCloudPoison4E17B0(target, source, cloud *Object, damage int32, r PlayerDamageRuntime4E17B0) (bool, bool) {
	const typ = object.DamagePoison
	reject := func(why string) (bool, bool) {
		return playerDamageUnsupported4E17B0(r, why, target, source, cloud, damage, typ)
	}
	if target.Flags().HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if r.Frame == nil {
			return reject("missing player cloud POISON invulnerability frame")
		}
		if byte(r.Frame())&3 == 0 {
			if r.Audio == nil {
				return reject("missing player cloud POISON invulnerability audio")
			}
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	if r.CoopMode != nil && r.CoopMode() && source != nil && source.FindOwnerChainPlayer() == target {
		return true, false
	}
	if target.UpdateData == nil {
		return reject("missing player cloud POISON update")
	}
	update := target.UpdateDataPlayer()
	if update.Player == nil {
		return reject("missing player cloud POISON player info")
	}
	if update.Player.Field3680&1 != 0 {
		return true, false
	}
	armorFlags, weaponFlags := update.Player.ArmorEquip, update.Player.WeaponEquip
	update.Field76 = 0
	if update.Player.ObserveTarget() != nil {
		if r.ObserveClear == nil {
			return reject("missing player cloud POISON observe service")
		}
		r.ObserveClear(target)
	}
	liveLayout := func() bool { return target.Class().Has(object.ClassPlayer) && target.UpdateData != nil }
	// ObserveClear may turn the admitted nonmissile cloud into a missile.
	// Reflect uses its current position; the later ordinary prefix retains
	// its separate previous-position check.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && cloud.Class().Has(object.ClassMissile) {
		if r.BlockDirection == nil {
			return reject("missing player cloud POISON Reflect direction")
		}
		if r.BlockDirection(target, cloud.PosVec) {
			if r.ProjectileReflect == nil {
				return reject("missing player cloud POISON Reflect service")
			}
			r.ProjectileReflect(cloud, target)
			// 004E193C reloads subclass after reflection. Transfer depends
			// on bit0x40 even when reflection cleared the missile class.
			if uint32(cloud.SubClass())&0x40 == 0 {
				if r.ClearOwner == nil {
					return reject("missing player cloud POISON Reflect clear-owner")
				}
				r.ClearOwner(cloud)
				if r.SetOwner == nil {
					return reject("missing player cloud POISON Reflect set-owner")
				}
				r.SetOwner(target, cloud)
			}
			// 004E1956 reads both fields after ClearOwner/SetOwner.
			if cloud.Class().Has(object.ClassMissile) && uint32(cloud.SubClass())&2 != 0 {
				if r.ChangeOwner == nil {
					return reject("missing player cloud POISON Reflect change-owner")
				}
				r.ChangeOwner(cloud, target)
			}
			if r.Audio == nil {
				return reject("missing player cloud POISON Reflect audio")
			}
			r.Audio(122, target)
			return true, false
		}
	}
	if source != nil {
		pos := cloud.PrevPos // 004E1A49 precedes exclusion callbacks.
		if r.BlockSourceExcluded == nil {
			return reject("missing player cloud POISON exclusion")
		}
		excluded := r.BlockSourceExcluded(cloud)
		if !liveLayout() {
			return reject("unsupported live player cloud POISON marker record")
		}
		if source != cloud {
			update.Field76, update.Field75 = 1, uint32(cloud.TypeInd)
		}
		if !excluded {
			if r.BlockDirection == nil {
				return reject("missing player cloud POISON direction")
			}
			front := r.BlockDirection(target, pos)
			if !liveLayout() {
				return reject("unsupported live player cloud POISON block record")
			}
			if front {
				// This is the shared 004E1B98..004E1BFD ordinary shield tail.
				// Both stock cloud types are excluded; custom shapes may block.
				if update.State == PlayerState16 && armorFlags&0x3000000 != 0 {
					return playerDamageWorldFlameShield4E17B0(target, source, cloud, damage, typ, r)
				}
				if weaponFlags&0x400 != 0 && cloud.Class().Has(object.ClassMissile) {
					if blocked, h, result := playerDamagePlayerCloudGreatSword4E17B0(target, source, cloud, damage, &update.State, r); blocked {
						return h, result
					}
				}
				if weaponFlags&0x400 == 0 && update.State == PlayerState1 && armorFlags&0x3000000 != 0 {
					if r.BerserkShieldBlock == nil {
						return reject("missing player cloud POISON berserker shield")
					}
					if r.BerserkShieldBlock(target) {
						return playerDamageWorldFlameShield4E17B0(target, source, cloud, damage, typ, r)
					}
				}
			}
		}
	}
	if !liveLayout() {
		return reject("unsupported live player cloud POISON raw marker record")
	}
	if update.Field76 == 0 {
		update.Field76, update.Field75 = 2, uint32(typ)
	}
	if r.GodMode != nil && r.GodMode() && target.Class().Has(object.ClassPlayer) {
		return true, true
	}
	effective := damage
	if r.QuestMode != nil && r.QuestMode() {
		if r.QuestDamageScale == nil {
			return reject("missing player cloud POISON Quest scale")
		}
		scale := r.QuestDamageScale()
		scaled := experienceSpill32TowardZero(experienceMul53TowardZero(float64(scale), float64(effective)))
		// 004E2063 uses signed FISTP under the startup control word.
		effective = netFXTruncateSignedDWORD419A70(scaled)
		if damage > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing player cloud POISON default service")
	}
	return true, r.DefaultDamage(target, source, cloud, effective, typ)
}

// 004E1C35..004E1D08 after the player cloud's marker/exclusion/facing
// prefix. State stays on the entry update; subsequent state stores and
// inventory selection follow the live object after each external callback.
func playerDamagePlayerCloudGreatSword4E17B0(
	target, source, cloud *Object, damage int32, cachedState *PlayerState,
	r PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	const typ = object.DamagePoison
	reject := func(reason string) (bool, bool, bool) {
		h, ok := playerDamageUnsupported4E17B0(r, reason, target, source, cloud, damage, typ)
		return true, h, ok
	}
	if cachedState == nil {
		return reject("missing player cloud POISON GreatSword cached state")
	}
	switch *cachedState {
	case PlayerState13, PlayerState18, PlayerState19, PlayerState20:
	default:
		return false, false, false
	}
	if r.ProjectileReflect == nil {
		return reject("missing player cloud POISON GreatSword reflection")
	}
	r.ProjectileReflect(cloud, target)
	if cloud.Class().Has(object.ClassMissile) && uint32(cloud.SubClass())&2 == 0 {
		if r.ClearOwner == nil {
			return reject("missing player cloud POISON GreatSword clear-owner")
		}
		r.ClearOwner(cloud)
		if r.SetOwner == nil {
			return reject("missing player cloud POISON GreatSword set-owner")
		}
		r.SetOwner(target, cloud)
	}
	if r.Audio == nil {
		return reject("missing player cloud POISON GreatSword audio")
	}
	r.Audio(890, target)
	livePlayer := func() bool {
		return target.Class().Has(object.ClassPlayer) && target.UpdateData != nil && target.UpdateDataPlayer().Player != nil
	}
	if !livePlayer() {
		return reject("unsupported live player cloud POISON GreatSword RNG layout")
	}
	if r.Melee.RandomInt == nil {
		return reject("missing player cloud POISON GreatSword RNG")
	}
	state := PlayerState(r.Melee.RandomInt(18, 20))
	if !livePlayer() {
		return reject("unsupported live player cloud POISON GreatSword state layout")
	}
	if r.PlayerSetState == nil {
		return reject("missing player cloud POISON GreatSword state")
	}
	r.PlayerSetState(target, state)
	if r.BlockDamagePercent == nil {
		return reject("missing player cloud POISON GreatSword balance")
	}
	amount := playerDamageBlockAmount4E17B0(r.BlockDamagePercent(), damage)
	item := playerDamageGreatSwordItem4E22A0(target)
	if r.Melee.DamageBlockWeapon == nil {
		return reject("missing player cloud POISON GreatSword durability")
	}
	if item != nil && item.HealthData != nil &&
		(r.Melee.CanDamageBlockWeapon == nil || !r.Melee.CanDamageBlockWeapon(item)) {
		return reject("unsupported live player cloud POISON GreatSword durability")
	}
	if !r.Melee.DamageBlockWeapon(item, target, source, cloud, amount, typ) && r.Unsupported != nil {
		r.Unsupported("player cloud POISON GreatSword durability failed", target, source, cloud, damage, typ)
	}
	if item == nil {
		return reject("original player cloud POISON GreatSword nil-item fault")
	}
	// Inspect the selected item after wear, even if wear relinked inventory.
	if item.Flags().Has(object.FlagDestroyed) {
		if !livePlayer() {
			return reject("unsupported live player cloud POISON GreatSword broken-item layout")
		}
		if r.PlayerSetState == nil {
			return reject("missing player cloud POISON GreatSword idle state")
		}
		r.PlayerSetState(target, PlayerState13)
	}
	return true, true, false
}
