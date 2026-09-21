//go:build !server

package opennox

import (
	"bytes"
	_ "embed"
	"image"
	"math"
	"testing"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

//go:embed common/memmap/nox/blobdata/blob_587000.dat
var undeadKillerBlob587000Test []byte

func TestUndeadKillerSqrtTable4B7A80(t *testing.T) {
	const offset = 155956
	if len(undeadKillerBlob587000Test) < offset+len(undeadKillerSqrtTable4B7A80) {
		t.Fatalf("blob_587000.dat is only %d bytes", len(undeadKillerBlob587000Test))
	}
	want := undeadKillerBlob587000Test[offset : offset+len(undeadKillerSqrtTable4B7A80)]
	if !bytes.Equal(undeadKillerSqrtTable4B7A80[:], want) {
		t.Fatal("native square-root table differs from byte_587000[155956:156212]")
	}
}

func TestUndeadKillerDistance4B7A80(t *testing.T) {
	tests := []struct {
		name string
		from image.Point
		toX  uint32
		toY  uint32
		want uint32
	}{
		{"zero", image.Pt(10, 20), 10, 20, 0},
		{"three four five", image.Point{}, 3, 4, 5},
		{"radius transition", image.Point{}, 96, 128, 160},
		{"large approximation", image.Point{}, 300, 400, 498},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := undeadKillerDistance4B7A80(tc.from, tc.toX, tc.toY); got != tc.want {
				t.Fatalf("distance = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestUndeadKillerLifetime4B7A80(t *testing.T) {
	if !undeadKillerAlive4B7A80(170, 100) {
		t.Fatal("70-frame endpoint should remain alive")
	}
	if undeadKillerAlive4B7A80(171, 100) {
		t.Fatal("71-frame endpoint should expire")
	}
	if !undeadKillerAlive4B7A80(3, math.MaxUint32-5) {
		t.Fatal("wrapped nine-frame lifetime should remain alive")
	}
}

func TestDrawUndeadKiller4B7A80HighAddress(t *testing.T) {
	source := &client.Drawable{
		PosVec:    image.Pt(0x10002, 0x1fffc),
		ZVal:      uint16(int16(7)),
		ZVal2:     uint16(int16(5)),
		AnimStart: math.MaxUint32 - 5,
		Field_81:  0x10002 + 96,
		Field_82:  0x1fffc + 128,
	}
	orb := &client.Drawable{}
	orb.UnionEffect().Field_110 = 0x11223344
	orb.UnionEffect().Field_111 = 0xaabbccdd
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(orb)) <= uintptr(^uint32(0)) {
			t.Skipf("allocator returned a low drawable address: source=%p orb=%p", source, orb)
		}
	}

	randomValues := []int{90, -5, 5, 7, -3, 9}
	randomRanges := [][2]int{{0, 100}, {-5, 5}, {-5, 5}, {6, 10}, {-5, 5}, {3, 10}}
	randomIndex, activateCalls, deleteCalls, glowCalls := 0, 0, 0, 0
	vp := &noxrender.Viewport{Screen: image.Rect(20, 30, 660, 510), World: image.Rect(100, 200, 740, 680)}
	hooks := undeadKillerDrawHooks4B7A80{
		frame: func() uint32 { return 3 },
		random: func(minimum, maximum int) int {
			if got := [2]int{minimum, maximum}; got != randomRanges[randomIndex] {
				t.Fatalf("random range %d = %v, want %v", randomIndex, got, randomRanges[randomIndex])
			}
			value := randomValues[randomIndex]
			randomIndex++
			return value
		},
		spawn: func(typeID int, pos image.Point) *client.Drawable {
			if typeID != 77 || pos != image.Pt(65533, -2) {
				t.Fatalf("spawn = type %d at %v, want type 77 at (65533,-2)", typeID, pos)
			}
			return orb
		},
		activate: func(got *client.Drawable) {
			activateCalls++
			if got != orb {
				t.Fatalf("activated %p, want %p", got, orb)
			}
		},
		delete: func(*client.Drawable) { deleteCalls++ },
		drawGlow: func(pos image.Point, color noxcolor.RGBA5551, inner, outer int) {
			glowCalls++
			wantPoint := vp.ToScreenPos(source.PosVec).Add(image.Pt(0, -12))
			if pos != wantPoint || color != undeadKillerGlow4B7A80 || inner != 4 || outer != 12 {
				t.Fatalf("glow = pos:%v color:%#x radii:(%d,%d), want pos:%v color:%#x radii:(4,12)", pos, color, inner, outer, wantPoint, undeadKillerGlow4B7A80)
			}
		},
	}

	if got := drawUndeadKiller4B7A80(source, vp, 77, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	effect := orb.UnionEffect()
	wantEndpoints := uint32(uint16(source.Field_81)) | uint32(uint16(source.Field_82))<<16
	if effect.Field_108 != wantEndpoints || effect.Field_110 != 0x07223344 || effect.Field_111 != 0xaa000009 {
		t.Fatalf("orb state = endpoints:%#x field110:%#x field111:%#x", effect.Field_108, effect.Field_110, effect.Field_111)
	}
	if randomIndex != len(randomValues) || activateCalls != 1 || deleteCalls != 0 || glowCalls != 1 {
		t.Fatalf("calls = random:%d activate:%d delete:%d glow:%d", randomIndex, activateCalls, deleteCalls, glowCalls)
	}
}

func TestDrawUndeadKillerExpires4B7A80(t *testing.T) {
	source := &client.Drawable{AnimStart: 100}
	deleted := 0
	hooks := undeadKillerDrawHooks4B7A80{
		frame: func() uint32 { return 171 },
		delete: func(got *client.Drawable) {
			deleted++
			if got != source {
				t.Fatalf("deleted %p, want %p", got, source)
			}
		},
		random:   func(int, int) int { panic("expired effect requested randomness") },
		spawn:    func(int, image.Point) *client.Drawable { panic("expired effect spawned an orb") },
		activate: func(*client.Drawable) { panic("expired effect activated an orb") },
		drawGlow: func(image.Point, noxcolor.RGBA5551, int, int) { panic("expired effect drew a glow") },
	}
	if got := drawUndeadKiller4B7A80(source, &noxrender.Viewport{}, 77, hooks); got != 0 || deleted != 1 {
		t.Fatalf("expired result = %d, deletes = %d; want 0, 1", got, deleted)
	}
}

func TestUndeadKillerDrawDispatch4B7A80(t *testing.T) {
	fn := legacy.Get_nox_thing_undead_killer_draw()
	if fn == nil {
		t.Fatal("undead-killer callback pointer is nil")
	}
	dr := &client.Drawable{}
	if got, ok := (*Client)(nil).callUndeadKillerDraw4B7A80(dr, nil); ok || got != 0 {
		t.Fatalf("non-undead dispatch = (%d, %t), want (0, false)", got, ok)
	}
	dr.DrawFuncPtr = fn
	if dr.DrawFuncPtr != fn {
		t.Fatal("undead-killer callback pointer was not retained at native width")
	}
}
