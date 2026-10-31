package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterMissileNativeRuntime532540 supplies external world services while
// trajectory, prediction, action-stack admission and RNG use engine code.
// ObservePush only observes the real result; it cannot decide admission.
type MonsterMissileNativeRuntime532540 struct {
	AudioEvent        func(uint32, *Object)
	BuffOff           func(*Object, EnchantID)
	PlayerAttack      func(*Object) int
	NewObjectByTypeID func(string) *Object
	MapTraceRay       func(types.Pointf, types.Pointf, MapTraceFlags) bool
	CreateObjectAt    func(*Object, *Object, types.Pointf)
	DelayedDelete     func(*Object)
	CancelAction      func(ai.ActionType) func(*Object)
	EndAction         func(ai.ActionType) func(*Object)
	ObservePush       func(ai.ActionType, *MonsterUpdateData, *AIStackItem)
}
