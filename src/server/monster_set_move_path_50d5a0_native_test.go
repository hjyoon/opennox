package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Independent FLD/FSUB, retained Y-square, X-square and FADD reference for
// GAME.EXE 0050D5BA..0050D5EB / 0050D660..0050D67B / 0050D6BD..0050D6D8.
// Inputs are binary32, but no binary32 spill occurs before these comparisons.
// The original CRT/game loop select precision 53 and round toward zero.
func monsterSetMovePathRefSquare50D5A0(from, to types.Pointf) *big.Float {
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	input := func(v float32) *big.Float { return chop().SetFloat64(float64(v)) }
	dx := chop().Sub(input(from.X), input(to.X))
	dy := chop().Sub(input(from.Y), input(to.Y))
	return chop().Add(chop().Mul(dy, dy), chop().Mul(dx, dx))
}

func monsterSetMovePathRefArrival50D5A0(from, to types.Pointf) bool {
	square := monsterSetMovePathRefSquare50D5A0(from, to)
	// Reuse only the independent integer/rational square-root reference, not
	// the production math.Sqrt/residual or any production arithmetic helper.
	root := monsterMoveForceRefSqrt50D3B0(square)
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	bias := chop().SetFloat64(math.Float64frombits(0x3f847ae140000000))
	return chop().Add(root, bias).Cmp(chop().SetInt64(8)) <= 0
}

func monsterSetMovePathAssertRefresh50D5A0(t *testing.T, unit *Object, update *MonsterUpdateData, waypoint *Waypoint, last, target types.Pointf, direct bool, frame, previous uint32) {
	t.Helper()
	*update = MonsterUpdateData{Field0: 0x12345678, Field1: 0x89abcdef,
		Field67: 7, Field70: previous, Field71: 0x12340000, Field127: 0x76543210}
	unit.PosVec = types.Ptf(1000, 1000)
	threshold := int64(2500)
	if direct {
		update.Field2, update.Field68 = 1, last
	} else {
		threshold = 10000
		update.Field74, update.Waypoints[0] = 1, waypoint
		update.Field92, update.Field93 = math.Float32bits(last.X), math.Float32bits(last.Y)
	}
	wantRefresh := frame-previous > 10 && monsterSetMovePathRefSquare50D5A0(last, target).Cmp(new(big.Float).SetInt64(threshold)) > 0
	beforeUnit := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	wantUpdate := *update
	wantWaypoint := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(waypoint), unsafe.Sizeof(*waypoint))
	if !direct {
		if wantRefresh {
			wantUpdate.Field74, wantUpdate.Field91 = 0, 0
			wantUpdate.Field92, wantUpdate.Field93 = math.Float32bits(target.X), math.Float32bits(target.Y)
		} else {
			// The unchanged 0050D2E0 appends its detailed-path target when
			// the low status byte is zero, even if its hook leaves count zero.
			wantUpdate.Path[0] = waypoint.PosVec
		}
	}
	traces, finds, detailed, moves, statusCalls := 0, 0, 0, 0, 0
	complete := monsterCreatureSetMovePath50D5A0(unit, target, monsterMovePathHooks50D5A0{
		frame: func() uint32 { return frame },
		trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
			traces++
			if from != unit.PosVec || to != target || flags != 0 {
				t.Fatalf("trace changed: %v -> %v flags=%d", from, to, flags)
			}
			return direct
		},
		findWaypoint: func(got *Object, pos *types.Pointf) *Waypoint {
			finds++
			want := unit.PosVec
			if finds == 2 {
				want = target
			}
			if got != unit || *pos != want {
				t.Fatalf("coarse search order: #%d unit=%p pos=%v want=%v", finds, got, *pos, want)
			}
			return nil
		},
		setPathStatus: func(monsterWaypointPathStatus547F70) { statusCalls++ },
		setDetailedPath: func(got *Object, pos *types.Pointf) {
			detailed++
			want := target
			if !direct && !wantRefresh {
				want = waypoint.PosVec
				if pos != &waypoint.PosVec {
					t.Fatal("existing waypoint target pointer replaced")
				}
			}
			if got != unit || *pos != want {
				t.Fatalf("detailed target: unit=%p pos=%v want=%v", got, *pos, want)
			}
		},
		actuallyMove: func(got *Object) bool {
			moves++
			if got != unit {
				t.Fatal("movement unit replaced")
			}
			return false
		},
	})
	wantFinds, wantDetailed, wantMoves := 0, 1, 1
	if direct && !wantRefresh {
		wantDetailed = 0
	}
	if !direct && wantRefresh {
		wantFinds, wantMoves = 2, 0
	}
	if complete || traces != 1 || finds != wantFinds || detailed != wantDetailed || moves != wantMoves || statusCalls != 0 ||
		!bytes.Equal(beforeUnit, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
		!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) ||
		!bytes.Equal(wantWaypoint, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(waypoint), unsafe.Sizeof(*waypoint))) {
		t.Fatalf("refresh direct=%t last=%v target=%v age=%d: complete=%t trace/find/detail/move/status=%d/%d/%d/%d/%d want=1/%d/%d/%d/0 refresh=%t square=%s",
			direct, last, target, frame-previous, complete, traces, finds, detailed, moves, statusCalls, wantFinds, wantDetailed, wantMoves, wantRefresh, monsterSetMovePathRefSquare50D5A0(last, target).Text('g', 20))
	}
}

