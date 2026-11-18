package legacy

/*
#include <stdio.h>
#include "GAME1.h"

typedef int (*game1_facade_fn)(nox_tileDef_t*);
typedef nox_video_bag_image_t* (*game1_icon_fn)(char);
typedef int (*game1_status_fn)(nox_playerInfo*, int);
typedef void (*game1_respawn_fn)(nox_playerInfo*, uint8_t);
typedef nox_playerInfo* (*game1_equip_fn)(uint8_t, int, uint32_t, const uint8_t*);
typedef nox_playerInfo* (*game1_dequip_fn)(uint8_t, int, uint32_t);
_Static_assert(_Generic(&nox_xxx_checkFacade_4117E0, game1_facade_fn:1, default:0), "facade ABI");
_Static_assert(_Generic(&sub_413420, game1_icon_fn:1, default:0), "enchantment image ABI");
_Static_assert(_Generic(&nox_xxx_netNeedTimestampStatus_4174F0, game1_status_fn:1, default:0), "player status ABI");
_Static_assert(_Generic(&nox_xxx_cliPlayerRespawn_417680, game1_respawn_fn:1, default:0), "client respawn ABI");
_Static_assert(_Generic(&nox_xxx_clientEquipWeaponArmor_417AA0, game1_equip_fn:1, default:0), "const equipment modifier ABI");
_Static_assert(_Generic(&sub_417B80, game1_dequip_fn:1, default:0), "client dequip ABI");

static game1_facade_fn volatile game1_facade_indirect = nox_xxx_checkFacade_4117E0;
static game1_icon_fn volatile game1_icon_indirect = sub_413420;
static game1_status_fn volatile game1_status_indirect = nox_xxx_netNeedTimestampStatus_4174F0;
static game1_respawn_fn volatile game1_respawn_indirect = nox_xxx_cliPlayerRespawn_417680;
static game1_equip_fn volatile game1_equip_indirect = nox_xxx_clientEquipWeaponArmor_417AA0;
static game1_dequip_fn volatile game1_dequip_indirect = sub_417B80;

static int call_game1_facade(int indirect, const char* name) {
	nox_tileDef_t tile = {0};
	snprintf(tile.name, sizeof(tile.name), "%s", name);
	return indirect ? game1_facade_indirect(&tile) : nox_xxx_checkFacade_4117E0(&tile);
}

static nox_video_bag_image_t* call_game1_icon(int indirect, char flag) {
	return indirect ? game1_icon_indirect(flag) : sub_413420(flag);
}

static int call_game1_status(int indirect, nox_playerInfo* player, int mask) {
	return indirect ? game1_status_indirect(player, mask) : nox_xxx_netNeedTimestampStatus_4174F0(player, mask);
}

static nox_playerInfo* call_game1_equip(int indirect, uint8_t op, int code, uint32_t type, const uint8_t* ids) {
	return indirect ? game1_equip_indirect(op, code, type, ids) : nox_xxx_clientEquipWeaponArmor_417AA0(op, code, type, ids);
}

static nox_playerInfo* call_game1_dequip(int indirect, uint8_t op, int code, uint32_t type) {
	return indirect ? game1_dequip_indirect(op, code, type) : sub_417B80(op, code, type);
}

static void call_game1_respawn(int indirect, nox_playerInfo* player, uint8_t mask) {
	if (indirect) {
		game1_respawn_indirect(player, mask);
	} else {
		nox_xxx_cliPlayerRespawn_417680(player, mask);
	}
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func callGame1FacadeNative(route int, name *byte) int {
	return int(C.call_game1_facade(C.int(route), (*C.char)(unsafe.Pointer(name))))
}

func callGame1IconNative(route int, flag byte) unsafe.Pointer {
	return unsafe.Pointer(C.call_game1_icon(C.int(route), C.char(flag)))
}

func callGame1StatusNative(route int, player *server.Player, mask uint32) int32 {
	return int32(C.call_game1_status(C.int(route), (*nox_playerInfo)(player.C()), C.int(int32(mask))))
}

func callGame1EquipNative(route int, op byte, code int32, itemType uint32, ids *[4]byte) *server.Player {
	return asPlayerS(C.call_game1_equip(C.int(route), C.uint8_t(op), C.int(code), C.uint32_t(itemType),
		(*C.uint8_t)(unsafe.Pointer(ids))))
}

func callGame1DequipNative(route int, op byte, code int32, itemType uint32) *server.Player {
	return asPlayerS(C.call_game1_dequip(C.int(route), C.uint8_t(op), C.int(code), C.uint32_t(itemType)))
}

func callGame1RespawnNative(route int, player *server.Player, mask byte) {
	C.call_game1_respawn(C.int(route), (*nox_playerInfo)(player.C()), C.uint8_t(mask))
}
