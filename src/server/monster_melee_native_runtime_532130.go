package server

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"unsafe"
)

// MonsterMeleeNativeRuntime532130 supplies external engine services. Nil
// world, enemy and action lookups use the server's ordinary bindings.
// ObservePush observes the real push result and never decides admission.
type MonsterMeleeNativeRuntime532130 struct {
	AudioEvent   func(uint32, *Object)
	BuffOff      func(*Object, EnchantID)
	PlayerAttack func(*Object) int
	CanStrike    func(unsafe.Pointer) bool
	Strike       func(*Object, unsafe.Pointer) int
	EachInRect   func(types.Rectf, func(*Object) bool)
	IsEnemy      func(*Object, *Object) bool
	CancelAction func(ai.ActionType) func(*Object)
	EndAction    func(ai.ActionType) func(*Object)
	ObservePush  func(ai.ActionType, *MonsterUpdateData, *AIStackItem)
}

// MonsterMeleeBlockerState532390 uses the stable native pointer side slot
// and the separate packed binary32 distance. It never overlays a DWORD.
type MonsterMeleeBlockerState532390 struct {
	Selected *unsafe.Pointer
	Distance *float32
}
