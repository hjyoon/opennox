package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestMonsterActionFlee544760RejectsNaNSpeed(t *testing.T) {
	unit := fleeMonsterTestObject544760(t)
	unit.SpeedBase = float32(math.NaN())
	pops, moves := 0, 0
	monsterActionFlee544760(unit, monsterActionFleeHooks544760{
		frame: func() uint32 { return 1 }, tickRate: func() uint32 { return 30 },
		pop:          func() int { pops++; return 0 },
		actuallyMove: func(*Object) bool { moves++; return false },
	})
	if pops != 1 || moves != 0 {
		t.Fatalf("NaN speed pop/move = %d/%d, want 1/0", pops, moves)
	}
}

func TestMonsterActionFlee544760PreservesOtherArguments(t *testing.T) {
	for _, count := range []int{-1, 0, 1, 2} {
		t.Run(string(rune('b'+count)), func(t *testing.T) {
			unit := fleeMonsterTestObject544760(t)
			update := unit.UpdateDataMonster()
			update.CurrentEnemy = &Object{PosVec: types.Ptf(120, 220)}
			head := update.AIStackHead()
			head.SetArgs(types.Ptf(300, 400), uint32(0xdeadbeef), uint32(0xaabbccdd))
			monsterActionFlee544760(unit, monsterActionFleeHooks544760{
				frame: func() uint32 { return 100 }, tickRate: func() uint32 { return 30 },
				pop: func() int { return 0 },
				generatePath: func(_ []types.Pointf, _ *Object, target *types.Pointf) int {
					if *target != update.CurrentEnemy.PosVec {
						t.Fatalf("target = %v", *target)
					}
					*target = types.Ptf(700, 800)
					return count
				},
			})
			want := unit.PosVec
			if count > 1 {
				want = types.Ptf(700, 800)
			}
			if head.ArgPos(0) != want || head.ArgU32(2) != 0xdeadbeef || head.ArgU32(3) != 0xaabbccdd {
				t.Fatalf("head args = %v/%#x/%#x, want %v/deadbeef/aabbccdd", head.ArgPos(0), head.ArgU32(2), head.ArgU32(3), want)
			}
		})
	}
}

func TestMonsterActionFlee544760RangeUsesUnspilledCoordinates(t *testing.T) {
	unit := fleeMonsterTestObject544760(t)
	unit.PosVec = types.Ptf(0, 0)
	update := unit.UpdateDataMonster()
	update.CurrentEnemy = &Object{PosVec: types.Ptf(math.Nextafter32(1, 0), 0.0003)}
	update.FleeRange, update.Field2, update.Field70 = 1, 3, 4
	monsterActionFlee544760(unit, monsterActionFleeHooks544760{
		frame: func() uint32 { return 20 }, tickRate: func() uint32 { return 30 },
		pop: func() int { return 0 },
	})
	if update.Field2 != 0 {
		t.Fatalf("rounded range retained stale path = %d", update.Field2)
	}
}
