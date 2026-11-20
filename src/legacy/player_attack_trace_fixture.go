package legacy

/*
#include "GAME4_3.h"
#include "GAME4_1.h"

typedef int (*player_trace_fn)(nox_object_t*, nox_player_attack_info_native_t*);
typedef void (*player_hit_fn)(nox_object_t*, nox_player_attack_info_native_t*);
typedef void (*player_select_fn)(nox_object_t*, nox_object_t*);
typedef float2* (*unit_radius_scan_fn)(float2*, float, nox_unit_radius_callback_t, void*);
typedef void (*unit_radius_candidate_fn)(nox_object_t*, nox_unit_radius_context_t*);
_Static_assert(_Generic(&nox_xxx_playerTraceAttack_538330, player_trace_fn: 1, default: 0), "public attack trace ABI");
_Static_assert(_Generic(&nox_xxx_playerTraceAttackNative_538330, player_trace_fn: 1, default: 0), "native attack trace ABI");
_Static_assert(_Generic(&sub_538510, player_hit_fn: 1, default: 0), "attack damage candidate ABI");
_Static_assert(_Generic(&sub_5386A0, player_select_fn: 1, default: 0), "attack select candidate ABI");
_Static_assert(_Generic(&sub_518040, unit_radius_scan_fn: 1, default: 0), "unit radius scan ABI");
_Static_assert(_Generic(&sub_5180B0, unit_radius_candidate_fn: 1, default: 0), "unit radius candidate ABI");
_Static_assert(sizeof(nox_unit_radius_context_t) == (sizeof(void*) == 4 ? 16 : 32), "unit radius context width");

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

typedef struct unit_radius_probe {
	uintptr_t calls;
	uintptr_t candidate;
	uintptr_t data;
	uintptr_t context;
} unit_radius_probe;

static void record_unit_radius_candidate(nox_object_t* candidate, void* data) {
	unit_radius_probe* probe = data;
	++probe->calls;
	probe->candidate = (uintptr_t)candidate;
	probe->data = (uintptr_t)data;
}

static float2* call_unit_radius_scan_checked(float2* center, float radius, unit_radius_probe* probe) {
	unit_radius_scan_fn volatile fn = sub_518040;
	return fn(center, radius, probe ? record_unit_radius_candidate : NULL, probe);
}

static void call_unit_radius_candidate_checked(nox_object_t* candidate, float2* center, float radius, unit_radius_probe* probe) {
	unit_radius_candidate_fn volatile fn = sub_5180B0;
	if (!probe) {
		fn(candidate, NULL);
		return;
	}
	nox_unit_radius_context_t context = {center, radius, record_unit_radius_candidate, probe};
	probe->context = (uintptr_t)&context;
	fn(candidate, &context);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/types"
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

type unitRadiusProbe518040 struct {
	calls, candidate, data, context uintptr
}

func callUnitRadiusScanC518040(center *types.Pointf, radius float32, probe *unitRadiusProbe518040) *types.Pointf {
	p := C.call_unit_radius_scan_checked((*C.float2)(unsafe.Pointer(center)), C.float(radius), (*C.unit_radius_probe)(unsafe.Pointer(probe)))
	return (*types.Pointf)(unsafe.Pointer(p))
}

func callUnitRadiusCandidateC5180B0(candidate *server.Object, center *types.Pointf, radius float32, probe *unitRadiusProbe518040) {
	C.call_unit_radius_candidate_checked(asObjectC(candidate), (*C.float2)(unsafe.Pointer(center)), C.float(radius), (*C.unit_radius_probe)(unsafe.Pointer(probe)))
}
