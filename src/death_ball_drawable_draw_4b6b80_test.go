//go:build !server

package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func TestSparkDrawColors4B6970(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   unsafe.Pointer
	}{
		{"red", legacy.Get_nox_thing_red_spark_draw()},
		{"blue", legacy.Get_nox_thing_blue_spark_draw()},
		{"cyan", legacy.Get_nox_thing_cyan_spark_draw()},
		{"green", legacy.Get_nox_thing_green_spark_draw()},
		{"yellow", legacy.Get_nox_thing_yellow_spark_draw()},
		{"violet", legacy.Get_nox_thing_violet_spark_draw()},
		{"white", legacy.Get_nox_thing_white_spark_draw()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bright, dim, ok := sparkDrawColors4B6970(tc.fn)
			if !ok || bright == 0 || dim == 0 {
				t.Fatalf("colors = %v/%v/%t", bright, dim, ok)
			}
		})
	}
	if _, _, ok := sparkDrawColors4B6970(nil); ok {
		t.Fatal("nil draw function matched a spark")
	}
}

func TestSparkleDrawColors4B6770(t *testing.T) {
	tests := []struct {
		name       string
		fn         unsafe.Pointer
		roll       int
		wantBright uint16
		wantDim    uint16
	}{
		{"magic bright", legacy.Get_nox_thing_magic_sparkle_draw(), 5, uint16(manaBombOrbBright4B6B80), uint16(blueSparkBright4B6880)},
		{"magic dim", legacy.Get_nox_thing_magic_sparkle_draw(), 4, uint16(blueSparkBright4B6880), uint16(blueSparkDim4B6880)},
		{"pixie bright", legacy.Get_nox_thing_pixie_dust_draw(), 5, uint16(manaBombOrbBright4B6B80), uint16(pixieSparkBright4B6770)},
		{"pixie dim", legacy.Get_nox_thing_pixie_dust_draw(), 4, uint16(pixieSparkBright4B6770), uint16(pixieSparkDim4B6770)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			bright, dim, ok := sparkleDrawColors4B6770(tc.fn, func(min, max int) int {
				calls++
				if min != 0 || max != 10 {
					t.Fatalf("random range = (%d, %d)", min, max)
				}
				return tc.roll
			})
			if !ok || uint16(bright) != tc.wantBright || uint16(dim) != tc.wantDim || calls != 1 {
				t.Fatalf("colors = %#x/%#x, ok=%t, calls=%d", bright, dim, ok, calls)
			}
		})
	}
	calls := 0
	if _, _, ok := sparkleDrawColors4B6770(nil, func(int, int) int {
		calls++
		return 0
	}); ok || calls != 0 {
		t.Fatalf("unknown callback matched or consumed randomness: ok=%t, calls=%d", ok, calls)
	}
}

func TestSparkleLifetime4B6770HighAddress(t *testing.T) {
	dr := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	effect := dr.UnionEffect()
	effect.Field_111 = 100
	effect.Field_112 = 125
	remaining, duration, alive := sparkleLifetime4B6770(dr, 100)
	if remaining != 24 || duration != 25 || !alive {
		t.Fatalf("first frame lifetime = (%d, %d, %t)", remaining, duration, alive)
	}
	remaining, duration, alive = sparkleLifetime4B6770(dr, 124)
	if remaining != 1 || duration != 25 || !alive {
		t.Fatalf("last live frame lifetime = (%d, %d, %t)", remaining, duration, alive)
	}
	remaining, duration, alive = sparkleLifetime4B6770(dr, 125)
	if remaining != 0 || duration != 25 || alive {
		t.Fatalf("expired lifetime = (%d, %d, %t)", remaining, duration, alive)
	}
	effect.Field_111 = 200
	effect.Field_112 = 200
	if _, _, alive := sparkleLifetime4B6770(dr, 199); alive {
		t.Fatal("zero-duration sparkle reported alive")
	}
}

func TestPixieDrawState4B6E80HighAddress(t *testing.T) {
	dr := &client.Drawable{
		PosVec:  image.Pt(130, 250),
		Field_8: 90,
		Field_9: 245,
		ZVal:    10,
		ZVal2:   5,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	advancePixieZ4B6E80(dr, 49)
	if dr.ZVal != 9 {
		t.Fatalf("descending Z = %d, want 9", dr.ZVal)
	}
	advancePixieZ4B6E80(dr, 50)
	if dr.ZVal != 10 {
		t.Fatalf("ascending Z = %d, want 10", dr.ZVal)
	}
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 200, 740, 680),
	}
	point, tail := pixieDrawPoints4B6E80(dr, vp)
	if point != image.Pt(40, 55) || tail != image.Pt(20, 55) {
		t.Fatalf("pixie points = %v -> %v, want (40,55) -> (20,55)", point, tail)
	}
	dr.ZVal = 0
	advancePixieZ4B6E80(dr, 0)
	if dr.ZVal != 0 {
		t.Fatalf("minimum Z changed to %d", dr.ZVal)
	}
	dr.ZVal = 35
	advancePixieZ4B6E80(dr, 100)
	if dr.ZVal != 35 {
		t.Fatalf("maximum Z changed to %d", dr.ZVal)
	}
}

