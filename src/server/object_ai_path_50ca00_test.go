package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestAIPathHasNoEnemiesAround50CA60PreservesCellAndTraversal(t *testing.T) {
	type objectID uint64
	const (
		self   objectID = 0x1_0000_0011
		friend objectID = 0x1_0000_0022
		enemy  objectID = 0x1_0000_0033
		later  objectID = 0x1_0000_0044
	)

	var (
		gotPos      types.Pointf
		gotRadius   float32
		visited     []objectID
		enemyChecks [][2]objectID
	)
	got := aiPathHasNoEnemiesAround50CA60(self, 7, -3, aiPathEnemySearchHooks50CA60[objectID]{
		eachInCircle: func(pos types.Pointf, radius float32, visit func(objectID) bool) {
			gotPos = pos
			gotRadius = radius
			for _, candidate := range []objectID{friend, enemy, later} {
				visited = append(visited, candidate)
				if !visit(candidate) {
					t.Fatal("enemy callback stopped the circle iterator")
				}
			}
		},
		isEnemyTo: func(source, candidate objectID) bool {
			enemyChecks = append(enemyChecks, [2]objectID{source, candidate})
			return candidate == enemy
		},
	})
	if got {
		t.Fatal("enemy search returned clear, want enemy found")
	}
	wantPos := types.Pointf{X: 172.5, Y: -57.5}
	if gotPos != wantPos {
		t.Fatalf("circle center = %v, want %v", gotPos, wantPos)
	}
	if gotRadius != 100 {
		t.Fatalf("circle radius = %v, want 100", gotRadius)
	}
	if want := []objectID{friend, enemy, later}; !equalObjectIDs50CA60(visited, want) {
		t.Fatalf("visited objects = %#v, want %#v", visited, want)
	}
	wantChecks := [][2]objectID{{self, friend}, {self, enemy}}
	if len(enemyChecks) != len(wantChecks) {
		t.Fatalf("enemy checks = %#v, want %#v", enemyChecks, wantChecks)
	}
	for i, want := range wantChecks {
		if enemyChecks[i] != want {
			t.Fatalf("enemy check %d = %#v, want %#v", i, enemyChecks[i], want)
		}
	}
}

func TestAIPathHasNoEnemiesAround50CA60ReturnsClear(t *testing.T) {
	const self uint64 = 0x1_0000_0001
	checks := 0
	got := aiPathHasNoEnemiesAround50CA60(self, 0, 0, aiPathEnemySearchHooks50CA60[uint64]{
		eachInCircle: func(pos types.Pointf, radius float32, visit func(uint64) bool) {
			if got, want := [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}, [2]uint32{0x41380000, 0x41380000}; got != want {
				t.Fatalf("origin cell center bits = %#x, want %#x", got, want)
			}
			if radius != 100 {
				t.Fatalf("circle radius = %v, want 100", radius)
			}
			for _, candidate := range []uint64{0x1_0000_0002, 0x1_0000_0003} {
				if !visit(candidate) {
					t.Fatal("clear callback stopped the circle iterator")
				}
			}
		},
		isEnemyTo: func(source, candidate uint64) bool {
			if source != self {
				t.Fatalf("enemy source = %#x, want %#x", source, self)
			}
			checks++
			return false
		},
	})
	if !got {
		t.Fatal("enemy search reported an enemy, want clear")
	}
	if checks != 2 {
		t.Fatalf("enemy checks = %d, want 2", checks)
	}
}

func equalObjectIDs50CA60[T comparable](got, want []T) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
