package server

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Independent arithmetic reference for the original gameplay rounding mode.
// Atan2 is the same cross-platform transcendental approximation as the port;
// the add/multiply, binary32 spill and integer truncation use math/big rather
// than production residual helpers. Literal original-byte cases below also
// cover the actual observed angle, without calling this reference.
func monsterMoveDirectionReference50D3B0(segment types.Pointf) Dir16 {
	angle := math.Atan2(float64(segment.Y), float64(segment.X))
	if math.IsNaN(angle) {
		return 0
	}
	chop := func() *big.Float { return new(big.Float).SetPrec(53).SetMode(big.ToZero) }
	input := func(v float64) *big.Float { return chop().SetFloat64(v) }
	value := chop().Add(input(angle), input(float64(math.Float32frombits(0x40c90fdb))))
	value = chop().Mul(value, input(float64(math.Float32frombits(0x4222f983))))
	value = chop().Add(value, input(0.5))
	stored := input(float64(monsterMoveForceRefSpill50D3B0(value)))
	integer, _ := stored.Int64()
	return Dir16(uint8(integer))
}

// Inputs are the 45 unchanged headless Troll/RedApple movement calls at
// a398c9b70. Expected directions, force words, cursors, and false returns
// come from GAME.EXE 0050D3B0..0050D59D, including 00509ED0 and 00419A70,
// with FPCW=0x0e7f. Original SHA-256:
// 0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a.
// This fixes selected arithmetic evidence, not the unresolved food timeout.
func TestMonsterMoveDirection50D3B0NativeCapturedOriginal(t *testing.T) {
	pathBits := [7][2]uint32{
		{0x454edb33, 0x450b4b33},
		{0x45504b33, 0x450cbb33},
		{0x4551bb33, 0x450e2b33},
		{0x45532b33, 0x450f9b33},
		{0x45549b33, 0x45107800},
		{0x45560b33, 0x45107800},
		{0x45564000, 0x45104000},
	}
	cases := []struct {
		frame             uint32
		pos               [2]uint32
		start, wantCursor uint32
		wantDirection     Dir16
		wantForce         [2]uint32
	}{
		{594, [2]uint32{0x454e4000, 0x450a4000}, 0, 0, 32, [2]uint32{0x3f842a4b, 0x3fe38aeb}},
		{595, [2]uint32{0x454e48af, 0x450a4ef4}, 0, 0, 32, [2]uint32{0x3f84298e, 0x3fe388f4}},
		{596, [2]uint32{0x454e5b58, 0x450a6f14}, 0, 0, 32, [2]uint32{0x3f8426be, 0x3fe3845c}},
		{597, [2]uint32{0x454e73e5, 0x450a9957}, 0, 0, 32, [2]uint32{0x3f8420db, 0x3fe37c27}},
		{598, [2]uint32{0x454e8feb, 0x450ac997}, 0, 0, 32, [2]uint32{0x3f84177c, 0x3fe36b10}},
		{599, [2]uint32{0x454eadfd, 0x450afd5d}, 0, 1, 32, [2]uint32{0x3fb2eb41, 0x3fc10ba4}},
		{600, [2]uint32{0x454ed059, 0x450b30f6}, 1, 1, 32, [2]uint32{0x3fb65edd, 0x3fbdc703}},
		{601, [2]uint32{0x454ef72e, 0x450b62fa}, 1, 1, 32, [2]uint32{0x3fb8f502, 0x3fbb3e4f}},
		{602, [2]uint32{0x454f20f3, 0x450b93c5}, 1, 1, 32, [2]uint32{0x3fbaf9ef, 0x3fb9354d}},
		{603, [2]uint32{0x454f4caf, 0x450bc39e}, 1, 1, 32, [2]uint32{0x3fbca212, 0x3fb77e8a}},
		{604, [2]uint32{0x454f79c2, 0x450bf2b9}, 1, 1, 32, [2]uint32{0x3fbe1618, 0x3fb5f32e}},
		{605, [2]uint32{0x454fa7c8, 0x450c2138}, 1, 1, 32, [2]uint32{0x3fbf7ac5, 0x3fb46be7}},
		{606, [2]uint32{0x454fd683, 0x450c4f33}, 1, 1, 32, [2]uint32{0x3fc10653, 0x3fb2a761}},
		{607, [2]uint32{0x455005cf, 0x450c7cb5}, 1, 2, 32, [2]uint32{0x3fbb9717, 0x3fb8a1ae}},
		{608, [2]uint32{0x45503524, 0x450caa44}, 2, 2, 32, [2]uint32{0x3fbb566c, 0x3fb8e04c}},
		{609, [2]uint32{0x45506447, 0x450cd815}, 2, 2, 32, [2]uint32{0x3fbb2388, 0x3fb91000}},
		{610, [2]uint32{0x45509347, 0x450d0614}, 2, 2, 32, [2]uint32{0x3fbaf9fd, 0x3fb934f5}},
		{611, [2]uint32{0x4550c22d, 0x450d3431}, 2, 2, 32, [2]uint32{0x3fbad5c2, 0x3fb9529e}},
		{612, [2]uint32{0x4550f101, 0x450d6264}, 2, 2, 32, [2]uint32{0x3fbab32f, 0x3fb96b63}},
		{613, [2]uint32{0x45511fc7, 0x450d90a6}, 2, 2, 32, [2]uint32{0x3fba8d31, 0x3fb98187}},
		{614, [2]uint32{0x45514e80, 0x450dbef4}, 2, 2, 32, [2]uint32{0x3fba5bfe, 0x3fb9951d}},
		{615, [2]uint32{0x45517d2d, 0x450ded4b}, 2, 3, 32, [2]uint32{0x3fba2444, 0x3fba1748}},
		{616, [2]uint32{0x4551abcd, 0x450e1bb1}, 3, 3, 32, [2]uint32{0x3fba1576, 0x3fba230d}},
		{617, [2]uint32{0x4551da62, 0x450e4a24}, 3, 3, 32, [2]uint32{0x3fba0931, 0x3fba2b6f}},
		{618, [2]uint32{0x455208f0, 0x450e78a1}, 3, 3, 32, [2]uint32{0x3fb9fe70, 0x3fba310f}},
		{619, [2]uint32{0x45523779, 0x450ea724}, 3, 3, 32, [2]uint32{0x3fb9f3c7, 0x3fba34a1}},
		{620, [2]uint32{0x455265fd, 0x450ed5aa}, 3, 3, 32, [2]uint32{0x3fb9e7df, 0x3fba361d}},
		{621, [2]uint32{0x4552947e, 0x450f0435}, 3, 3, 32, [2]uint32{0x3fb9d990, 0x3fba3396}},
		{622, [2]uint32{0x4552c2fa, 0x450f32c0}, 3, 3, 32, [2]uint32{0x3fb9c2f9, 0x3fba2a5a}},
		{623, [2]uint32{0x4552f173, 0x450f614b}, 3, 4, 22, [2]uint32{0x3fdc3510, 0x3f902746}},
		{624, [2]uint32{0x4553222c, 0x450f8d13}, 4, 4, 22, [2]uint32{0x3fdf5e75, 0x3f8b2e65}},
		{625, [2]uint32{0x455355b1, 0x450fb55c}, 4, 4, 22, [2]uint32{0x3fe1ddcd, 0x3f870f1e}},
		{626, [2]uint32{0x45538b27, 0x450fdb20}, 4, 4, 22, [2]uint32{0x3fe3f551, 0x3f837391}},
		{627, [2]uint32{0x4553c1fa, 0x450fff05}, 4, 4, 22, [2]uint32{0x3fe5dccb, 0x3f800526}},
		{628, [2]uint32{0x4553f9d0, 0x45102172}, 4, 4, 22, [2]uint32{0x3fe7d01b, 0x3f78a6af}},
		{629, [2]uint32{0x45543272, 0x451042a3}, 4, 5, 0, [2]uint32{0x4002c399, 0x3e6c2a46}},
		{630, [2]uint32{0x45546d92, 0x45105cbd}, 5, 5, 0, [2]uint32{0x40034d95, 0x3e0a76cf}},
		{631, [2]uint32{0x4554ab54, 0x45106e5e}, 5, 5, 0, [2]uint32{0x40038797, 0x3d6672b7}},
		{632, [2]uint32{0x4554e601, 0x45107a90}, 5, 5, 0, [2]uint32{0x40038fda, 0xbc932ded}},
		{633, [2]uint32{0x45550be4, 0x45108606}, 5, 5, 0, [2]uint32{0x40035bc0, 0xbde6e2b9}},
		{634, [2]uint32{0x45552ad3, 0x45108faa}, 5, 5, 0, [2]uint32{0x4002d1e4, 0xbe5cc107}},
		{635, [2]uint32{0x4555551c, 0x451093b1}, 5, 5, 0, [2]uint32{0x4002078b, 0xbe9e31c4}},
		{636, [2]uint32{0x4555859f, 0x4510936a}, 5, 5, 0, [2]uint32{0x4000cccb, 0xbed377b4}},
		{637, [2]uint32{0x4555b06a, 0x4510925d}, 5, 6, 223, [2]uint32{0x3fe4271b, 0xbf82df49}},
		{638, [2]uint32{0x4555d491, 0x45108bff}, 6, 6, 223, [2]uint32{0x3fd6ad3d, 0xbf97db82}},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("frame-%d", tc.frame), func(t *testing.T) {
			unit, update := monsterMoveSelectionFixture50D3B0(t)
			def, freeDef := alloc.New(MonsterDef{})
			t.Cleanup(freeDef)
			*def = MonsterDef{RunMultiplier96: math.Float32frombits(0x3f800000)}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(def)) <= math.MaxUint32 {
				t.Fatal("native monster definition below 4 GiB")
			}
			unit.PosVec = types.Ptf(math.Float32frombits(tc.pos[0]), math.Float32frombits(tc.pos[1]))
			unit.SpeedCur = math.Float32frombits(0x4003a384)
			update.Field2, update.Field67 = 7, tc.start
			update.StatusFlags, update.MonsterDef = 0x10000, def
			for i, bits := range pathBits {
				update.Path[i] = types.Ptf(math.Float32frombits(bits[0]), math.Float32frombits(bits[1]))
			}
			beforeUnit, beforeUpdate, beforeDef := *unit, *update, *def
			beforeUnit.Direction1, beforeUnit.Direction2 = tc.wantDirection, tc.wantDirection
			beforeUnit.ForceVec = types.Ptf(math.Float32frombits(tc.wantForce[0]), math.Float32frombits(tc.wantForce[1]))
			beforeUpdate.Field67 = tc.wantCursor
			calls := 0
			complete := monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
				i := int(tc.start) + calls
				calls++
				if i >= len(pathBits) || from != unit.PosVec || to != update.Path[i] || flags != 132 ||
					memmap.Uint32(0x5D4594, 2386204) != 0xfedcba98 {
					t.Fatal("captured ray arguments, order, or debug-store prefix changed")
				}
				return true
			})
			if complete || calls != 7-int(tc.start) || *def != beforeDef ||
				memmap.Uint32(0x5D4594, 2386204) != tc.wantCursor ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)),
					monsterMoveSelectionBytes50D3B0(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
				!bytes.Equal(monsterMoveSelectionBytes50D3B0(unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)),
					monsterMoveSelectionBytes50D3B0(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
				t.Fatalf("direction=%d/%d want=%d, force=(%08x,%08x) want=(%08x,%08x), cursor=%d want=%d, complete=%t, rays=%d",
					unit.Direction1, unit.Direction2, tc.wantDirection,
					math.Float32bits(unit.ForceVec.X), math.Float32bits(unit.ForceVec.Y),
					tc.wantForce[0], tc.wantForce[1], update.Field67, tc.wantCursor, complete, calls)
			}
		})
	}
}
