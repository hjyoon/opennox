package server

import (
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

const (
	playerDamageInvulnerableEnchant4E17B0 = EnchantID(23)
	playerDamageReflectEnchant4E17B0      = EnchantID(27)
	playerDamageShieldEnchant4E17B0       = EnchantID(26)
	playerDamageInvisibleEnchant4E17B0    = EnchantID(0)
	playerDamageInvulnerableSound4E17B0   = 71
)

// PlayerDamageRuntime4E17B0 contains the services called by the native-width
// PlayerDamage slice. Unsupported is reported before this slice changes any
// object state, so a caller can keep an unported branch visible without
// entering the PE32 callback on a 64-bit host.
type PlayerDamageRuntime4E17B0 struct {
	Frame               func() uint32
	CoopMode            func() bool
	GameplayFlag1       func() bool
	QuestMode           func() bool
	QuestDamageScale    func() float32
	GodMode             func() bool
	IsEnemy             func(*Object, *Object) bool
	SentryGlobeType     uint16
	GameBallType        uint16
	GameBallOnDamage    func(*Object, *Object, int32)
	Audio               func(int, *Object)
	BuffOff             func(*Object, EnchantID)
	ObserveClear        func(*Object)
	ItemArmorValue      func(*Object) float32
	ApplyArmorDefend    func(*ModifierEff, *Object, *Object, *Object, *Object, *float32) bool
	CanDamageArmor      func(*Object) bool
	DamageArmor         func(*Object, *Object, *Object, int32, object.DamageType) bool
	ReportArmorHealth   func(*Object, *Object, uint16, uint16)
	CanApplyLateDefend  func(*ModifierEff) bool
	ApplyLateDefend     func(*ModifierEff, *Object, *Object, *Object, *Object, int32, object.DamageType) int32
	BlockSourceExcluded func(*Object) bool
	BlockDirection      func(*Object, types.Pointf) bool
	BerserkShieldBlock  func(*Object) bool
	ProjectileReflect   func(*Object, *Object)
	ClearOwner          func(*Object)
	SetOwner            func(*Object, *Object)
	ChangeOwner         func(*Object, *Object)
	PointFX             func(int, types.Pointf)
	BlockDamagePercent  func() float64
	CanDamageBlockItem  func(*Object) bool
	DamageBlockItem     func(*Object, *Object, *Object, *Object, float32, object.DamageType) bool
	PlayerSetState      func(*Object, PlayerState) bool
	FireProtection      func(*Object) float64
	ElectricArmorScale  func(*Object) float32
	BalanceFloatInd     func(string, int) float64
	AdjustHP            func(*Object, int32)
	VampirismFX         func(int, image.Point, image.Point, uint16)
	PlayerDamageSound   func(*Object, *Object)
	PlayerDamageSoundC  unsafe.Pointer
	ShieldReduce        func(*Object, *int32, object.DamageType, *Object)
	DamageClear         func(*Object, int32)
	DefaultDamage       func(*Object, *Object, *Object, int32, object.DamageType) bool
	Unsupported         func(string, *Object, *Object, *Object, int32, object.DamageType)
}

func playerDamageReflectShield4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	if !target.HasEnchant(playerDamageReflectEnchant4E17B0) {
		return false, false, false
	}
	attack := weapon
	if attack == nil {
		attack = source
	}
	if attack == nil {
		return false, false, false
	}
	missile := attack.ObjClass.Has(object.ClassMissile)
	electric := typ == object.DamageZapRay || typ == object.DamageAirborneElectric
	if !missile && !electric {
		return false, false, false
	}
	if runtime.BlockDirection == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing Reflect Shield direction service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if !runtime.BlockDirection(target, attack.PosVec) {
		return false, false, false
	}
	transferOwner := missile && uint32(attack.ObjSubClass)&0x40 == 0
	changeOwner := missile && uint32(attack.ObjSubClass)&2 != 0
	pointFX := typ == object.DamageZapRay
	if runtime.Audio == nil ||
		(missile && runtime.ProjectileReflect == nil) ||
		(transferOwner && (runtime.ClearOwner == nil || runtime.SetOwner == nil)) ||
		(changeOwner && runtime.ChangeOwner == nil) ||
		(pointFX && runtime.PointFX == nil) {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing Reflect Shield effect service", target, source, weapon, damage, typ)
		return true, handled, result
	}

	update := target.UpdateDataPlayer()
	update.Field76 = 0
	if update.Player.ObserveTarget() != nil && runtime.ObserveClear != nil {
		runtime.ObserveClear(target)
	}
	if missile {
		runtime.ProjectileReflect(attack, target)
		if transferOwner {
			runtime.ClearOwner(attack)
			runtime.SetOwner(target, attack)
		}
		if changeOwner {
			runtime.ChangeOwner(attack, target)
		}
	}
	if pointFX {
		runtime.PointFX(132, target.PosVec)
	}
	runtime.Audio(122, target)
	return true, true, false
}

