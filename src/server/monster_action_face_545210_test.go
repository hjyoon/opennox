package server

import (
	"testing"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func faceActionTestObject545210(t *testing.T, action ai.ActionType) *Object {
	t.Helper()
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(action)
	unit.PosVec = types.Ptf(10, 20)
	return unit
}

func TestMonsterActionFaceLocation545210TurnsByOriginalStep(t *testing.T) {
	unit := faceActionTestObject545210(t, ai.ACTION_FACE_LOCATION)
	unit.Direction1 = 0
	unit.Direction2 = 0
	unit.UpdateDataMonster().AIStack[0].SetArgs(types.Ptf(10, 120))
	pops := 0

	monsterActionFaceLocation545210(unit, func() int { pops++; return 7 })
	if unit.Direction2 != 8 {
		t.Fatalf("Direction2 = %d, want 8", unit.Direction2)
	}
	if pops != 0 {
		t.Fatalf("pop calls = %d, want 0", pops)
	}

	unit.UpdateDataMonster().AIStack[0].SetArgs(types.Ptf(10, -80))
	monsterActionFaceLocation545210(unit, func() int { pops++; return 7 })
	if unit.Direction2 != 0 {
		t.Fatalf("Direction2 after clockwise turn = %d, want 0", unit.Direction2)
	}
	if pops != 0 {
		t.Fatalf("pop calls = %d, want 0", pops)
	}
}

func TestMonsterActionFacePoint545240TurnsBeforeCompletion(t *testing.T) {
	unit := faceActionTestObject545210(t, ai.ACTION_FACE_LOCATION)
	unit.Direction1 = 0
	unit.Direction2 = 0
	pops := 0
	got := monsterActionFacePoint545240(unit, types.Ptf(110, 20), func() int {
		pops++
		return 23
	})
	if got != 23 || pops != 1 {
		t.Fatalf("result/pop calls = %d/%d, want 23/1", got, pops)
	}
	if unit.Direction2 != 8 {
		t.Fatalf("Direction2 = %d, want original pre-pop turn to 8", unit.Direction2)
	}
}

func TestMonsterActionFaceObject545300PreservesNativePointer(t *testing.T) {
	unit := faceActionTestObject545210(t, ai.ACTION_FACE_OBJECT)
	unit.Direction1 = 0
	unit.Direction2 = 0
	target := monsterActionTestObject50A910(t)
	target.PosVec = types.Ptf(10, 120)
	unit.UpdateDataMonster().AIStack[0].SetArgs(target)
	pops := 0

	monsterActionFaceObject545300(unit, func() int { pops++; return 0 })
	if got := unit.UpdateDataMonster().AIStack[0].ArgObj(0); got != target {
		t.Fatalf("target pointer = %p, want %p", got, target)
	}
	if unit.Direction2 != 8 || pops != 0 {
		t.Fatalf("Direction2/pop calls = %d/%d, want 8/0", unit.Direction2, pops)
	}
}

func TestMonsterActionFaceObject545300NilTargetPops(t *testing.T) {
	unit := faceActionTestObject545210(t, ai.ACTION_FACE_OBJECT)
	pops := 0
	got := monsterActionFaceObject545300(unit, func() int { pops++; return 31 })
	if got != 31 || pops != 1 {
		t.Fatalf("result/pop calls = %d/%d, want 31/1", got, pops)
	}
}

func TestMonsterActionFaceAngle545340(t *testing.T) {
	unit := faceActionTestObject545210(t, ai.ACTION_FACE_ANGLE)
	unit.Direction1 = 0
	unit.Direction2 = 252
	unit.UpdateDataMonster().AIStack[0].SetArgs(uint32(64))
	pops := 0
	monsterActionFaceAngle545340(unit, func() int { pops++; return 0 })
	if unit.Direction2 != 4 || pops != 0 {
		t.Fatalf("Direction2/pop calls = %d/%d, want 4/0", unit.Direction2, pops)
	}
}

func TestMonsterActionSetAngle5453E0WrapsAndPops(t *testing.T) {
	unit := faceActionTestObject545210(t, ai.ACTION_SET_ANGLE)
	unit.Direction1 = 3
	unit.Direction2 = 5
	unit.UpdateDataMonster().AIStack[0].SetArgs(int32(-1))
	pops := 0
	got := monsterActionSetAngle5453E0(unit, func() int { pops++; return 11 })
	if unit.Direction1 != 255 || unit.Direction2 != 255 {
		t.Fatalf("directions = %d/%d, want 255/255", unit.Direction1, unit.Direction2)
	}
	if got != 11 || pops != 1 {
		t.Fatalf("result/pop calls = %d/%d, want 11/1", got, pops)
	}
}
