package legacy

import (
	"bytes"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func pointOccupancyObjectBytes54EF00(unit *server.Object) []byte {
	out := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	at, size := int(unsafe.Offsetof(unit.Field62)), int(unsafe.Sizeof(unit.Field62))
	clear(out[at : at+size])
	return out
}

func TestPointOccupancyProduction54EF00(t *testing.T) {
	if occupancyOriginalParentABI54EF00() != 64 || occupancyOriginalCandidateABI54EF60() != 64 || pointShapeOriginalABI547DB0() != 64 {
		t.Fatal("all three native public signatures required")
	}
	for _, tc := range []struct {
		name             string
		kind             server.ShapeKind
		position, point  types.Pointf
		radius, radius2  float32
		class            object.Class
		flags            uint32
		inside, occupied bool
	}{
		{"circle/center", server.ShapeKindCircle, types.Ptf(300, 300), types.Ptf(300, 300), 10, 100, object.ClassMonster, 0, true, true},
		{"circle/boundary", server.ShapeKindCircle, types.Ptf(300, 300), types.Ptf(310, 300), 10, 100, object.ClassMonster, 0, true, true},
		{"circle/outside", server.ShapeKindCircle, types.Ptf(300, 300), types.Ptf(312, 300), 10, 100, object.ClassMonster, 0, false, false},
		{"circle/unspilled-difference", server.ShapeKindCircle, types.Ptf(1, 0), types.Ptf(math.Float32frombits(0xb3000000), 0), 1, 1, object.ClassMonster, 0, false, false},
		{"circle/unordered-radius", server.ShapeKindCircle, types.Ptf(300, 300), types.Ptf(300, 300), 10, math.Float32frombits(0x7fc12345), object.ClassMonster, 0, true, true},
		{"box/center", server.ShapeKindBox, types.Ptf(300, 300), types.Ptf(300, 300), 0, 0, object.ClassMonster, 0, true, true},
		{"box/inside", server.ShapeKindBox, types.Ptf(300, 300), types.Ptf(309, 300), 0, 0, object.ClassMonster, 0, true, true},
		{"box/east-edge", server.ShapeKindBox, types.Ptf(300, 300), types.Ptf(310, 300), 0, 0, object.ClassMonster, 0, false, false},
		{"box/west-edge", server.ShapeKindBox, types.Ptf(300, 300), types.Ptf(290, 300), 0, 0, object.ClassMonster, 0, false, false},
		{"box/north-edge", server.ShapeKindBox, types.Ptf(300, 300), types.Ptf(300, 310), 0, 0, object.ClassMonster, 0, false, false},
		{"box/south-edge", server.ShapeKindBox, types.Ptf(300, 300), types.Ptf(300, 290), 0, 0, object.ClassMonster, 0, false, false},
		{"class/generator", server.ShapeKindCircle, types.Ptf(300, 300), types.Ptf(300, 300), 10, 100, object.ClassMonsterGenerator, 0, true, false},
		{"field5/disabled", server.ShapeKindCircle, types.Ptf(300, 300), types.Ptf(300, 300), 10, 100, object.ClassMonster, 0x800, true, false},
		{"field5/unrelated", server.ShapeKindCircle, types.Ptf(300, 300), types.Ptf(300, 300), 10, 100, object.ClassMonster, 0x80000000, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, unit, _ := monsterLookAtFixture5125A0(t)
			srv.Map.Init()
			t.Cleanup(srv.Map.Free)
			oldServer := GetServer
			GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
			t.Cleanup(func() { GetServer = oldServer })
			oldRuntime := pointOccupancyRuntimeOverride54EF00
			pointOccupancyRuntimeOverride54EF00 = nil
			t.Cleanup(func() { pointOccupancyRuntimeOverride54EF00 = oldRuntime })
			point, free := alloc.New(types.Pointf{})
			t.Cleanup(free)
			*point = tc.point
			unit.ObjClass, unit.ObjFlags, unit.Field5 = tc.class, object.FlagActive, tc.flags
			unit.PosVec, unit.NewPos = tc.position, tc.position
			unit.Shape = server.Shape{Kind: tc.kind}
			unit.Shape.Circle.R, unit.Shape.Circle.R2 = tc.radius, tc.radius2
			if tc.kind == server.ShapeKindBox {
				unit.Shape.Box.LeftTop, unit.Shape.Box.LeftBottom = 0, -10
				unit.Shape.Box.LeftBottom2, unit.Shape.Box.LeftTop2 = -10, 0
				unit.Shape.Box.RightTop, unit.Shape.Box.RightBottom = 10, 0
				unit.Shape.Box.RightBottom2, unit.Shape.Box.RightTop2 = 0, 10
			}
			srv.Map.AddObjectToIndex(unit)
			before := pointOccupancyObjectBytes54EF00(unit)
			beforePoint := *point
			scratch := memmap.PtrUint32(0x5D4594, 2491708)
			oldScratch := *scratch
			t.Cleanup(func() { *scratch = oldScratch })
			if got := pointShapeOriginalC547DB0(unit, point); (got == 1) != tc.inside {
				t.Fatal("actual native predicate", got, tc.inside)
			}
			for _, path := range []string{"C-parent", "Go-server", "C-callback"} {
				t.Run(path, func(t *testing.T) {
					*scratch = 0x89abcdef
					want := uint32(0)
					if tc.occupied {
						want = 1
					}
					switch path {
					case "C-parent":
						if got := occupancyOriginalParentC54EF00(point); got != want {
							t.Fatal("actual indexed C parent", got, want)
						}
					case "Go-server":
						if got := srv.PointOccupied54EF00(point); got != tc.occupied {
							t.Fatal("ordinary indexed server", got, tc.occupied)
						}
					case "C-callback":
						if !tc.occupied {
							want = 0x89abcdef
						}
						occupancyOriginalCandidateC54EF60(unit, point)
					}
					if *scratch != want || *point != beforePoint || !bytes.Equal(pointOccupancyObjectBytes54EF00(unit), before) {
						t.Fatal("scratch, whole native object except real map visit tokens, or caller point changed", *scratch, want)
					}
				})
			}
		})
	}
}
