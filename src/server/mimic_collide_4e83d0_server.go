package server

import (
	"math"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type mimicCollideNativeDeps4E83D0 struct {
	isEnemy         func(*Object, *Object) bool
	actionScheduled func(*Object, ai.ActionType) bool
	pushAction      func(*Object, ai.ActionType) *AIStackItem
	frame           func() uint32
	monsterCollide  func(*Object, *Object, unsafe.Pointer) unsafe.Pointer
}

func mimicCollideNative4E83D0(
	mimic, other *Object,
	collision unsafe.Pointer,
	deps mimicCollideNativeDeps4E83D0,
) unsafe.Pointer {
	return mimicCollide4E83D0(mimic, other, collision, mimicCollideHooks4E83D0[
		*Object,
		*AIStackItem,
		unsafe.Pointer,
		unsafe.Pointer,
	]{
		flags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		classLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		isEnemy: func(mimic, other *Object) int32 {
			if deps.isEnemy(mimic, other) {
				return 1
			}
			return 0
		},
		actionScheduled: func(obj *Object, action uint32) int32 {
			if deps.actionScheduled(obj, ai.ActionType(action)) {
				return 1
			}
			return 0
		},
		pushAction: func(obj *Object, action uint32) *AIStackItem {
			return deps.pushAction(obj, ai.ActionType(action))
		},
		frame: deps.frame,
		storeActionArg: func(action *AIStackItem, index int, value uint32) {
			action.Args[index] = uintptr(value)
		},
		posXBits: func(obj *Object) uint32 {
			return math.Float32bits(obj.PosVec.X)
		},
		posYBits: func(obj *Object) uint32 {
			return math.Float32bits(obj.PosVec.Y)
		},
		monsterCollide: deps.monsterCollide,
	})
}

// MimicCollide4E83D0 binds the original Mimic collision state transition to
// native Object and AIStackItem fields. collision is forwarded for ABI
// fidelity even though 004E83B0 itself does not inspect it.
func (s *Server) MimicCollide4E83D0(
	mimic, other *Object,
	collision unsafe.Pointer,
	call MonsterCollideScriptCaller4E83B0,
) unsafe.Pointer {
	return mimicCollideNative4E83D0(mimic, other, collision, mimicCollideNativeDeps4E83D0{
		isEnemy:         s.IsEnemyTo,
		actionScheduled: (*Object).MonsterActionIsScheduled,
		pushAction: func(obj *Object, action ai.ActionType) *AIStackItem {
			return obj.MonsterPushAction(action)
		},
		frame: s.Frame,
		monsterCollide: func(mimic, other *Object, _ unsafe.Pointer) unsafe.Pointer {
			return MonsterCollideScript4E83B0(mimic, other, call)
		},
	})
}
