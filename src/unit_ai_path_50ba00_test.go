package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

func TestAIPathGridCell50BA00UsesBinary32X87Conversion(t *testing.T) {
	s := &Server{Server: new(server.Server)}
	tests := []struct {
		name string
		pos  types.Pointf
		want ntype.Point32
	}{
		{name: "positive ties", pos: types.Ptf(34.5, 57.5), want: ntype.Point32{X: 2, Y: 2}},
		{name: "negative ties", pos: types.Ptf(-34.5, -57.5), want: ntype.Point32{X: -2, Y: -2}},
		{name: "invalid", pos: types.Ptf(float32(math.NaN()), float32(math.Inf(1))), want: ntype.Point32{X: math.MinInt32, Y: math.MinInt32}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := s.aiPathGridCell50BA00(&test.pos)
			if got.X != test.want.X || got.Y != test.want.Y {
				t.Fatalf("grid(%v) = (%d, %d), want (%d, %d)", test.pos, got.X, got.Y, test.want.X, test.want.Y)
			}
		})
	}
}

func TestAIPathDoorDirectionMask50BA00UsesNativeObjectFlags(t *testing.T) {
	obj := &server.Object{ObjSubClass: math.MaxUint32, ObjFlags: 0}
	if got := aiPathDoorDirectionMask50BA00(obj); got != 0xD8 {
		t.Fatalf("ground mask = %#x, want 0xd8", got)
	}

	obj.ObjSubClass = 0
	obj.ObjFlags = object.FlagAirborne
	if got := aiPathDoorDirectionMask50BA00(obj); got != 0x98 {
		t.Fatalf("airborne mask = %#x, want 0x98", got)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		if got, want := unsafe.Offsetof(obj.ObjSubClass), uintptr(16); got != want {
			t.Fatalf("native ObjSubClass offset = %d, want %d", got, want)
		}
		if got, want := unsafe.Offsetof(obj.ObjFlags), uintptr(20); got != want {
			t.Fatalf("native ObjFlags offset = %d, want %d", got, want)
		}
	}
}

func TestAIPathTileProbe50C830UsesNativeFourPointTable(t *testing.T) {
	wantOffsets := [...]types.Pointf{
		{X: 11.5, Y: 0},
		{X: 23, Y: 11.5},
		{X: 0, Y: 11.5},
		{X: 11.5, Y: 23},
	}
	if aiPathTileProbes50C830 != wantOffsets {
		t.Fatalf("AI tile probe offsets = %v, want %v", aiPathTileProbes50C830, wantOffsets)
	}
	wantBits := [...][2]uint32{
		{0x41380000, 0x00000000},
		{0x41b80000, 0x41380000},
		{0x00000000, 0x41380000},
		{0x41380000, 0x41b80000},
	}
	for i, p := range aiPathTileProbes50C830 {
		if got := [2]uint32{math.Float32bits(p.X), math.Float32bits(p.Y)}; got != wantBits[i] {
			t.Fatalf("AI tile probe offset %d bits = %#x, want %#x", i, got, wantBits[i])
		}
	}

	wantPoints := [...]types.Pointf{
		{X: 172.5, Y: -69},
		{X: 184, Y: -57.5},
		{X: 161, Y: -57.5},
		{X: 172.5, Y: -46},
	}
	var gotPoints []types.Pointf
	if got := aiPathTileProbe50C830(7, -3, func(p types.Pointf) int {
		gotPoints = append(gotPoints, p)
		return 0
	}); got != 1 {
		t.Fatalf("clear AI tile probes = %d, want 1", got)
	}
	if len(gotPoints) != len(wantPoints) {
		t.Fatalf("AI tile probe count = %d, want %d", len(gotPoints), len(wantPoints))
	}
	for i, want := range wantPoints {
		if gotPoints[i] != want {
			t.Fatalf("AI tile probe %d = %v, want %v", i, gotPoints[i], want)
		}
	}
}

func TestAIPathTileProbe50C830StopsAtBlockingTile(t *testing.T) {
	for block := 0; block < 4; block++ {
		seen := 0
		got := aiPathTileProbe50C830(0, 0, func(types.Pointf) int {
			cur := seen
			seen++
			if cur == block {
				return 6
			}
			return 0
		})
		if got != 0 {
			t.Fatalf("blocking probe %d result = %d, want 0", block, got)
		}
		if seen != block+1 {
			t.Fatalf("blocking probe %d visited %d probes, want %d", block, seen, block+1)
		}
	}
}
