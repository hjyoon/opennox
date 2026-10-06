package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func monsterMoveSelectionFixture50D3B0(t *testing.T) (*Object, *MonsterUpdateData) {
	t.Helper()
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	*unit = Object{ObjClass: object.ClassMonster, SpeedCur: 2, Direction1: 37, Direction2: 91, ForceVec: types.Ptf(7, -9)}
	update, freeUpdate := alloc.New(MonsterUpdateData{})
	t.Cleanup(freeUpdate)
	*update = MonsterUpdateData{Field0: 0x12345678, Field1: 0x89abcdef, Field70: 333, Field71: 444, Field127: 555, StatusFlags: object.MonStatusFrustrated}
	unit.UpdateData = unsafe.Pointer(update)
	for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("path selection fixture below 4 GiB: %p", p)
		}
	}
	debug := memmap.PtrUint32(0x5D4594, 2386204)
	oldDebug := *debug
	*debug = 0xfedcba98
	t.Cleanup(func() { *debug = oldDebug })
	return unit, update
}

func monsterMoveSelectionBytes50D3B0(p unsafe.Pointer, size uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(size))...)
}

// Independent precision-53 / ToZero FLD, FSUB, Y-square, X-square, FADD,
// FCOM and binary32 best-distance FST model of GAME.EXE 0050D42A..0050D48C.
// This reference does not use the production residual/Nextafter helpers.
func monsterMoveSelectionReference50D3B0(pos types.Pointf, path []types.Pointf, start int, clear []bool) (selected int, complete bool) {
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	input := func(v float32) *big.Float { return chop().SetFloat64(float64(v)) }
	selected, closePoint := -1, -1
	best := input(10000000)
	for i := start; i < len(path); i++ {
		if !clear[i] {
			continue
		}
		dx := chop().Sub(input(path[i].X), input(pos.X))
		dy := chop().Sub(input(path[i].Y), input(pos.Y))
		distance := chop().Add(chop().Mul(dy, dy), chop().Mul(dx, dx))
		if distance.Cmp(input(64)) > 0 {
			if selected < 0 || best.Cmp(distance) > 0 {
				selected = i
				if distance.Cmp(input(math.MaxFloat32)) > 0 {
					best = input(math.MaxFloat32)
				} else {
					best = new(big.Float).SetPrec(24).SetMode(big.ToZero).Set(distance)
					if _, accuracy := best.Float32(); accuracy != big.Exact {
						panic("selection reference binary32 spill not exact")
					}
				}
			}
			continue
		}
		if i == len(path)-1 {
			return -1, true
		}
		closePoint = i
	}
	if selected < 0 {
		selected = closePoint
	}
	return selected, false
}

