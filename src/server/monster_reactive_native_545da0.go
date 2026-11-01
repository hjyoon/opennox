package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterReactiveRuntime545DA0 supplies the original world services lazily.
// Raw entries retain the original unguarded object/update-data contract.
type MonsterReactiveRuntime545DA0 struct {
	Frame    func() uint32
	TickRate func() uint32
	Random   func(int, int) int
	Push     func(*Object, ai.ActionType) *AIStackItem
}

func (unit *Object) MonsterRecentSound545DA0(runtime MonsterReactiveRuntime545DA0) bool {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if update.Field97 == 0 {
		return false
	}
	if runtime.Frame()-update.Field101 >= 3*runtime.TickRate() {
		return false
	}
	runtime.Push(unit, ai.DEPENDENCY_NO_INTERESTING_SOUND)
	runtime.Push(unit, ai.DEPENDENCY_NO_VISIBLE_ENEMY)
	if item := runtime.Push(unit, ai.ACTION_WAIT); item != nil {
		fps := runtime.TickRate()
		delay := runtime.Random(int(int32(fps)), int(int32(fps*2)))
		item.Args[0] = uintptr(runtime.Frame() + uint32(delay))
	}
	if item := runtime.Push(unit, ai.ACTION_FACE_LOCATION); item != nil {
		// Keep the entry update pointer, but load coordinates after Push.
		item.Args[0] = uintptr(math.Float32bits(update.Field99X))
		item.Args[1] = uintptr(math.Float32bits(update.Field99Y))
	}
	update.Field97 = 0
	return true
}

func (unit *Object) MonsterLookAtDamager5466B0(runtime MonsterReactiveRuntime545DA0) bool {
	if !(*MonsterUpdateData)(unit.UpdateData).StatusFlags.Has(object.MonStatusInjured) {
		return false
	}
	if item := runtime.Push(unit, ai.ACTION_FACE_LOCATION); item != nil {
		item.Args[0] = uintptr(math.Float32bits(unit.Pos132.X))
		item.Args[1] = uintptr(math.Float32bits(unit.Pos132.Y))
	}
	return true
}
