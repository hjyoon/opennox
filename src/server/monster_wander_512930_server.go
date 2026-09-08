package server

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func monsterWanderStoreArgU32512930(action *AIStackItem, index int, value uint32) {
	action.Args[index] = action.Args[index]&^uintptr(0xffffffff) | uintptr(value)
}

func monsterWanderStoreArgLow512930(action *AIStackItem, index int, value uint8) {
	action.Args[index] = action.Args[index]&^uintptr(0xff) | uintptr(value)
}

func monsterWanderNative512930(unit *Object) {
	monsterWander512930(unit, monsterWanderHooks512930[*Object, *MonsterUpdateData, *AIStackItem]{
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		loadUpdate: func(unit *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(unit.UpdateData)
		},
		clearActionStack: func(unit *Object) {
			unit.ClearActionStack()
		},
		pushAction: func(unit *Object, action uint32) *AIStackItem {
			return unit.MonsterPushAction(ai.ActionType(action))
		},
		storeArgU32: monsterWanderStoreArgU32512930,
		loadRoamFlagsLow: func(update *MonsterUpdateData) uint8 {
			return uint8(update.Field333)
		},
		storeArgLow: monsterWanderStoreArgLow512930,
	})
}

// ScriptMonsterRoam512930 binds GAME.EXE 00512930 to native-width Object,
// MonsterUpdateData, and AIStackItem pointers. It intentionally adds no null
// guards because the original faults on a null unit or missing UpdateData.
func (*Server) ScriptMonsterRoam512930(unit *Object) {
	monsterWanderNative512930(unit)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MonsterUpdateData{}.Field333)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(AIStackItem{}.Action)]
	_ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-unsafe.Sizeof(AIStackItem{}.Args[0])]
)
