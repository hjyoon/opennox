package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

type MonsterGeneratorCollideScriptCaller4EBE10 func(
	block *ScriptCallback,
	caller *Object,
	trigger *Object,
	event ScriptEventType,
) unsafe.Pointer

// MonsterGeneratorCollide4EBE10 binds GAME.EXE 004EBE10 to native-width
// Object and MonsterGenUpdateData layouts. The collision point is part of the
// registered callback ABI but is not read by the original function.
func (*Server) MonsterGeneratorCollide4EBE10(
	source, target *Object,
	collision *types.Pointf,
	call MonsterGeneratorCollideScriptCaller4EBE10,
) {
	_ = collision
	monsterGeneratorCollide4EBE10(source, target, monsterGeneratorCollideHooks4EBE10[
		*Object,
		*MonsterGenUpdateData,
		*ScriptCallback,
		unsafe.Pointer,
	]{
		loadTargetClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadSourceUpdate: func(obj *Object) *MonsterGenUpdateData {
			// Do not call UpdateDataMonsterGen: GAME.EXE has no source class gate.
			return (*MonsterGenUpdateData)(obj.UpdateData)
		},
		collisionBlock: func(update *MonsterGenUpdateData) *ScriptCallback {
			return &update.ScriptCollision
		},
		scriptCallback: func(block *ScriptCallback, caller, trigger *Object) unsafe.Pointer {
			return call(block, caller, trigger, NoxEventGeneratorCollide)
		},
	})
}
