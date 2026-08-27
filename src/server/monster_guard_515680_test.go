package server

import (
	"math"
	"sync/atomic"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterGoPatrol515680RestoresGuardStack(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })

	unit := monsterActionTestObject50A910(t)
	unit.serverHandle = s.handle
	update := unit.UpdateDataMonster()
	update.AIStackInd = 0
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_WAIT), Args: [4]uintptr{123}, Field5: 1}
	update.SightRange = -1
	p1 := types.Ptf(1529.25, 5045.5)
	p2 := types.Ptf(1497.75, 5090.125)
	distance := math.Float32frombits(0x42f68000)

	if !s.MonsterGoPatrol515680(unit, p1, p2, distance) {
		t.Fatal("live monster guard request was rejected")
	}
	if update.AIStackInd != 0 {
		t.Fatalf("AIStackInd = %d, want 0", update.AIStackInd)
	}
	head := update.AIStackHead()
	if head.Type() != ai.ACTION_GUARD {
		t.Fatalf("head = %v, want GUARD", head.Type())
	}
	if got := head.ArgPos(0); got != p1 {
		t.Fatalf("guard point = %v, want %v", got, p1)
	}
	wantDirection := uint32(DirFromVec(types.Ptf(p2.X-p1.X, p2.Y-p1.Y)))
	if got := head.ArgU32(2); got != wantDirection {
		t.Fatalf("guard direction = %d, want %d", got, wantDirection)
	}
	if math.Float32bits(update.SightRange) != math.Float32bits(distance) {
		t.Fatalf("sight range bits = %#x, want %#x", math.Float32bits(update.SightRange), math.Float32bits(distance))
	}
}

func TestMonsterGoPatrol515680RejectsInvalidUnits(t *testing.T) {
	s := new(Server)
	tests := []struct {
		name string
		unit *Object
	}{
		{name: "nil"},
		{name: "non monster", unit: &Object{ObjClass: object.ClassPlayer}},
		{name: "missing update", unit: &Object{ObjClass: object.ClassMonster}},
		{name: "dead", unit: func() *Object {
			unit := monsterActionTestObject50A910(t)
			unit.ObjFlags = object.FlagDead
			return unit
		}()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if s.MonsterGoPatrol515680(tc.unit, types.Pointf{}, types.Pointf{}, 1) {
				t.Fatal("invalid guard request was accepted")
			}
		})
	}
}
