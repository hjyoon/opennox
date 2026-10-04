package server

import (
	"math"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// PlayerDamageMeleeRuntime4E17B0 supplies the extra services used by ordinary
// melee defenses. This is a Go callback bundle, not a persisted PE32 layout.
type PlayerDamageMeleeRuntime4E17B0 struct {
	RandomInt             func(int, int) int
	StaffWalkingBlock     func() bool
	MonsterBlockAction    func(*Object)
	MonsterPopBlockAction func(*Object)
	CanDamageBlockWeapon  func(*Object) bool
	DamageBlockWeapon     func(*Object, *Object, *Object, *Object, float32, object.DamageType) bool
}

// playerDamageMeleeShape4E17B0 identifies the ordinary armed/unarmed unit
// attacks admitted by the melee slice of GAME.EXE 004E17B0/004E0B30. Charge,
// missiles and spell projectiles retain their separately ported paths.
func playerDamageMeleeShape4E17B0(source, weapon *Object, typ object.DamageType) bool {
	if source == nil || (!source.Class().Has(object.ClassPlayer) &&
		(!source.Class().Has(object.ClassMonster) || source.UpdateData == nil)) {
		return false
	}
	if weapon == nil {
		return (typ == object.DamageClaw || typ == object.DamageCrush) &&
			(source.Class().Has(object.ClassPlayer) || uint32(source.SubClass())&0x10 != 0)
	}
	if source == weapon || weapon.Class().HasAny(object.MaskUnits|object.ClassMissile) ||
		!weapon.Class().HasAny(object.ClassWeapon|object.ClassWand) ||
		(typ != object.DamageBlade && typ != object.DamageCrush) {
		return false
	}
	// A spell staff can qualify only through its native WandUseData flags.
	// Do not enter 004E1400's original nil-fault boundary for malformed data.
	if weapon.Class().Has(object.ClassWand) && uint32(weapon.SubClass())&0x047f0000 != 0 && weapon.UseData.Ptr == nil {
		return false
	}
	return defaultDamageAttackQualifies4E1400(source, weapon)
}

// defaultDamageFriendlyException4E1470 restores the War Hammer exception in
// GAME.EXE 004E1470. This skips only the melee gate, not the earlier campaign
// owner gate in DefaultDamage.
func defaultDamageFriendlyException4E1470(weapon *Object) bool {
	return weapon != nil && weapon.Class().Has(object.ClassWeapon) && uint32(weapon.SubClass())&0x4000 != 0
}

// playerDamageMonsterBlockReady534340 is the action-ID predicate at 00534340.
// The original permits idle/wait/guard, face actions and an existing weapon
// block; attacking, casting and shield-block actions do not qualify.
func playerDamageMonsterBlockReady534340(target *Object) bool {
	switch target.MonsterActionGet50A020() {
	case ai.ACTION_IDLE, ai.ACTION_WAIT, ai.ACTION_GUARD, ai.ACTION_WEAPON_BLOCK,
		ai.ACTION_FACE_LOCATION, ai.ACTION_FACE_OBJECT, ai.ACTION_FACE_ANGLE:
		return true
	}
	return false
}

type playerDamageMeleeBlock4E17B0 struct {
	item   *Object
	sound  int
	weapon bool
}

func playerDamageMeleeBlockPlan4E17B0(
	target, source, weapon *Object,
	typ object.DamageType,
	player bool, state PlayerState, weaponFlags, armorFlags uint32,
	r PlayerDamageRuntime4E17B0,
) (playerDamageMeleeBlock4E17B0, string) {
	var plan playerDamageMeleeBlock4E17B0
	shield := armorFlags&0x3000000 != 0 &&
		((player && state == PlayerState16) || (!player && target.MonsterActionGet50A020() == ai.ACTION_BLOCK_ATTACK))
	canBerserk := player && state == PlayerState1 && weaponFlags&0x400 == 0 && armorFlags&0x3000000 != 0
	if !shield && !canBerserk && (typ != object.DamageBlade || weaponFlags&(0x400|0x7ff8000) == 0) {
		return plan, ""
	}
	if r.BlockSourceExcluded == nil || r.BlockDirection == nil {
		return plan, "missing melee block direction service"
	}
	attack := weapon
	if attack == nil {
		attack = source
	}
	if r.BlockSourceExcluded(attack) || !r.BlockDirection(target, attack.PrevPos) {
		return plan, ""
	}
	ready := state == PlayerState13 || state == PlayerState18 || state == PlayerState19 || state == PlayerState20
	if !player {
		ready = playerDamageMonsterBlockReady534340(target)
	}
	mask := uint32(2)
	if !shield && weaponFlags&0x400 != 0 && typ == object.DamageBlade {
		if !ready {
			return plan, ""
		}
		plan.sound, plan.weapon, mask = 890, true, 0x400
	} else if !shield {
		if canBerserk {
			if r.BerserkShieldBlock == nil {
				return plan, "missing berserker shield service"
			}
			shield = r.BerserkShieldBlock(target)
		}
		if !shield && weaponFlags&0x7ff8000 != 0 && typ == object.DamageBlade {
			if player {
				// 004E1D35..004E1D40 accepts idle or an existing staff
				// block, not the GreatSword's three block animations.
				ready = state == PlayerState13 || state == PlayerState21
				if state == PlayerState0 {
					if r.Melee.StaffWalkingBlock == nil {
						return plan, "missing walking staff block service"
					}
					ready = r.Melee.StaffWalkingBlock()
				}
			}
			if !ready {
				return plan, ""
			}
			plan.sound, plan.weapon, mask = 894, true, 0x7ff8000
		} else if !shield {
			return plan, ""
		}
	}
	if shield {
		plan.sound = 878
	}
	// 004E22A0/004E2330 use the first equipped subclass match, not the
	// cached equipped-weapon pointer. A missing item is an original PE32
	// nil-fault boundary, so reject it before committing the hit prefix.
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Flags().Has(object.FlagEquipped) && uint32(item.SubClass())&mask != 0 {
			plan.item = item
			break
		}
	}
	if plan.item == nil {
		return plan, "missing equipped melee block item"
	}
	if r.Audio == nil || r.BlockDamagePercent == nil ||
		(player && r.PlayerSetState == nil) || (!player && r.Melee.MonsterPopBlockAction == nil) {
		return plan, "missing melee block effect service"
	}
	if plan.weapon {
		if r.Melee.CanDamageBlockWeapon == nil || !r.Melee.CanDamageBlockWeapon(plan.item) || r.Melee.DamageBlockWeapon == nil ||
			(player && plan.sound == 890 && r.Melee.RandomInt == nil) || (!player && r.Melee.MonsterBlockAction == nil) {
			return plan, "melee block weapon callback"
		}
	} else if r.CanDamageBlockItem == nil || !r.CanDamageBlockItem(plan.item) || r.DamageBlockItem == nil {
		return plan, "melee shield durability callback"
	}
	return plan, ""
}

