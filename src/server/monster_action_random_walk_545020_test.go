package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func randomWalkTestObject545020(t *testing.T) *Object {
	t.Helper()
	unit, freeUnit := alloc.New(Object{})
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	t.Cleanup(freeUnit)
	unit.ObjClass = object.ClassMonster
	unit.UpdateData = unsafe.Pointer(update)
	unit.SpeedCur = 2.5
	unit.Direction1 = 250
	unit.Direction2 = 1
	unit.PosVec = types.Ptf(100, 200)
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_RANDOM_WALK)
	update.MonsterDef = &MonsterDef{RunMultiplier96: 1.75}
	return unit
}

func TestMonsterActionRandomWalk545020WrapsAndAvoidsLava(t *testing.T) {
	unit := randomWalkTestObject545020(t)
	var probe types.Pointf
	moveAudioCalls := 0
	if !monsterActionRandomWalk545020(unit, monsterActionRandomWalkHooks545020{
		random: func(min, max int) int {
			if min != -20 || max != 20 {
				t.Fatalf("random bounds = %d..%d", min, max)
			}
			return 20
		},
		tileAt: func(got types.Pointf) int {
			probe = got
			return 6
		},
		moveAudio: func(got *Object) {
			if got != unit {
				t.Fatalf("move-audio unit = %p, want %p", got, unit)
			}
			moveAudioCalls++
		},
	}) {
		t.Fatal("random-walk action was not handled")
	}
	// 250 + 20 wraps to 14; lava rotates that direction by another 64.
	const wantDirection = Dir16(78)
	if unit.Direction1 != wantDirection || unit.Direction2 != wantDirection {
		t.Fatalf("directions = %d/%d, want %d", unit.Direction1, unit.Direction2, wantDirection)
	}
	probeCosine, probeSine := SinCosDir(14)
	wantProbe := types.Ptf(100+30*probeCosine, 200+30*probeSine)
	if probe != wantProbe {
		t.Fatalf("tile probe = %v, want %v", probe, wantProbe)
	}
	cosine, sine := SinCosDir(byte(wantDirection))
	if math.Float32bits(unit.ForceVec.X) != math.Float32bits(float32(2.5*float64(cosine))) ||
		math.Float32bits(unit.ForceVec.Y) != math.Float32bits(float32(2.5*float64(sine))) {
		t.Fatalf("force = %v", unit.ForceVec)
	}
	if moveAudioCalls != 1 {
		t.Fatalf("move-audio calls = %d, want 1", moveAudioCalls)
	}
}

func TestMonsterActionRandomWalk545020RunningAndFlying(t *testing.T) {
	unit := randomWalkTestObject545020(t)
	unit.ObjSubClass = object.SubClass(0x400)
	unit.UpdateDataMonster().StatusFlags = object.MonStatusRunning
	tileCalls := 0
	if !monsterActionRandomWalk545020(unit, monsterActionRandomWalkHooks545020{
		random: func(int, int) int { return -20 },
		tileAt: func(types.Pointf) int {
			tileCalls++
			return 6
		},
	}) {
		t.Fatal("running random-walk action was not handled")
	}
	if tileCalls != 0 {
		t.Fatalf("flying monster tile probes = %d, want 0", tileCalls)
	}
	const wantDirection = byte(230)
	cosine, sine := SinCosDir(wantDirection)
	speed := float64(float32(2.5)) * float64(float32(1.75))
	if unit.Direction1 != Dir16(wantDirection) || unit.Direction2 != Dir16(wantDirection) ||
		math.Float32bits(unit.ForceVec.X) != math.Float32bits(float32(speed*float64(cosine))) ||
		math.Float32bits(unit.ForceVec.Y) != math.Float32bits(float32(speed*float64(sine))) {
		t.Fatalf("running state = dir %d/%d force %v", unit.Direction1, unit.Direction2, unit.ForceVec)
	}
}

func TestMonsterActionRandomWalk545020RejectsWrongHead(t *testing.T) {
	unit := randomWalkTestObject545020(t)
	unit.UpdateDataMonster().AIStack[0].Action = uint32(ai.ACTION_GUARD)
	before := *unit
	if monsterActionRandomWalk545020(unit, monsterActionRandomWalkHooks545020{random: func(int, int) int { return 0 }}) {
		t.Fatal("non-random-walk head was handled")
	}
	if *unit != before {
		t.Fatal("rejected action changed the unit")
	}
}
