package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// These are the unchanged Wiz02c Bookcase6 position, speed and waypoint
// chain. The other orientations exercise the same failure on all four axes.
// Position integration here is a declared controller fixture; the campaign
// E2E separately exercises the original plate, collisions, rendering and Save.
func TestMoverUpdateNativeBookcase6AxisAlignedWaypoints54F740(t *testing.T) {
	for _, swap := range []bool{false, true} {
		for _, sx := range []float32{1, -1} {
			for _, sy := range []float32{1, -1} {
				t.Run(fmt.Sprintf("swap=%t/x=%g/y=%g", swap, sx, sy), func(t *testing.T) {
					point := func(x, y float32) types.Pointf {
						if swap {
							x, y = y, x
						}
						return types.Ptf(sx*x, sy*y)
					}
					s := newMoverStateServer(t)
					source := playerCloudGreatSwordAllocateR78[Object](t)
					target := playerCloudGreatSwordAllocateR78[Object](t)
					data := playerCloudGreatSwordAllocateR78[MoverUpdateData](t)
					*source = Object{serverHandle: s.handle, ObjFlags: object.Flags(moverUpdateActiveFlag54F740), UpdateData: unsafe.Pointer(data)}
					*target = Object{serverHandle: s.handle, ObjFlags: object.Flags(moverUpdateTargetRequiredFlag54F740), PosVec: point(1554, 4746)}
					*data = MoverUpdateData{Field_1: math.Float32frombits(20), Field_2: 137, Field_8: 1974}
					positions := []types.Pointf{point(1554, 4747), point(1566, 4759), point(1591, 4734)}
					waypoints := make(map[uint32]*Waypoint)
					for i, pos := range positions {
						wp := playerCloudGreatSwordAllocateR78[Waypoint](t)
						*wp = Waypoint{Index: uint32(137 + i), PosVec: pos}
						waypoints[wp.Index] = wp
						if i > 0 {
							prev := waypoints[wp.Index-1]
							prev.PointsCnt, prev.Points[0].Waypoint = 1, wp
						}
					}
					var removed int
					deps := moverUpdateNativeDeps54F740{
						objectByExtent: func(id uint32) *Object {
							if id != 1974 {
								t.Fatalf("target extent = %d", id)
							}
							return target
						},
						waypointByID: func(id uint32) *Waypoint { return waypoints[id] },
						randomInt: func(_, _ int) int {
							t.Fatal("the original Bookcase6 path has no random branch")
							return 0
						},
						move: func(obj *Object, pos types.Pointf) {
							obj.PosVec, obj.NewPos, obj.PrevPos = pos, pos, pos
						},
						removeUpdatable: func(obj *Object) {
							if obj != source {
								t.Fatal("removed the target instead of its controller")
							}
							removed++
						},
					}
					var visited []uint32
					for tick := 0; tick < 120 && data.Field_0 != 3; tick++ {
						moverUpdateNative54F740(source, deps)
						if wp := source.MoverWaypointFor(data, 3); wp != nil && (len(visited) == 0 || visited[len(visited)-1] != wp.Index) {
							visited = append(visited, wp.Index)
						}
						if data.Field_0 == 1 {
							source.PosVec = source.PosVec.Add(source.VelVec)
						}
						if data.Field_3 != 0 || data.Field_5 != 0 || data.Field_7 != 0 || data.Field_8 != 1974 {
							t.Fatal("native movement corrupted the fixed PE32 record")
						}
					}
					if data.Field_0 != 3 || target.PosVec != positions[2] || !reflect.DeepEqual(visited, []uint32{137, 138, 139}) {
						t.Fatalf("state/position/path = %d/%v/%v; want stopped at %v through [137 138 139]", data.Field_0, target.PosVec, visited, positions[2])
					}
					moverUpdateNative54F740(source, deps)
					if removed != 1 || target.PosVec != positions[2] {
						t.Fatal("terminal controller did not stop without moving the bookcase")
					}
				})
			}
		}
	}
}

func TestMoverUpdate54F740AxisArrivalBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name     string
		velocity types.Pointf
		waypoint types.Pointf
		stopped  bool
	}{
		{"vertical-passed", types.Ptf(0, 1), types.Ptf(10, 19), true},
		{"vertical-arrived", types.Ptf(0, 1), types.Ptf(10, 20), true},
		{"vertical-ahead", types.Ptf(0, 1), types.Ptf(10, 21), false},
		{"vertical-other-axis-misaligned", types.Ptf(0, 1), types.Ptf(9, 19), false},
		{"horizontal-passed", types.Ptf(1, 0), types.Ptf(9, 20), true},
		{"horizontal-arrived", types.Ptf(1, 0), types.Ptf(10, 20), true},
		{"horizontal-ahead", types.Ptf(1, 0), types.Ptf(11, 20), false},
		{"horizontal-other-axis-misaligned", types.Ptf(1, 0), types.Ptf(9, 19), false},
		{"stationary", types.Pointf{}, types.Ptf(10, 19), false},
		{"negative-zero", types.Ptf(math.Float32frombits(0x80000000), 1), types.Ptf(10, 19), true},
		{"nan-x", types.Ptf(float32(math.NaN()), 1), types.Ptf(10, 19), false},
		{"nan-y", types.Ptf(1, float32(math.NaN())), types.Ptf(9, 20), false},
		{"infinite-x", types.Ptf(float32(math.Inf(1)), 0), types.Ptf(9, 20), false},
		{"infinite-y", types.Ptf(0, float32(math.Inf(1))), types.Ptf(10, 19), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, source, target := newMoverUpdateFixture54F740()
			f.afterMove = func(_ int, obj *moverUpdateTestObject54F740, pos types.Pointf) {
				obj.pos = pos
			}
			source.data.state, source.flags, source.speedCur = 1, moverUpdateActiveFlag54F740, 5
			source.pos, source.velocity = types.Ptf(10, 20), tc.velocity
			source.data.waypoints[3] = &moverUpdateTestWaypoint54F740{name: "end", pos: tc.waypoint}
			moverUpdate54F740(source, f.hooks())
			if got := source.data.state == 3; got != tc.stopped {
				t.Fatalf("stopped = %v, want %v", got, tc.stopped)
			}
			if tc.stopped && target.pos != tc.waypoint {
				t.Fatalf("bookcase = %v, want endpoint %v", target.pos, tc.waypoint)
			}
		})
	}
}
