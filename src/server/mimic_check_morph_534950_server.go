package server

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

type mimicCheckMorphNativeDeps534950 struct {
	frame      func() uint32
	tickRate   func() uint32
	pushAction func(*Object, ai.ActionType) *AIStackItem
	audio      func(sound.ID, *Object, int, uint32)
}

func mimicCheckMorphNative534950(obj *Object, d mimicCheckMorphNativeDeps534950) {
	mimicCheckMorph534950(obj, mimicCheckMorphHooks534950[*Object, *MonsterUpdateData, *AIStackItem]{
		update: func(obj *Object) *MonsterUpdateData {
			return (*MonsterUpdateData)(obj.UpdateData)
		},
		stackIndex: func(update *MonsterUpdateData) int8 {
			return update.AIStackInd
		},
		head: func(update *MonsterUpdateData, index int8) *AIStackItem {
			return &update.AIStack[index]
		},
		action:   func(head *AIStackItem) uint32 { return head.Action },
		targetX:  func(head *AIStackItem) float32 { return head.ArgF32(0) },
		posX:     func(obj *Object) float32 { return obj.PosVec.X },
		targetY:  func(head *AIStackItem) float32 { return head.ArgF32(1) },
		posY:     func(obj *Object) float32 { return obj.PosVec.Y },
		status:   func(update *MonsterUpdateData) uint32 { return uint32(update.StatusFlags) },
		frame:    d.frame,
		timer:    func(update *MonsterUpdateData) uint32 { return update.Field137 },
		tickRate: d.tickRate,
		pushAction: func(obj *Object, action uint32) *AIStackItem {
			return d.pushAction(obj, ai.ActionType(action))
		},
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			d.audio(sound.ID(id), obj, int(kind), code)
		},
	})
}

// MimicCheckMorph534950 binds the original morph check to native-width
// Object, MonsterUpdateData and AIStackItem fields. The existing typed push
// preserves action-reset and StackChanged effects; audio remains 00501960's
// EventObj dependency. The caller, as in GAME.EXE, selects Mimic objects.
func (s *Server) MimicCheckMorph534950(obj *Object) {
	mimicCheckMorphNative534950(obj, mimicCheckMorphNativeDeps534950{
		frame:    s.Frame,
		tickRate: s.TickRate,
		pushAction: func(obj *Object, action ai.ActionType) *AIStackItem {
			return obj.MonsterPushAction(action)
		},
		audio: func(id sound.ID, obj *Object, kind int, code uint32) {
			s.Audio.EventObj(id, obj, kind, code)
		},
	})
}
