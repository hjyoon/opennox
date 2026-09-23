package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterPickupObjectRangeSquared544B90 = float32(75 * 75)

type monsterActionPickupObjectHooks544B90 struct {
	canInteract    func(*Object, *Object, int) bool
	placeInventory func(*Object, *Object, int, int) bool
	useByNetCode   func(*Object, *Object) int32
	pop            func() int
}

// monsterActionPickupObject544B90 restores GAME.EXE 00544B90 without passing
// either the monster or its AI-stack target through the original PE32 int ABI.
// The action always completes, even when the target disappeared or cannot be
// reached. A successfully reached health potion is used after the pickup call,
// matching the original callback order and ignored pickup return value.
func monsterActionPickupObject544B90(unit *Object, hooks monsterActionPickupObjectHooks544B90) int {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) || hooks.pop == nil {
		return 0
	}
	head := unit.UpdateDataMonster().AIStackHead()
	if head == nil || head.Type() != ai.ACTION_PICKUP_OBJECT {
		return 0
	}
	target := head.ArgObj(0)
	if target != nil {
		delta := target.PosVec.Sub(unit.PosVec)
		if delta.X*delta.X+delta.Y*delta.Y < monsterPickupObjectRangeSquared544B90 &&
			hooks.canInteract != nil && hooks.canInteract(unit, target, 0) {
			if hooks.placeInventory != nil {
				hooks.placeInventory(unit, target, 1, 1)
			}
			// The original reloads the object argument from the same stack item
			// after inventory placement instead of retaining its earlier value.
			target = head.ArgObj(0)
			if target != nil && target.ObjSubClass.AsFood().Has(object.FoodHealthPotion) && hooks.useByNetCode != nil {
				hooks.useByNetCode(unit, target)
			}
		}
	}
	return hooks.pop()
}

// MonsterActionPickupObject544B90 binds the native-width pickup action to the
// live visibility, inventory, item-use, and action-stack services.
func (s *Server) MonsterActionPickupObject544B90(unit *Object, placeInventory func(*Object, *Object, int, int) bool) int {
	return monsterActionPickupObject544B90(unit, monsterActionPickupObjectHooks544B90{
		canInteract:    s.CanInteract,
		placeInventory: placeInventory,
		useByNetCode:   s.UseByNetCode53F8E0,
		pop:            unit.MonsterPopAction,
	})
}
