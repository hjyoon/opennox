package legacy

/*
#include <stdint.h>
#include "defs.h"
*/
import "C"

import (
	"encoding/binary"
	"strings"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

// game5MonsterFunction keeps function pointers in the native MonsterDef
// fields. The old public entries wrote them through fixed PE32 offsets.
func game5MonsterFunction(def *server.MonsterDef, name string, table map[string]unsafe.Pointer, field *unsafe.Pointer) C.int {
	if def == nil {
		return 0
	}
	if strings.EqualFold(name, "NULL") {
		*(*uintptr)(unsafe.Pointer(field)) = 0
		return 1
	}
	fn, ok := table[name]
	if !ok {
		return 0
	}
	*(*uintptr)(unsafe.Pointer(field)) = uintptr(fn)
	return 1
}

//export nox_xxx_monsterLoadStrikeFn_549040
func nox_xxx_monsterLoadStrikeFn_549040(raw unsafe.Pointer, name *C.char) C.int {
	if raw == nil || name == nil {
		return 0
	}
	def := (*server.MonsterDef)(raw)
	return game5MonsterFunction(def, GoString(name), monsterStrikeFunctions, &def.MeleeStrikeFunc236)
}

//export nox_xxx_monsterLoadDieFn_5490E0
func nox_xxx_monsterLoadDieFn_5490E0(raw unsafe.Pointer, name *C.char) C.int {
	if raw == nil || name == nil {
		return 0
	}
	def := (*server.MonsterDef)(raw)
	return game5MonsterFunction(def, GoString(name), monsterDieFunctions, &def.DieFunc228)
}

//export nox_xxx_monsterLoadDeadFn_549180
func nox_xxx_monsterLoadDeadFn_549180(raw unsafe.Pointer, name *C.char) C.int {
	if raw == nil || name == nil {
		return 0
	}
	def := (*server.MonsterDef)(raw)
	return game5MonsterFunction(def, GoString(name), monsterDeadFunctions, &def.DeadFunc232)
}

//export nox_xxx_monsterMainAIFn_547210
func nox_xxx_monsterMainAIFn_547210(raw *nox_object_t) {
	Nox_xxx_monsterMainAIFn_547210(asObjectS(raw))
}

//export nox_xxx_monsterAutoSpells_54C0C0
func nox_xxx_monsterAutoSpells_54C0C0(raw *nox_object_t) C.short {
	return C.short(GetServer().S().MonsterAutoSpells54C0C0(asObjectS(raw)))
}

//export nox_xxx_createWeapon_54C710
func nox_xxx_createWeapon_54C710(raw *nox_object_t) { weaponCreateCall54C710(asObjectS(raw)) }

//export sub_54C950
func sub_54C950(raw *nox_object_t) { armorCreateCall54C950(asObjectS(raw)) }

//export nox_xxx_createFnObelisk_54CA10
func nox_xxx_createFnObelisk_54CA10(raw *nox_object_t) C.int {
	server.ObeliskCreateNative54CA10(asObjectS(raw))
	return 0
}

//export nox_xxx_createFnAnim_54CA50
func nox_xxx_createFnAnim_54CA50(raw *nox_object_t) { server.AnimCreateNative54CA50(asObjectS(raw)) }

//export nox_xxx_createTrigger_54CA60
func nox_xxx_createTrigger_54CA60(raw *nox_object_t) *C.uint8_t {
	obj := asObjectS(raw)
	server.TriggerCreateNative54CA60(obj)
	return (*C.uint8_t)(obj.UpdateData)
}

//export nox_xxx_createMonsterGen_54CA90
func nox_xxx_createMonsterGen_54CA90(raw *nox_object_t) *C.uint32_t {
	obj := asObjectS(raw)
	server.MonsterGeneratorCreateNative54CA90(obj)
	return (*C.uint32_t)(obj.UpdateData)
}

//export nox_xxx_createRewardMarker_54CAC0
func nox_xxx_createRewardMarker_54CAC0(raw *nox_object_t) *C.uint32_t {
	obj := asObjectS(raw)
	server.RewardMarkerCreateNative54CAC0(obj)
	return (*C.uint32_t)(obj.InitData)
}

//export nox_xxx_diePlayer_54D2B0
func nox_xxx_diePlayer_54D2B0(raw *nox_object_t) C.int {
	// Death callbacks discard EAX. Zero defines its previously incidental word;
	// the native competitive/quest implementations own every gameplay effect.
	playerDieCall54D2B0(asObjectS(raw))
	return 0
}

//export nox_xxx_playerHandleElimDeath_54D7A0
func nox_xxx_playerHandleElimDeath_54D7A0(victim, killer *nox_object_t) {
	server.PlayerHandleElimDeath54D7A0(asObjectS(victim), asObjectS(killer), playerUpdateScoreRuntime54D980(GetServer()))
}

//export nox_xxx_playerUpdateScore_54D980
func nox_xxx_playerUpdateScore_54D980(victim, killer, assist *nox_object_t, tracking C.uint32_t) {
	server.PlayerUpdateScore54D980(asObjectS(victim), asObjectS(killer), asObjectS(assist), uint32(tracking), playerUpdateScoreRuntime54D980(GetServer()))
}

//export nox_xxx_playerHandleKotrDeath_54DC40
func nox_xxx_playerHandleKotrDeath_54DC40(victim, killer *nox_object_t) {
	server.PlayerHandleKotrDeath54DC40(asObjectS(victim), asObjectS(killer), playerHandleKotrDeathRuntime54DC40(GetServer()))
}

//export nox_xxx_netNotifyPlayerDied_54DF00
func nox_xxx_netNotifyPlayerDied_54DF00(raw *nox_object_t) {
	var packet [3]byte
	packet[0] = byte(netmsg.MSG_PLAYER_DIED)
	binary.LittleEndian.PutUint16(packet[1:], uint16(asObjectS(raw).NetCode))
	GetServer().S().NetSendPacketXxx1(255, packet[:], nil, 0)
}

//export nox_xxx_createSpark_54FD80
func nox_xxx_createSpark_54FD80(x, y C.float, kind, lifetime C.int, vx, vy, z C.float, owner *nox_object_t) *C.float {
	return (*C.float)(unsafe.Pointer(Nox_xxx_createSpark_54FD80(float32(x), float32(y), int(kind), int(lifetime), float32(vx), float32(vy), float32(z), asObjectS(owner))))
}

//export nox_xxx_updateMonsterGenerator_54E930
func nox_xxx_updateMonsterGenerator_54E930(raw *nox_object_t) C.char {
	monsterGeneratorUpdateCall54E930(asObjectS(raw))
	return 0
}

//export nox_xxx_mobGeneratorPick_54EBA0
func nox_xxx_mobGeneratorPick_54EBA0(generator *nox_object_t, destination *C.float2, template *nox_object_t) C.int {
	if GetServer().S().MonsterGeneratorPick54EBA0(asObjectS(generator), asObjectS(template), (*types.Pointf)(unsafe.Pointer(destination)), monsterGeneratorRuntime54E930()) {
		return 1
	}
	return 0
}

//export nox_xxx_mgenSetCreaturePos_54ED50
func nox_xxx_mgenSetCreaturePos_54ED50(generator *nox_object_t, destination *C.float2, target, template *nox_object_t) C.int {
	if GetServer().S().MonsterGeneratorSetCreaturePos54ED50(asObjectS(generator), asObjectS(template), asObjectS(target), (*types.Pointf)(unsafe.Pointer(destination)), monsterGeneratorRuntime54E930()) {
		return 1
	}
	return 0
}

//export sub_54EF90
func sub_54EF90(radius C.float, center *C.float2, template *nox_object_t, destination *C.float2) C.int {
	if GetServer().S().MonsterGeneratorRandomPosition54EF90(float32(radius), *(*types.Pointf)(unsafe.Pointer(center)), asObjectS(template), (*types.Pointf)(unsafe.Pointer(destination)), monsterGeneratorRuntime54E930()) {
		return 1
	}
	return 0
}

//export nox_xxx_mobGeneratorSpawn_54F070
func nox_xxx_mobGeneratorSpawn_54F070(generator, template *nox_object_t, destination *C.float2) *C.uint32_t {
	return (*C.uint32_t)(unsafe.Pointer(GetServer().S().MonsterGeneratorSpawn54F070(asObjectS(generator), asObjectS(template), *(*types.Pointf)(unsafe.Pointer(destination)), monsterGeneratorRuntime54E930())))
}

//export nox_xxx_unitCreatureCopyUC_54F2B0
func nox_xxx_unitCreatureCopyUC_54F2B0(template, spawned *nox_object_t) {
	GetServer().S().MonsterGeneratorCopyCreature54F2B0(asObjectS(template), asObjectS(spawned), monsterGeneratorRuntime54E930())
}
