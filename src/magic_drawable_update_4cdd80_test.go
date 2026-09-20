package opennox

import (
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateMagicDrawable4CDD80HighAddress(t *testing.T) {
	magic := &client.Drawable{
		PosVec:  image.Pt(140, 240),
		Field_8: 100,
		Field_9: 200,
		ZVal:    5,
		ZVal2:   6,
	}
	magicEffect := magic.UnionEffect()
	magicEffect.Field_108 = 110
	magicEffect.Field_109 = 220
	spawnedDrawables := make([]client.Drawable, 5)
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(magic)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(&spawnedDrawables[0])) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}

	drawMarker := byte(0)
	drawFunc := unsafe.Pointer(&drawMarker)
	typeNames := make([]string, 0, 2)
	typeIDs := map[string]int{"MagicTailLink": 41, "BlueSpark": 42}
	randomValues := []int{-1, 1, 11, -2, 2, 12, -3, 3, -4, 4, 13}
	randomIndex := 0
	spawned := 0
	activated := make([]*client.Drawable, 0, 4)
	frames := 0
	decayed := false
	hooks := magicDrawableHooks4CDD80{
		typeID: func(name string) int {
			typeNames = append(typeNames, name)
			return typeIDs[name]
		},
		random: func(min, max int) int {
			if randomIndex >= len(randomValues) {
				t.Fatalf("unexpected random call (%d, %d)", min, max)
			}
			value := randomValues[randomIndex]
			if min == -8 && max == 8 {
				if value < min || value > max {
					t.Fatalf("random value %d outside (%d, %d)", value, min, max)
				}
			} else if min != 10 || max != 20 || value < min || value > max {
				t.Fatalf("unexpected random call/value (%d, %d) = %d", min, max, value)
			}
			randomIndex++
			return value
		},
		frame: func() uint32 {
			frames++
			return 700
		},
		fps: func() uint32 { return 30 },
		spawn: func(typ int, pos image.Point) *client.Drawable {
			index := spawned
			spawned++
			wantPositions := []image.Point{
				image.Pt(110, 220),
				image.Pt(99, 201),
				image.Pt(108, 212),
				image.Pt(117, 223),
				image.Pt(126, 234),
			}
			wantType := 42
			if index == 0 {
				wantType = 41
			}
			if typ != wantType || pos != wantPositions[index] {
				t.Fatalf("spawn %d = %d/%v, want %d/%v", index, typ, pos, wantType, wantPositions[index])
			}
			if index == 3 {
				return nil
			}
			spawnedDrawables[index].PosVec = pos
			return &spawnedDrawables[index]
		},
		drawFunc: drawFunc,
		activate: func(dr *client.Drawable) {
			activated = append(activated, dr)
		},
		decay: func(dr *client.Drawable, lifetime int) {
			if dr != &spawnedDrawables[0] || lifetime != 30 {
				t.Fatalf("decay = %p/%d", dr, lifetime)
			}
			decayed = true
		},
	}

	if got := updateMagicDrawable4CDD80(magic, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if want := []string{"MagicTailLink", "BlueSpark"}; !reflect.DeepEqual(typeNames, want) {
		t.Fatalf("type names = %v, want %v", typeNames, want)
	}
	if spawned != 5 || randomIndex != len(randomValues) || frames != 6 || !decayed {
		t.Fatalf("spawned=%d random=%d frames=%d decayed=%t", spawned, randomIndex, frames, decayed)
	}
	if want := []*client.Drawable{
		&spawnedDrawables[0], &spawnedDrawables[1], &spawnedDrawables[2], &spawnedDrawables[4],
	}; !reflect.DeepEqual(activated, want) {
		t.Fatalf("activated = %p, want %p", activated, want)
	}
	if tailEffect := spawnedDrawables[0].UnionEffect(); tailEffect.Field_108 != 140 || tailEffect.Field_109 != 240 ||
		magicEffect.Field_108 != 140 || magicEffect.Field_109 != 240 {
		t.Fatalf("tail/magic positions = %+v/%+v", tailEffect, magicEffect)
	}

	for index, lifetime := range map[int]uint32{1: 11, 2: 12, 4: 13} {
		spark := &spawnedDrawables[index]
		effect := spark.UnionEffect()
		if spark.DrawFuncPtr != drawFunc || spark.LightFlags != 2 ||
			spark.LightColor.R != 128 || spark.LightColor.G != 128 || spark.LightColor.B != 255 ||
			effect.Field_108 != uint32(140)<<12 || effect.Field_109 != uint32(240)<<12 ||
			effect.Field_110 != 0 || effect.Field_111 != 700 || effect.Field_112 != 700+lifetime ||
			spark.Field_74_4 != 0 || spark.ZVal != 5 || spark.ZVal2 != 6 || spark.VelZ != 0 {
			t.Fatalf("spark %d not initialized: %+v, %+v", index, spark, effect)
		}
	}
}

func TestUpdateMagicDrawable4CDD80TailSpawnFailure(t *testing.T) {
	magic := &client.Drawable{PosVec: image.Pt(30, 40)}
	effect := magic.UnionEffect()
	effect.Field_108 = 10
	effect.Field_109 = 20
	spawned, activated, decayed := 0, false, false
	got := updateMagicDrawable4CDD80(magic, magicDrawableHooks4CDD80{
		typeID: func(name string) int {
			if name == "MagicTailLink" {
				return 1
			}
			return 2
		},
		random: func(min, max int) int { return 0 },
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawned++
			return nil
		},
		activate: func(*client.Drawable) { activated = true },
		decay:    func(*client.Drawable, int) { decayed = true },
	})
	if got != 1 || spawned != 5 || activated || decayed || effect.Field_108 != 10 || effect.Field_109 != 20 {
		t.Fatalf("got=%d spawned=%d activated=%t decayed=%t effect=%+v", got, spawned, activated, decayed, effect)
	}
}

func TestUpdateMagicDrawable4CDD80TailThreshold(t *testing.T) {
	magic := &client.Drawable{PosVec: image.Pt(11, 13)}
	effect := magic.UnionEffect()
	effect.Field_108 = 1
	effect.Field_109 = 3
	tailSpawned, sparkSpawns := false, 0
	got := updateMagicDrawable4CDD80(magic, magicDrawableHooks4CDD80{
		typeID: func(name string) int {
			if name == "MagicTailLink" {
				tailSpawned = true
				return 1
			}
			return 2
		},
		random: func(min, max int) int { return 0 },
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ == 1 {
				tailSpawned = true
			} else {
				sparkSpawns++
			}
			return nil
		},
	})
	if got != 1 || tailSpawned || sparkSpawns != 4 || effect.Field_108 != 1 || effect.Field_109 != 3 {
		t.Fatalf("got=%d tail=%t sparks=%d effect=%+v", got, tailSpawned, sparkSpawns, effect)
	}
}