func TestMonsterMoveSelection50D3B0NativeOriginalBranches(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pos     types.Pointf
		path    []types.Pointf
		start   int
		blocked uint32
	}{
		{name: "zero-last", path: []types.Pointf{{}}},
		{name: "equal-eight-last", path: []types.Pointf{{X: 8}}},
		{name: "retained-chop-eight", path: []types.Pointf{{X: math.Float32frombits(0x33c00000), Y: 8}}},
		{name: "retained-chop-eight-negative", path: []types.Pointf{{X: math.Float32frombits(0xb3c00000), Y: -8}}},
		{name: "above-eight", path: []types.Pointf{{X: math.Float32frombits(0x33c00001), Y: math.Float32frombits(0x41000001)}}},
		{name: "equal-far-distance", path: []types.Pointf{{X: 20}, {Y: 20}}},
		{name: "best-distance-chop-tie", path: []types.Pointf{{X: 10, Y: 0.002}, {X: 10, Y: 0.001}}},
		{name: "best-distance-chop-tie-negative", path: []types.Pointf{{X: -10, Y: -0.002}, {X: -10, Y: -0.001}}},
		{name: "closer-far", path: []types.Pointf{{X: 50}, {Y: 20}}},
		{name: "far-preferred-to-close", path: []types.Pointf{{X: 20}, {X: 3}, {X: 100}}, blocked: 4},
		{name: "last-close-fallback", path: []types.Pointf{{X: 2}, {Y: 3}, {X: 100}}, blocked: 4},
		{name: "last-close-completes-before-far", path: []types.Pointf{{X: 20}, {X: 3}}},
		{name: "blocked-last", path: []types.Pointf{{X: 20}, {X: 3}}, blocked: 2},
		{name: "all-blocked", path: []types.Pointf{{X: 20}, {X: 3}}, blocked: 3},
		{name: "start-one", path: []types.Pointf{{X: 9}, {X: 50}, {Y: 20}}, start: 1},
		{name: "max-finite-square", path: []types.Pointf{{X: math.MaxFloat32}, {X: math.MaxFloat32, Y: math.MaxFloat32}}},
		{name: "max-finite-opposite-sign", pos: types.Ptf(-math.MaxFloat32, math.MaxFloat32), path: []types.Pointf{{X: math.MaxFloat32, Y: -math.MaxFloat32}, {X: 0, Y: 0}}},
		{name: "subnormal-last", path: []types.Pointf{{X: math.Float32frombits(1), Y: math.Float32frombits(0x80000001)}}},
		{name: "signed-zero-last", pos: types.Ptf(math.Float32frombits(0x80000000), 0), path: []types.Pointf{{Y: math.Float32frombits(0x80000000)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			unit.PosVec = tc.pos
			update.Field2, update.Field67 = uint32(len(tc.path)), uint32(tc.start)
			copy(update.Path[:], tc.path)
			clear := make([]bool, len(tc.path))
			for i := range clear {
				clear[i] = tc.blocked&(1<<i) == 0
			}
			selected, complete := monsterMoveSelectionReference50D3B0(tc.pos, tc.path, tc.start, clear)
			beforeUnit, beforeUpdate := *unit, *update
			calls := 0
			got := monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
				i := tc.start + calls
				calls++
				if from != tc.pos || to != tc.path[i] || flags != 132 || memmap.Uint32(0x5D4594, 2386204) != 0xfedcba98 {
					t.Fatal("ray trace lost snapshot, order, flags, or debug-store prefix")
				}
				return clear[i]
			})
			wantDebug := uint32(0xfedcba98)
			if selected >= 0 {
				beforeUpdate.Field67 = uint32(selected)
				wantDebug = uint32(selected)
				// Force normalization is intentionally unchanged by this selection
				// restoration. Its separate rounding discrepancy is not certified here.
				beforeUnit.Direction1, beforeUnit.Direction2 = unit.Direction1, unit.Direction2
				beforeUnit.ForceVec = unit.ForceVec
			} else {
				beforeUpdate.Field2 = 0
			}
			if got != complete || memmap.Uint32(0x5D4594, 2386204) != wantDebug || calls != len(tc.path)-tc.start ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) {
				t.Fatalf("selection=%d complete=%t/%t index=%d count=%d debug=%#x/%#x traces=%d", selected, got, complete, update.Field67, update.Field2, memmap.Uint32(0x5D4594, 2386204), wantDebug, calls)
			}
		})
	}
}

func TestMonsterMoveSelection50D3B0NativeTraceReloads(t *testing.T) {
	for _, tc := range []struct {
		name             string
		count, liveCount uint32
		points           [3]types.Pointf
		firstClear       bool
		wantCalls        int
		wantSelected     int
		wantComplete     bool
	}{
		{"grow-far", 1, 2, [3]types.Pointf{{X: 100}, {X: 20}}, true, 2, 1, false},
		{"shrink-far", 3, 1, [3]types.Pointf{{X: 100}, {X: 20}, {X: 30}}, true, 1, 0, false},
		{"shrink-close-completes", 3, 1, [3]types.Pointf{{X: 1}, {X: 20}, {X: 30}}, true, 1, -1, true},
		{"zero-after-far", 3, 0, [3]types.Pointf{{X: 100}, {X: 20}, {X: 30}}, true, 1, 0, false},
		{"signed-negative-count-after-far", 3, 0x80000000, [3]types.Pointf{{X: 100}, {X: 20}, {X: 30}}, true, 1, 0, false},
		{"wrapped-count-after-far", 3, 0xffffffff, [3]types.Pointf{{X: 100}, {X: 20}, {X: 30}}, true, 1, 0, false},
		{"zero-after-close", 3, 0, [3]types.Pointf{{X: 1}, {X: 20}, {X: 30}}, true, 1, 0, false},
		{"shrink-after-blocked", 3, 1, [3]types.Pointf{{X: 100}, {X: 20}, {X: 30}}, false, 1, -1, false},
		{"grow-after-blocked-completes", 1, 3, [3]types.Pointf{{X: 100}, {X: 20}, {X: 1}}, false, 3, -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			update.Field2 = tc.count
			copy(update.Path[:], tc.points[:])
			beforeUnit, beforeUpdate := *unit, *update
			calls := 0
			got := monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
				if calls >= tc.wantCalls {
					t.Fatal("trace did not reload live path count")
				}
				if from != (types.Pointf{}) || to != tc.points[calls] || flags != 132 {
					t.Fatal("unexpected trace argument")
				}
				calls++
				if calls == 1 {
					update.Field2 = tc.liveCount
					return tc.firstClear
				}
				return true
			})
			beforeUpdate.Field2 = tc.liveCount
			wantDebug := uint32(0xfedcba98)
			if tc.wantSelected >= 0 {
				beforeUpdate.Field67 = uint32(tc.wantSelected)
				wantDebug = uint32(tc.wantSelected)
				beforeUnit.Direction1, beforeUnit.Direction2, beforeUnit.ForceVec = unit.Direction1, unit.Direction2, unit.ForceVec
			} else {
				beforeUpdate.Field2 = 0
			}
			if got != tc.wantComplete || calls != tc.wantCalls || memmap.Uint32(0x5D4594, 2386204) != wantDebug ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) {
				t.Fatalf("trace reload complete=%t/%t calls=%d/%d index=%d count=%d debug=%#x/%#x", got, tc.wantComplete, calls, tc.wantCalls, update.Field67, update.Field2, memmap.Uint32(0x5D4594, 2386204), wantDebug)
			}
		})
	}
}

