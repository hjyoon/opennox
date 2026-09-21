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

func TestIsDrawableUpdateCallback49BD70(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   unsafe.Pointer
	}{
		{"death ball", legacy.Get_nox_xxx_updDrawDBall_4CDF80()},
		{"death ball fragment", legacy.Get_sub_4CE0A0()},
		{"death ball charge", legacy.Get_nox_xxx_updDrawDBallCharge_4CE0C0()},
		{"magic", legacy.Get_nox_xxx_updDrawMagic_4CDD80()},
		{"vortex source", legacy.Get_nox_xxx_updDrawVortexSource_4CC950()},
		{"meteor", legacy.Get_sub_4CCD00()},
		{"fist", legacy.Get_nox_xxx_updDrawFist_4CCDB0()},
		{"color light", legacy.Get_nox_xxx_updDrawColorlight_4CE390()},
		{"undead killer", legacy.Get_nox_xxx_updDrawUndeadKiller_4CCCF0()},
		{"monster generator", legacy.Get_nox_xxx_updDrawMonsterGen_4BC920()},
		{"cloud", legacy.Get_nox_xxx_updDrawCloud_4CE1D0()},
		{"small cloud", legacy.Get_sub_4CE360()},
		{"linear orb", legacy.Get_sub_4CA650()},
		{"charm", legacy.Get_sub_4CD400()},
		{"titan fireball", legacy.Get_sub_4CCE70()},
		{"strong fireball", legacy.Get_sub_4CD090()},
		{"fireball", legacy.Get_sub_4CD0C0()},
		{"weak fireball", legacy.Get_sub_4CD0F0()},
		{"pitiful fireball", legacy.Get_sub_4CD120()},
		{"heal", legacy.Get_sub_4CD450()},
		{"drain mana", legacy.Get_sub_4CD690()},
		{"mana bomb charge", legacy.Get_nox_xxx_updDrawManabombCharge_4CCAC0()},
		{"teleport wake", legacy.Get_nox_xxx_updDrawTeleportWake_4CD8D0()},
		{"sparkle trail", legacy.Get_nox_xxx_updDrawSparkleTrail_4CDBF0()},
		{"magic missile", legacy.Get_nox_xxx_updDrawMagicMissile_4CD9E0()},
		{"mana bomb orb", legacy.Get_sub_4CA720()},
		{"secondary cloud", legacy.Get_sub_4CE340()},
		{"predicted linear", legacy.Get_nox_xxx_sprite_4CA540()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fn == nil {
				t.Fatal("callback pointer is nil")
			}
			if !isDrawableUpdateCallback49BD70(tc.fn) {
				t.Fatalf("client update %p was not recognized", tc.fn)
			}
		})
	}
	if isDrawableUpdateCallback49BD70(nil) {
		t.Fatal("nil was recognized as a client update")
	}
	if isDrawableUpdateCallback49BD70(legacy.Get_nox_thing_pixie_draw()) {
		t.Fatal("Pixie draw callback was recognized as a client update")
	}
}