func playerDamageMeleeApplyBlock4E17B0(
	plan playerDamageMeleeBlock4E17B0,
	target, source, weapon *Object, damage int32, typ object.DamageType,
	player bool, r PlayerDamageRuntime4E17B0,
) {
	r.Audio(plan.sound, target)
	if plan.weapon {
		if player {
			state := PlayerState21
			if plan.sound == 890 {
				state = PlayerState(r.Melee.RandomInt(18, 20))
			}
			r.PlayerSetState(target, state)
		} else {
			r.Melee.MonsterBlockAction(target)
		}
	}
	effective := weapon
	if effective == nil {
		effective = source
	}
	amount := float32(r.BlockDamagePercent() * float64(damage))
	apply := r.DamageBlockItem
	if plan.weapon {
		apply = r.Melee.DamageBlockWeapon
	}
	if !apply(plan.item, target, source, effective, amount, typ) && r.Unsupported != nil {
		r.Unsupported("melee block durability failed", target, source, weapon, damage, typ)
	}
	if plan.item.Flags().Has(object.FlagDestroyed) {
		if player {
			r.PlayerSetState(target, PlayerState13)
		} else {
			r.Melee.MonsterPopBlockAction(target)
		}
	}
}

// PlayerDamageMeleeNative4E17B0 restores the ordinary BLADE/CRUSH and unarmed
// CLAW/CRUSH slice for players and NPC-subclass monsters. Armor and block
// defenses also serve unit-sourced SIMPLE CRUSH, including stock Fists whose
// type IDs bypass ordinary shield blocking via BlockSourceExcluded.
// durability precede the original GodMode/Quest/DefaultDamage tail. Reflect
// Shield does not intercept these non-missile, non-electric hits in GAME.EXE.
// The native possession prefix is kept fail-closed until its live update-data
// replacement ordering is ported separately.
func PlayerDamageMeleeNative4E17B0(
	target, source, weapon *Object, damage int32, typ object.DamageType,
	r PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if target == nil || (!playerDamageMeleeShape4E17B0(source, weapon, typ) &&
		!playerDamageSimpleCrushShape4E17B0(source, weapon, typ)) || !target.Class().HasAny(object.MaskUnits) {
		return playerDamageUnsupported4E17B0(r, "unsupported ordinary melee shape", target, source, weapon, damage, typ)
	}
	if target.Flags().HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	frame := uint32(0)
	if r.Frame != nil {
		frame = r.Frame()
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if byte(frame)&3 == 0 && r.Audio != nil {
			r.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	player := target.Class().Has(object.ClassPlayer)
	if player && r.CoopMode != nil && r.CoopMode() && source.FindOwnerChainPlayer() == target {
		return true, false
	}
	if target.UpdateData == nil || (!player && uint32(target.SubClass())&0x10 == 0) {
		return playerDamageUnsupported4E17B0(r, "unsupported melee target update", target, source, weapon, damage, typ)
	}
	var carry, marker, markerType *uint32
	var armorValue float32
	var weaponFlags, armorFlags uint32
	var state PlayerState
	if player {
		ud := target.UpdateDataPlayer()
		if ud.Player == nil {
			return playerDamageUnsupported4E17B0(r, "nil player info", target, source, weapon, damage, typ)
		}
		if ud.Player.Field3680&1 != 0 {
			return true, false
		}
		if ud.Player.ObserveTarget() != nil {
			return playerDamageUnsupported4E17B0(r, "possessed player melee", target, source, weapon, damage, typ)
		}
		carry, marker, markerType = &ud.Field21, &ud.Field76, &ud.Field75
		armorValue = math.Float32frombits(ud.Field57)
		weaponFlags, armorFlags, state = ud.Player.WeaponEquip, ud.Player.ArmorEquip, ud.State
	} else {
		ud := target.UpdateDataMonster()
		carry, marker, markerType = &ud.Field1, &ud.Field547, &ud.Field546
		armorValue = math.Float32frombits(ud.Field518)
		weaponFlags, armorFlags = ud.WeaponEquipFlags, ud.ArmorEquipFlags
	}
	block, reason := playerDamageMeleeBlockPlan4E17B0(target, source, weapon, typ, player, state, weaponFlags, armorFlags, r)
	if reason != "" {
		return playerDamageUnsupported4E17B0(r, reason, target, source, weapon, damage, typ)
	}
	if block.item != nil {
		*marker = 1
		attack := weapon
		if attack == nil {
			attack = source
		}
		*markerType = uint32(attack.TypeInd)
		playerDamageMeleeApplyBlock4E17B0(block, target, source, weapon, damage, typ, player, r)
		return true, false
	}
	quest := r.QuestMode != nil && r.QuestMode()
	if r.DefaultDamage == nil || (quest && r.QuestDamageScale == nil) {
		return playerDamageUnsupported4E17B0(r, "missing melee damage tail service", target, source, weapon, damage, typ)
	}
	absorption := float64(armorValue)
	if typ == object.DamageCrush {
		absorption *= 0.5
	}
	scaled := float32((1.0 - absorption) * float64(damage))
	accumulated := scaled + math.Float32frombits(*carry)
	effective := playerDamageRound4E17B0(accumulated)
	remaining := damage - effective
	if !playerDamageArmorReady4E17B0(target, r) {
		return playerDamageUnsupported4E17B0(r, "melee armor durability callback", target, source, weapon, damage, typ)
	}
	// The hit marker is already 1 at 004E1A4D/004E1AD7; it is visible to
	// armor effects and persists unless a durability callback clears it.
	*marker = 1
	attack := weapon
	if attack == nil {
		attack = source
	}
	*markerType = uint32(attack.TypeInd)
	*carry = math.Float32bits(accumulated - float32(effective))
	playerDamageApplyArmor4E17B0(target, source, weapon, remaining, typ, r)
	if *marker == 0 {
		*marker, *markerType = 2, uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if player && r.GodMode != nil && r.GodMode() {
		return true, true
	}
	if quest {
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(r.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, r.DefaultDamage(target, source, weapon, effective, typ)
}