func TestMonsterMoveSelection50D3B0NativeCachedOriginAndUpdate(t *testing.T) {
	for _, replaceUpdate := range []bool{false, true} {
		t.Run(fmt.Sprintf("replace-update-%t", replaceUpdate), func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			replacement, freeReplacement := alloc.New(MonsterUpdateData{})
			t.Cleanup(freeReplacement)
			*replacement = MonsterUpdateData{Field2: 1, Field67: 7, Field70: 999, Path: [32]types.Pointf{{X: -1000, Y: 999}}}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(replacement)) <= math.MaxUint32 {
				t.Fatal("replacement update below 4 GiB")
			}
			update.Field2 = 2
			update.Path[0], update.Path[1] = types.Ptf(100, 0), types.Ptf(120, 0)
			beforeUnit, beforeUpdate, beforeReplacement := *unit, *update, *replacement
			calls := 0
			if monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
				if from != (types.Pointf{}) || to != types.Ptf(100+float32(20*calls), 0) || flags != 132 {
					t.Fatal("trace did not preserve entry origin or point snapshot")
				}
				calls++
				if calls == 1 {
					unit.PosVec = types.Ptf(90, 0)
					update.Path[0] = types.Ptf(200, 0)
					if replaceUpdate {
						unit.UpdateData = unsafe.Pointer(replacement)
					}
				}
				return true
			}) {
				t.Fatal("far path completed")
			}
			// The pre-trace endpoint 100 is ten from the live unit at 90, so
			// slot zero wins. Its live endpoint 200 supplies normalization.
			distance := float32(110 + monsterMoveDistanceBias50D3B0)
			beforeUnit.PosVec = types.Ptf(90, 0)
			beforeUnit.Direction1, beforeUnit.Direction2 = 128, 128
			beforeUnit.ForceVec = types.Ptf(float32(float64(2)*110/float64(distance)), 0)
			if replaceUpdate {
				beforeUnit.UpdateData = unsafe.Pointer(replacement)
			}
			beforeUpdate.Path[0] = types.Ptf(200, 0)
			if calls != 2 || update.Field67 != 0 || memmap.Uint32(0x5D4594, 2386204) != 0 ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeReplacement), unsafe.Sizeof(beforeReplacement)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(replacement), unsafe.Sizeof(*replacement))) {
				t.Fatalf("cached state calls=%d index=%d debug=%#x force=%v", calls, update.Field67, memmap.Uint32(0x5D4594, 2386204), unit.ForceVec)
			}
		})
	}
}

func TestMonsterMoveSelection50D3B0NativeUnorderedCompletion(t *testing.T) {
	for _, word := range []uint32{0x7fc12345, 0xffc12345, 0x7f800001, 0xff800001} {
		t.Run(fmt.Sprintf("word-%08x", word), func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			update.Field2, update.Field67 = 2, 1
			update.Path[1] = types.Ptf(math.Float32frombits(word), 20)
			beforeUnit, beforeUpdate := *unit, *update
			calls := 0
			if !monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
				calls++
				return from == (types.Pointf{}) && math.Float32bits(to.X) == word && to.Y == 20 && flags == 132
			}) {
				t.Fatal("unordered last distance did not follow original completion branch")
			}
			beforeUpdate.Field2 = 0
			if calls != 1 || memmap.Uint32(0x5D4594, 2386204) != 0xfedcba98 ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
				t.Fatal("unordered completion changed force, direction, or cached path bytes")
			}
		})
	}
}