func TestFinishBlueRainSparkDraw4B7060HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(321, 654), VelZ: 5}
	spark := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(source)) <= uintptr(^uint32(0)) ||
			uintptr(unsafe.Pointer(spark)) <= uintptr(^uint32(0))) {
		t.Skip("allocator returned a low address")
	}
	spawned, activated, deleted, frames := 0, 0, 0, 0
	hooks := blueRainSparkDrawHooks4B7060{
		typeID: func(name string) int {
			if name != "WhiteSpark" {
				t.Fatalf("type name = %q", name)
			}
			return 73
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawned++
			if typ != 73 || pos != source.PosVec {
				t.Fatalf("spawn = type %d at %v", typ, pos)
			}
			return spark
		},
		random: func(min, max int) int {
			switch {
			case min == 0 && max == 255:
				return 41
			case min == 1 && max == 1611:
				return 1200
			case min == 10 && max == 96:
				return 30
			case min == 5 && max == 15:
				return 12
			case min == 0 && max == 8:
				return 7
			default:
				t.Fatalf("unexpected random range (%d, %d)", min, max)
				return 0
			}
		},
		frame: func() uint32 {
			frame := uint32(700 + frames)
			frames++
			return frame
		},
		activate: func(got *client.Drawable) {
			activated++
			if got != spark {
				t.Fatalf("activated %p, want %p", got, spark)
			}
		},
		delete: func(got *client.Drawable) {
			deleted++
			if got != source {
				t.Fatalf("deleted %p, want %p", got, source)
			}
		},
	}
	if got := finishBlueRainSparkDraw4B7060(source, 1, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	effect := spark.UnionEffect()
	if spawned != 1 || activated != 1 || deleted != 1 || frames != 2 ||
		effect.Field_108 != uint32(source.PosVec.X)<<12 ||
		effect.Field_109 != uint32(source.PosVec.Y)<<12 || effect.Field_110 != 1200 ||
		effect.Field_111 != 701 || effect.Field_112 != 730 || spark.Field_74_4 != 41 ||
		spark.ZVal != 12 || spark.ZVal2 != 0 || spark.VelZ != 7 {
		t.Fatalf("replacement not initialized: spawned=%d activated=%d deleted=%d frames=%d spark=%+v effect=%+v",
			spawned, activated, deleted, frames, spark, effect)
	}

	source.VelZ = 4
	if got := finishBlueRainSparkDraw4B7060(source, 1, hooks); got != 1 {
		t.Fatalf("slow spark result = %d, want 1", got)
	}
	if spawned != 1 || deleted != 1 {
		t.Fatalf("slow spark caused side effects: spawned=%d deleted=%d", spawned, deleted)
	}
	if got := finishBlueRainSparkDraw4B7060(source, 0, hooks); got != 0 {
		t.Fatalf("expired spark result = %d, want 0", got)
	}
}

func TestCharmOrbFields4B6B80NativeUnion(t *testing.T) {
	dr := &client.Drawable{}
	dr.UnionEffect().Field_111 = 0xaa331207
	radius, tick, countdown := charmOrbFields4B6B80(dr)
	if radius != 7 || tick != 18 || countdown != 51 {
		t.Fatalf("CharmOrb fields = (%d, %d, %d)", radius, tick, countdown)
	}
	setCharmOrbRadiusAndCountdown4B6B80(dr, 6, 50)
	if got := dr.UnionEffect().Field_111; got != 0xaa321206 {
		t.Fatalf("CharmOrb union = %#x, want %#x", got, uint32(0xaa321206))
	}
}

func TestMovingGlowOrbStep4B6B80HighAddress(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(100, 200)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	effect := dr.UnionEffect()
	effect.Field_108 = uint32(130) | uint32(240)<<16
	effect.Field_110 = uint32(10) << 24
	pos, done := movingGlowOrbStep4B6B80(dr)
	if done || pos != image.Pt(105, 207) {
		t.Fatalf("moving glow orb step = %v, done=%v", pos, done)
	}
	dr.PosVec = image.Pt(126, 236)
	pos, done = movingGlowOrbStep4B6B80(dr)
	if !done || pos != dr.PosVec {
		t.Fatalf("near glow orb step = %v, done=%v", pos, done)
	}
}

func TestAdvanceDeathBallSpark4B6970HighAddress(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(100, 200), ZVal: 22, VelZ: 3}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	effect := dr.UnionEffect()
	effect.Field_108 = 100 << 12
	effect.Field_109 = 200 << 12
	effect.Field_110 = 1500
	effect.Field_111 = 100
	effect.Field_112 = 125
	pos, remaining, duration := advanceDeathBallSpark4B6970(dr, 100)
	if pos != image.Pt(105, 200) || remaining != 24 || duration != 25 ||
		effect.Field_108 != (100<<12)+1500*16 || effect.Field_109 != 200<<12 ||
		dr.ZVal != 25 || dr.VelZ != 3 {
		t.Fatalf("spark after first frame: pos=%v remaining=%d duration=%d fields=%+v z=%d vz=%d",
			pos, remaining, duration, effect, dr.ZVal, dr.VelZ)
	}
	_, remaining, _ = advanceDeathBallSpark4B6970(dr, 101)
	if remaining != 24 || dr.ZVal != 28 || dr.VelZ != 2 {
		t.Fatalf("spark after odd frame: remaining=%d z=%d vz=%d", remaining, dr.ZVal, dr.VelZ)
	}
}
