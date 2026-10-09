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
	// Ordinary clouds cannot enter Reflect. Do not route a callback-introduced
	// live missile through the nonmissile prefix or a PE32 fallback.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) && cloud.Class().Has(object.ClassMissile) {
		return reject("unsupported live player cloud POISON Reflect missile")
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
					return reject("unsupported live player cloud POISON GreatSword missile")
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
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if damage > 0 && effective < 1 {
			effective = 1
		}
	}
	if r.DefaultDamage == nil {
		return reject("missing player cloud POISON default service")
	}
	return true, r.DefaultDamage(target, source, cloud, effective, typ)
}
