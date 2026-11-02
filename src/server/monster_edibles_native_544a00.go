package server

import "github.com/opennox/opennox/v1/common/unit/ai"

// MonsterSearchEdibleNative544A00 implements the raw C search contract. Unlike
// ordinary guarded searches, reentrant raw entries share their live scratch.
func (unit *Object) MonsterSearchEdibleNative544A00(radius float32, state *MonsterEdibleSearchState544A00, runtime MonsterEdibleNativeRuntime544A00) *Object {
	state.BestObject = nil
	state.BestDistance = monsterSearchEdibleMaxDistance544A00
	runtime.EachCircle(&unit.PosVec, radius, func(candidate, owner *Object) {
		candidate.MonsterSearchEdibleCandidateNative544A40(owner, state, runtime)
	}, unit)
	return state.BestObject
}

func (candidate *Object) MonsterSearchEdibleCandidateNative544A40(unit *Object, state *MonsterEdibleSearchState544A00, runtime MonsterEdibleNativeRuntime544A00) {
	monsterSearchEdibleCandidate544A40(candidate, unit, state, runtime.CanInteract, runtime.Online)
}

func (unit *Object) MonsterRetreatCheckEdiblesNative5455E0(state *MonsterEdibleSearchState544A00, runtime MonsterEdibleNativeRuntime544A00) *AIStackItem {
	return monsterRetreatCheckEdibles5455E0(unit, monsterActionRetreatHooks545440{
		quest: runtime.Quest(),
		searchFood: func(unit *Object, radius float32) *Object {
			return unit.MonsterSearchEdibleNative544A00(radius, state, runtime)
		},
		push: func(action ai.ActionType, _ ...any) *AIStackItem { return runtime.Push(unit, action) },
	})
}
