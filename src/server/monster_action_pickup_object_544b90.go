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
// An absent entry target or failed range/visibility check completes the action.
// Callback-invalidated targets preserve the original fault prefix instead.
// Placement and use reload the entry-cached slot, ignoring callback results.
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
		// 00544BB4..00544BD9 retains each operation at x87 precision 53
		// under gameplay ToZero, with no binary32 spill or FMA contraction.
		dx := monsterMoveToRunAddChop53_544434(float64(target.PosVec.X), -float64(unit.PosVec.X))
		dy := monsterMoveToRunAddChop53_544434(float64(target.PosVec.Y), -float64(unit.PosVec.Y))
		ySquared := monsterMoveToRunSquareChop53_544434(dy)
		xSquared := monsterMoveToRunSquareChop53_544434(dx)
		distance := monsterMoveToRunAddChop53_544434(ySquared, xSquared)
		// The original tests C0 alone: unordered also enters visibility.
		if !(distance >= float64(monsterPickupObjectRangeSquared544B90)) && hooks.canInteract(unit, target, 0) {
			hooks.placeInventory(unit, head.ArgObj(0), 1, 1)
			// Both reloads use the cached item even if a callback changes the
			// unit's update record or stack index. The byte read has no nil gate.
			target = head.ArgObj(0)
			if byte(target.ObjSubClass)&0x10 != 0 {
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
