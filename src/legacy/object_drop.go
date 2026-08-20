package legacy

/*
#include "GAME3_3.h"
#include "GAME4_3.h"
#include "aud_event_drop_4ee2f0.h"
#include "crown_drop_4ed5e0.h"
#include "food_drop_4ede50.h"
#include "glyph_drop_4ed500.h"
#include "potion_drop_4edde0.h"
#include "trap_drop_4ed580.h"
#include "treasure_drop_4ed710.h"
*/
import "C"
import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func init() {
	server.RegisterObjectDropC("DefaultDrop", C.nox_xxx_dropDefault_4ED290)
	server.RegisterObjectDropC("ArmorDrop", C.nox_xxx_dropArmor_53EB70)
	server.RegisterObjectDropC("WeaponDrop", C.nox_xxx_dropWeapon_53AB10)
	server.RegisterObjectDrop("TreasureDrop", C.nox_xxx_dropTreasure_4ED710, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return treasureDropCall4ED710(obj, obj2, pos)
	})
	server.RegisterObjectDrop("GlyphDrop", C.nox_GlyphDrop_4ED500, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return glyphDropCall4ED500(obj, obj2, pos)
	})
	server.RegisterObjectDrop("PotionDrop", C.sub_4EDDE0, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return potionDropCall4EDDE0(obj, obj2, pos)
	})
	server.RegisterObjectDrop("TrapDrop", C.nox_xxx_dropTrap_4ED580, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return trapDropCall4ED580(obj, obj2, pos)
	})
	server.RegisterObjectDrop("FoodDrop", C.nox_xxx_dropFood_4EDE50, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return foodDropCall4EDE50(obj, obj2, pos)
	})
	server.RegisterObjectDrop("CrownDrop", C.nox_xxx_dropCrown_4ED5E0, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return crownDropCall4ED5E0(obj, obj2, pos)
	})
	server.RegisterObjectDrop("AudEventDrop", C.nox_objectDropAudEvent_4EE2F0, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return audEventDropCall4EE2F0(obj, obj2, pos)
	})
	server.RegisterObjectDrop("AnkhTradableDrop", C.nox_xxx_dropAnkhTradable_4EE370, func(obj, obj2 *server.Object, pos *types.Pointf) int32 {
		return ankhTradableDropCall4EE370(obj, obj2, pos)
	})
}

func Nox_xxx_dropDefault_4ED290(obj1 *server.Object, obj2 *server.Object, a3 *types.Pointf) int {
	return int(defaultDropCall4ED290(obj1, obj2, a3))
}
