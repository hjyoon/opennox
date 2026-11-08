package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func aiPathBoundsCircleTest(class object.Class, flags object.Flags, x, y int32, radius float32) *Object {
	obj := aiPathCircleObject50AFA0(class, flags, x, y)
	obj.Shape.Circle.R = radius
	obj.Shape.Circle.R2 = radius * radius
	obj.Nox_xxx_objectUnkUpdateCoords_4E7290()
	return obj
}

func TestAIPathIndexObjectBoundaryOverlap(t *testing.T) {
	edges := []struct {
		name string
		x, y int32
	}{
		{"left", 0, 25}, {"right", aiMapIndexSize - 1, 25},
		{"top", 25, 0}, {"bottom", 25, aiMapIndexSize - 1},
		{"low-corner", 0, 0}, {"high-corner", aiMapIndexSize - 1, aiMapIndexSize - 1},
	}
	for _, edge := range edges {
		t.Run(edge.name, func(t *testing.T) {
			paths := &serverAIPaths{mapIndexGen: 73}
			obj := aiPathBoundsCircleTest(object.ClassObstacle, 0, edge.x, edge.y, 35)
			paths.IndexObject(obj)
			cell := paths.MapIndex(int(edge.x), int(edge.y))
			if cell.IndexGen4 != 73 || cell.Flags8 != AIIndexObject|AIIndexObjectTall {
				t.Fatalf("in-bounds overlap lost: generation=%d flags=%#x", cell.IndexGen4, cell.Flags8)
			}
			if got := paths.MapIndex(128, 128).Flags8; got != 0 {
				t.Fatalf("unrelated cell changed: %#x", got)
			}
		})
	}
}

func TestAIPathIndexObjectTransitionClassBounds(t *testing.T) {
	tests := []struct {
		name  string
		class object.Class
		flags object.Flags
		want  AIMapIndexFlags
	}{
		{"marked-complex-immobile-obstacle", object.ClassObstacle | object.ClassComplex | object.ClassImmobile, object.FlagMarked, AIIndexObject | AIIndexObjectTall},
		{"destroyed-short-obstacle", object.ClassObstacle | object.ClassComplex | object.ClassImmobile, object.FlagMarked | object.FlagDestroyed | object.FlagShort, AIIndexObject},
		{"no-update-complex", object.ClassComplex, object.FlagMarked | object.FlagNoUpdate, AIIndexObject | AIIndexObjectTall},
		{"fire-below", object.ClassFire, object.FlagMarked | object.FlagBelow, AIIndexFire},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			paths := &serverAIPaths{mapIndexGen: 91}
			obj := aiPathBoundsCircleTest(test.class, test.flags, aiMapIndexSize-1, 25, 35)
			paths.IndexObject(obj)
			cell := paths.MapIndex(aiMapIndexSize-1, 25)
			if cell.IndexGen4 != 91 || cell.Flags8 != test.want {
				t.Fatalf("class/flags admission changed: generation=%d flags=%#x want=%#x", cell.IndexGen4, cell.Flags8, test.want)
			}
		})
	}
}

