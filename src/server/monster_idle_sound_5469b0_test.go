package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestMonsterIdleSound5469B0SchedulesAndPlays(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.Aggression = 0.5
	update.Field131 = math.Float32bits(300)
	update.Field132 = 100
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	soundSet, freeSoundSet := alloc.Make([]uint32{}, 5)
	defer freeSoundSet()
	soundSet[4] = 77
	update.SoundSet122 = unsafe.Pointer(&soundSet[0])
	var randomCalls, playCalls int
	monsterIdleSound5469B0(unit, monsterIdleSoundHooks5469B0{
		frame:    func() uint32 { return 100 },
		tickRate: func() uint32 { return 30 },
		random: func(min, max int) int {
			randomCalls++
			if min != 600 || max != 1800 {
				t.Fatalf("random bounds = %d..%d, want 600..1800", min, max)
			}
			return 900
		},
		play: func(id sound.ID, gotUnit *Object) {
			playCalls++
			if id != 77 || gotUnit != unit {
				t.Fatalf("play = %d/%p, want 77/%p", id, gotUnit, unit)
			}
		},
	})
	if randomCalls != 1 || playCalls != 1 || update.Field132 != 1000 {
		t.Fatalf("calls/deadline = %d/%d/%d, want 1/1/1000", randomCalls, playCalls, update.Field132)
	}
}

func TestMonsterIdleSound5469B0Gates(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Object, *MonsterUpdateData)
	}{
		{name: "not monster", setup: func(unit *Object, _ *MonsterUpdateData) { unit.ObjClass = object.ClassPlayer }},
		{name: "attacks at will", setup: func(_ *Object, update *MonsterUpdateData) { update.Aggression = 0.7 }},
		{name: "has enemy", setup: func(_ *Object, update *MonsterUpdateData) { update.CurrentEnemy = &Object{} }},
		{name: "far player", setup: func(_ *Object, update *MonsterUpdateData) { update.Field131 = math.Float32bits(301) }},
		{name: "active action", setup: func(_ *Object, update *MonsterUpdateData) { update.AIStack[0].Action = uint32(ai.ACTION_MOVE_TO) }},
		{name: "before deadline", setup: func(_ *Object, update *MonsterUpdateData) { update.Field132 = 101 }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			update := unit.UpdateDataMonster()
			update.Aggression = 0.5
			update.Field131 = math.Float32bits(100)
			update.Field132 = 100
			update.AIStackInd = 0
			update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
			tc.setup(unit, update)
			monsterIdleSound5469B0(unit, monsterIdleSoundHooks5469B0{
				frame:    func() uint32 { return 100 },
				tickRate: func() uint32 { return 30 },
				random: func(int, int) int {
					t.Fatal("random called past gate")
					return 0
				},
				play: func(sound.ID, *Object) { t.Fatal("play called past gate") },
			})
		})
	}
}
