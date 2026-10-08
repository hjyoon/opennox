package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

// Independent precision-53 / ToZero model of 0050D305..0050D329. The
// original retains both differences and squares in x87 registers, without
// a binary32 spill, and tests C0 alone (including unordered comparisons).
func monsterAdvanceWaypointRefArrival50D2E0(pos, target types.Pointf) bool {
	for _, v := range []float32{pos.X, pos.Y, target.X, target.Y} {
		if math.IsNaN(float64(v)) {
			return true
		}
	}
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	input := func(v float32) *big.Float { return chop().SetFloat64(float64(v)) }
	dx := chop().Sub(input(target.X), input(pos.X))
	dy := chop().Sub(input(target.Y), input(pos.Y))
	square := chop().Add(chop().Mul(dy, dy), chop().Mul(dx, dx))
	return square.Cmp(chop().SetInt64(64)) < 0
}

func TestMonsterAdvanceWaypoint50D2E0NativeOriginalPrecision(t *testing.T) {
	// These ten geometries, both coarse-route lengths and four callback
	// outcomes are also checked against the sealed original 206-byte body.
	for _, geometry := range []struct {
		name   string
		pos    [2]uint32
		target [2]uint32
	}{
		{"chop-positive-x", [2]uint32{0x1e3ce508, 0}, [2]uint32{0x41000000, 0}},
		{"chop-negative-x", [2]uint32{0x9e3ce508, 0}, [2]uint32{0xc1000000, 0}},
		{"chop-positive-y", [2]uint32{0, 0x1e3ce508}, [2]uint32{0, 0x41000000}},
		{"chop-negative-y", [2]uint32{0, 0x9e3ce508}, [2]uint32{0, 0xc1000000}},
		{"equal-eight", [2]uint32{}, [2]uint32{0x41000000, 0}},
		{"below-eight", [2]uint32{}, [2]uint32{0x40ffffff, 0}},
		{"above-eight", [2]uint32{}, [2]uint32{0x41000001, 0}},
		{"near-diagonal", [2]uint32{}, [2]uint32{0x40400000, 0x40800000}},
		{"far-diagonal", [2]uint32{}, [2]uint32{0x40c00000, 0x40c00000}},
		{"unordered", [2]uint32{0x7fc12345, 0}, [2]uint32{0x41000000, 0}},
	} {
		for _, last := range []bool{false, true} {
			for _, outcome := range []struct {
				status uint32
				move   bool
			}{{0, false}, {0, true}, {1, true}, {2, true}} {
				t.Run(fmt.Sprintf("%s/last-%t/status-%d/move-%t", geometry.name, last, outcome.status, outcome.move), func(t *testing.T) {
					unit, update := monsterMoveSelectionFixture50D3B0(t)
					first, second := monsterSetMovePathWaypoint50D5A0(t), monsterSetMovePathWaypoint50D5A0(t)
					unit.PosVec = types.Ptf(math.Float32frombits(geometry.pos[0]), math.Float32frombits(geometry.pos[1]))
					first.PosVec = types.Ptf(math.Float32frombits(geometry.target[0]), math.Float32frombits(geometry.target[1]))
					second.PosVec = types.Ptf(40, 50)
					update.Field2, update.Field74, update.Field91 = 0, 2, 0
					if last {
						update.Field74 = 1
					}
					update.Field71 = 0x12340000 | outcome.status
					update.Waypoints[0], update.Waypoints[1] = first, second
					update.Path[1] = types.Ptf(math.Float32frombits(0xdeadbeef), math.Float32frombits(0xbaadf00d))
					update.Path[2] = types.Ptf(123, -456)
					update.AIStack[7].Args = [4]uintptr{0x11223344, 0x55667788, 0x99aabbcc, 0xddeeff00}
					beforeUnit := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
					beforeFirst := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(first), unsafe.Sizeof(*first))
					beforeSecond := monsterMoveSelectionBytes50D3B0(unsafe.Pointer(second), unsafe.Sizeof(*second))
					wantUpdate := *update
					arrived := monsterAdvanceWaypointRefArrival50D2E0(unit.PosVec, first.PosVec)
					wantComplete := arrived && last
					wantPoint := first
					wantCalls := "detail,move"
					if wantComplete {
						wantUpdate.Field74 = 0
						wantCalls = ""
					} else {
						if arrived {
							wantUpdate.Field91 = 1
							wantPoint = second
						}
						wantUpdate.Field2 = 1
						if outcome.status == 0 {
							wantUpdate.Path[1] = wantPoint.PosVec
						}
						wantComplete = outcome.move && outcome.status == 2
					}
					var calls []string
					complete := monsterAdvanceWaypointPath50D2E0(unit, monsterMovePathHooks50D5A0{
						setDetailedPath: func(got *Object, pos *types.Pointf) {
							calls = append(calls, "detail")
							if got != unit || pos != &wantPoint.PosVec || update.Field91 != wantUpdate.Field91 {
								t.Fatalf("detailed path lost full pointer or arrival cursor: unit=%p point=%p cursor=%d want=%p/%p/%d", got, pos, update.Field91, unit, &wantPoint.PosVec, wantUpdate.Field91)
							}
							update.Field2 = 1
						},
						actuallyMove: func(got *Object) bool {
							calls = append(calls, "move")
							if got != unit {
								t.Fatal("movement lost full unit pointer")
							}
							return outcome.move
						},
					})
					if complete != wantComplete || strings.Join(calls, ",") != wantCalls ||
						!bytes.Equal(beforeUnit, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
						!bytes.Equal(beforeFirst, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(first), unsafe.Sizeof(*first))) ||
						!bytes.Equal(beforeSecond, monsterMoveSelectionBytes50D3B0(unsafe.Pointer(second), unsafe.Sizeof(*second))) ||
						!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
						t.Fatalf("arrival=%t complete=%t/%t count=%d cursor=%d detail-count=%d status=%#x calls=%v want=%q; only original named stores are allowed", arrived, complete, wantComplete, update.Field74, update.Field91, update.Field2, update.Field71, calls, wantCalls)
					}
				})
			}
		}
	}
}
