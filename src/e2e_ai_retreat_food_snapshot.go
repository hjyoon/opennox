package opennox

import (
	"fmt"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// A read-only, post-update diagnostic. Coordinates, counters and action IDs
// are evidence, not a replacement for any original action/dependency gate.
type e2eAIRetreatFoodState struct {
	frame, fps, progressFrame, cornerFrame uint32
	position, target, progressPosition     types.Pointf
	currentHP, maximumHP                   uint16
	pathCount, pathCursor, pathStatus      uint32
	status                                 object.MonsterStatus
	head                                   ai.ActionType
	headTarget                             uintptr
	stack                                  []ai.ActionType
}

func e2eAIRetreatFoodSnapshot(unit, food *server.Object, frame, fps uint32) (e2eAIRetreatFoodState, bool) {
	if unit == nil || food == nil || unit.UpdateData == nil || unit.HealthData == nil ||
		!unit.Class().Has(object.ClassMonster) {
		return e2eAIRetreatFoodState{}, false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return e2eAIRetreatFoodState{}, false
	}
	items := update.GetAIStack()
	if len(items) == 0 {
		return e2eAIRetreatFoodState{}, false
	}
	head := &items[len(items)-1]
	state := e2eAIRetreatFoodState{
		frame: frame, fps: fps, progressFrame: update.Field124, cornerFrame: update.Field127,
		position: unit.PosVec, target: food.PosVec,
		progressPosition: types.Ptf(math.Float32frombits(update.Field125), math.Float32frombits(update.Field126)),
		currentHP:        unit.HealthData.Cur, maximumHP: unit.HealthData.Max,
		pathCount: update.Field2, pathCursor: update.Field67, pathStatus: update.Field71,
		status: update.StatusFlags, head: head.Type(),
		stack: make([]ai.ActionType, len(items)),
	}
	if state.head == ai.ACTION_MOVE_TO {
		state.headTarget = head.Args[2] // retain the full native word, no dereference
	}
	for i := range items {
		state.stack[i] = items[i].Type()
	}
	return state, true
}

func (state e2eAIRetreatFoodState) String() string {
	dx := float64(state.target.X) - float64(state.position.X)
	dy := float64(state.target.Y) - float64(state.position.Y)
	return fmt.Sprintf("executed-frame=%d fps=%d head=%s tracked=%#x HP=%d/%d pos=%v target=%v distance-squared=%.8f progress-frame=%d progress-pos=%v corner-frame=%d corner-age=%d path=%d/%d/%#x status=%#x stack=%v",
		state.frame, state.fps, state.head, state.headTarget, state.currentHP, state.maximumHP,
		state.position, state.target, dx*dx+dy*dy, state.progressFrame, state.progressPosition,
		state.cornerFrame, state.frame-state.cornerFrame, state.pathCount, state.pathCursor, state.pathStatus, uint32(state.status), state.stack)
}