func playerDamageShieldItem4E17B0(target *Object) *Object {
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.ObjFlags.Has(object.FlagEquipped) && uint32(item.ObjSubClass)&2 != 0 {
			return item
		}
	}
	return nil
}

func playerDamageShieldBlock4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (applicable, handled, result bool) {
	update := target.UpdateDataPlayer()
	player := update.Player
	if player.ArmorEquip&0x3000000 == 0 {
		return false, false, false
	}
	shieldStance := update.State == PlayerState16
	if !shieldStance && update.State == PlayerState1 && player.WeaponEquip&0x400 == 0 {
		if runtime.BerserkShieldBlock == nil {
			handled, result = playerDamageUnsupported4E17B0(runtime, "missing berserker shield service", target, source, weapon, damage, typ)
			return true, handled, result
		}
		shieldStance = runtime.BerserkShieldBlock(target)
	}
	if !shieldStance {
		return false, false, false
	}
	attack := weapon
	if attack == nil {
		attack = source
	}
	if attack == nil {
		return false, false, false
	}
	if runtime.BlockSourceExcluded == nil || runtime.BlockDirection == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing shield direction service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if runtime.BlockSourceExcluded(attack) || !runtime.BlockDirection(target, attack.PrevPos) {
		return false, false, false
	}
	reflectProjectile := attack.ObjClass.Has(object.ClassMissile) && uint32(attack.ObjSubClass)&0x70 == 0
	transferOwner := reflectProjectile && uint32(attack.ObjSubClass)&2 == 0
	if reflectProjectile && runtime.ProjectileReflect == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing projectile reflection service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if transferOwner && (runtime.ClearOwner == nil || runtime.SetOwner == nil) {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing projectile owner service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	if runtime.Audio == nil || runtime.BlockDamagePercent == nil || runtime.DamageBlockItem == nil {
		handled, result = playerDamageUnsupported4E17B0(runtime, "missing shield damage service", target, source, weapon, damage, typ)
		return true, handled, result
	}
	shield := playerDamageShieldItem4E17B0(target)
	if shield != nil && (runtime.CanDamageBlockItem == nil || !runtime.CanDamageBlockItem(shield) || runtime.PlayerSetState == nil) {
		handled, result = playerDamageUnsupported4E17B0(runtime, "shield durability callback", target, source, weapon, damage, typ)
		return true, handled, result
	}
	update.Field76 = 0
	if player.ObserveTarget() != nil && runtime.ObserveClear != nil {
		runtime.ObserveClear(target)
	}
	if typ == object.DamageImpale && source != nil && weapon != nil && source != weapon {
		// 004E1A4D latches a distinct weapon before the block audio/effects.
		// Existing non-PIERCE slices keep their separate marker contracts.
		update.Field76 = 1
		update.Field75 = uint32(weapon.TypeInd)
	}
	runtime.Audio(878, target)
	if reflectProjectile {
		runtime.ProjectileReflect(attack, target)
		if transferOwner {
			runtime.ClearOwner(attack)
			runtime.SetOwner(target, attack)
		}
	}
	if shield != nil {
		amount := float32(runtime.BlockDamagePercent() * float64(damage))
		if !runtime.DamageBlockItem(shield, target, source, weapon, amount, typ) && runtime.Unsupported != nil {
			runtime.Unsupported("shield durability failed", target, source, weapon, damage, typ)
		}
		if shield.ObjFlags.Has(object.FlagDestroyed) {
			runtime.PlayerSetState(target, PlayerState13)
		}
	}
	return true, true, false
}

type playerDamageItemCarry4E17B0 struct {
	item   *Object
	value  *float32
	next   float32
	damage int32
}

type playerDamageLateDefend4E1320 struct {
	item     *Object
	modifier *ModifierEff
}

func playerDamageUnsupported4E17B0(
	runtime PlayerDamageRuntime4E17B0,
	reason string,
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
) (bool, bool) {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, target, source, weapon, damage, typ)
	}
	return false, false
}

