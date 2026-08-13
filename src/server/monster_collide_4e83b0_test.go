package server

import (
	"math"
	"reflect"
	"testing"
)

type monsterCollideTestObject4E83B0 struct {
	name   string
	update *monsterCollideTestUpdate4E83B0
}

type monsterCollideTestUpdate4E83B0 struct {
	block *monsterCollideTestBlock4E83B0
}

type monsterCollideTestBlock4E83B0 struct {
	id int32
}

func TestMonsterCollide4E83B0ReadCallOrderArgumentsAndExactReturn(t *testing.T) {
	for _, result := range []int64{math.MinInt64, -1, 0, 1, math.MaxInt64} {
		events := []string{}
		block := &monsterCollideTestBlock4E83B0{id: 22}
		monster := &monsterCollideTestObject4E83B0{
			name:   "monster",
			update: &monsterCollideTestUpdate4E83B0{block: block},
		}
		other := &monsterCollideTestObject4E83B0{name: "other"}
		got := monsterCollide4E83B0(monster, other, monsterCollideHooks4E83B0[
			*monsterCollideTestObject4E83B0,
			*monsterCollideTestUpdate4E83B0,
			*monsterCollideTestBlock4E83B0,
			int64,
		]{
			updateData: func(got *monsterCollideTestObject4E83B0) *monsterCollideTestUpdate4E83B0 {
				events = append(events, "update")
				if got != monster {
					t.Fatal("update-data received a different monster")
				}
				return got.update
			},
			collisionBlock: func(got *monsterCollideTestUpdate4E83B0) *monsterCollideTestBlock4E83B0 {
				events = append(events, "block")
				if got != monster.update {
					t.Fatal("collision-block received different update data")
				}
				return got.block
			},
			scriptCallback: func(gotBlock *monsterCollideTestBlock4E83B0, caller, trigger *monsterCollideTestObject4E83B0) int64 {
				events = append(events, "call")
				if gotBlock != block || caller != other || trigger != monster {
					t.Fatalf("callback args = (%p, %p, %p), want (%p, %p, %p)", gotBlock, caller, trigger, block, other, monster)
				}
				return result
			},
		})
		if got != result {
			t.Fatalf("result = %d, want %d", got, result)
		}
		if want := []string{"update", "block", "call"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestMonsterCollide4E83B0NilUpdateFaultsBeforeCallback(t *testing.T) {
	events := []string{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update returned without a panic")
		}
		if want := []string{"update", "block"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	monsterCollide4E83B0(
		&monsterCollideTestObject4E83B0{name: "monster"},
		(*monsterCollideTestObject4E83B0)(nil),
		monsterCollideHooks4E83B0[
			*monsterCollideTestObject4E83B0,
			*monsterCollideTestUpdate4E83B0,
			*monsterCollideTestBlock4E83B0,
			int,
		]{
			updateData: func(obj *monsterCollideTestObject4E83B0) *monsterCollideTestUpdate4E83B0 {
				events = append(events, "update")
				return obj.update
			},
			collisionBlock: func(update *monsterCollideTestUpdate4E83B0) *monsterCollideTestBlock4E83B0 {
				events = append(events, "block")
				return update.block
			},
			scriptCallback: func(*monsterCollideTestBlock4E83B0, *monsterCollideTestObject4E83B0, *monsterCollideTestObject4E83B0) int {
				t.Fatal("nil update reached callback")
				return 0
			},
		},
	)
}
