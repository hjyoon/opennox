package legacy

/*
#include "defs.h"
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "pickup_default_4f31e0.h"
#include "pickup_food_4f3350.h"
#include "pickup_use_4f34d0.h"
#include "pickup_trap_4f3510.h"
#include "pickup_treasure_4f3580.h"
#include "pickup_potion_4f37d0.h"
#include "pickup_gold_4f3a60.h"
#include "pickup_ammo_4f3b00.h"
#include "pickup_spellbook_4f3c60.h"
#include "pickup_abilitybook_4f3ce0.h"
#include "aud_event_pickup_4f3d50.h"
*/
import "C"
import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_pickupDefault_4F31E0         server.PickupFunc
	Nox_xxx_pickupFood_4F3350            func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupUse_4F34D0             func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupTrap_4F3510            func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupTreasure_4F3580        func(*server.Object, *server.Object, int32, int32) int32
	Nox_objectPickupAudEvent_4F3D50      func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupPotion_4F37D0          func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupGold_4F3A60            func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupAmmo_4F3B00            func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupSpellbook_4F3C60       func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_pickupAbilitybook_4F3CE0     func(*server.Object, *server.Object, int32, int32) int32
	Nox_xxx_playerClassCanUseItem_57B3D0 func(item *server.Object, cl player.Class) bool
	Sub_57B370                           func(cl object.Class, sub object.SubClass, typ int) byte
)

func init() {
	server.RegisterObjectPickup("DefaultPickup", C.nox_xxx_pickupDefault_4F31E0, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupDefault_4F31E0(who, it, a3, a4)
	})
	server.RegisterObjectPickup("FoodPickup", C.nox_xxx_pickupFood_4F3350, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupFood_4F3350(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("UsePickup", C.nox_xxx_pickupUse_4F34D0, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupUse_4F34D0(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("TrapPickup", C.nox_xxx_pickupTrap_4F3510, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupTrap_4F3510(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("TreasurePickup", C.nox_xxx_pickupTreasure_4F3580, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupTreasure_4F3580(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickupC("ArmorPickup", C.nox_xxx_pickupArmor_53E7F0)
	server.RegisterObjectPickupC("WeaponPickup", C.sub_53A720)
	server.RegisterObjectPickupC("OblivionPickup", C.nox_xxx_sendMsgOblivionPickup_53A9C0)
	server.RegisterObjectPickup("PotionPickup", C.nox_xxx_pickupPotion_4F37D0, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupPotion_4F37D0(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("GoldPickup", C.nox_xxx_pickupGold_4F3A60, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupGold_4F3A60(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("AmmoPickup", C.nox_xxx_pickupAmmo_4F3B00, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupAmmo_4F3B00(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("SpellBookPickup", C.nox_xxx_pickupSpellbook_4F3C60, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupSpellbook_4F3C60(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("AbilityBookPickup", C.nox_xxx_pickupAbilitybook_4F3CE0, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_xxx_pickupAbilitybook_4F3CE0(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("CrownPickup", C.sub_4F3400, func(who, crown *server.Object, a3, a4 int) bool {
		s := GetServer().S()
		return crownPickupCall4F3400(s, who, crown, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickup("AudEventPickup", C.nox_objectPickupAudEvent_4F3D50, func(who, it *server.Object, a3, a4 int) bool {
		return Nox_objectPickupAudEvent_4F3D50(who, it, int32(a3), int32(a4)) != 0
	})
	server.RegisterObjectPickupC("AnkhTradablePickup", C.sub_4F3DD0)
}

//export sub_57B370
func sub_57B370(cl, sub, typ int32) byte {
	return Sub_57B370(object.Class(cl), object.SubClass(sub), int(typ))
}

//export sub_419E10
func sub_419E10(u *nox_object_t, a2 int32) {
	GetServer().S().Players.SetXxx(asObjectS(u), a2)
}

//export sub_419E60
func sub_419E60(u *nox_object_t) int32 {
	return int32(bool2int(GetServer().S().Players.CheckXxx(asObjectS(u))))
}

//export sub_419EA0
func sub_419EA0() int32 {
	return int32(bool2int(GetServer().S().Players.AnyXxx()))
}

//export nox_xxx_playerClassCanUseItem_57B3D0
func nox_xxx_playerClassCanUseItem_57B3D0(item *nox_object_t, cl int8) int32 {
	return int32(bool2int(Nox_xxx_playerClassCanUseItem_57B3D0(asObjectS(item), player.Class(cl))))
}