func playerDamageRound4E17B0(value float32) int32 {
	return int32(math.RoundToEven(float64(value)))
}

func playerDamageOwnsType4E17B0(owner *Object, typeInd uint16) bool {
	for obj := owner.Field129; obj != nil; obj = obj.Field128 {
		if obj.TypeInd == typeInd {
			return true
		}
	}
	return false
}

func playerDamagePlanLateDefend4E1320(
	target *Object,
	runtime PlayerDamageRuntime4E17B0,
) ([]playerDamageLateDefend4E1320, bool) {
	var plan []playerDamageLateDefend4E1320
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) ||
			!item.ObjClass.HasAny(object.ClassFlag|object.ClassWeapon|object.ClassArmor|object.ClassWand) ||
			item.InitData == nil {
			continue
		}
		modifiers := item.InitDataModifier().Modifiers
		for i := 2; i < len(modifiers); i++ {
			modifier := modifiers[i]
			if modifier == nil || modifier.Defend76.Fnc == nil {
				continue
			}
			if runtime.CanApplyLateDefend == nil || runtime.ApplyLateDefend == nil ||
				!runtime.CanApplyLateDefend(modifier) {
				return nil, false
			}
			plan = append(plan, playerDamageLateDefend4E1320{item: item, modifier: modifier})
		}
	}
	return plan, true
}

func playerDamagePlanArmorCarry4E17B0(
	target, source, weapon *Object,
	armorValue float32,
	remaining int32,
	runtime PlayerDamageRuntime4E17B0,
) ([]playerDamageItemCarry4E17B0, bool) {
	if remaining == 0 {
		return nil, true
	}
	var plan []playerDamageItemCarry4E17B0
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjClass.Has(object.ClassArmor) || !item.ObjFlags.Has(object.FlagEquipped) || item.HealthData == nil {
			continue
		}
		if item.UpdateData == nil || item.InitData == nil {
			return nil, false
		}
		if armorValue == 0 || runtime.ItemArmorValue == nil {
			return nil, false
		}
		portion := float32(float64(runtime.ItemArmorValue(item)) / float64(armorValue) * float64(remaining))
		modifier := item.InitDataModifier().Modifiers[1]
		if modifier != nil && modifier.Defend76.Fnc != nil {
			if runtime.ApplyArmorDefend == nil ||
				!runtime.ApplyArmorDefend(modifier, item, target, weapon, source, &portion) {
				return nil, false
			}
		}
		value := (*float32)(item.UpdateData)
		total := portion + *value
		damage := playerDamageRound4E17B0(total)
		if damage > 0 && (runtime.CanDamageArmor == nil || !runtime.CanDamageArmor(item) || runtime.DamageArmor == nil) {
			return nil, false
		}
		plan = append(plan, playerDamageItemCarry4E17B0{
			item: item, value: value, next: total - float32(damage), damage: damage,
		})
	}
	return plan, true
}

