package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// BombCollideData is the registered eight-byte BombCollide record. GAME.EXE
// 004E96F0 does not inspect or modify any byte of it.
type BombCollideData struct {
	Reserved [8]uint8
}

type BombCollideScriptCaller4E96F0 func(
	block *ScriptCallback,
	caller *Object,
	trigger *Object,
	event ScriptEventType,
) unsafe.Pointer

type BombCollideRuntime4E96F0 struct {
	ScriptCallback BombCollideScriptCaller4E96F0
	DamageClear    func(*Object, int32)
}

type bombCollideNativeDeps4E96F0 struct {
	gameModeCoop    func() int32
	firstPlayerUnit func() *Object
	scriptCallback  BombCollideScriptCaller4E96F0
	damageClear     func(*Object, int32)
}

func bombCollideNative4E96F0(
	bomb, other *Object,
	collision unsafe.Pointer,
	deps bombCollideNativeDeps4E96F0,
) {
	bombCollide4E96F0(bomb, other, collision, bombCollideHooks4E96F0[
		*Object,
		*MonsterUpdateData,
		*ScriptCallback,
	]{
		loadUpdateData: func(obj *Object) *MonsterUpdateData {
			// Do not use UpdateDataMonster: 004E96F0 has no class check.
			return (*MonsterUpdateData)(obj.UpdateData)
		},
		gameModeCoop:    deps.gameModeCoop,
		firstPlayerUnit: deps.firstPlayerUnit,
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		collisionBlock: func(update *MonsterUpdateData) *ScriptCallback {
			return &update.ScriptCollision
		},
		scriptCallback: func(block *ScriptCallback, caller, trigger *Object) {
			_ = deps.scriptCallback(block, caller, trigger, NoxEventBombCollide)
		},
		classLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		unitsOnSameTeam: func(first, second *Object) int32 {
			if UnitsHaveSameTeam4EC520(first, second) {
				return 1
			}
			return 0
		},
		storeCollideUnit: func(update *MonsterUpdateData, other *Object) {
			update.BombCollideTarget = other
		},
		damageClear: deps.damageClear,
	})
}

// BombCollide4E96F0 binds the callback to native-width Object and
// MonsterUpdateData layouts. Script and damage callbacks retain their existing
// runtime boundaries while all object references crossing them remain typed.
func (s *Server) BombCollide4E96F0(
	bomb, other *Object,
	collision unsafe.Pointer,
	runtime BombCollideRuntime4E96F0,
) {
	bombCollideNative4E96F0(bomb, other, collision, bombCollideNativeDeps4E96F0{
		gameModeCoop: func() int32 {
			if noxflags.HasGame(noxflags.GameModeCoop) {
				return 1
			}
			return 0
		},
		firstPlayerUnit: s.Players.FirstUnit,
		scriptCallback:  runtime.ScriptCallback,
		damageClear:     runtime.DamageClear,
	})
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(BombCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(BombCollideData{}.Reserved)]
)
