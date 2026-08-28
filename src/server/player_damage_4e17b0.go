package server

import (
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
	Frame              func() uint32
	CoopMode           func() bool
	QuestMode          func() bool
	QuestDamageScale   func() float32
	GodMode            func() bool
	IsEnemy            func(*Object, *Object) bool
	Audio              func(int, *Object)
	BuffOff            func(*Object, EnchantID)
	ObserveClear       func(*Object)
	ItemArmorValue     func(*Object) float32
	CanDamageArmor     func(*Object) bool
	DamageArmor        func(*Object, *Object, *Object, int32, object.DamageType) bool
	ReportArmorHealth  func(*Object, *Object, uint16, uint16)
	FireProtection     func(*Object) float64
	PlayerDamageSound  func(*Object, *Object)
	PlayerDamageSoundC unsafe.Pointer
	DamageClear        func(*Object, int32)
	Unsupported        func(string, *Object, *Object, *Object, int32, object.DamageType)
}

type playerDamageItemCarry4E17B0 struct {
	item   *Object
	value  *float32
	next   float32
	damage int32
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

func playerDamageHasLateDefendEffect4E17B0(target *Object) bool {
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) ||
			!item.ObjClass.HasAny(object.ClassFlag|object.ClassWeapon|object.ClassArmor|object.ClassWand) ||
			item.InitData == nil {
			continue
		}
		modifiers := item.InitDataModifier().Modifiers
		for i := 2; i < len(modifiers); i++ {
			if modifiers[i] != nil && modifiers[i].Defend76.Fnc != nil {
				return true
			}
		}
	}
	return false
}

func playerDamagePlanArmorCarry4E17B0(
	target *Object,
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
		if item.UpdateData == nil {
			return nil, false
		}
		if armorValue == 0 || runtime.ItemArmorValue == nil {
			return nil, false
		}
		if item.InitData != nil {
			modifier := item.InitDataModifier().Modifiers[1]
			if modifier != nil && modifier.Defend76.Fnc != nil {
				return nil, false
			}
		}
		value := (*float32)(item.UpdateData)
		portion := float32(float64(runtime.ItemArmorValue(item)) / float64(armorValue) * float64(remaining))
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

// PlayerDamageNative4E17B0 restores the ordinary Spider BITE and source-less
// LAVA branches of GAME.EXE 004E17B0 together with their relevant
// unit-default-damage tails. It returns handled=false before mutation for
// spell, projectile, block, and modifier branches that remain separate ports.
func PlayerDamageNative4E17B0(
	target, source, weapon *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageRuntime4E17B0,
) (handled, result bool) {
	if target == nil || !target.ObjClass.Has(object.ClassPlayer) || target.UpdateData == nil {
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
	lava := typ == object.DamageLava && damage > 0 && source == nil && weapon == nil
	bite := typ == object.DamageBite && damage > 0 && source != nil && weapon != nil && source == weapon &&
		source.ObjClass.Has(object.ClassMonster) && source.UpdateData != nil
	if !lava && !bite {
		return playerDamageUnsupported4E17B0(runtime, "unsupported player damage shape", target, source, weapon, damage, typ)
	}
	quest := runtime.QuestMode != nil && runtime.QuestMode()
	if bite && quest {
		return playerDamageUnsupported4E17B0(runtime, "quest damage scaling", target, source, weapon, damage, typ)
	}
	if target.HasEnchant(playerDamageShieldEnchant4E17B0) ||
		(bite && (target.HasEnchant(playerDamageReflectEnchant4E17B0) || source.HasEnchant(EnchantID(13)))) {
		return playerDamageUnsupported4E17B0(runtime, "combat enchant", target, source, weapon, damage, typ)
	}
	if bite && (player.ArmorEquip&0x3000000 != 0 || player.WeaponEquip&(0x400|0x7ff8000) != 0) {
		return playerDamageUnsupported4E17B0(runtime, "active block equipment", target, source, weapon, damage, typ)
	}
	if target.DamageSound != nil && target.DamageSound != runtime.PlayerDamageSoundC {
		return playerDamageUnsupported4E17B0(runtime, "custom player damage sound", target, source, weapon, damage, typ)
	}
	if playerDamageHasLateDefendEffect4E17B0(target) {
		return playerDamageUnsupported4E17B0(runtime, "late equipped-item defend effect", target, source, weapon, damage, typ)
	}
	if bite && (runtime.IsEnemy == nil || !runtime.IsEnemy(target, source)) {
		return playerDamageUnsupported4E17B0(runtime, "non-enemy source", target, source, weapon, damage, typ)
	}
	if lava && (runtime.FireProtection == nil || (quest && runtime.QuestDamageScale == nil)) {
		return playerDamageUnsupported4E17B0(runtime, "missing lava damage service", target, source, weapon, damage, typ)
	}

	armorValue := math.Float32frombits(update.Field57)
	effective := damage
	remaining := damage
	accumulated := math.Float32frombits(update.Field21)
	if bite {
		armored := float32((1.0 - float64(armorValue)) * float64(damage))
		accumulated = armored + accumulated
		effective = playerDamageRound4E17B0(accumulated)
		remaining = damage - effective
	}
	itemPlan, ok := playerDamagePlanArmorCarry4E17B0(target, armorValue, remaining, runtime)
	if !ok {
		return playerDamageUnsupported4E17B0(runtime, "armor durability callback", target, source, weapon, damage, typ)
	}
	if effective == 0 {
		effective = 1
	}
	if runtime.DamageClear == nil || runtime.BuffOff == nil {
		return playerDamageUnsupported4E17B0(runtime, "missing native damage service", target, source, weapon, damage, typ)
	}

	update.Field76 = 0
	if player.ObserveTarget() != nil && runtime.ObserveClear != nil {
		runtime.ObserveClear(target)
	}
	if bite {
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

	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	if lava {
		if quest {
			scaled := float32(float64(runtime.QuestDamageScale()) * float64(effective))
			effective = playerDamageRound4E17B0(scaled)
			if damage > 0 && effective < 1 {
				effective = 1
			}
		}
		// PlayerDamage calls DefaultDamage after the armor pass, so the
		// invulnerability gate is observed a second time in the original.
		if target.HasEnchant(playerDamageInvulnerableEnchant4E17B0) {
			if byte(frame)&3 == 0 && runtime.Audio != nil {
				runtime.Audio(playerDamageInvulnerableSound4E17B0, target)
			}
			return true, true
		}
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
		target.Pos132 = types.Pointf{}
	} else {
		target.Pos132 = weapon.PrevPos
	}
	runtime.BuffOff(target, playerDamageInvisibleEnchant4E17B0)
	target.Obj130 = weapon
	target.Field131 = uint32(typ)
	target.Frame134 = frame

	if lava {
		if runtime.PlayerDamageSound != nil {
			runtime.PlayerDamageSound(target, nil)
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
		if monsterUpdate.Field130 == 0 {
			monsterUpdate.Field130 = frame
		}
	}
	runtime.DamageClear(target, effective)
	return true, true
}
