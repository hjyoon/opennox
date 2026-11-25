package legacy

/*
#include "server__script__builtin.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export nox_server_gotoHome
func nox_server_gotoHome(unit *C.nox_object_t) {
	Nox_server_gotoHome(asObjectS((*nox_object_t)(unit)))
}

//export nox_script_SetQuestInt_514BE0
func nox_script_SetQuestInt_514BE0() C.int {
	key := nox_script_pop()
	value := nox_script_pop()
	QuestJournalSetInt500540(GoString(nox_script_getString_512E40(key)), value)
	return 0
}

//export nox_script_SetQuestFloat_514C10
func nox_script_SetQuestFloat_514C10() C.int {
	key := nox_script_pop()
	bits := nox_script_pop()
	questJournalSet500540(GoString(nox_script_getString_512E40(key)), 1, uint32(bits))
	return 0
}

func scriptBuiltinPopObject(vm NoxScript) *server.Object {
	return vm.ScriptToObject(int(vm.PopI32()))
}

//export nox_script_GiveExp_516190
func nox_script_GiveExp_516190() C.int {
	vm := GetServer().NoxScriptC()
	award := vm.PopF32()
	if unit := scriptBuiltinPopObject(vm); unit != nil {
		Nox_xxx_plyrGiveExp_4EF3A0_exp_level(unit, award)
	}
	return 0
}

func Nox_script_StartupScreen_516600_A() {
	runtime := GetServer()
	s := runtime.S()
	var host *server.Object
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.Players.NextUnit(unit) {
		update := unit.UpdateDataPlayer()
		if update != nil && update.Player != nil && update.Player.PlayerInd == server.HostPlayerIndex {
			host = unit
			break
		}
	}
	server.JournalEntriesRemoveByMask4277B0(host, 0xE)
	for item := host.FirstItem(); item != nil; {
		next := item.NextItem()
		if item.Class().Has(object.ClassKey) {
			runtime.DelayedDelete(item)
		}
		item = next
	}
	*memmap.PtrUint32(0x5D4594, 2386832) = 1
}

//export nox_script_StartupScreen_516600_A
func nox_script_StartupScreen_516600_A() {
	Nox_script_StartupScreen_516600_A()
}

func scriptBuiltinHostUnit(s *server.Server) *server.Object {
	if host := s.Players.ByInd(server.HostPlayerIndex); host != nil {
		return host.PlayerUnit
	}
	return nil
}

//export nox_script_MakeFriendly_516720
func nox_script_MakeFriendly_516720() C.int {
	runtime := GetServer()
	s := runtime.S()
	host := scriptBuiltinHostUnit(s)
	s.ScriptMakeFriendly516720(host, scriptBuiltinPopObject(runtime.NoxScriptC()))
	return 0
}

//export nox_script_MakeEnemy_516760
func nox_script_MakeEnemy_516760() C.int {
	runtime := GetServer()
	runtime.S().ScriptMakeEnemy516760(scriptBuiltinPopObject(runtime.NoxScriptC()))
	return 0
}

//export nox_script_builtin_516790
func nox_script_builtin_516790(defaultValue unsafe.Pointer) C.int {
	vm := GetServer().NoxScriptC()
	if unit := scriptBuiltinPopObject(vm); unit != nil {
		vm.PushBool(server.ScriptUnknownB8516790(unit))
	} else {
		vm.PushI32(int32(uintptr(defaultValue)))
	}
	return 0
}

//export nox_script_BecomePet_5167D0
func nox_script_BecomePet_5167D0() C.int {
	runtime := GetServer()
	host := scriptBuiltinHostUnit(runtime.S())
	if unit := scriptBuiltinPopObject(runtime.NoxScriptC()); host != nil && unit != nil {
		unitBecomePetRuntime4E7B00(host, unit)
	}
	return 0
}

//export nox_script_BecomeEnemy_516810
func nox_script_BecomeEnemy_516810() C.int {
	runtime := GetServer()
	host := scriptBuiltinHostUnit(runtime.S())
	if unit := scriptBuiltinPopObject(runtime.NoxScriptC()); host != nil && unit != nil {
		unitBecomeEnemyRuntime4E7B60(host, unit)
	}
	return 0
}

//export nox_script_builtin_516850
func nox_script_builtin_516850(defaultValue unsafe.Pointer) C.int {
	vm := GetServer().NoxScriptC()
	if unit := scriptBuiltinPopObject(vm); unit != nil {
		vm.PushBool(server.ScriptUnknownB9516850(unit))
	} else {
		vm.PushI32(int32(uintptr(defaultValue)))
	}
	return 0
}

//export nox_script_OblivionGive_516890
func nox_script_OblivionGive_516890() C.int {
	runtime := GetServer()
	host := scriptBuiltinHostUnit(runtime.S())
	upgrade := int(runtime.NoxScriptC().PopI32())
	server.ScriptSetHalberd516890(host, upgrade, server.ScriptSetHalberdRuntime516890{
		DelayedDelete: runtime.DelayedDelete,
		Respawn: func(owner *server.Object, typeID string) *server.Object {
			return Nox_xxx_playerRespawnItem_4EF750(owner, typeID, nil, 1, 1)
		},
		TryEquip: func(owner, item *server.Object) {
			Nox_xxx_playerTryEquip_4F2F70(owner, item)
		},
	})
	return 0
}
