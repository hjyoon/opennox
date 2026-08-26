package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestMonsterHurtSound532800SchedulesAndPlaysAtNativeWidth(t *testing.T) {
	unit := monsterActionTestObject50A910(t)
	update := unit.UpdateDataMonster()
	update.Field133 = 100
	soundSet, freeSoundSet := alloc.Make([]uint32{}, 3)
	defer freeSoundSet()
	soundSet[2] = 731
	update.SoundSet122 = unsafe.Pointer(&soundSet[0])
	var randomCalls, playCalls int
	monsterHurtSound532800(unit, monsterHurtSoundHooks532800{
		frame:    func() uint32 { return 100 },
		tickRate: func() uint32 { return 30 },
		random: func(minimum, maximum int) int {
			randomCalls++
			if minimum != 60 || maximum != 120 {
				t.Fatalf("random bounds = %d..%d, want 60..120", minimum, maximum)
			}
			return 79
		},
		play: func(id sound.ID, gotUnit *Object) {
			playCalls++
			if id != 731 || gotUnit != unit {
				t.Fatalf("play = %d/%p, want 731/%p", id, gotUnit, unit)
			}
		},
	})
	if randomCalls != 1 || playCalls != 1 || update.Field133 != 179 {
		t.Fatalf("calls/deadline = %d/%d/%d, want 1/1/179", randomCalls, playCalls, update.Field133)
	}
}

func TestMonsterHurtSound532800Gates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*Object, *MonsterUpdateData)
	}{
		{name: "not monster", setup: func(unit *Object, _ *MonsterUpdateData) { unit.ObjClass = object.ClassPlayer }},
		{name: "before deadline", setup: func(_ *Object, update *MonsterUpdateData) { update.Field133 = 101 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			update := unit.UpdateDataMonster()
			tc.setup(unit, update)
			monsterHurtSound532800(unit, monsterHurtSoundHooks532800{
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
