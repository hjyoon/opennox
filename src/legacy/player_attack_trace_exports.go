package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

var publicAttackTrace538330 server.PlayerAttackTraceState538330

func playerAttackTraceRuntime538330() server.PlayerAttackTraceRuntime538330 {
	srv := GetServer()
	s := srv.S()
	return server.PlayerAttackTraceRuntime538330{
		Scan: func(center *types.Pointf, radius float32, visit func(*server.Object)) {
			server.PlayerAttackScanCircle518040(center, radius, s.Map.EachObjInRect, circleBoxCollisionNative54A990, visit)
		},
		Selection: playerAttackTargetRuntime5386A0(s),
		Direction: func(from types.Pointf, direction int32, to types.Pointf) uint8 {
			return uint8(twoPointsAndDirection4E6E50(from, direction, to))
		},
		Trace: func(from, to types.Pointf) bool {
			return s.MapTraceRay(from, to, server.MapTraceFlag1|server.MapTraceFlag3)
		},
		PreHit: func(target, owner, weapon *server.Object, attack *server.PlayerAttackData538330) {
			if weapon != nil && weapon.InitData != nil {
				s.PlayerPreAttackEffects538290(target, owner, weapon, unsafe.Pointer(attack))
			}
		},
		Damage: func(target, owner, weapon *server.Object, damage int32, damageType uint8) {
			if target.Damage != nil {
				target.CallDamage(owner, weapon, int(damage), object.DamageType(damageType))
			}
		},
		ReportPlayerFX: func() bool { return noxflags.HasGame(noxflags.GameModeCoop) },
		PointFX: func(pos *types.Pointf) {
			nox_xxx_netSendPointFx_522FF0(C.char(-117), (*C.float2)(unsafe.Pointer(pos)))
		},
		Force: func(origin *types.Pointf, target *server.Object, force float32) {
			applyForceExport52DF80(target, unsafe.Pointer(origin), force)
		},
		Walls: func(bounds [4]int32, center *types.Pointf, radius float32, damage int32, damageType uint8, source *server.Object) {
			rect := image.Rectangle{
				Min: image.Pt(int(bounds[0]), int(bounds[1])),
				Max: image.Pt(int(bounds[2]), int(bounds[3])),
			}
			srv.Nox_xxx_mapDamageToWalls_534FC0(rect, *center, radius, int(damage), object.DamageType(damageType), source)
		},
		ItemDamage: func() float64 { return s.Balance.Float("ItemDamagePercentage") },
		WeaponDamage: func(weapon, owner, source, effective *server.Object, damage float32, damageType uint8) {
			playerDamageWeaponNative4E1560(weapon, owner, source, effective, damage, object.DamageType(damageType))
		},
	}
}

//export nox_xxx_playerTraceAttack_538330
func nox_xxx_playerTraceAttack_538330(attacker *nox_object_t, info *C.nox_player_attack_info_native_t) C.int {
	if attacker == nil || info == nil {
		return 0
	}
	attack := (*server.PlayerAttackData538330)(unsafe.Pointer(info))
	return C.int(publicAttackTrace538330.Trace(asObjectS(attacker), attack, playerAttackTraceRuntime538330()))
}

//export nox_xxx_playerTraceAttackNative_538330
func nox_xxx_playerTraceAttackNative_538330(attacker *nox_object_t, info *C.nox_player_attack_info_native_t) C.int {
	if attacker == nil || info == nil {
		return 0
	}
	attack := (*server.PlayerAttackData538330)(unsafe.Pointer(info))
	var state server.PlayerAttackTraceState538330
	return C.int(state.Trace(asObjectS(attacker), attack, playerAttackTraceRuntime538330()))
}

//export sub_538510
func sub_538510(candidate *nox_object_t, info *C.nox_player_attack_info_native_t) {
	if candidate == nil || info == nil {
		return
	}
	attack := (*server.PlayerAttackData538330)(unsafe.Pointer(info))
	publicAttackTrace538330.Hit(asObjectS(candidate), attack, playerAttackTraceRuntime538330())
}

//export sub_5386A0
func sub_5386A0(candidate, attacker *nox_object_t) {
	if candidate == nil || attacker == nil || candidate == attacker {
		return
	}
	publicAttackTrace538330.Select(asObjectS(candidate), asObjectS(attacker), playerAttackTargetRuntime5386A0(GetServer().S()))
}
