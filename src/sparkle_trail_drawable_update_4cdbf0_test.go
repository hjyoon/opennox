package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestUpdateSparkleTrailDrawable4CDBF0HighAddress(t *testing.T) {
	trail := &client.Drawable{
		PosVec:  image.Pt(120, 220),
		Field_8: 100,
		Field_9: 200,
		ZVal:    11,
		ZVal2:   12,
	}
	spawnedDrawables := make([]client.Drawable, 5)
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(trail)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(&spawnedDrawables[0])) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}
	drawMarker := byte(0)
	drawFunc := unsafe.Pointer(&drawMarker)
	spawned, activated, frames := 0, 0, 0
	hooks := sparkleTrailDrawableHooks4CDBF0{
		typeID: func(name string) int {
			if name != "BlueSpark" {
				t.Fatalf("type name = %q", name)
			}
			return 37
		},
		random: func(min, max int) int {
			switch {
			case min == -3 && max == 3:
				return 2
			case min == 2 && max == 10:
				return 7
			default:
				t.Fatalf("unexpected random range (%d, %d)", min, max)
				return 0
			}
		},
		frame: func() uint32 {
			frames++
			return 500
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 37 {
				t.Fatalf("type = %d", typ)
			}
			index := spawned
			spawned++
			want := image.Pt(102+4*index, 202+4*index)
			if pos != want {
				t.Fatalf("spark %d position = %v, want %v", spawned, pos, want)
			}
			if index == 2 {
				return nil
			}
			spawnedDrawables[index].PosVec = pos
			return &spawnedDrawables[index]
		},
		drawFunc: drawFunc,
		activate: func(spark *client.Drawable) {
			activated++
			effect := spark.UnionEffect()
			if spark.DrawFuncPtr != drawFunc || spark.LightFlags != 2 ||
				spark.LightColor.R != 255 || spark.LightColor.G != 200 || spark.LightColor.B != 75 ||
				effect.Field_108 != uint32(spark.PosVec.X)<<12 ||
				effect.Field_109 != uint32(spark.PosVec.Y)<<12 ||
				effect.Field_110 != 0 || effect.Field_111 != 500 || effect.Field_112 != 507 ||
				spark.Field_74_4 != 0 || spark.ZVal != 11 || spark.ZVal2 != 12 || spark.VelZ != 0 {
				t.Fatalf("spark %d not initialized: %+v, %+v", activated, spark, effect)
			}
		},
	}
	if got := updateSparkleTrailDrawable4CDBF0(trail, hooks); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if spawned != 5 || activated != 4 || frames != 8 {
		t.Fatalf("spawned=%d activated=%d frames=%d", spawned, activated, frames)
	}
}