func monsterSetMovePathWaypoint50D5A0(t *testing.T) *Waypoint {
	t.Helper()
	waypoint, free := alloc.New(Waypoint{})
	t.Cleanup(free)
	*waypoint = Waypoint{PosVec: types.Ptf(400, 600), Flags: 1, Flags2: 0x80}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(waypoint)) <= math.MaxUint32 {
		t.Fatalf("native waypoint below 4 GiB: %p", waypoint)
	}
	return waypoint
}

func TestMonsterSetMovePath50D5A0NativeRefreshArithmetic(t *testing.T) {
	for _, direct := range []bool{true, false} {
		limit, tiny := float32(50), float32(0.0000006)
		if !direct {
			limit, tiny = 100, 0.000001
		}
		for _, tc := range []struct {
			name            string
			last            types.Pointf
			frame, previous uint32
		}{
			{"retained-chop-positive", types.Ptf(limit, tiny), 111, 100},
			{"retained-chop-negative", types.Ptf(-limit, -tiny), 111, 100},
			{"retained-chop-swapped", types.Ptf(tiny, limit), 111, 100},
			{"equal-threshold", types.Ptf(limit, 0), 111, 100},
			{"below-threshold", types.Ptf(math.Nextafter32(limit, 0), tiny), 111, 100},
			{"above-threshold", types.Ptf(math.Nextafter32(limit, math.MaxFloat32), tiny), 111, 100},
			{"ordered-greater", types.Ptf(limit, 0.001), 111, 100},
			{"age-ten", types.Ptf(limit+1, tiny), 110, 100},
			{"wrap-age-ten", types.Ptf(limit+1, tiny), 4, math.MaxUint32 - 5},
			{"wrap-age-eleven", types.Ptf(limit+1, tiny), 5, math.MaxUint32 - 5},
			{"finite-largest", types.Ptf(math.MaxFloat32, -math.MaxFloat32), 111, 100},
			{"subnormal-gap", types.Ptf(math.Float32frombits(1), math.Float32frombits(0x80000001)), 111, 100},
		} {
			t.Run(fmt.Sprintf("direct-%t/%s", direct, tc.name), func(t *testing.T) {
				unit, update := monsterMoveSelectionFixture50D3B0(t)
				monsterSetMovePathAssertRefresh50D5A0(t, unit, update, monsterSetMovePathWaypoint50D5A0(t), tc.last, types.Pointf{}, direct, tc.frame, tc.previous)
			})
		}
	}
}

func TestMonsterSetMovePath50D5A0NativeArrivalArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name        string
		pos, target types.Pointf
	}{
		{"zero", types.Pointf{}, types.Pointf{}},
		{"signed-zero", types.Ptf(math.Float32frombits(0x80000000), 0), types.Ptf(0, math.Float32frombits(0x80000000))},
		{"subnormal", types.Pointf{}, types.Ptf(math.Float32frombits(1), math.Float32frombits(0x80000001))},
		{"below-biased-eight", types.Pointf{}, types.Ptf(7.98, 0)},
		{"near-biased-eight", types.Pointf{}, types.Ptf(7.99, 0)},
		{"above-biased-eight", types.Pointf{}, types.Ptf(math.Nextafter32(7.99, 8), 0)},
		{"negative-biased-eight", types.Pointf{}, types.Ptf(-7.99, 0)},
		{"equal-eight", types.Pointf{}, types.Ptf(0, 8)},
		{"retained-sqrt", types.Pointf{}, types.Ptf(5.65, 5.65)},
		{"map-coordinates", types.Ptf(2300, 1400), types.Ptf(2305, 1406)},
		{"large-cancellation", types.Ptf(math.MaxFloat32, -math.MaxFloat32), types.Ptf(math.MaxFloat32, -math.MaxFloat32)},
		{"large-finite-difference", types.Ptf(-math.MaxFloat32, math.MaxFloat32), types.Ptf(math.MaxFloat32, -math.MaxFloat32)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			unit.PosVec = tc.pos
			update.Field2, update.Field74 = 1, 0
			beforeUnit := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
			beforeUpdate := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))
			traces, moves := 0, 0
			want := monsterSetMovePathRefArrival50D5A0(tc.target, tc.pos)
			got := monsterCreatureSetMovePath50D5A0(unit, tc.target, monsterMovePathHooks50D5A0{
				frame: func() uint32 { return update.Field70 },
				trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
					traces++
					if from != tc.pos || to != tc.target || flags != 0 {
						t.Fatal("arrival arithmetic changed the trace snapshot")
					}
					return true
				},
				setDetailedPath: func(*Object, *types.Pointf) { t.Fatal("fresh existing path rebuilt") },
				actuallyMove:    func(*Object) bool { moves++; return false },
			})
			wantCalls := 1
			if want {
				wantCalls = 0
			}
			if got != want || traces != wantCalls || moves != wantCalls ||
				!bytes.Equal(beforeUnit, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
				!bytes.Equal(beforeUpdate, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
				t.Fatalf("arrival=%t want=%t trace/move=%d/%d want=%d; records must remain unchanged", got, want, traces, moves, wantCalls)
			}
		})
	}
}

