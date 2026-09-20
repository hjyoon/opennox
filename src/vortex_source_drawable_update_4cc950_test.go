package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateVortexSourceDrawable4CC950HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(100, 200), ZVal: 17}
	orb := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(orb)) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}

	randomValues := []int{64, 3, 51}
	randomIndex := 0
	activated, sightDestroyed := false, false
	wantPosition := source.PosVec.Add(image.Pt(
		50*sincosTable16[64].X/16,
		50*sincosTable16[64].Y/16,
	))
	hooks := vortexSourceDrawableHooks4CC950{
		typeID: func(name string) int {
			if name != "WhiteVortexOrb" {
				t.Fatalf("type name = %q", name)
			}
			return 73
		},
		random: func(min, max int) int {
			wantRanges := [][2]int{{0, 255}, {2, 3}, {0, 100}}
			if randomIndex >= len(randomValues) || [2]int{min, max} != wantRanges[randomIndex] {
				t.Fatalf("random call %d = (%d, %d)", randomIndex, min, max)
			}
			value := randomValues[randomIndex]
			randomIndex++
			return value
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 73 || pos != wantPosition {
				t.Fatalf("spawn = %d/%v, want 73/%v", typ, pos, wantPosition)
			}
			orb.PosVec = pos
			return orb
		},
		dimColor:    0x11223344,
		brightColor: 0x55667788,
		activate: func(got *client.Drawable) {
			if got != orb || sightDestroyed {
				t.Fatalf("activate = %p after sight destroy=%t", got, sightDestroyed)
			}
			activated = true
		},
		sightDestroy: func(got *client.Drawable) {
			if got != orb || !activated {
				t.Fatalf("sight destroy = %p after activate=%t", got, activated)
			}
			sightDestroyed = true
		},
	}

	if got := updateVortexSourceDrawable4CC950(source, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if randomIndex != len(randomValues) || !activated || !sightDestroyed || source.ZVal != 0 {
		t.Fatalf("random=%d activated=%t sight=%t z=%d", randomIndex, activated, sightDestroyed, source.ZVal)
	}
	effect := orb.UnionEffect()
	wantMotion := uint32(64) | uint32(0xfd)<<8 | 50<<16 | 1<<24
	if effect.Field_108 != hooks.dimColor || effect.Field_109 != hooks.brightColor ||
		effect.Field_110 != 100 || effect.Field_111 != 200 || effect.Field_112 != wantMotion {
		t.Fatalf("orb effect = %+v, want motion %#x", effect, wantMotion)
	}
}

func TestUpdateVortexSourceDrawable4CC950PositiveSpin(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(-10, -20), ZVal: 9}
	orb := &client.Drawable{}
	values := []int{255, 2, 50}
	index := 0
	got := updateVortexSourceDrawable4CC950(source, vortexSourceDrawableHooks4CC950{
		typeID: func(string) int { return 1 },
		random: func(int, int) int {
			value := values[index]
			index++
			return value
		},
		spawn:        func(int, image.Point) *client.Drawable { return orb },
		activate:     func(*client.Drawable) {},
		sightDestroy: func(*client.Drawable) {},
	})
	if got != 1 || source.ZVal != 0 || byte(orb.UnionEffect().Field_112>>8) != 2 {
		t.Fatalf("got=%d z=%d motion=%#x", got, source.ZVal, orb.UnionEffect().Field_112)
	}
}

func TestUpdateVortexSourceDrawable4CC950SpawnFailure(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(10, 20), ZVal: 7}
	randomCalls, activated, sightDestroyed := 0, false, false
	got := updateVortexSourceDrawable4CC950(source, vortexSourceDrawableHooks4CC950{
		typeID: func(string) int { return 1 },
		random: func(int, int) int {
			randomCalls++
			return 0
		},
		spawn:        func(int, image.Point) *client.Drawable { return nil },
		activate:     func(*client.Drawable) { activated = true },
		sightDestroy: func(*client.Drawable) { sightDestroyed = true },
	})
	if got != 1 || randomCalls != 1 || source.ZVal != 7 || activated || sightDestroyed {
		t.Fatalf("got=%d random=%d z=%d activated=%t sight=%t", got, randomCalls, source.ZVal, activated, sightDestroyed)
	}
}
