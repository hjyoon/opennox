package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterInterestingSoundNative5466F0 keeps the original unguarded raw
// entry contract. Its terrain services receive the cached MUD coordinate
// pointer; ordinary AI admission remains in the server wrapper.
func (unit *Object) MonsterInterestingSoundNative5466F0(
	reactive MonsterReactiveRuntime545DA0,
	tileAt func(*types.Pointf) int,
	pathReachable func(*Object, *types.Pointf) bool,
	trace func(types.Pointf, types.Pointf) bool,
) int {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if update.Field97 == 0 {
		return 0
	}
	if reactive.Frame()-update.Field101 >= 3*reactive.TickRate() {
		return 0
	}
	point := (*types.Pointf)(unsafe.Pointer(&update.Field99X))
	if unit.ObjSubClass&(1<<10) != 0 || tileAt(point) != 6 {
		if pathReachable(unit, point) {
			if trace(unit.PosVec, *point) {
				// The real child reloads the object's current MUD and clocks.
				unit.MonsterRecentSound545DA0(reactive)
			} else {
				reactive.Push(unit, ai.DEPENDENCY_NO_INTERESTING_SOUND)
				reactive.Push(unit, ai.DEPENDENCY_NO_VISIBLE_ENEMY)
				reactive.Push(unit, ai.DEPENDENCY_NOT_FRUSTRATED)
				if item := reactive.Push(unit, ai.DEPENDENCY_LOCATION_IS_SAFE); item != nil {
					item.Args[0] = uintptr(math.Float32bits(update.Field99X))
					item.Args[1] = uintptr(math.Float32bits(update.Field99Y))
				}
				if item := reactive.Push(unit, ai.ACTION_MOVE_TO); item != nil {
					item.Args[0] = uintptr(math.Float32bits(update.Field99X))
					item.Args[1] = uintptr(math.Float32bits(update.Field99Y))
					item.Args[2] = 0
				}
			}
		}
	}
	// Even when the child reloaded another MUD or rejected a changed marker,
	// the admitted outer entry still clears its own cached MUD and succeeds.
	update.Field97 = 0
	return 1
}
