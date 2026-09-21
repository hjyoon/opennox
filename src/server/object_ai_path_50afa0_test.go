package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func aiPathCircleObject50AFA0(class object.Class, flags object.Flags, cellX, cellY int32) *Object {
	base := types.Ptf(float32(cellX)*23, float32(cellY)*23)
	center := types.Ptf(base.X+11.5, base.Y+11.5)
	obj := &Object{
		ObjClass:  class,
		ObjFlags:  flags,
		PosVec:    center,
		NewPos:    center,
		CollideP1: base,
		CollideP2: base,
	}
	obj.Shape.Kind = ShapeKindCircle
	obj.Shape.Circle.R = 1
	obj.Shape.Circle.R2 = 1
	return obj
}

func TestAIPathCellSamples50AFA0MatchExecutableBits(t *testing.T) {
	want := [9][2]uint32{
		{0x40133333, 0x40133333},
		{0x41380000, 0x40133333},
		{0x41a59999, 0x40133333},
		{0x40133333, 0x41380000},
		{0x41380000, 0x41380000},
		{0x41a59999, 0x41380000},
		{0x40133333, 0x41a59999},
		{0x41380000, 0x41a59999},
		{0x41a59999, 0x41a59999},
	}
	for i, sample := range aiPathCellSamples50AFA0 {
		got := [2]uint32{math.Float32bits(sample.X), math.Float32bits(sample.Y)}
		if got != want[i] {
			t.Errorf("sample %d bits = %#x/%#x, want %#x/%#x",
				i, got[0], got[1], want[i][0], want[i][1])
		}
	}
	if got := math.Float32bits(aiPathDangerousMargin50B2C0); got != 0x41380000 {
		t.Fatalf("dangerous margin bits = %#08x, want 0x41380000", got)
	}
}

func TestAIPathGridCell50AFA0UsesX87Rounding(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{11.5, 0},
		{34.5, 2},
		{57.5, 2},
		{80.5, 4},
		{-34.5, -2},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.NaN()), math.MinInt32},
	}
	for _, test := range tests {
		if got := aiPathGridCell50AFA0(test.value); got != test.want {
			t.Errorf("cell(%v) = %d, want %d", test.value, got, test.want)
		}
	}
}

func TestAIPathRebuild50AFA0SpecialPriorityAndWaypointRounding(t *testing.T) {
	s := newElevatorStateServer(t)
	paths := &serverAIPaths{s: s}

	hole := &Object{ObjClass: object.ClassHole | object.ClassTransporter, PosVec: types.Ptf(34.5, 80.5)}
	transporter := &Object{ObjClass: object.ClassTransporter, PosVec: types.Ptf(103.5, 126.5)}
	elevator := &Object{ObjClass: object.ClassElevator, PosVec: types.Ptf(149.5, 172.5)}
	shaft := &Object{ObjClass: object.ClassElevatorShaft, PosVec: types.Ptf(195.5, 218.5)}
	doorHole := &Object{ObjClass: object.ClassDoor | object.ClassHole, PosVec: types.Ptf(241.5, 241.5)}
	hole.ObjNext = transporter
	transporter.ObjNext = elevator
	elevator.ObjNext = shaft
	shaft.ObjNext = doorHole
	s.Objs.List = hole
	s.WPs.List = &Waypoint{PosVec: types.Ptf(264.5, 287.5), Flags2: 0x80}

	paths.Sub50AFA0()

	tests := []struct {
		x, y int
		want AIMapIndexFlags
	}{
		{2, 4, AIIndexHole},
		{4, 6, AIIndexTransporter},
		{6, 8, AIIndexElevator},
		{8, 10, AIIndexElevatorShaft},
		{10, 10, 0},
		{12, 12, AIIndexWaypoint},
	}
	for _, test := range tests {
		if got := paths.MapIndex(test.x, test.y).Flags8; got != test.want {
			t.Errorf("cell (%d,%d) flags = %#x, want %#x", test.x, test.y, got, test.want)
		}
	}
}

func TestAIPathRebuild50AFA0ReadsNativeObjectFlags(t *testing.T) {
	if got := unsafe.Offsetof(Object{}.ObjFlags); got != 20 {
		t.Fatalf("native ObjFlags offset = %d, want 20", got)
	}
	s := newElevatorStateServer(t)
	paths := &serverAIPaths{s: s}

	skipped := aiPathCircleObject50AFA0(object.ClassImmobile, object.FlagNoCollide, 20, 20)
	skipped.ObjSubClass = 0
	short := aiPathCircleObject50AFA0(object.ClassImmobile, object.FlagShort, 22, 22)
	short.ObjSubClass = 0
	skipped.ObjNext = short
	s.Objs.List = skipped

	paths.Sub50AFA0()

	if got := paths.MapIndex(20, 20).Flags8; got != 0 {
		t.Fatalf("no-collide object flags = %#x, want 0", got)
	}
	if got := paths.MapIndex(22, 22).Flags8; got != AIIndexOccupied {
		t.Fatalf("short object flags = %#x, want %#x", got, AIIndexOccupied)
	}
}

