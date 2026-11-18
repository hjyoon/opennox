package legacy

/*
#include "GAME4_3.h"

typedef int (*npc_weapon_change_fn)(nox_object_t*, nox_object_t*);
typedef void (*quiver_dequip_fn)(nox_object_t*, int, int);
typedef void (*shield_dequip_fn)(nox_object_t*);
_Static_assert(_Generic(&nox_xxx_NPCEquipWeapon_53A2C0, npc_weapon_change_fn:1, default:0), "weapon equip ABI");
_Static_assert(_Generic(&nox_xxx_equipWeaponNPC_53A030, npc_weapon_change_fn:1, default:0), "weapon dequip ABI");
_Static_assert(_Generic(&sub_53A0F0, quiver_dequip_fn:1, default:0), "quiver dequip ABI");
_Static_assert(_Generic(&sub_53A3D0, shield_dequip_fn:1, default:0), "shield dequip ABI");

static npc_weapon_change_fn volatile equip_indirect = nox_xxx_NPCEquipWeapon_53A2C0;
static npc_weapon_change_fn volatile dequip_indirect = nox_xxx_equipWeaponNPC_53A030;
static quiver_dequip_fn volatile quiver_indirect = sub_53A0F0;
static shield_dequip_fn volatile shield_indirect = sub_53A3D0;

static int call_npc_equip(int indirect, nox_object_t* owner, nox_object_t* item) {
	return indirect ? equip_indirect(owner, item) : nox_xxx_NPCEquipWeapon_53A2C0(owner, item);
}
static int call_npc_dequip(int indirect, nox_object_t* owner, nox_object_t* item) {
	return indirect ? dequip_indirect(owner, item) : nox_xxx_equipWeaponNPC_53A030(owner, item);
}
static void call_quiver_dequip(int indirect, nox_object_t* owner, int report, int broadcast) {
	if (indirect) {
		quiver_indirect(owner, report, broadcast);
	} else {
		sub_53A0F0(owner, report, broadcast);
	}
}
static void call_shield_dequip(int indirect, nox_object_t* owner) {
	if (indirect) {
		shield_indirect(owner);
	} else {
		sub_53A3D0(owner);
	}
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func callNPCEquipWeaponNative53A2C0(route int, owner, item *server.Object) int {
	return int(C.call_npc_equip(C.int(route), asObjectC(owner), asObjectC(item)))
}

func callNPCDequipWeaponNative53A030(route int, owner, item *server.Object) int {
	return int(C.call_npc_dequip(C.int(route), asObjectC(owner), asObjectC(item)))
}

func callDequipQuiverNative53A0F0(route int, owner *server.Object, report, broadcast int32) {
	C.call_quiver_dequip(C.int(route), asObjectC(owner), C.int(report), C.int(broadcast))
}

func callDequipShieldsNative53A3D0(route int, owner *server.Object) {
	C.call_shield_dequip(C.int(route), asObjectC(owner))
}
