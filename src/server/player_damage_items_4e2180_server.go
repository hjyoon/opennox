package server

import (
	"math"

	"github.com/opennox/libs/object"
)

type PlayerDamageItemsRuntime4E2180 struct {
	ItemArmorValue func(*Object) float64
	EquipDamage    func(*Object, *Object, *Object, *Object, float32, object.DamageType)
}

// PlayerDamageItems4E2180 distributes armor wear using native Object links.
// The runtime's EquipDamage is the original 004E16D0 callee, not a precomputed
// list of item damage amounts. Player and NPC update records keep their own
// armor layouts, with the Player branch taking precedence for mixed classes.
func PlayerDamageItems4E2180(
	target, source, effective *Object,
	damage int32,
	typ object.DamageType,
	runtime PlayerDamageItemsRuntime4E2180,
) {
	playerDamageItems4E2180(target, source, effective, damage, int32(typ), playerDamageItemsHooks4E2180[*Object]{
		loadClass:       func(obj *Object) uint32 { return uint32(obj.ObjClass) },
		loadSubclassLow: func(obj *Object) uint8 { return uint8(obj.ObjSubClass) },
		loadPlayerArmor: func(obj *Object) float32 {
			return math.Float32frombits(obj.UpdateDataPlayer().Field57)
		},
		loadMonsterArmor: func(obj *Object) float32 {
			return math.Float32frombits(obj.UpdateDataMonster().Field518)
		},
		loadFirstItem:  func(obj *Object) *Object { return obj.InvFirstItem },
		loadFlags:      func(obj *Object) uint32 { return uint32(obj.ObjFlags) },
		itemArmorValue: runtime.ItemArmorValue,
		equipDamage: func(item, owner, source, effective *Object, amount float32, typ int32) {
			runtime.EquipDamage(item, owner, source, effective, amount, object.DamageType(typ))
		},
		loadNextItem: func(obj *Object) *Object { return obj.InvNextItem },
	})
}
