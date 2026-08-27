package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
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
	GodMode            func() bool
	IsEnemy            func(*Object, *Object) bool
	Audio              func(int, *Object)
	BuffOff            func(*Object, EnchantID)
	ObserveClear       func(*Object)
	ItemArmorValue     func(*Object) float32
	PlayerDamageSound  func(*Object, *Object)
	PlayerDamageSoundC unsafe.Pointer
	DamageClear        func(*Object, int32)
	Unsupported        func(string, *Object, *Object, *Object, int32, object.DamageType)
}

type playerDamageItemCarry4E17B0 struct {
	value *float32
	next  float32
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
	if armorValue == 0 || runtime.ItemArmorValue == nil {
		return nil, false
	}
	var plan []playerDamageItemCarry4E17B0
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjClass.Has(object.ClassArmor) || !item.ObjFlags.Has(object.FlagEquipped) || item.HealthData == nil {
			continue
		}
		if item.UpdateData == nil {
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
		next := portion + *value
		if playerDamageRound4E17B0(next) > 0 {
			// ArmorDamage_4E1500 and its destruction report form their own
			// callback slice. Do not partly apply PlayerDamage before it is
			// available at native pointer width.
			return nil, false
		}
		plan = append(plan, playerDamageItemCarry4E17B0{value: value, next: next})
	}
	return plan, true
}

// PlayerDamageNative4E17B0 restores the ordinary Spider BITE branch of
// GAME.EXE 004E17B0 together with the relevant unit-default-damage tail. It
// returns handled=false before mutation for spell, projectile, block,
// modifier, armor-break, and quest-scaling branches that remain separate
// ports.
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
	if typ != object.DamageBite || damage <= 0 || source == nil || weapon == nil || source != weapon ||
		!source.ObjClass.Has(object.ClassMonster) || source.UpdateData == nil {
		return playerDamageUnsupported4E17B0(runtime, "non-Spider BITE shape", target, source, weapon, damage, typ)
	}
	if runtime.QuestMode != nil && runtime.QuestMode() {
		return playerDamageUnsupported4E17B0(runtime, "quest damage scaling", target, source, weapon, damage, typ)
	}
	if target.HasEnchant(playerDamageReflectEnchant4E17B0) || target.HasEnchant(playerDamageShieldEnchant4E17B0) ||
		source.HasEnchant(EnchantID(13)) {
		return playerDamageUnsupported4E17B0(runtime, "combat enchant", target, source, weapon, damage, typ)
	}
	if player.ArmorEquip&0x3000000 != 0 || player.WeaponEquip&(0x400|0x7ff8000) != 0 {
		return playerDamageUnsupported4E17B0(runtime, "active block equipment", target, source, weapon, damage, typ)
	}
	if target.DamageSound != nil && target.DamageSound != runtime.PlayerDamageSoundC {
		return playerDamageUnsupported4E17B0(runtime, "custom player damage sound", target, source, weapon, damage, typ)
	}
	if playerDamageHasLateDefendEffect4E17B0(target) {
		return playerDamageUnsupported4E17B0(runtime, "late equipped-item defend effect", target, source, weapon, damage, typ)
	}
	if runtime.IsEnemy == nil || !runtime.IsEnemy(target, source) {
		return playerDamageUnsupported4E17B0(runtime, "non-enemy source", target, source, weapon, damage, typ)
	}

	armorValue := math.Float32frombits(update.Field57)
	fraction := math.Float32frombits(update.Field21)
	armored := float32((1.0 - float64(armorValue)) * float64(damage))
	accumulated := armored + fraction
	effective := playerDamageRound4E17B0(accumulated)
	remaining := damage - effective
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
	update.Field21 = math.Float32bits(accumulated - float32(playerDamageRound4E17B0(accumulated)))
	for _, item := range itemPlan {
		*item.value = item.next
	}
	update.Field76 = 2
	update.Field75 = math.Float32bits(float32(typ))

	if runtime.GodMode != nil && runtime.GodMode() {
		return true, true
	}
	target.Pos132 = weapon.PrevPos
	runtime.BuffOff(target, playerDamageInvisibleEnchant4E17B0)
	target.Obj130 = weapon
	target.Field131 = uint32(typ)
	target.Frame134 = frame

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
	runtime.DamageClear(target, effective)
	return true, true
}
