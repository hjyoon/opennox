package server

import "unsafe"

type MonsterCollideScriptCaller4E83B0 func(
	block *ScriptCallback,
	caller *Object,
	trigger *Object,
	event ScriptEventType,
) unsafe.Pointer

// MonsterCollideScript4E83B0 binds the original three-argument script callback
// call to the native MonsterUpdateData layout. Event 22 is metadata required by
// the modern script bridge; GAME.EXE's 00502490 receives only the first three
// arguments.
func MonsterCollideScript4E83B0(
	monster *Object,
	other *Object,
	call MonsterCollideScriptCaller4E83B0,
) unsafe.Pointer {
	return monsterCollide4E83B0(monster, other, monsterCollideHooks4E83B0[
		*Object,
		*MonsterUpdateData,
		*ScriptCallback,
		unsafe.Pointer,
	]{
		updateData: func(obj *Object) *MonsterUpdateData {
			// Do not use UpdateDataMonster: 004E83B0 has no class check.
			return (*MonsterUpdateData)(obj.UpdateData)
		},
		collisionBlock: func(update *MonsterUpdateData) *ScriptCallback {
			return &update.ScriptCollision
		},
		scriptCallback: func(block *ScriptCallback, caller, trigger *Object) unsafe.Pointer {
			return call(block, caller, trigger, NoxEventMonsterCollide)
		},
	})
}
