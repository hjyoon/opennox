package legacy

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectNeedSyncCMatchesGo(t *testing.T) {
	got := &server.Object{Field37: 0x13579BDF, Field38: 0x2468ACE0}
	want := *got
	want.NeedSync()

	objectNeedSyncC(got)
	if got.Field38 != want.Field38 {
		t.Fatalf("Field38: C = %#08x, Go = %#08x", got.Field38, want.Field38)
	}
	if got.Field37 != want.Field37 {
		t.Fatalf("C overwrote Field37: got %#08x, want %#08x", got.Field37, want.Field37)
	}
}

func TestObjectStatusMaskCMatchesGo(t *testing.T) {
	const (
		val1 = uint32(0x020000)
		val2 = uint32(0x000002)
	)
	for _, set := range []bool{false, true} {
		t.Run(map[bool]string{false: "clear", true: "set"}[set], func(t *testing.T) {
			got := &server.Object{Field37: 0x80000015, Field38: 0xA5A5A5A5}
			for i := range got.Field140 {
				got.Field140[i] = 0x5A000000 | val1 | val2 | uint32(i<<4)
			}
			want := *got
			want.Sub_4E4500(val1, val2, set)

			objectStatusMaskC(got, val1, val2, set)
			if got.Field140 != want.Field140 {
				for i := range got.Field140 {
					if got.Field140[i] != want.Field140[i] {
						t.Errorf("Field140[%d]: C = %#08x, Go = %#08x", i, got.Field140[i], want.Field140[i])
					}
				}
			}
			if got.Field37 != want.Field37 || got.Field38 != want.Field38 {
				t.Fatalf("C overwrote adjacent state: Field37=%#08x Field38=%#08x", got.Field37, got.Field38)
			}
		})
	}
}

func TestObjectRaiseCMatchesGo(t *testing.T) {
	tests := []struct {
		name string
		from float32
		to   float32
	}{
		{name: "changed", from: 2.5, to: 7.25},
		{name: "equal", from: 2.5, to: 2.5},
		{name: "negative zero", from: float32(math.Copysign(0, -1)), to: 0},
		{name: "source NaN is unordered", from: math.Float32frombits(0x7FC00001), to: 7.25},
		{name: "target NaN is unordered", from: 2.5, to: math.Float32frombits(0x7FC00002)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := &server.Object{
				ObjClass: object.ClassPlayer,
				ZVal:     tc.from,
				Field37:  0x13579BDF,
				Field38:  0x2468ACE0,
			}
			for i := range got.Field140 {
				got.Field140[i] = 0xA5000000 | uint32(i<<4) | uint32(i&7)
			}
			want := *got
			want.Raise(tc.to)

			objectRaiseC(got, tc.to)
			if math.Float32bits(got.ZVal) != math.Float32bits(want.ZVal) {
				t.Errorf("ZVal: C = %#08x, Go = %#08x", math.Float32bits(got.ZVal), math.Float32bits(want.ZVal))
			}
			if got.Field38 != want.Field38 {
				t.Errorf("Field38: C = %#08x, Go = %#08x", got.Field38, want.Field38)
			}
			if got.Field140 != want.Field140 {
				t.Errorf("Field140 differs: C = %#v, Go = %#v", got.Field140, want.Field140)
			}
			if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass {
				t.Fatal("C overwrote state outside the original function contract")
			}
		})
	}
}

func TestObjectMarkAnimFrameCMatchesGo(t *testing.T) {
	got := &server.Object{
		ObjClass: object.ClassImmobile,
		Field33:  0xAAAAAAAA,
		Field37:  0x13579BDF,
		Field38:  0x2468ACE0,
	}
	for i := range got.Field140 {
		got.Field140[i] = 0x5A000000 | uint32(i<<4) | uint32(i&7)
	}
	want := *got
	want.MarkAnimFrame(0xFEDCBA98)

	objectMarkAnimFrameC(got, 0xFEDCBA98)
	if got.Field33 != want.Field33 || got.Field38 != want.Field38 {
		t.Errorf("frame/sync: C = (%#08x, %#08x), Go = (%#08x, %#08x)", got.Field33, got.Field38, want.Field33, want.Field38)
	}
	if got.Field140 != want.Field140 {
		t.Errorf("Field140 differs: C = %#v, Go = %#v", got.Field140, want.Field140)
	}
	if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass {
		t.Fatal("C overwrote state outside the original function contract")
	}
}
