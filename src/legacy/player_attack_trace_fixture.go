package legacy

/*
#include "GAME4_3.h"

typedef int (*player_trace_fn)(nox_object_t*, nox_player_attack_info_native_t*);
typedef void (*player_hit_fn)(nox_object_t*, nox_player_attack_info_native_t*);
typedef void (*player_select_fn)(nox_object_t*, nox_object_t*);
_Static_assert(_Generic(&nox_xxx_playerTraceAttack_538330, player_trace_fn: 1, default: 0), "public attack trace ABI");
_Static_assert(_Generic(&nox_xxx_playerTraceAttackNative_538330, player_trace_fn: 1, default: 0), "native attack trace ABI");
_Static_assert(_Generic(&sub_538510, player_hit_fn: 1, default: 0), "attack damage candidate ABI");
_Static_assert(_Generic(&sub_5386A0, player_select_fn: 1, default: 0), "attack select candidate ABI");

enum {
	attack_record_size = sizeof(nox_player_attack_info_native_t),
	attack_owner_offset = offsetof(nox_player_attack_info_native_t, owner),
	attack_center_offset = offsetof(nox_player_attack_info_native_t, pos_x),
	attack_weapon_offset = offsetof(nox_player_attack_info_native_t, weapon),
	attack_mask_offset = offsetof(nox_player_attack_info_native_t, field_32),
};

static int call_player_trace_checked(nox_object_t* attacker, nox_player_attack_info_native_t* attack, int live) {
	player_trace_fn volatile fn = live ? nox_xxx_playerTraceAttackNative_538330 : nox_xxx_playerTraceAttack_538330;
	return fn(attacker, attack);
}

static void call_player_hit_checked(nox_object_t* candidate, nox_player_attack_info_native_t* attack, int indirect) {
	player_hit_fn volatile fn = sub_538510;
	if (indirect) {
		fn(candidate, attack);
	} else {
		sub_538510(candidate, attack);
	}
}

static void call_player_select_checked(nox_object_t* candidate, nox_object_t* attacker) {
	player_select_fn volatile fn = sub_5386A0;
	fn(candidate, attacker);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func attackTraceCLayout538330() [5]uintptr {
	return [5]uintptr{C.attack_record_size, C.attack_owner_offset, C.attack_center_offset, C.attack_weapon_offset, C.attack_mask_offset}
}

func callAttackTraceC538330(attacker *server.Object, info *server.PlayerAttackData538330, live bool) int {
	return int(C.call_player_trace_checked(asObjectC(attacker), (*C.nox_player_attack_info_native_t)(unsafe.Pointer(info)), C.int(bool2int(live))))
}

func callAttackHitC538510(target *server.Object, info *server.PlayerAttackData538330, indirect bool) {
	C.call_player_hit_checked(asObjectC(target), (*C.nox_player_attack_info_native_t)(unsafe.Pointer(info)), C.int(bool2int(indirect)))
}

func callAttackSelectC5386A0(candidate, attacker *server.Object) {
	C.call_player_select_checked(asObjectC(candidate), asObjectC(attacker))
}
