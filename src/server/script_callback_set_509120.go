package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
	ns4 "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const scriptCallbackNameStride509120 = uintptr(128)

type scriptCallbackTarget509120 struct {
	nameOffset   uintptr
	bindFunction func() func(int32)
}

func scriptCallbackTargetForEvent509120(obj *Object, event int) (scriptCallbackTarget509120, bool) {
	if event == 14 {
		return scriptCallbackTarget509120{
			bindFunction: func() func(int32) {
				return func(function int32) { obj.ScriptPickup.Func = function }
			},
		}, true
	}

	switch class := obj.Class(); {
	case class.Has(object.ClassTrigger):
		switch event {
		case 0:
			return scriptCallbackTarget509120{
				nameOffset: 4 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataTrigger()
					return func(function int32) { data.ScriptCollide.Func = function }
				},
			}, true
		case 1:
			return scriptCallbackTarget509120{
				nameOffset: 2 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataTrigger()
					return func(function int32) { data.ScriptActivate.Func = function }
				},
			}, true
		case 2:
			return scriptCallbackTarget509120{
				nameOffset: 3 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataTrigger()
					return func(function int32) { data.ScriptDeactivate.Func = function }
				},
			}, true
		}
	case class.Has(object.ClassMonster):
		switch ns4.ObjectEvent(event) {
		case ns4.EventEnemySighted:
			return scriptCallbackTarget509120{
				nameOffset: 5 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptEnemySighted.Func = function }
				},
			}, true
		case ns4.EventLookingForEnemy:
			return scriptCallbackTarget509120{
				nameOffset: 6 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptLookingForEnemy.Func = function }
				},
			}, true
		case ns4.EventDeath:
			return scriptCallbackTarget509120{
				nameOffset: 7 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptDeath.Func = function }
				},
			}, true
		case ns4.EventChangeFocus:
			return scriptCallbackTarget509120{
				nameOffset: 8 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptChangeFocus.Func = function }
				},
			}, true
		case ns4.EventIsHit:
			return scriptCallbackTarget509120{
				nameOffset: 9 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptIsHit.Func = function }
				},
			}, true
		case ns4.EventRetreat:
			return scriptCallbackTarget509120{
				nameOffset: 10 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptRetreat.Func = function }
				},
			}, true
		case ns4.EventCollision:
			return scriptCallbackTarget509120{
				nameOffset: 11 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptCollision.Func = function }
				},
			}, true
		case ns4.EventEnemyHeard:
			return scriptCallbackTarget509120{
				nameOffset: 12 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptHearEnemy.Func = function }
				},
			}, true
		case ns4.EventEndOfWaypoint:
			return scriptCallbackTarget509120{
				nameOffset: 13 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptEndOfWaypoint.Func = function }
				},
			}, true
		case ns4.EventLostEnemy:
			return scriptCallbackTarget509120{
				nameOffset: 14 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonster()
					return func(function int32) { data.ScriptLostEnemy.Func = function }
				},
			}, true
		}
	case class.Has(object.ClassHole):
		if event == 12 {
			return scriptCallbackTarget509120{
				nameOffset: scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := (*HoleCollideData)(obj.CollideData)
					return func(function int32) { data.Script.Func = function }
				},
			}, true
		}
	case class.Has(object.ClassMonsterGenerator):
		switch event {
		case 15:
			return scriptCallbackTarget509120{
				nameOffset: 15 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonsterGen()
					return func(function int32) { data.FuncInd52 = uint32(function) }
				},
			}, true
		case 16:
			return scriptCallbackTarget509120{
				nameOffset: 16 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonsterGen()
					return func(function int32) { data.FuncInd60 = uint32(function) }
				},
			}, true
		case 17:
			return scriptCallbackTarget509120{
				nameOffset: 18 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonsterGen()
					return func(function int32) { data.FuncInd68 = uint32(function) }
				},
			}, true
		case 18:
			return scriptCallbackTarget509120{
				nameOffset: 17 * scriptCallbackNameStride509120,
				bindFunction: func() func(int32) {
					data := obj.UpdateDataMonsterGen()
					return func(function int32) { data.ScriptCollision.Func = function }
				},
			}, true
		}
	}
	return scriptCallbackTarget509120{}, false
}

func scriptCallbackStoreName509120(base unsafe.Pointer, offset uintptr, name string) {
	destination := unsafe.Slice((*byte)(unsafe.Add(base, offset)), len(name)+1)
	copy(destination, name)
	destination[len(name)] = 0
}

func scriptCallbackSet509120(
	obj *Object,
	event int,
	name string,
	storeName bool,
	indexByName func(string) int32,
) bool {
	if obj.Field189 == nil {
		return false
	}
	target, ok := scriptCallbackTargetForEvent509120(obj, event)
	if !ok {
		return false
	}
	if storeName {
		scriptCallbackStoreName509120(obj.Field189, target.nameOffset, name)
		return true
	}
	setFunction := target.bindFunction()
	setFunction(indexByName(name))
	return true
}

// Nox_script_objCallbackSet_509120 restores GAME.EXE 00509120 using native
// Object, update-data, collide-data, and callback pointers. The name table is
// still the original sequence of 128-byte slots selected by GameFlag22/23.
func (s *NoxScriptVM) Nox_script_objCallbackSet_509120(obj *Object, event int, name string) {
	scriptCallbackSet509120(
		obj,
		event,
		name,
		noxflags.HasGame(noxflags.GameFlag22|noxflags.GameFlag23),
		func(name string) int32 { return int32(s.ScriptIndexByName(name)) },
	)
}
