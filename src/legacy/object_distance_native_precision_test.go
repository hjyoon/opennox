package legacy

import (
	"bytes"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Literal retained-register results from the independent precision-53/ToZero
// model, including the final double return. These calls exercise the Go
// implementation and exported Go body; they are not a C round-trip claim.
func TestObjectDistance4E6C00NativeRetainedEntries(t *testing.T) {
	for _, tc := range []struct {
		name         string
		x, y, bx, by uint32
		kind         server.ShapeKind
		r, w, h      uint32
		br, bw, bh   uint32
		want         uint64
	}{
		{name: "sqrt-two", x: 0x3f800000, y: 0x3f800000, want: 0x3ff6a09e667f3bcc},
		{name: "sqrt-five", x: 0x3f800000, y: 0x40000000, want: 0x4001e3779b97f4a7},
		{name: "circle-subtractions", x: 0x3f800000, y: 0x40000000, kind: server.ShapeKindCircle, r: 0x3e000000, br: 0x3e800000, want: 0x3ffdc6ef372fe94e},
		{name: "box-subtractions", x: 0x3f800000, y: 0x40000000, kind: server.ShapeKindBox, w: 0x3e800000, h: 0x3d800000, bw: 0x3e000000, bh: 0x3f000000, want: 0x3ffdc6ef372fe94e},
		{name: "retained-x-subtract", x: 0x40800000, y: 0x40400000, bx: 0x25800000, want: 0x4013ffffffffffff},
		{name: "retained-y-subtract", x: 0x40400000, y: 0x40800000, by: 0x25800000, want: 0x4013ffffffffffff},
		{name: "retained-circle-surface", x: 0x40800000, y: 0x40400000, bx: 0x25800000, kind: server.ShapeKindCircle, r: 0x3e000000, br: 0x3e800000, want: 0x40127fffffffffff},
		{name: "wide-negative-circle-surface", x: 0xff7fffff, y: 0xc0000000, bx: 0x3f800000, by: 0x40400000, kind: server.ShapeKindCircle, r: 0x3e000000, br: 0x3e800000, want: 0x47efffffdffffffe},
		{name: "unknown-full-DWORD-kind", x: 0x40800000, y: 0x40400000, bx: 0x25800000, kind: server.ShapeKind(0x60000002), r: 0x3f800000, want: 0x4013ffffffffffff},
		{name: "exact-three-four-five", x: 0x40400000, y: 0x40800000, want: 0x4014000000000000},
		{name: "overlap-floor", x: 0x3f800000, kind: server.ShapeKindCircle, r: 0x40000000, want: 0x3f847ae140000000},
		{name: "unordered-position-floor", x: 0x7fc12345, want: 0x3f847ae140000000},
		{name: "positive-infinity", x: 0x7f800000, want: 0x7ff0000000000000},
		{name: "equal-infinity-floor", x: 0x7f800000, bx: 0x7f800000, want: 0x3f847ae140000000},
		{name: "box-width-unordered-selects-height", x: 0x41200000, kind: server.ShapeKindBox, w: 0x7fc12345, h: 0x40800000, want: 0x4020000000000000},
		{name: "box-height-unordered-floor", x: 0x41200000, kind: server.ShapeKindBox, w: 0x40800000, h: 0x7fc12345, want: 0x3f847ae140000000},
	} {
		for _, entry := range []string{"implementation", "export-body"} {
			t.Run(tc.name+"/"+entry, func(t *testing.T) {
				a, freeA := alloc.New(server.Object{})
				b, freeB := alloc.New(server.Object{})
				defer freeA()
				defer freeB()
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(a)) <= math.MaxUint32 || uintptr(unsafe.Pointer(b)) <= math.MaxUint32) {
					t.Fatalf("native distance objects below 4 GiB: %p %p", a, b)
				}
				a.PosVec = types.Ptf(math.Float32frombits(tc.x), math.Float32frombits(tc.y))
				b.PosVec = types.Ptf(math.Float32frombits(tc.bx), math.Float32frombits(tc.by))
				a.Shape, b.Shape = server.Shape{Kind: tc.kind}, server.Shape{Kind: tc.kind}
				a.Shape.Circle.R, b.Shape.Circle.R = math.Float32frombits(tc.r), math.Float32frombits(tc.br)
				a.Shape.Box.W, a.Shape.Box.H = math.Float32frombits(tc.w), math.Float32frombits(tc.h)
				b.Shape.Box.W, b.Shape.Box.H = math.Float32frombits(tc.bw), math.Float32frombits(tc.bh)
				a.ObjClass, b.ObjClass = object.ClassFood, object.ClassMissile
				a.ObjFlags, b.ObjFlags = object.FlagDead|object.FlagNoUpdate, object.FlagDestroyed
				aBefore := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(a)), int(unsafe.Sizeof(*a))))
				bBefore := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(b)), int(unsafe.Sizeof(*b))))
				var got float64
				if entry == "implementation" {
					got = objectDistance_4E6C00(a, b)
				} else {
					got = float64(nox_xxx_calcDistance_4E6C00(asObjectC(a), asObjectC(b)))
				}
				if math.Float64bits(got) != tc.want {
					t.Errorf("retained %s result=%016x want=%016x", entry, math.Float64bits(got), tc.want)
				}
				if !bytes.Equal(aBefore, unsafe.Slice((*byte)(unsafe.Pointer(a)), len(aBefore))) || !bytes.Equal(bBefore, unsafe.Slice((*byte)(unsafe.Pointer(b)), len(bBefore))) {
					t.Fatal("distance entry changed native object bytes")
				}
			})
		}
	}
}