func TestAIPathIndexObject50B2C0NativeFlagsAndClasses(t *testing.T) {
	paths := &serverAIPaths{mapIndexGen: 77}

	obstacle := aiPathCircleObject50AFA0(object.ClassObstacle, 0, 30, 30)
	paths.IndexObject(obstacle)
	cell := paths.MapIndex(30, 30)
	if cell.IndexGen4 != 77 || cell.Flags8 != AIIndexObject|AIIndexObjectTall {
		t.Fatalf("obstacle generation/flags = %d/%#x, want 77/%#x",
			cell.IndexGen4, cell.Flags8, AIIndexObject|AIIndexObjectTall)
	}

	short := aiPathCircleObject50AFA0(object.ClassObstacle, object.FlagShort, 32, 32)
	short.ObjSubClass = 0
	paths.IndexObject(short)
	if got := paths.MapIndex(32, 32).Flags8; got != AIIndexObject {
		t.Fatalf("short obstacle flags = %#x, want %#x", got, AIIndexObject)
	}

	skipped := aiPathCircleObject50AFA0(object.ClassObstacle, object.FlagNoCollide, 34, 34)
	skipped.ObjSubClass = 0
	paths.IndexObject(skipped)
	if got := paths.MapIndex(34, 34).Flags8; got != 0 {
		t.Fatalf("no-collide obstacle flags = %#x, want 0", got)
	}

	fire := aiPathCircleObject50AFA0(object.ClassFire, object.FlagBelow, 36, 36)
	paths.IndexObject(fire)
	if got := paths.MapIndex(36, 36).Flags8; got != AIIndexFire {
		t.Fatalf("fire flags = %#x, want %#x", got, AIIndexFire)
	}
}

func TestAIPathIndexObject50B2C0RestoresNativeDangerousGeometry(t *testing.T) {
	paths := &serverAIPaths{mapIndexGen: 91}
	obj := aiPathCircleObject50AFA0(object.ClassObstacle|object.ClassDangerous, 0, 40, 40)
	obj.Shape.Circle.R = 2.25
	obj.Shape.Circle.R2 = 2.25 * 2.25
	obj.ZSize1 = math.Float32frombits(0x41234567)
	obj.ZSize2 = math.Float32frombits(0x42345678)
	obj.Nox_xxx_objectUnkUpdateCoords_4E7290()
	wantShape := obj.Shape
	wantZ1, wantZ2 := obj.ZSize1, obj.ZSize2
	wantP1, wantP2 := obj.CollideP1, obj.CollideP2

	paths.IndexObject(obj)

	if obj.Shape != wantShape || math.Float32bits(obj.ZSize1) != math.Float32bits(wantZ1) ||
		math.Float32bits(obj.ZSize2) != math.Float32bits(wantZ2) {
		t.Fatalf("dangerous geometry was not restored: shape=%+v z=%#x/%#x",
			obj.Shape, math.Float32bits(obj.ZSize1), math.Float32bits(obj.ZSize2))
	}
	if obj.CollideP1 != wantP1 || obj.CollideP2 != wantP2 {
		t.Fatalf("restored collider = %v..%v, want %v..%v",
			obj.CollideP1, obj.CollideP2, wantP1, wantP2)
	}
	if got := paths.MapIndex(40, 40); got.IndexGen4 != 91 || got.Flags8&AIIndexObject == 0 {
		t.Fatalf("dangerous cell generation/flags = %d/%#x, want generation 91 and object bit",
			got.IndexGen4, got.Flags8)
	}
}

func TestAIPathExpandDangerousShape50B2C0MatchesX87CirclePrecision(t *testing.T) {
	shape := Shape{Kind: ShapeKindCircle}
	shape.Circle.R = math.Float32frombits(0x3f000006)

	aiPathExpandDangerousShape50B2C0(&shape)

	if got := math.Float32bits(shape.Circle.R); got != 0x41400000 {
		t.Fatalf("expanded radius bits = %#08x, want 0x41400000", got)
	}
	if got := math.Float32bits(shape.Circle.R2); got != 0x43100001 {
		t.Fatalf("expanded radius squared bits = %#08x, want x87 result 0x43100001", got)
	}
}

func TestAIPathExpandDangerousShape50B2C0RecalculatesBox(t *testing.T) {
	shape := Shape{Kind: ShapeKindBox}
	shape.Box.W = 13.25
	shape.Box.H = 7.75
	shape.Box.Calc()
	want := shape
	want.Box.W += 2 * aiPathDangerousMargin50B2C0
	want.Box.H += 2 * aiPathDangerousMargin50B2C0
	want.Box.Calc()

	aiPathExpandDangerousShape50B2C0(&shape)

	if shape != want {
		t.Fatalf("expanded box = %+v, want recalculated %+v", shape.Box, want.Box)
	}
}
