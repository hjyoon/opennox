package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterPatrolRequest515680 is the request layout shared with the C entry.
type MonsterPatrolRequest515680 struct {
	Origin      types.Pointf
	FacingPoint types.Pointf
	SightRange  float32
}

// MonsterGoPatrolFromRecord515680 applies the legacy C request. The
// direction delta is evaluated before the action stack callbacks. The
// action arguments and sight range are read from the record after them.
func MonsterGoPatrolFromRecord515680(unit *Object, request *MonsterPatrolRequest515680) {
	if unit == nil || request == nil || unit.UpdateData == nil ||
		!unit.ObjClass.Has(object.ClassMonster) || unit.ObjFlags.Has(object.FlagDead) {
		return
	}
	update := unit.UpdateDataMonster()
	add := pointOnLineAddChop53_57C8A0
	spill := pointOnLineSpill57C8A0
	dx := spill(add(float64(request.FacingPoint.X), -float64(request.Origin.X)))
	dy := spill(add(float64(request.FacingPoint.Y), -float64(request.Origin.Y)))

	unit.ClearActionStack()
	if action := unit.MonsterPushAction(ai.ACTION_GUARD); action != nil {
		origin := request.Origin
		direction := directionFromVector509ED0(dx, dy)
		action.SetArgs(origin, direction)
	}
	update.SightRange = request.SightRange
}