func TestAIPathDangerousBoundsRestore(t *testing.T) {
	for _, test := range []struct {
		name string
		x, y int32
	}{
		{"expanded-edge", aiMapIndexSize - 1, 25}, {"empty-intersection", aiMapIndexSize + 4, 25},
	} {
		t.Run(test.name, func(t *testing.T) {
			paths := &serverAIPaths{mapIndexGen: 37}
			obj := aiPathBoundsCircleTest(object.ClassObstacle|object.ClassDangerous, object.FlagMarked, test.x, test.y, 2.25)
			obj.ZSize1, obj.ZSize2 = math.Float32frombits(0x41234567), math.Float32frombits(0x42345678)
			obj.ObjNext = new(Object)
			shape, p1, p2, next := obj.Shape, obj.CollideP1, obj.CollideP2, obj.ObjNext
			z1, z2 := math.Float32bits(obj.ZSize1), math.Float32bits(obj.ZSize2)
			runtime.GC()
			paths.IndexObject(obj)
			if obj.Shape != shape || obj.CollideP1 != p1 || obj.CollideP2 != p2 ||
				math.Float32bits(obj.ZSize1) != z1 || math.Float32bits(obj.ZSize2) != z2 || obj.ObjNext != next {
				t.Fatal("dangerous geometry or full native object link was not restored")
			}
			if test.x < aiMapIndexSize {
				if got := paths.MapIndex(int(test.x), int(test.y)); got.IndexGen4 != 37 || got.Flags8&AIIndexObject == 0 {
					t.Fatal("expanded in-bounds overlap disappeared")
				}
			} else {
				for y := range paths.mapIndex {
					for x := range paths.mapIndex[y] {
						if paths.mapIndex[y][x] != (AIMapIndexNode{}) {
							t.Fatal("fully outside dangerous object indexed a border cell")
						}
					}
				}
			}
			t.Logf("typed object pointer=%#x pointer-width=%d", uintptr(unsafe.Pointer(obj)), unsafe.Sizeof(obj))
			runtime.KeepAlive(obj)
		})
	}
}

func TestAIPathRebuildTransitionBoundaryObjects(t *testing.T) {
	srv := newElevatorStateServer(t)
	paths := &serverAIPaths{s: srv}
	outside := aiPathBoundsCircleTest(object.ClassObstacle|object.ClassComplex, object.FlagMarked|object.FlagDestroyed, aiMapIndexSize+4, 25, 2)
	immobile := aiPathBoundsCircleTest(object.ClassImmobile, object.FlagMarked, 0, 25, 35)
	combined := aiPathBoundsCircleTest(object.ClassObstacle|object.ClassImmobile|object.ClassComplex, object.FlagMarked, aiMapIndexSize-1, 25, 35)
	inside := aiPathBoundsCircleTest(object.ClassObstacle, 0, 40, 40, 1)
	outside.ObjNext, immobile.ObjNext, combined.ObjNext = immobile, combined, inside
	srv.Objs.List = outside
	runtime.GC()
	paths.Sub50AFA0()
	if paths.MapIndex(0, 25).Flags8 != AIIndexOccupied|AIIndexOccupiedTall {
		t.Fatal("immobile in-bounds overlap was dropped")
	}
	if paths.MapIndex(aiMapIndexSize-1, 25).Flags8 != AIIndexOccupied|AIIndexOccupiedTall|AIIndexObject|AIIndexObjectTall {
		t.Fatal("combined dynamic/static obstacle overlap was dropped")
	}
	if paths.MapIndex(40, 40).Flags8 != AIIndexObject|AIIndexObjectTall || outside.ObjNext != immobile || immobile.ObjNext != combined || combined.ObjNext != inside {
		t.Fatal("rebuild did not continue through the full native object list")
	}
	// A second map rebuild must clear old border cells and index the new
	// position; a previous out-of-grid object must not stop list traversal.
	immobile.PosVec, immobile.NewPos = types.Ptf(70*23+11.5, 70*23+11.5), types.Ptf(70*23+11.5, 70*23+11.5)
	immobile.Nox_xxx_objectUnkUpdateCoords_4E7290()
	paths.Sub50AFA0()
	if paths.MapIndex(0, 25).Flags8 != 0 || paths.MapIndex(70, 70).Flags8 != AIIndexOccupied|AIIndexOccupiedTall {
		t.Fatal("map rebuild retained stale border occupancy or lost moved occupancy")
	}
	runtime.KeepAlive(outside)
}

