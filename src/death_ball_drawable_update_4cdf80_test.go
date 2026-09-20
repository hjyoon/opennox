package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateDeathBallCharge4CE0C0HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(65530, 65534)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", source)
	}
	spawned, activated, randomCalls := 0, 0, 0
	hooks := deathBallDrawableHooks4CDF80{
		typeID: func(name string) int {
			if name != "CharmOrb" {
				t.Fatalf("type name = %q", name)
			}
			return 93
		},
		random: func(min, max int) int {
			randomCalls++
			switch {
			case min == 0 && max == 255:
				return 0
			case min == 2 && max == 8:
				return 2
			case min == 0 && max == 100:
				return 0
			case min == 6 && max == 10:
				return 8
			case min == -20 && max == 20:
				if randomCalls%7 == 5 {
					return -7 // Y offset is sampled before X.
				}
				return 5
			case min == 3 && max == 10:
				return 4
			default:
				t.Fatalf("unexpected random range (%d, %d)", min, max)
				return 0
			}
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawned++
			if typ != 93 || pos != image.Pt(31, 65527) {
				t.Fatalf("orb %d = type %d, pos %v", spawned, typ, pos)
			}
			return &client.Drawable{PosVec: pos}
		},
		activate: func(orb *client.Drawable) {
			activated++
			effect := orb.UnionEffect()
			if effect.Field_108 != 0xfffefffa || effect.Field_110 != 0x08000000 || effect.Field_111 != 4 {
				t.Fatalf("orb %d union = (%#x, %#x, %#x)", activated,
					effect.Field_108, effect.Field_110, effect.Field_111)
			}
		},
	}
	if got := updateDeathBallCharge4CE0C0(source, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if spawned != 10 || activated != 10 || randomCalls != 70 {
		t.Fatalf("spawned=%d activated=%d random calls=%d", spawned, activated, randomCalls)
	}
}

func TestUpdateDeathBallSparks4CDF80HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(1551, 987)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", source)
	}
	spawned, activated, frames := 0, 0, 0
	hooks := deathBallDrawableHooks4CDF80{
		typeID: func(name string) int {
			if name != "DeathBallSpark" {
				t.Fatalf("type name = %q", name)
			}
			return 94
		},
		random: func(min, max int) int {
			switch {
			case min == 0 && max == 255:
				return 41
			case min == 1000 && max == 3000:
				return 1500
			case min == 10 && max == 40:
				return 25
			case min == 0 && max == 4:
				return 3
			default:
				t.Fatalf("unexpected random range (%d, %d)", min, max)
				return 0
			}
		},
		frame: func() uint32 {
			frame := uint32(500 + frames)
			frames++
			return frame
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawned++
			if typ != 94 || pos != source.PosVec {
				t.Fatalf("spark %d = type %d, pos %v", spawned, typ, pos)
			}
			if spawned == 2 {
				return nil
			}
			return &client.Drawable{PosVec: pos}
		},
		activate: func(spark *client.Drawable) {
			activated++
			effect := spark.UnionEffect()
			wantFrame := uint32(500 + 2*(activated-1))
			if effect.Field_108 != uint32(source.PosVec.X)<<12 ||
				effect.Field_109 != uint32(source.PosVec.Y)<<12 ||
				effect.Field_110 != 1500 || effect.Field_111 != wantFrame+1 ||
				effect.Field_112 != wantFrame+25 || spark.Field_74_4 != 41 ||
				spark.ZVal != 22 || spark.VelZ != 3 {
				t.Fatalf("spark %d is not initialized: %+v, %+v", activated, spark, effect)
			}
		},
	}
	if got := updateDeathBallSparks4CDF80(source, 3, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if spawned != 3 || activated != 2 || frames != 4 {
		t.Fatalf("spawned=%d activated=%d frames=%d", spawned, activated, frames)
	}
}

func TestLinearOrbStep4CA650HighAddress(t *testing.T) {
	orb := &client.Drawable{PosVec: image.Pt(120, 100)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(orb)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", orb)
	}
	orb.UnionEffect().Field_108 = 100 | 100<<16
	orb.UnionEffect().Field_110 = 8 << 24
	if next, done := linearOrbStep4CA650(orb); next != image.Pt(113, 100) || done {
		t.Fatalf("first step = %v, done %v", next, done)
	}
	orb.PosVec = image.Pt(109, 100)
	if _, done := linearOrbStep4CA650(orb); !done {
		t.Fatal("orb within 10 pixels was not removed")
	}
	orb.PosVec = image.Pt(111, 100)
	orb.UnionEffect().Field_110 = 20 << 24
	if next, done := linearOrbStep4CA650(orb); next.X >= 100 || !done {
		t.Fatalf("overshoot = %v, done %v", next, done)
	}
}

func TestDrawableUpdateRejectsUnknownLegacyCallbacks(t *testing.T) {
	marker := new(byte)
	dr := &client.Drawable{
		ClientUpdateFuncPtr: unsafe.Pointer(marker),
		Field_115:           unsafe.Pointer(marker),
	}
	c := new(Client)

	if got := c.callDrawableUpdate49BD70(nil, dr); got != 1 {
		t.Fatalf("unknown primary update = %d, want 1", got)
	}
	// The secondary dispatcher must also ignore unknown native callbacks.
	// This call is primarily a regression assertion that it does not enter C.
	c.callDrawableSecondaryUpdate49BD70(nil, dr)
}
