package opennox

import (
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateMagicMissileDrawable4CD9E0HighAddress(t *testing.T) {
	missile := &client.Drawable{
		PosVec:  image.Pt(130, 240),
		Field_8: 100,
		Field_9: 200,
	}
	missileEffect := missile.UnionEffect()
	missileEffect.Field_108 = 110
	missileEffect.Field_109 = 220
	spawnedDrawables := make([]client.Drawable, 3)
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(missile)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(&spawnedDrawables[0])) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}

	typeNames := make([]string, 0, 3)
	typeIDs := map[string]int{
		"Spark":                31,
		"Puff":                 32,
		"MagicMissileTailLink": 33,
	}
	randomValues := []int{25, 50, 77, 6, 4, 75, 20}
	randomRanges := [][2]int{
		{0, 100}, {0, 100}, {0, 255}, {3, 10}, {0, 6},
		{0, 100}, {0, 100},
	}
	randomIndex := 0
	spawned := 0
	activated := make([]*client.Drawable, 0, 2)
	frames := 0
	decayed := false
	hooks := magicMissileDrawableHooks4CD9E0{
		sparkCount: func() int32 { return 2 },
		typeID: func(name string) int {
			typeNames = append(typeNames, name)
			return typeIDs[name]
		},
		random: func(min, max int) int {
			if randomIndex >= len(randomValues) {
				t.Fatalf("unexpected random call (%d, %d)", min, max)
			}
			if got := [2]int{min, max}; got != randomRanges[randomIndex] {
				t.Fatalf("random range %d = %v, want %v", randomIndex, got, randomRanges[randomIndex])
			}
			value := randomValues[randomIndex]
			randomIndex++
			return value
		},
		frame: func() uint32 {
			frames++
			return 500
		},
		fps: func() uint32 { return 30 },
		spawn: func(typ int, pos image.Point) *client.Drawable {
			index := spawned
			spawned++
			switch index {
			case 0:
				if typ != 31 || pos != image.Pt(107, 220) {
					t.Fatalf("first spark spawn = %d/%v", typ, pos)
				}
				spawnedDrawables[index].PosVec = pos
				return &spawnedDrawables[index]
			case 1:
				if typ != 31 || pos != image.Pt(122, 208) {
					t.Fatalf("second spark spawn = %d/%v", typ, pos)
				}
				return nil
			case 2:
				if typ != 33 || pos != image.Pt(110, 220) {
					t.Fatalf("tail spawn = %d/%v", typ, pos)
				}
				spawnedDrawables[index].PosVec = pos
				return &spawnedDrawables[index]
			default:
				t.Fatalf("unexpected spawn %d/%v", typ, pos)
				return nil
			}
		},
		activate: func(dr *client.Drawable) {
			activated = append(activated, dr)
		},
		decay: func(dr *client.Drawable, lifetime int) {
			if dr != &spawnedDrawables[2] || lifetime != 10 {
				t.Fatalf("decay = %p/%d", dr, lifetime)
			}
			decayed = true
		},
	}

	if got := updateMagicMissileDrawable4CD9E0(missile, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if want := []string{"Spark", "Puff", "MagicMissileTailLink"}; !reflect.DeepEqual(typeNames, want) {
		t.Fatalf("type names = %v, want %v", typeNames, want)
	}
	if spawned != 3 || randomIndex != len(randomValues) || frames != 2 || !decayed {
		t.Fatalf("spawned=%d random=%d frames=%d decayed=%t", spawned, randomIndex, frames, decayed)
	}
	if want := []*client.Drawable{&spawnedDrawables[0], &spawnedDrawables[2]}; !reflect.DeepEqual(activated, want) {
		t.Fatalf("activated = %p, want %p", activated, want)
	}

	spark := &spawnedDrawables[0]
	sparkEffect := spark.UnionEffect()
	if sparkEffect.Field_108 != uint32(107)<<12 || sparkEffect.Field_109 != uint32(220)<<12 ||
		sparkEffect.Field_110 != 0 || sparkEffect.Field_111 != 500 || sparkEffect.Field_112 != 506 ||
		spark.Field_74_4 != 77 || spark.ZVal != 20 || spark.VelZ != 4 {
		t.Fatalf("spark not initialized: %+v, %+v", spark, sparkEffect)
	}
	tailEffect := spawnedDrawables[2].UnionEffect()
	if tailEffect.Field_108 != 130 || tailEffect.Field_109 != 240 ||
		missileEffect.Field_108 != 130 || missileEffect.Field_109 != 240 {
		t.Fatalf("tail/missile positions = %+v/%+v", tailEffect, missileEffect)
	}
}

func TestUpdateMagicMissileDrawable4CD9E0TailSpawnFailure(t *testing.T) {
	missile := &client.Drawable{PosVec: image.Pt(30, 40)}
	effect := missile.UnionEffect()
	effect.Field_108 = 10
	effect.Field_109 = 20
	activated, decayed := false, false
	got := updateMagicMissileDrawable4CD9E0(missile, magicMissileDrawableHooks4CD9E0{
		sparkCount: func() int32 { return 0 },
		typeID:     func(string) int { return 1 },
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if pos != image.Pt(10, 20) {
				t.Fatalf("tail position = %v", pos)
			}
			return nil
		},
		activate: func(*client.Drawable) { activated = true },
		decay:    func(*client.Drawable, int) { decayed = true },
	})
	if got != 1 || activated || decayed || effect.Field_108 != 10 || effect.Field_109 != 20 {
		t.Fatalf("got=%d activated=%t decayed=%t effect=%+v", got, activated, decayed, effect)
	}
}

func TestUpdateMagicMissileDrawable4CD9E0TailThreshold(t *testing.T) {
	missile := &client.Drawable{PosVec: image.Pt(11, 13)}
	effect := missile.UnionEffect()
	effect.Field_108 = 1
	effect.Field_109 = 3
	spawned := false
	got := updateMagicMissileDrawable4CD9E0(missile, magicMissileDrawableHooks4CD9E0{
		sparkCount: func() int32 { return 0 },
		typeID:     func(string) int { return 1 },
		spawn: func(int, image.Point) *client.Drawable {
			spawned = true
			return nil
		},
	})
	if got != 1 || spawned {
		t.Fatalf("got=%d spawned=%t", got, spawned)
	}
}
