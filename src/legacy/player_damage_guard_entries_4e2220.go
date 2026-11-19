package legacy

/*
#include "GAME3_3.h"
*/
import "C"

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func playerDamageElectricArmorSpecial4E2220() float32 {
	return *memmap.PtrFloat32(0x587000, 201108)
}

func playerDamageElectricArmorScale4E2220(s *server.Server, target *server.Object) float32 {
	return server.PlayerDamageElectricArmorScale4E2220(target, playerDamageElectricArmorSpecial4E2220,
		server.ItemHasMaterial7Modifier4133D0, func(item *server.Object) float32 {
			if def := s.Modif.Nox_xxx_equipClothFindDefByTT413270(int(item.TypeInd)); def != nil {
				return def.DamageCoeffOrArmor64
			}
			return 0
		})
}

//export sub_4E2220
func sub_4E2220(target *nox_object_t) C.double {
	return C.double(server.PlayerDamageElectricArmorScale4E2220(asObjectS(target), playerDamageElectricArmorSpecial4E2220,
		func(item *server.Object) bool { return rawMaterial7Native4133D0(item.CObj()) != 0 },
		func(item *server.Object) float32 { return float32(armorCoefficientNative415BD0(item.CObj())) }))
}

type playerBlockWearRuntime4E22A0 struct {
	damage          func(item, target, source, effective *server.Object, amount float32, typ object.DamageType) bool
	setPlayerState  func(*server.Object, server.PlayerState) bool
	popMonsterState func(*server.Object)
}

func playerBlockWear4E22A0(target, source, weapon *server.Object, mask uint32, amount float32,
	typ object.DamageType, allowMissingArmor bool, runtime playerBlockWearRuntime4E22A0) bool {
	item := target.InvFirstItem
	for ; item != nil; item = item.InvNextItem {
		if item.Flags().Has(object.FlagEquipped) && uint32(item.SubClass())&mask != 0 {
			break
		}
	}
	effective := weapon
	if effective == nil {
		effective = source
	}
	runtime.damage(item, target, source, effective, amount, typ)
	// Retain the existing C armor adapter's missing-item check after the call.
	// The weapon entry reads the selected item's flags unconditionally.
	if allowMissingArmor && item == nil {
		return false
	}
	if !item.ObjFlags.Has(object.FlagDestroyed) {
		return false
	}
	if target.Class().Has(object.ClassPlayer) {
		runtime.setPlayerState(target, server.PlayerState13)
	} else {
		runtime.popMonsterState(target)
	}
	return true
}

//export sub_4E22A0
func sub_4E22A0(target, source, weapon *nox_object_t, mask C.uint32_t, amount C.float, typ C.int) C.int {
	return C.int(bool2int(playerBlockWear4E22A0(asObjectS(target), asObjectS(source), asObjectS(weapon),
		uint32(mask), float32(amount), object.DamageType(typ), false, playerBlockWearRuntime4E22A0{
			damage:          playerDamageWeaponNative4E1560,
			setPlayerState:  Nox_xxx_playerSetState_4FA020,
			popMonsterState: func(o *server.Object) { o.MonsterPopAction() },
		})))
}

//export sub_4E2330
func sub_4E2330(target, source, weapon *nox_object_t, mask C.uint32_t, amount C.float, typ C.int) C.int {
	return C.int(bool2int(playerBlockWear4E22A0(asObjectS(target), asObjectS(source), asObjectS(weapon),
		uint32(mask), float32(amount), object.DamageType(typ), true, playerBlockWearRuntime4E22A0{
			damage:          equipDamageNative4E16D0,
			setPlayerState:  Nox_xxx_playerSetState_4FA020,
			popMonsterState: func(o *server.Object) { o.MonsterPopAction() },
		})))
}
