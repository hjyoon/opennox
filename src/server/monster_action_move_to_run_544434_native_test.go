package server

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// GAME.EXE 00544434 spills only Field329*3 for the near squared threshold.
// Its far threshold retains the unspilled product plus 30. Both FCOMPPs
// compare retained centre-distance squares, not a binary32 square root.
// TEST C0 first stops on less or unordered; TEST C0|C3 then starts only
// on ordered greater. The original run helpers honor AlwaysRun/NeverRun.
func TestMonsterActionMoveToRun544434NativeOriginalBands(t *testing.T) {
	unit, freeUnit := alloc.New(Object{ObjClass: object.ClassMonster, SpeedBase: 2})
	t.Cleanup(freeUnit)
	*unit = Object{ObjClass: object.ClassMonster, SpeedBase: 2}
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	owner, freeOwner := alloc.New(Object{ObjClass: object.ClassPlayer})
	t.Cleanup(freeOwner)
	owner.ObjClass = object.ClassPlayer
	unit.UpdateData = unsafe.Pointer(update)
	for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(owner)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native MOVE_TO fixture below 4 GiB: %p", p)
		}
	}
	snapshot := func(p unsafe.Pointer, size uintptr) []byte {
		return append([]byte(nil), unsafe.Slice((*byte)(p), int(size))...)
	}
	for _, tc := range []struct {
		name                          string
		unitX, unitY, x, y, rangeWord uint32
		band                          int
	}{
		{name: "near-zero", rangeWord: 0x41200000, band: -1},
		{name: "near-integer", x: 0x41a00000, rangeWord: 0x41200000, band: -1},
		{name: "equal-near", x: 0x41f00000, rangeWord: 0x41200000},
		{name: "below-near-hidden-by-f32-sqrt", x: 0x41efffff, y: 0x3c23d70a, rangeWord: 0x41200000, band: -1},
		{name: "retained-diagonal-above-near", x: 0x41f00000, y: 0x39000000, rangeWord: 0x41200000},
		{name: "middle-band", x: 0x42200000, rangeWord: 0x41200000},
		{name: "equal-far", x: 0x42700000, rangeWord: 0x41200000},
		{name: "retained-diagonal-above-far", x: 0x42700000, y: 0x39000000, rangeWord: 0x41200000, band: 1},
		{name: "far-integer", x: 0x42800000, rangeWord: 0x41200000, band: 1},
		{name: "chop-near-product-spill", x: 0x41f00001, rangeWord: 0x41200001},
		{name: "unspilled-far-product", x: 0x42700001, rangeWord: 0x41200001, band: 1},
		{name: "negative-near-squared", x: 0x41200000, rangeWord: 0xc1200000, band: -1},
		{name: "negative-far-squared", x: 0x42200000, rangeWord: 0xc1200000, band: 1},
		{name: "negative-large-near", x: 0x42200000, rangeWord: 0xc1a00000, band: -1},
		{name: "equal-negative-near", x: 0x42700000, rangeWord: 0xc1a00000, band: 1},
		{name: "zero-range-equal-near"},
		{name: "zero-range-equal-far", x: 0x41f00000},
		{name: "zero-range-above-far", x: 0x41f00000, y: 0x39000000, band: 1},
		{name: "signed-zero", unitX: 0x80000000, y: 0x80000000, rangeWord: 0x80000000},
		{name: "smallest-subnormal-near", rangeWord: 1, band: -1},
		{name: "equal-smallest-subnormal-near", x: 3, rangeWord: 1},
		{name: "finite-square-not-f32-overflow", x: 0x7f7fffff, rangeWord: 0x7f800000, band: -1},
		{name: "near-overflow-spill-chops-to-max-finite", x: 0x7f7fffff, rangeWord: 0x7f7fffff},
		{name: "negative-overflow-near-square", x: 0x3f800000, rangeWord: 0xff7fffff, band: -1},
		{name: "positive-infinite-target", x: 0x7f800000, rangeWord: 0x41200000, band: 1},
		{name: "negative-infinite-target", x: 0xff800000, rangeWord: 0x41200000, band: 1},
		{name: "equal-infinite-squares", x: 0xff800000, rangeWord: 0x7f800000},
		{name: "negative-infinite-range", x: 0x3f800000, rangeWord: 0xff800000, band: -1},
		{name: "target-qnan", x: 0x7fc12345, rangeWord: 0x41200000, band: -1},
		{name: "target-negative-snan", y: 0xff800001, rangeWord: 0x41200000, band: -1},
		{name: "unit-qnan", unitY: 0xffc12345, rangeWord: 0x41200000, band: -1},
		{name: "range-qnan", x: 0x3f800000, rangeWord: 0x7fc12345, band: -1},
		{name: "range-negative-snan", x: 0x3f800000, rangeWord: 0xff800001, band: -1},
		{name: "infinity-minus-infinity", unitX: 0x7f800000, x: 0x7f800000, rangeWord: 0x41200000, band: -1},
	} {
		for _, action := range []ai.ActionType{ai.ACTION_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_FIGHT} {
			for _, conditions := range []int{0, 2} {
				for flags := 0; flags < 8; flags++ {
					t.Run(fmt.Sprintf("%s/%s/conditions-%d/flags-%d", tc.name, action, conditions, flags), func(t *testing.T) {
						status := object.MonStatusFrustrated | object.MonStatusCanSeeFriends
						if flags&1 != 0 {
							status |= object.MonStatusRunning
						}
						if flags&2 != 0 {
							status |= object.MonStatusAlwaysRun
						}
						if flags&4 != 0 {
							status |= object.MonStatusNeverRun
						}
						*update = MonsterUpdateData{Field329: math.Float32frombits(tc.rangeWord), StatusFlags: status, AIStackInd: int8(conditions + 1), Field97: 91, Field101: 92, CurrentEnemy: owner}
						update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_ESCORT), Args: [4]uintptr{uintptr(unsafe.Pointer(owner)), 0x12345678, 0x89abcdef, 0xff800001}, Field5: 0xfedcba98}
						for i := 1; i <= conditions; i++ {
							update.AIStack[i].Action = uint32(ai.DEPENDENCY_TIME)
						}
						head := update.AIStackHead()
						head.Action, head.Field5 = uint32(action), 0x12345678
						target := types.Pointf{X: math.Float32frombits(tc.x), Y: math.Float32frombits(tc.y)}
						head.SetArgs(target, owner, uint32(0xfedcba98))
						unit.PosVec = types.Pointf{X: math.Float32frombits(tc.unitX), Y: math.Float32frombits(tc.unitY)}
						beforeUnit := snapshot(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
						beforeUpdate := *update
						want := status
						if tc.band < 0 && !status.Has(object.MonStatusAlwaysRun) {
							want &^= object.MonStatusRunning
						}
						if tc.band > 0 && !status.Has(object.MonStatusNeverRun) {
							want |= object.MonStatusRunning
						}
						pathCalls, audioCalls := 0, 0
						hooks := monsterActionMoveToHooks5443F0{
							frame: func() uint32 { return 100 }, tickRate: func() uint32 { return 30 },
							random: func(int, int) int { t.Fatal("run-band decision called random"); return 0 },
							setMovePath: func(got *Object, pos types.Pointf) bool {
								pathCalls++
								if got != unit || math.Float32bits(pos.X) != tc.x || math.Float32bits(pos.Y) != tc.y {
									t.Fatal("path callback lost native unit or raw target")
								}
								if update.StatusFlags != want {
									t.Fatalf("run band before path=%#x want=%#x (band %d)", uint32(update.StatusFlags), uint32(want), tc.band)
								}
								return false
							},
							moveAudio: func(got *Object) {
								audioCalls++
								if got != unit {
									t.Fatal("audio callback lost native unit")
								}
							},
							push: func(ai.ActionType, ...any) *AIStackItem { t.Fatal("run-band decision pushed action"); return nil },
							pop:  func() int { t.Fatal("run-band decision popped action"); return 0 },
						}
						if !monsterActionMoveToForAction5443F0(unit, action, hooks) || pathCalls != 1 || audioCalls != 1 {
							t.Fatalf("MOVE_TO path/audio calls=%d/%d", pathCalls, audioCalls)
						}
						beforeUpdate.StatusFlags = want
						if !bytes.Equal(beforeUnit, snapshot(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
							!bytes.Equal(snapshot(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), snapshot(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
							t.Fatal("run-band decision changed raw inputs, stack, cached update or native identity")
						}
					})
				}
			}
		}
	}
}
