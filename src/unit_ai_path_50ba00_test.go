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
