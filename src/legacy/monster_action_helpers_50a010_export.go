package legacy

/*
// The shared header still contains unrelated Win32-only layout assertions.
#define _Static_assert(...)
#include "GAME4_1.h"
#undef _Static_assert
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func monsterActionIsConditionExportCall50A010(action int32) int32 {
	return int32(C.nox_xxx_monsterActionIsCondition_50A010(C.int(action)))
}

func monsterActionGetExportCall50A020(unit *server.Object) ai.ActionType {
	return ai.ActionType(uint32(C.nox_xxx_mobActionGet_50A020(asObjectC(unit))))
}

func monsterActionPreviousExportCall50A040(unit *server.Object) ai.ActionType {
	return ai.ActionType(uint32(C.sub_50A040(asObjectC(unit))))
}

func monsterActionPushIfChangedExportCall50A360(unit *server.Object, action ai.ActionType) unsafe.Pointer {
	return C.nox_xxx_monsterAction_50A360(asObjectC(unit), C.int(int32(action)))
}

//export nox_xxx_monsterActionIsCondition_50A010
func nox_xxx_monsterActionIsCondition_50A010(action C.int) C.int {
	if ai.ActionType(uint32(action)).IsCondition() {
		return 1
	}
	return 0
}

//export nox_xxx_mobActionGet_50A020
func nox_xxx_mobActionGet_50A020(unit *nox_object_t) C.int {
	return C.int(int32(asObjectS(unit).MonsterActionGet50A020()))
}

//export sub_50A040
func sub_50A040(unit *nox_object_t) C.int {
	return C.int(int32(asObjectS(unit).MonsterActionPrevious50A040()))
}

//export nox_xxx_monsterAction_50A360
func nox_xxx_monsterAction_50A360(unit *nox_object_t, action C.int) unsafe.Pointer {
	return asObjectS(unit).MonsterActionPushIfChanged50A360(ai.ActionType(uint32(action))).C()
}