func TestMonsterSetMovePath50D5A0NativeUnorderedArrival(t *testing.T) {
	for _, bits := range []uint32{0x7fc12345, 0xffc12345, 0x7f812345, 0xff812345} {
		t.Run(fmt.Sprintf("%08x", bits), func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			unit.PosVec = types.Ptf(math.Float32frombits(bits), 0)
			beforeUnit := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
			beforeUpdate := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))
			if !monsterCreatureSetMovePath50D5A0(unit, types.Ptf(100, 0), monsterMovePathHooks50D5A0{
				frame: func() uint32 { t.Fatal("unordered arrival called frame"); return 0 },
				trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
					t.Fatal("unordered arrival called trace")
					return false
				},
			}) || !bytes.Equal(beforeUnit, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
				!bytes.Equal(beforeUpdate, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
				t.Fatal("FCOM C0|C3 unordered arrival changed")
			}
		})
	}
}

func TestMonsterSetMovePath50D5A0NativeIndependentRefreshMatrix(t *testing.T) {
	for _, direct := range []bool{true, false} {
		for _, boundary := range []bool{true, false} {
			t.Run(fmt.Sprintf("direct-%t/boundary-%t", direct, boundary), func(t *testing.T) {
				unit, update := monsterMoveSelectionFixture50D3B0(t)
				waypoint := monsterSetMovePathWaypoint50D5A0(t)
				word := uint32(0x50d5a001)
				next := func() float32 {
					word = 1664525*word + 1013904223
					bits := word
					if bits&0x7f800000 == 0x7f800000 {
						bits &^= 0x00800000
					}
					return math.Float32frombits(bits)
				}
				for i := 0; i < 512; i++ {
					last, target := types.Ptf(next(), next()), types.Ptf(next(), next())
					if boundary {
						limit, tiny := float32(50), float32(0.0000006)
						if !direct {
							limit, tiny = 100, 0.000001
						}
						last, target = types.Ptf(limit, tiny), types.Pointf{}
						if i%4 == 1 {
							last.X = math.Nextafter32(limit, 0)
						} else if i%4 == 2 {
							last.X = math.Nextafter32(limit, math.MaxFloat32)
						} else if i%4 == 3 {
							last.Y *= 2
						}
						if i&4 != 0 {
							last.X, last.Y = -last.X, -last.Y
						}
						if i&8 != 0 {
							last.X, last.Y = last.Y, last.X
						}
					}
					monsterSetMovePathAssertRefresh50D5A0(t, unit, update, waypoint, last, target, direct, 5, math.MaxUint32-5)
				}
			})
		}
	}
}

func TestMonsterSetMovePath50D5A0ReferenceCalibration(t *testing.T) {
	for _, tc := range []struct {
		name       string
		pos        types.Pointf
		wantSquare int64
		nearestUp  bool
	}{
		{"direct-boundary", types.Ptf(50, 0.0000006), 2500, true},
		{"coarse-boundary", types.Ptf(100, 0.000001), 10000, true},
		{"three-four-five", types.Ptf(3, 4), 25, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := monsterSetMovePathRefSquare50D5A0(tc.pos, types.Pointf{})
			x, y := float64(tc.pos.X), float64(tc.pos.Y)
			if got.Cmp(new(big.Float).SetInt64(tc.wantSquare)) != 0 || (x*x+y*y > float64(tc.wantSquare)) != tc.nearestUp {
				t.Fatalf("independent reference=%s want=%d, nearest=%g", got.Text('g', 20), tc.wantSquare, x*x+y*y)
			}
		})
	}
}