// playerDamageMonster4E17B0 restores the native-width monster half of
// PlayerDamage used by scripted NPCs. In particular, War01A's wizard setpiece
// sends Bryan's MorningStar CRUSH and the wizard's weaponless
// AIRBORNE_ELECTRIC hits through this callback. The original accepts monsters
// whose subclass has bit 0x10, shares their fractional-damage accumulator with
// players, damages equipped armor, and then enters DefaultDamage.
func playerDamageMonster4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if target.UpdateData == nil || uint32(target.SubClass())&0x10 == 0 {
		return playerDamageUnsupported4E17B0(runtime, "unsupported monster target", target, source, weapon, damage, typ)
	}
	if target.ObjFlags.HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	frame := uint32(0)
	if runtime.Frame != nil {
		frame = runtime.Frame()
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}

	update := target.UpdateDataMonster()
	crush := typ == object.DamageCrush && source != nil && source.Class().Has(object.ClassMonster) &&
		source.UpdateData != nil && weapon != nil && weapon.Class().Has(object.ClassWeapon)
	airborneElectric := typ == object.DamageAirborneElectric && source != nil &&
		source.Class().Has(object.ClassMonster) && source.UpdateData != nil && weapon == nil
	if !crush && !airborneElectric {
		return playerDamageUnsupported4E17B0(runtime, "unsupported monster damage shape", target, source, weapon, damage, typ)
	}
	// Reflect Shield precedes the damage-type switch and monster shield blocks
	// require their own action/equipment effects. Keep those uncommon branches
	// fail-closed until they have native-width effect ports.
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) {
		return playerDamageUnsupported4E17B0(runtime, "monster Reflect Shield", target, source, weapon, damage, typ)
	}
	if crush && update.ArmorEquipFlags&0x3000000 != 0 && target.MonsterActionGet50A020() == 21 {
		return playerDamageUnsupported4E17B0(runtime, "monster shield block", target, source, weapon, damage, typ)
	}
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	if airborneElectric && runtime.ElectricArmorScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing electric armor service", target, source, weapon, damage, typ)
	}
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, source, weapon, damage, typ)
	}

	armorValue := math.Float32frombits(update.Field518)
	accumulated := math.Float32frombits(update.Field1)
	remaining := damage
	if crush {
		scaled := float32((1.0 - float64(armorValue)*0.5) * float64(damage))
		accumulated += scaled
	} else {
		scaled := float32(float64(runtime.ElectricArmorScale(target)) * float64(damage))
		accumulated += scaled
	}
	effective := playerDamageRound4E17B0(accumulated)
	if crush {
		remaining = damage - effective
	}
	itemPlan, ok := playerDamagePlanArmorCarry4E17B0(target, source, weapon, armorValue, remaining, runtime)
	if !ok {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}

	update.Field547 = 0
	update.Field1 = math.Float32bits(accumulated - float32(playerDamageRound4E17B0(accumulated)))
	for _, planned := range itemPlan {
		*planned.value = planned.next
		if planned.damage <= 0 {
			continue
		}
		health := planned.item.HealthData
		before := health.Cur
		runtime.DamageArmor(planned.item, source, weapon, planned.damage, typ)
		after := health.Cur
		if before != after && runtime.ReportArmorHealth != nil {
			runtime.ReportArmorHealth(target, planned.item, before, after)
		}
	}
	if weapon != nil && weapon != source {
		update.Field547 = 1
		update.Field546 = uint32(weapon.TypeInd)
	} else if weapon == nil && (typ == object.DamageClaw || typ == object.DamageCrush) {
		update.Field547 = 1
		update.Field546 = uint32(source.TypeInd)
	}
	if update.Field547 == 0 {
		update.Field547 = 2
		update.Field546 = uint32(typ)
	}
	if quest {
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

// playerDamageElectricPlayer4E17B0 restores the cases 9/17 at 004E1DF1.
// Electric armor scales HP damage, but 004E1E1E distributes the original
// damage (not the absorbed difference) to equipped armor. The caller has
// already checked the Player, observer, Coop and Reflect Shield gates.
func playerDamageElectricPlayer4E17B0(
	target, source *Object, damage int32, typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if runtime.ElectricArmorScale == nil || runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing player electric service", target, source, nil, damage, typ)
	}
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, nil, damage, typ)
	}
	update := target.UpdateDataPlayer()
	scaled := float32(float64(runtime.ElectricArmorScale(target)) * float64(damage))
	accumulated := scaled + math.Float32frombits(update.Field21)
	effective := playerDamageRound4E17B0(accumulated)
	armorValue := math.Float32frombits(update.Field57)
	itemPlan, ok := playerDamagePlanArmorCarry4E17B0(target, source, nil, armorValue, damage, runtime)
	if !ok {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, nil, damage, typ)
	}
	update.Field76 = 0
	if update.Player.ObserveTarget() != nil && runtime.ObserveClear != nil {
		runtime.ObserveClear(target)
	}
	update.Field21 = math.Float32bits(accumulated - float32(effective))
	for _, planned := range itemPlan {
		*planned.value = planned.next
		if planned.damage <= 0 {
			continue
		}
		health := planned.item.HealthData
		before := health.Cur
		runtime.DamageArmor(planned.item, source, nil, planned.damage, typ)
		after := health.Cur
		if before != after && runtime.ReportArmorHealth != nil {
			runtime.ReportArmorHealth(target, planned.item, before, after)
		}
	}
	if update.Field76 == 0 {
		update.Field76 = 2
		// 004E1E49 copies the incoming DWORD type, not float32(type).
		update.Field75 = uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if quest {
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, nil, effective, typ)
}