func TestMonsterMoveSelection50D3B0NativeEntryGuards(t *testing.T) {
	for _, tc := range []struct {
		name         string
		count, start uint32
		wantCalls    int
		wantSelected int
	}{
		{"empty-preserves-cursor", 0, 17, 0, -1},
		{"cursor-at-end", 1, 1, 0, -1},
		{"negative-cursor", 2, 0xffffffff, 0, -1},
		{"high-bit-cursor", 2, 0x80000000, 0, -1},
		{"negative-count", 0x80000000, 0, 0, -1},
		{"wrapped-count", 0xffffffff, 0, 0, -1},
		// Counts outside the native array preserve the existing safety bound;
		// equivalence to original out-of-bounds PE32 memory is not claimed.
		{"bounded-invalid-cursor", 33, 32, 0, -1},
		{"bounded-final-native-slot", 33, 31, 1, 31},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			update.Field2, update.Field67 = tc.count, tc.start
			update.Path[31] = types.Ptf(20, 0)
			beforeUnit, beforeUpdate := *unit, *update
			calls := 0
			if monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
				calls++
				if calls > tc.wantCalls || from != (types.Pointf{}) || to != types.Ptf(20, 0) || flags != 132 {
					t.Fatal("invalid path index or signed count reached ray trace")
				}
				return true
			}) {
				t.Fatal("entry guard completed path")
			}
			wantDebug := uint32(0xfedcba98)
			if tc.wantSelected >= 0 {
				beforeUpdate.Field67 = uint32(tc.wantSelected)
				wantDebug = uint32(tc.wantSelected)
				beforeUnit.Direction1, beforeUnit.Direction2, beforeUnit.ForceVec = unit.Direction1, unit.Direction2, unit.ForceVec
			} else {
				beforeUpdate.Field2 = 0
			}
			if calls != tc.wantCalls || memmap.Uint32(0x5D4594, 2386204) != wantDebug ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
				t.Fatal("entry guard changed unrelated state or cursor")
			}
		})
	}
}

func TestMonsterMoveSelection50D3B0NativeIndependentFiniteMatrix(t *testing.T) {
	for _, domain := range []string{"finite-binary32", "map-coordinates", "eight-boundary", "best-distance-spill"} {
		t.Run(domain, func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			word := uint32(0x50d3b001)
			next := func() float32 {
				word = 1664525*word + 1013904223
				bits := word
				if domain == "map-coordinates" {
					bits = word&0x807fffff | (uint32(128)+(word>>23)%12)<<23
				}
				if bits&0x7f800000 == 0x7f800000 {
					bits ^= 0x00800000
				}
				return math.Float32frombits(bits)
			}
			for sample := 0; sample < 1024; sample++ {
				unit.PosVec = types.Pointf{X: next(), Y: next()}
				count := 2 + sample%7
				*update = MonsterUpdateData{Field2: uint32(count), Field67: uint32(sample % count), Field70: 333, Field71: 444, StatusFlags: object.MonStatusFrustrated}
				clear := make([]bool, count)
				for i := 0; i < count; i++ {
					update.Path[i] = types.Pointf{X: next(), Y: next()}
					clear[i] = word&7 != 0
					if domain == "eight-boundary" {
						unit.PosVec = types.Pointf{}
						update.Path[i] = types.Pointf{X: math.Float32frombits(0x33c00000 + word%11), Y: math.Float32frombits(0x41000000 + word%3 - 1)}
					}
					if domain == "best-distance-spill" {
						unit.PosVec = types.Pointf{}
						update.Path[i] = types.Pointf{X: 10, Y: math.Float32frombits(0x3a800000 + word%0x900000)}
					}
				}
				start := int(update.Field67)
				selected, complete := monsterMoveSelectionReference50D3B0(unit.PosVec, update.Path[:count], start, clear)
				beforeUnit, beforeUpdate := *unit, *update
				*memmap.PtrUint32(0x5D4594, 2386204) = 0xfedcba98
				calls := 0
				got := monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
					i := start + calls
					calls++
					if i >= count || from != unit.PosVec || to != update.Path[i] || flags != 132 || memmap.Uint32(0x5D4594, 2386204) != 0xfedcba98 {
						t.Fatal("finite matrix trace arguments or debug prefix changed")
					}
					return clear[i]
				})
				wantDebug := uint32(0xfedcba98)
				if selected >= 0 {
					beforeUpdate.Field67 = uint32(selected)
					wantDebug = uint32(selected)
					beforeUnit.Direction1, beforeUnit.Direction2, beforeUnit.ForceVec = unit.Direction1, unit.Direction2, unit.ForceVec
				} else {
					beforeUpdate.Field2 = 0
				}
				if got != complete || calls != count-start || memmap.Uint32(0x5D4594, 2386204) != wantDebug ||
					!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
					!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)), monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
					t.Fatalf("sample %d selection=%d/%d complete=%t/%t count=%d traces=%d debug=%#x/%#x", sample, update.Field67, selected, got, complete, update.Field2, calls, memmap.Uint32(0x5D4594, 2386204), wantDebug)
				}
			}
		})
	}
}