func TestRestoreDrawableDrawFunc4B6B80(t *testing.T) {
	fallback := client.ThingDrawDefault
	canonical := legacy.Get_nox_thing_pixie_draw()
	if fallback == nil || canonical == nil {
		t.Fatalf("draw callbacks are not initialized: fallback=%p canonical=%p", fallback, canonical)
	}
	vortexUpdate := legacy.Get_nox_xxx_updDrawVortexSource_4CC950()

	t.Run("update-only type", func(t *testing.T) {
		dr := &client.Drawable{DrawFuncPtr: vortexUpdate}
		if !restoreDrawableDrawFunc4B6B80(dr, &client.ObjectType{}, fallback) {
			t.Fatal("corrupted draw callback was not repaired")
		}
		if dr.DrawFuncPtr != nil {
			t.Fatalf("draw callback = %p, want nil", dr.DrawFuncPtr)
		}
	})

	t.Run("canonical draw", func(t *testing.T) {
		dr := &client.Drawable{DrawFuncPtr: vortexUpdate}
		typ := &client.ObjectType{DrawFunc: canonical}
		if !restoreDrawableDrawFunc4B6B80(dr, typ, fallback) {
			t.Fatal("corrupted draw callback was not repaired")
		}
		if dr.DrawFuncPtr != canonical {
			t.Fatalf("draw callback = %p, want canonical %p", dr.DrawFuncPtr, canonical)
		}
	})

	t.Run("missing type", func(t *testing.T) {
		dr := &client.Drawable{DrawFuncPtr: vortexUpdate}
		if !restoreDrawableDrawFunc4B6B80(dr, nil, fallback) {
			t.Fatal("corrupted draw callback was not repaired")
		}
		if dr.DrawFuncPtr != fallback {
			t.Fatalf("draw callback = %p, want fallback %p", dr.DrawFuncPtr, fallback)
		}
	})

	t.Run("corrupted canonical and fallback", func(t *testing.T) {
		dr := &client.Drawable{DrawFuncPtr: vortexUpdate}
		typ := &client.ObjectType{DrawFunc: legacy.Get_nox_xxx_updDrawColorlight_4CE390()}
		if !restoreDrawableDrawFunc4B6B80(dr, typ, vortexUpdate) {
			t.Fatal("corrupted draw callback was not repaired")
		}
		if dr.DrawFuncPtr != nil {
			t.Fatalf("draw callback = %p, want nil", dr.DrawFuncPtr)
		}
	})

	t.Run("valid draw unchanged", func(t *testing.T) {
		dr := &client.Drawable{DrawFuncPtr: canonical}
		if restoreDrawableDrawFunc4B6B80(dr, &client.ObjectType{}, fallback) {
			t.Fatal("valid draw callback was reported as repaired")
		}
		if dr.DrawFuncPtr != canonical {
			t.Fatalf("draw callback = %p, want original %p", dr.DrawFuncPtr, canonical)
		}
	})

	t.Run("unknown callback", func(t *testing.T) {
		unknownValue := 0
		dr := &client.Drawable{DrawFuncPtr: unsafe.Pointer(&unknownValue)}
		if !restoreDrawableDrawFunc4B6B80(dr, &client.ObjectType{DrawFunc: canonical}, fallback) {
			t.Fatal("unknown draw callback was not repaired")
		}
		if dr.DrawFuncPtr != canonical {
			t.Fatalf("draw callback = %p, want canonical %p", dr.DrawFuncPtr, canonical)
		}
	})

	t.Run("matches client update slot", func(t *testing.T) {
		unknownValue := 0
		unknown := unsafe.Pointer(&unknownValue)
		dr := &client.Drawable{DrawFuncPtr: unknown, ClientUpdateFuncPtr: unknown}
		if !restoreDrawableDrawFunc4B6B80(dr, &client.ObjectType{DrawFunc: canonical}, fallback) {
			t.Fatal("client-update alias was not repaired")
		}
		if dr.DrawFuncPtr != canonical {
			t.Fatalf("draw callback = %p, want canonical %p", dr.DrawFuncPtr, canonical)
		}
	})
}

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

func TestGlowOrbColors4B6B80IncludesMoveOrbs(t *testing.T) {
	ids := glowOrbTypeIDs4B6B80{
		heal: 1, drainMana: 2, charm: 3, white: 4, manaBomb: 5, whiteMove: 6, blueMove: 7,
	}
	tests := []struct {
		name       string
		typeID     int
		wantBright uint16
		wantDim    uint16
	}{
		{"heal", ids.heal, uint16(healOrbBright4B6B80), uint16(healOrbDim4B6B80)},
		{"drain mana", ids.drainMana, uint16(drainManaOrbBright4B6B80), uint16(drainManaOrbDim4B6B80)},
		{"charm", ids.charm, uint16(charmOrbBright4B6B80), uint16(charmOrbDim4B6B80)},
		{"white", ids.white, uint16(manaBombOrbBright4B6B80), uint16(manaBombOrbDim4B6B80)},
		{"mana bomb", ids.manaBomb, uint16(manaBombOrbBright4B6B80), uint16(manaBombOrbDim4B6B80)},
		{"white move", ids.whiteMove, uint16(manaBombOrbBright4B6B80), uint16(manaBombOrbDim4B6B80)},
		{"blue move", ids.blueMove, uint16(drainManaOrbBright4B6B80), uint16(drainManaOrbDim4B6B80)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bright, dim, ok := glowOrbColors4B6B80(tc.typeID, ids)
			if !ok || uint16(bright) != tc.wantBright || uint16(dim) != tc.wantDim {
				t.Fatalf("colors = %#x/%#x, ok=%t", bright, dim, ok)
			}
		})
	}
	bright, dim, ok := glowOrbColors4B6B80(99, ids)
	if !ok || bright != healOrbBright4B6B80 || dim != healOrbDim4B6B80 {
		t.Fatalf("default colors = %#x/%#x, ok=%t", bright, dim, ok)
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
