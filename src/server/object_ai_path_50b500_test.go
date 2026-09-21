package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestAIPathIndexResetAndThrottle50B500(t *testing.T) {
	s := newElevatorStateServer(t)
	paths := &serverAIPaths{
		s:            s,
		mapIndexGen:  7,
		lastFrame:    20,
		calculated:   true,
		MapIndexLast: 9,
	}

	paths.Sub_50B500()
	if paths.calculated || paths.lastFrame != 20 || paths.mapIndexGen != 7 || paths.MapIndexLast != 9 {
		t.Fatalf("partial reset = calculated %t, last %d, generation %d, visit %d",
			paths.calculated, paths.lastFrame, paths.mapIndexGen, paths.MapIndexLast)
	}

	paths.calculated = true
	paths.Sub_50B510()
	if paths.calculated || paths.lastFrame != 0 || paths.mapIndexGen != 7 || paths.MapIndexLast != 9 {
		t.Fatalf("full reset = calculated %t, last %d, generation %d, visit %d",
			paths.calculated, paths.lastFrame, paths.mapIndexGen, paths.MapIndexLast)
	}

	s.SetFrame(14)
	paths.IndexObjects()
	if paths.calculated || paths.lastFrame != 0 || paths.mapIndexGen != 7 {
		t.Fatalf("early index = calculated %t, last %d, generation %d",
			paths.calculated, paths.lastFrame, paths.mapIndexGen)
	}

	s.SetFrame(15)
	paths.IndexObjects()
	if !paths.calculated || paths.lastFrame != 15 || paths.mapIndexGen != 8 {
		t.Fatalf("threshold index = calculated %t, last %d, generation %d",
			paths.calculated, paths.lastFrame, paths.mapIndexGen)
	}

	paths.calculated = false
	paths.lastFrame = math.MaxUint32 - 5
	s.SetFrame(9) // unsigned wrapped difference is exactly 15
	paths.IndexObjects()
	if !paths.calculated || paths.lastFrame != 9 || paths.mapIndexGen != 9 {
		t.Fatalf("wrapped index = calculated %t, last %d, generation %d",
			paths.calculated, paths.lastFrame, paths.mapIndexGen)
	}
}

func TestAIPathGridCell50B810UsesBinary32X87Conversion(t *testing.T) {
	paths := new(serverAIPaths)
	tests := []struct {
		pos  types.Pointf
		want [2]int
	}{
		{types.Ptf(34.5, 57.5), [2]int{2, 2}},
		{types.Ptf(-34.5, -57.5), [2]int{-2, -2}},
		{types.Ptf(80.5, 11.5), [2]int{4, 0}},
	}
	for _, test := range tests {
		x, y := paths.GridCell50B810(&test.pos)
		if x != test.want[0] || y != test.want[1] {
			t.Errorf("grid(%v) = (%d, %d), want (%d, %d)",
				test.pos, x, y, test.want[0], test.want[1])
		}
	}
}

func TestAIPathStaticPrecheck50B950UsesHeightSpecificBits(t *testing.T) {
	paths := new(serverAIPaths)
	ground := new(Object)
	airborne := &Object{ObjFlags: object.FlagAirborne}
	cell := paths.MapIndex(12, 34)

	cell.Flags8 = AIIndexOccupied
	if !paths.CheckIndexFlags(ground, 12, 34) || paths.CheckIndexFlags(airborne, 12, 34) {
		t.Fatalf("ground occupancy = ground %t, airborne %t, want true/false",
			paths.CheckIndexFlags(ground, 12, 34), paths.CheckIndexFlags(airborne, 12, 34))
	}

	cell.Flags8 = AIIndexOccupiedTall
	if paths.CheckIndexFlags(ground, 12, 34) || !paths.CheckIndexFlags(airborne, 12, 34) {
		t.Fatalf("tall occupancy = ground %t, airborne %t, want false/true",
			paths.CheckIndexFlags(ground, 12, 34), paths.CheckIndexFlags(airborne, 12, 34))
	}

	cell.Flags8 = AIIndexHole
	if paths.CheckIndexFlags(ground, 12, 34) || paths.CheckIndexFlags(airborne, 12, 34) {
		t.Fatalf("hole flag was mistaken for occupancy: ground %t, airborne %t",
			paths.CheckIndexFlags(ground, 12, 34), paths.CheckIndexFlags(airborne, 12, 34))
	}
}

func TestAIPathDynamicPrecheck50B8E0UsesGenerationAndHeight(t *testing.T) {
	paths := &serverAIPaths{mapIndexGen: 0x12345678}
	ground := new(Object)
	airborne := &Object{ObjFlags: object.FlagAirborne}
	fireproof := &Object{ObjSubClass: 1 << 10}
	cell := paths.MapIndex(23, 45)

	cell.IndexGen4 = paths.mapIndexGen - 1
	cell.Flags8 = AIIndexObject | AIIndexObjectTall | AIIndexFire
	if got := paths.sub50B8E0(ground, 23, 45); got != 0 {
		t.Fatalf("stale generation = %d, want 0", got)
	}

	cell.IndexGen4 = paths.mapIndexGen
	cell.Flags8 = AIIndexObject
	if got := paths.sub50B8E0(ground, 23, 45); got != 1 {
		t.Fatalf("ground object = %d, want 1", got)
	}
	if got := paths.sub50B8E0(airborne, 23, 45); got != 0 {
		t.Fatalf("short object for airborne = %d, want 0", got)
	}

	cell.Flags8 = AIIndexObjectTall
	if got := paths.sub50B8E0(airborne, 23, 45); got != 1 {
		t.Fatalf("tall object for airborne = %d, want 1", got)
	}
	if got := paths.sub50B8E0(ground, 23, 45); got != 0 {
		t.Fatalf("tall-only object for ground = %d, want 0", got)
	}

	cell.Flags8 = AIIndexFire
	if got := paths.sub50B8E0(ground, 23, 45); got != 1 {
		t.Fatalf("fire for ordinary object = %d, want 1", got)
	}
	if got := paths.sub50B8E0(fireproof, 23, 45); got != 0 {
		t.Fatalf("fire for subclass bit 10 = %d, want 0", got)
	}
}

func TestAIPathCombinedPrecheck50B8A0(t *testing.T) {
	paths := &serverAIPaths{mapIndexGen: 99}
	obj := new(Object)
	cell := paths.MapIndex(67, 89)

	if !paths.Nox_xxx_pathfind_preCheckWalls2_50B8A0(obj, 67, 89) {
		t.Fatal("clear cell was rejected")
	}
	cell.Flags8 = AIIndexOccupied
	if paths.Nox_xxx_pathfind_preCheckWalls2_50B8A0(obj, 67, 89) {
		t.Fatal("static occupied cell was accepted")
	}
	cell.Flags8 = AIIndexObject
	cell.IndexGen4 = paths.mapIndexGen
	if paths.Nox_xxx_pathfind_preCheckWalls2_50B8A0(obj, 67, 89) {
		t.Fatal("current-generation dynamic obstacle was accepted")
	}
}