// playerDamageMissilePierce4E17B0 restores switch case 3 at 004E1F84.
// The armor absorption was captured by the caller before direction/effect
// callbacks; 004E2180 reads the live armor value again for durability. The
// distinct missile marker precedes durability and survives the switch tail.
func playerDamageMissilePierce4E17B0(
	target, source, weapon *Object, update *PlayerUpdateData, armorValue float32,
	damage int32, typ object.DamageType, runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if runtime.DefaultDamage == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing default damage service", target, source, weapon, damage, typ)
	}
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	// GreatStaff can reflect missiles in states 13/18/19/20 before this
	// switch. Do not silently turn that separate, unported defense into HP
	// damage, including when its facing predicate has not been evaluated.
	if update.Player.WeaponEquip&0x400 != 0 &&
		(update.State == PlayerState13 || update.State == PlayerState18 || update.State == PlayerState19 || update.State == PlayerState20) {
		return playerDamageUnsupported4E17B0(runtime, "player GreatStaff block", target, source, weapon, damage, typ)
	}
	if applicable, h, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return h, result
	}
	if update.Player.ObserveTarget() != nil {
		// Possession can replace the live update/carry during ObserveClear.
		// That prefix requires its own ordered native slice.
		return playerDamageUnsupported4E17B0(runtime, "possessed player missile PIERCE", target, source, weapon, damage, typ)
	}
	scaled := float32((1.0 - float64(armorValue)) * float64(damage))
	live := target.UpdateDataPlayer()
	accumulated := scaled + math.Float32frombits(live.Field21)
	effective := playerDamageRound4E17B0(accumulated)
	remaining := damage - effective
	itemPlan, ok := playerDamagePlanArmorCarry4E17B0(target, source, weapon, math.Float32frombits(live.Field57), remaining, runtime)
	if !ok {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	update.Field76 = 0
	update.Field76 = 1
	update.Field75 = uint32(weapon.TypeInd)
	live.Field21 = math.Float32bits(accumulated - float32(effective))
	for _, planned := range itemPlan {
		*planned.value = planned.next
		if planned.damage <= 0 {
			continue
		}
		health := planned.item.HealthData
		before := health.Cur
		runtime.DamageArmor(planned.item, source, weapon, planned.damage, typ)
		after := health.Cur
		if before != after && runtime.ReportArmorHealth != nil {
			runtime.ReportArmorHealth(target, planned.item, before, after)
		}
	}
	if update.Field76 == 0 {
		update.Field76 = 2
		update.Field75 = uint32(typ)
	}
	if damage > 0 && effective == 0 {
		effective = 1
	}
	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if quest {
		before := effective
		effective = playerDamageRound4E17B0(float32(float64(runtime.QuestDamageScale()) * float64(effective)))
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	return true, runtime.DefaultDamage(target, source, weapon, effective, typ)
}

// PlayerDamageNative4E17B0 restores the ordinary Spider BITE, monster-fired
// missile IMPACT/PIERCE, Berserker Charge CRUSH, SentryGlobe ZAP_RAY, world FLAME,
// unarmed monster ELECTRIC/AIRBORNE_ELECTRIC, and source-less LAVA/POISON branches of
// GAME.EXE 004E17B0 together with their relevant unit-default-damage tails,
// plus the front-facing shield block and the common Quest damage scaling tail,
// and the early Reflect Shield and Coop self-damage gates. It returns
// handled=false before mutation for spell and modifier branches that remain
// separate ports.
func PlayerDamageNative4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if target == nil {
		return playerDamageUnsupported4E17B0(runtime, "non-player target", target, source, weapon, damage, typ)
	}
	if target.ObjClass.Has(object.ClassMonster) {
		return playerDamageMonster4E17B0(target, source, weapon, damage, typ, runtime)
	}
	if !target.ObjClass.Has(object.ClassPlayer) || target.UpdateData == nil {
		return playerDamageUnsupported4E17B0(runtime, "non-player target", target, source, weapon, damage, typ)
	}
	if target.ObjFlags.HasAny(object.FlagNoUpdate | object.FlagDead) {
		return true, false
	}
	frame := uint32(0)
	if runtime.Frame != nil {
		frame = runtime.Frame()
	}
	if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
		if byte(frame)&3 == 0 && runtime.Audio != nil {
			runtime.Audio(playerDamageInvulnerableSound4E17B0, target)
		}
		return true, true
	}
	update := target.UpdateDataPlayer()
	player := update.Player
	if player == nil {
		return playerDamageUnsupported4E17B0(runtime, "nil player info", target, source, weapon, damage, typ)
	}
	if player.Field3680&1 != 0 {
		return true, false
	}
	if runtime.CoopMode != nil && runtime.CoopMode() && source != nil &&
		source.FindOwnerChainPlayer() == target && typ != object.DamageManaBomb {
		return true, false
	}
	pierceArmorValue := math.Float32frombits(update.Field57)
	if applicable, handled, result := playerDamageReflectShield4E17B0(target, source, weapon, damage, typ, runtime); applicable {
		return handled, result
	}
	if typ == object.DamageImpale && damage > 0 && source != nil && source != weapon &&
		source.Class().Has(object.ClassMonster) && source.UpdateData != nil &&
		weapon != nil && weapon.Class().Has(object.ClassMissile) &&
		!weapon.Class().HasAny(object.MaskUnits|object.ClassWand) &&
		!defaultDamageAttackQualifies4E1400(source, weapon) {
		// Stock GolemArrow is also WEAPON, subclass 0x10. Admit its
		// ranged predicate without silently including melee Shock shapes.
		return playerDamageMissilePierce4E17B0(target, source, weapon, update, pierceArmorValue, damage, typ, runtime)
	}
	if (typ == object.DamageElectric || typ == object.DamageAirborneElectric) && damage > 0 &&
		source != nil && source.Class().Has(object.ClassMonster) && source.UpdateData != nil && weapon == nil {
		// Ordinary shield/sword blocks explicitly exclude type 9/17.
		return playerDamageElectricPlayer4E17B0(target, source, damage, typ, runtime)
	}
	lava := typ == object.DamageLava && damage > 0 && source == nil && weapon == nil
	poison := typ == object.DamagePoison && damage > 0 && source == nil && weapon == nil
	flame := typ == object.DamageFlame && damage > 0 && weapon != nil &&
		(source == nil || source == weapon) && weapon.ObjClass.Has(object.ClassFire)
	bite := typ == object.DamageBite && damage > 0 && source != nil && weapon != nil && source == weapon &&
		source.ObjClass.Has(object.ClassMonster) && source.UpdateData != nil
	missileImpact := typ == object.DamageImpact && damage > 0 && source != nil && weapon != nil && source != weapon &&
		source.ObjClass.Has(object.ClassMonster) && source.UpdateData != nil && weapon.ObjClass.Has(object.ClassMissile)
	playerCharge := typ == object.DamageCrush && damage > 0 && source != nil && source == weapon &&
		source.ObjClass.Has(object.ClassPlayer) && !source.ObjClass.HasAny(object.ClassMonster|object.ClassWeapon|object.ClassWand)
	sentryZapRayCandidate := typ == object.DamageZapRay && damage > 0 && source != nil && weapon != nil &&
		source.ObjClass.Has(object.ClassPlayer)
	if sentryZapRayCandidate && runtime.SentryGlobeType == 0 {
		return playerDamageUnsupported4E17B0(runtime, "missing SentryGlobe type", target, source, weapon, damage, typ)
	}
	sentryZapRay := sentryZapRayCandidate && weapon.TypeInd == runtime.SentryGlobeType
	if !flame && !lava && !poison && !bite && !missileImpact && !playerCharge && !sentryZapRay {
		return playerDamageUnsupported4E17B0(runtime, "unsupported player damage shape", target, source, weapon, damage, typ)
	}
	vampirism := (bite || missileImpact || playerCharge || sentryZapRay) && source.HasEnchant(damageVampirismEnchant4E0B30)
	if sentryZapRay {
		// sub_4E1400 is false for the actual SentryGlobe class. Keeping the
		// accepted shape equally narrow avoids silently skipping its separate
		// friendly-hit and Shock-retaliation branches for weapon-like objects.
		if weapon.ObjClass.HasAny(object.ClassMonster | object.ClassWeapon | object.ClassWand) {
			return playerDamageUnsupported4E17B0(runtime, "unexpected SentryGlobe class", target, source, weapon, damage, typ)
		}
	}
	if sentryZapRay || playerCharge {
		// PLAYER charge weapons also make sub_4E1400 false. Only the
		// general owner/friendly-fire gate applies; no melee Shock retaliation.
		if runtime.GameplayFlag1 == nil || runtime.IsEnemy == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing friendly-fire service", target, source, weapon, damage, typ)
		}
		if damage >= 20 && runtime.PlayerSetState == nil {
			return playerDamageUnsupported4E17B0(runtime, "missing player hurt-state service", target, source, weapon, damage, typ)
		}
	}
	if bite || missileImpact || playerCharge || sentryZapRay {
		if applicable, handled, result := playerDamageShieldBlock4E17B0(target, source, weapon, damage, typ, runtime); applicable {
			return handled, result
		}
	}
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	shielded := !poison && target.HasEnchant(playerDamageShieldEnchant4E17B0) &&
		(typ != object.DamageManaBomb || source != target)
	if shielded && runtime.ShieldReduce == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing Shield reduction service", target, source, weapon, damage, typ)
	}
	if target.DamageSound != nil && target.DamageSound != runtime.PlayerDamageSoundC {
		return playerDamageUnsupported4E17B0(runtime, "custom player damage sound", target, source, weapon, damage, typ)
	}
	lateDefendPlan, ok := playerDamagePlanLateDefend4E1320(target, runtime)
	if !ok {
		return playerDamageUnsupported4E17B0(runtime, "late equipped-item defend effect", target, source, weapon, damage, typ)
	}
	if bite && (runtime.IsEnemy == nil || !runtime.IsEnemy(target, source)) {
		return playerDamageUnsupported4E17B0(runtime, "non-enemy source", target, source, weapon, damage, typ)
	}
	if (flame || lava) && runtime.FireProtection == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing source-less damage service", target, source, weapon, damage, typ)
	}
	if quest && runtime.QuestDamageScale == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing quest damage service", target, source, weapon, damage, typ)
	}
	if vampirism && (runtime.Audio == nil || runtime.BalanceFloatInd == nil ||
		runtime.AdjustHP == nil || runtime.VampirismFX == nil) {
		return playerDamageUnsupported4E17B0(runtime, "missing Vampirism service", target, source, weapon, damage, typ)
	}

	armorValue := math.Float32frombits(update.Field57)
	effective := damage
	remaining := damage
	accumulated := math.Float32frombits(update.Field21)
	armorReduced := bite || missileImpact || playerCharge
	if armorReduced {
		absorption := float64(armorValue)
		if playerCharge {
			// Original CRUSH case 004E1EE8 uses half the armor absorption;
			// absorbed damage still goes through the usual durability pass.
			absorption *= 0.5
		}
		armored := float32((1.0 - absorption) * float64(damage))
		accumulated = armored + accumulated
		effective = playerDamageRound4E17B0(accumulated)
		remaining = damage - effective
	} else if poison || sentryZapRay {
		// POISON and ZAP_RAY are cases 5 and 16 in the original switch: they
		// change the player damage marker but do not run the armor-durability
		// pass.
		remaining = 0
	}
	itemPlan, ok := playerDamagePlanArmorCarry4E17B0(target, source, weapon, armorValue, remaining, runtime)
	if !ok {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	if effective == 0 {
		effective = 1
	}
	if runtime.DamageClear == nil || (!poison && !flame && runtime.BuffOff == nil) {
		return playerDamageUnsupported4E17B0(runtime, "missing native damage service", target, source, weapon, damage, typ)
	}
	// DefaultDamage releases a carried GameBall using damage after armor,
	// Quest scaling and late defend, but before Shield absorption. Validate
	// that tail before any stores; callback-dependent adjustments may raise
	// a sub-threshold amount, so they also require the service for carriers.
	mayDropBall := effective >= 30 || quest || len(lateDefendPlan) != 0 || flame || lava
	if mayDropBall && target.Field129 != nil {
		if runtime.GameBallType == 0 {
			return playerDamageUnsupported4E17B0(runtime, "missing GameBall type", target, source, weapon, damage, typ)
		}
		if playerDamageOwnsType4E17B0(target, runtime.GameBallType) {
			if runtime.GameBallOnDamage == nil {
				return playerDamageUnsupported4E17B0(runtime, "GameBall drop", target, source, weapon, damage, typ)
			}
			// The original high-damage release dereferences the attacker's
			// team after detaching the ball. Do not partially mutate a
			// source-less carrier on that still-unsupported faulting path.
			if source == nil {
				return playerDamageUnsupported4E17B0(runtime, "source-less GameBall drop", target, source, weapon, damage, typ)
			}
		}
	}

	update.Field76 = 0
	if player.ObserveTarget() != nil && runtime.ObserveClear != nil {
		runtime.ObserveClear(target)
	}
	if armorReduced {
		update.Field21 = math.Float32bits(accumulated - float32(playerDamageRound4E17B0(accumulated)))
	}
	for _, planned := range itemPlan {
		*planned.value = planned.next
		if planned.damage <= 0 {
			continue
		}
		health := planned.item.HealthData
		before := health.Cur
		runtime.DamageArmor(planned.item, source, weapon, planned.damage, typ)
		after := health.Cur
		if before != after && runtime.ReportArmorHealth != nil {
			runtime.ReportArmorHealth(target, planned.item, before, after)
		}
	}
	update.Field76 = 2
	update.Field75 = math.Float32bits(float32(typ))
	if playerCharge {
		// 004E1F42 stores literal DWORD 2, not IEEE float32(2).
		update.Field75 = uint32(object.DamageCrush)
	}

	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if quest {
		before := effective
		scaled := float32(float64(runtime.QuestDamageScale()) * float64(effective))
		effective = playerDamageRound4E17B0(scaled)
		if before > 0 && effective < 1 {
			effective = 1
		}
	}
	if (sentryZapRay || playerCharge) && !runtime.GameplayFlag1() {
		owner := source.FindOwnerChainPlayer()
		if owner != nil && owner.Class().HasAny(object.MaskUnits) &&
			!runtime.IsEnemy(target, owner) && (target != owner || quest) {
			return true, true
		}
	}
	if flame || lava || poison {
		// PlayerDamage calls DefaultDamage after the damage-type switch, so
		// the invulnerability gate is observed a second time in the original.
		if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
			if byte(frame)&3 == 0 && runtime.Audio != nil {
				runtime.Audio(playerDamageInvulnerableSound4E17B0, target)
			}
			return true, true
		}
		if flame || lava {
			protectionValue := runtime.FireProtection(target)
			if protectionValue != 0 && byte(frame)&3 == 0 && runtime.Audio != nil {
				runtime.Audio(104, target)
			}
			protection := float32(protectionValue)
			scaled := float32((1.0 - float64(protection)) * float64(effective))
			effective = playerDamageRound4E17B0(scaled)
			if effective == 0 {
				effective = 1
			}
		}
		target.Pos132 = types.Pointf{}
	} else {
		target.Pos132 = weapon.PrevPos
	}
	if !poison && !flame {
		runtime.BuffOff(target, playerDamageInvisibleEnchant4E17B0)
	}
	for _, planned := range lateDefendPlan {
		effective = runtime.ApplyLateDefend(
			planned.modifier, planned.item, target, weapon, source, effective, typ,
		)
	}
	target.Obj130 = weapon
	target.Field131 = uint32(typ)
	target.Frame134 = frame

	if lava || poison {
		if runtime.PlayerDamageSound != nil {
			runtime.PlayerDamageSound(target, nil)
		}
	} else if flame || playerCharge || sentryZapRay {
		if runtime.PlayerDamageSound != nil {
			runtime.PlayerDamageSound(target, weapon)
		}
	} else {
		monsterUpdate := source.UpdateDataMonster()
		monsterHasHitSound := false
		if monsterUpdate.SoundSet122 != nil {
			monsterHasHitSound = *(*uint32)(unsafe.Add(monsterUpdate.SoundSet122, 8*4)) != 0
		}
		if !monsterHasHitSound && runtime.PlayerDamageSound != nil {
			runtime.PlayerDamageSound(target, weapon)
		}
	}
	if vampirism {
		damageApplyVampirism4E0B30(
			source, target, weapon, effective,
			runtime.Audio, runtime.BalanceFloatInd, runtime.AdjustHP, runtime.VampirismFX,
		)
	}
	if runtime.GameBallOnDamage != nil {
		runtime.GameBallOnDamage(source, target, effective)
	}
	if bite || missileImpact {
		monsterUpdate := source.UpdateDataMonster()
		if monsterUpdate.Field130 == 0 {
			monsterUpdate.Field130 = frame
		}
	}
	if (sentryZapRay || playerCharge) && effective >= 20 && update.State != PlayerState1 && update.State != PlayerState15 {
		runtime.PlayerSetState(target, PlayerState30)
	}
	if shielded {
		shieldSource := weapon
		if shieldSource == nil {
			shieldSource = source
		}
		runtime.ShieldReduce(target, &effective, typ, shieldSource)
		if effective == 0 {
			return true, false
		}
	}
	runtime.DamageClear(target, effective)
	return true, true
}