func TestAIPathRebuildOutsideSpecialMarkers(t *testing.T) {
	for _, test := range []struct {
		name  string
		class object.Class
		want  AIMapIndexFlags
	}{
		{"hole", object.ClassHole, AIIndexHole}, {"transporter", object.ClassTransporter, AIIndexTransporter},
		{"elevator", object.ClassElevator, AIIndexElevator}, {"shaft", object.ClassElevatorShaft, AIIndexElevatorShaft},
		{"hole-priority", object.ClassHole | object.ClassTransporter | object.ClassImmobile, AIIndexHole},
	} {
		t.Run(test.name, func(t *testing.T) {
			srv := newElevatorStateServer(t)
			paths := &serverAIPaths{s: srv}
			invalid := &Object{ObjClass: test.class, PosVec: types.Ptf(-46, 34.5)}
			valid := &Object{ObjClass: test.class, PosVec: types.Ptf(34.5, 57.5)}
			invalid.ObjNext = valid
			srv.Objs.List = invalid
			paths.Sub50AFA0()
			if got := paths.MapIndex(1, 2).Flags8; got != test.want {
				t.Fatalf("valid marker/priority changed after invalid marker: got=%#x want=%#x", got, test.want)
			}
			if got := paths.MapIndex(0, 1).Flags8; got != 0 {
				t.Fatalf("invalid marker was clamped onto border: %#x", got)
			}
		})
	}
}

func TestAIPathRebuildOutsideWaypoints(t *testing.T) {
	srv := newElevatorStateServer(t)
	paths := &serverAIPaths{s: srv}
	negative := &Waypoint{PosVec: types.Ptf(-46, 34.5), Flags2: 0x80}
	far := &Waypoint{PosVec: types.Ptf(float32(aiMapIndexSize+4)*23, 34.5), Flags2: 0x80}
	invalid := &Waypoint{PosVec: types.Ptf(float32(math.NaN()), 34.5), Flags2: 0x80}
	valid := &Waypoint{PosVec: types.Ptf(34.5, 57.5), Flags2: 0x80}
	unflagged := &Waypoint{PosVec: types.Ptf(80.5, 103.5)}
	negative.WpNext, far.WpNext, invalid.WpNext, valid.WpNext = far, invalid, valid, unflagged
	srv.WPs.List = negative
	paths.Sub50AFA0()
	if paths.MapIndex(1, 2).Flags8 != AIIndexWaypoint || paths.MapIndex(0, 1).Flags8 != 0 || paths.MapIndex(3, 4).Flags8 != 0 {
		t.Fatal("waypoint bounds, visitation or flag filtering changed")
	}
}

func TestAIPathGridBoundsAreFiniteOrEmpty(t *testing.T) {
	tests := []struct {
		name   string
		p1, p2 types.Pointf
		empty  bool
	}{
		{"large-crossing", types.Ptf(-1e9, -1e9), types.Ptf(1e9, 1e9), false},
		{"outside-negative", types.Ptf(-2e9, -2e9), types.Ptf(-1e9, -1e9), true},
		{"outside-positive", types.Ptf(1e9, 1e9), types.Ptf(2e9, 2e9), true},
		{"reversed", types.Ptf(2000, 2000), types.Ptf(1000, 1000), true},
		{"NaN", types.Ptf(float32(math.NaN()), 0), types.Ptf(float32(math.NaN()), 1000), true},
		{"infinite", types.Ptf(float32(math.Inf(-1)), 0), types.Ptf(float32(math.Inf(1)), 1000), true},
		{"conversion-overflow", types.Ptf(math.MaxFloat32, 0), types.Ptf(math.MaxFloat32, 1000), true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			minX, minY, maxX, maxY := aiPathGridBounds50AFA0(test.p1, test.p2)
			empty := minX > maxX || minY > maxY
			if empty != test.empty {
				t.Fatalf("empty=%v want=%v bounds=%d,%d..%d,%d", empty, test.empty, minX, minY, maxX, maxY)
			}
			if !empty && (minX != 0 || minY != 0 || maxX != aiMapIndexSize-1 || maxY != aiMapIndexSize-1) {
				t.Fatal("large finite collider must intersect the entire finite grid; no unbounded int32 loop")
			}
		})
	}
}
