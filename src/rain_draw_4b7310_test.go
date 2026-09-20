//go:build !server

package opennox

import (
	"image"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func requireNativeDrawableAddress4B7310(t *testing.T, drawables ...*client.Drawable) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for _, dr := range drawables {
		if uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
			t.Skipf("allocator returned a low drawable address: %p", dr)
		}
	}
}

func TestRainDrawKindFor4B7310(t *testing.T) {
	tests := []struct {
		name string
		fn   unsafe.Pointer
		want rainDrawKind4B7310
	}{
		{"blue rain", legacy.Get_nox_thing_blue_rain_draw(), rainDrawBlueRain4B7810},
		{"level up", legacy.Get_nox_thing_levelup_draw(), rainDrawLevelUp4B7700},
		{"oblivion up", legacy.Get_nox_thing_oblivion_up_draw(), rainDrawOblivionUp4B77D0},
		{"rain orb", legacy.Get_nox_thing_rain_orb_draw(), rainDrawOrb4B7310},
		{"unknown", nil, rainDrawUnknown4B7310},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := rainDrawKindFor4B7310(tc.fn); got != tc.want {
				t.Fatalf("kind = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSpawnBlueRain4B7810HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(300, 400)}
	sparks := []*client.Drawable{{}, {}}
	requireNativeDrawableAddress4B7310(t, source, sparks[0], sparks[1])
	vp := &noxrender.Viewport{World: image.Rect(100, 200, 740, 680)}
	randomValues := []int{-4, 5, 100, 2, -3, 110}
	randomRanges := [][2]int{{-10, 10}, {-10, 10}, {90, 120}, {-10, 10}, {-10, 10}, {90, 120}}
	randomIndex, frameCalls, typeCalls, spawnCalls, activateCalls := 0, 0, 0, 0, 0
	hooks := rainSpawnHooks4B7310{
		typeID: func(name string) int {
			typeCalls++
			if name != "BlueRainSpark" {
				t.Fatalf("type name = %q", name)
			}
			return 71
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			wantPositions := []image.Point{{X: 296, Y: 405}, {X: 302, Y: 397}}
			if typ != 71 || pos != wantPositions[spawnCalls] {
				t.Fatalf("spawn %d = type %d at %v", spawnCalls, typ, pos)
			}
			dr := sparks[spawnCalls]
			spawnCalls++
			return dr
		},
		random: func(minimum, maximum int) int {
			if got := [2]int{minimum, maximum}; got != randomRanges[randomIndex] {
				t.Fatalf("random range %d = %v, want %v", randomIndex, got, randomRanges[randomIndex])
			}
			value := randomValues[randomIndex]
			randomIndex++
			return value
		},
		frame: func() uint32 {
			value := uint32(700 + frameCalls)
			frameCalls++
			return value
		},
		activate: func(*client.Drawable) { activateCalls++ },
	}
	if got := spawnBlueRain4B7810(source, vp, false, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if typeCalls != 1 || spawnCalls != 2 || activateCalls != 2 || randomIndex != 6 || frameCalls != 4 {
		t.Fatalf("calls = type:%d spawn:%d activate:%d random:%d frame:%d", typeCalls, spawnCalls, activateCalls, randomIndex, frameCalls)
	}
	checks := []struct {
		pos     image.Point
		z       uint16
		start   uint32
		expires uint32
	}{
		{image.Pt(296, 405), 205, 701, 800},
		{image.Pt(302, 397), 197, 703, 812},
	}
	for i, spark := range sparks {
		effect := spark.UnionEffect()
		want := checks[i]
		if effect.Field_108 != uint32(want.pos.X)<<12 || effect.Field_109 != uint32(want.pos.Y)<<12 ||
			effect.Field_110 != 0 || effect.Field_111 != want.start || effect.Field_112 != want.expires ||
			spark.Field_74_4 != 0 || spark.ZVal != want.z || spark.ZVal2 != 0 || spark.VelZ != -5 {
			t.Fatalf("spark %d not initialized: drawable=%+v effect=%+v", i, spark, effect)
		}
	}

	panicHooks := rainSpawnHooks4B7310{typeID: func(string) int { panic("disabled rain performed work") }}
	if got := spawnBlueRain4B7810(source, vp, true, panicHooks); got != 1 {
		t.Fatalf("disabled result = %d, want 1", got)
	}
}

func TestSpawnFallingRainOrbs4B7740HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(500, 600)}
	orbs := []*client.Drawable{{}, {}}
	orbs[0].UnionEffect().Field_110 = 0xab123456
	orbs[1].UnionEffect().Field_110 = 0xcd123456
	requireNativeDrawableAddress4B7310(t, source, orbs[0], orbs[1])
	vp := &noxrender.Viewport{World: image.Rect(100, 200, 740, 680)}
	randomValues := []int{-10, 10, 8, 3, 15, -15, 12, 10}
	randomRanges := [][2]int{{-15, 15}, {-15, 15}, {8, 12}, {3, 10}, {-15, 15}, {-15, 15}, {8, 12}, {3, 10}}
	randomIndex, spawnCalls, activateCalls := 0, 0, 0
	hooks := rainSpawnHooks4B7310{
		typeID: func(name string) int {
			if name != "RainOrbWhite" {
				t.Fatalf("type name = %q", name)
			}
			return 72
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			wantPositions := []image.Point{{X: 490, Y: 610}, {X: 515, Y: 585}}
			if typ != 72 || pos != wantPositions[spawnCalls] {
				t.Fatalf("spawn %d = type %d at %v", spawnCalls, typ, pos)
			}
			dr := orbs[spawnCalls]
			spawnCalls++
			return dr
		},
		random: func(minimum, maximum int) int {
			if got := [2]int{minimum, maximum}; got != randomRanges[randomIndex] {
				t.Fatalf("random range %d = %v, want %v", randomIndex, got, randomRanges[randomIndex])
			}
			value := randomValues[randomIndex]
			randomIndex++
			return value
		},
		activate: func(*client.Drawable) { activateCalls++ },
	}
	if got := spawnFallingRainOrbs4B7740(source, vp, "RainOrbWhite", hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantZ := []uint16{410, 385}
	wantVelocity := []int8{-8, -12}
	wantField110 := []uint32{0xab03019a, 0xcd0a0181}
	for i, orb := range orbs {
		effect := orb.UnionEffect()
		if orb.ZVal != wantZ[i] || orb.ZVal2 != 0 || orb.VelZ != wantVelocity[i] ||
			effect.Field_108 != 500 || effect.Field_109 != 600 || effect.Field_110 != wantField110[i] {
			t.Fatalf("orb %d not initialized: drawable=%+v effect=%+v", i, orb, effect)
		}
	}
	if spawnCalls != 2 || activateCalls != 2 || randomIndex != len(randomValues) {
		t.Fatalf("calls = spawn:%d activate:%d random:%d", spawnCalls, activateCalls, randomIndex)
	}
}

func TestAdvanceRainOrbFall4B7310HighAddress(t *testing.T) {
	dr := &client.Drawable{ZVal: 20, VelZ: -5}
	requireNativeDrawableAddress4B7310(t, dr)
	dr.UnionEffect().Field_110 = 0xab07000c
	z, radius, delta := advanceRainOrbFall4B7310(dr)
	if z != 20 || radius != 7 || delta != 8 || dr.ZVal != 15 || dr.UnionEffect().Field_110 != 0xab070014 {
		t.Fatalf("fall = z:%d radius:%d delta:%d next:%d packed:%#x", z, radius, delta, dr.ZVal, dr.UnionEffect().Field_110)
	}
}

func TestFinishRainOrb4B7310HighAddress(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(300, 400), TypeIDVal: 77}
	move := &client.Drawable{}
	move.UnionEffect().Field_110 = 0x00aabbcc
	move.UnionEffect().Field_111 = 0xdd112233
	source.UnionEffect().Field_108 = uint32(100)
	source.UnionEffect().Field_109 = uint32(200)
	requireNativeDrawableAddress4B7310(t, source, move)
	randomCalls, activated, deleted := 0, 0, 0
	hooks := rainOrbMoveHooks4B7310{
		typeID: func(name string) int {
			if name != "WhiteMoveOrb" {
				t.Fatalf("type name = %q", name)
			}
			return 88
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 88 || pos != image.Pt(300, 420) {
				t.Fatalf("spawn = type %d at %v", typ, pos)
			}
			return move
		},
		random: func(minimum, maximum int) int {
			randomCalls++
			switch [2]int{minimum, maximum} {
			case [2]int{6, 8}:
				return 7
			case [2]int{3, 10}:
				return 9
			default:
				t.Fatalf("random range = (%d, %d)", minimum, maximum)
				return 0
			}
		},
		activate: func(got *client.Drawable) {
			activated++
			if got != move {
				t.Fatalf("activated %p, want %p", got, move)
			}
		},
		delete: func(got *client.Drawable) {
			deleted++
			if got != source {
				t.Fatalf("deleted %p, want %p", got, source)
			}
		},
		direction: func(vector types.Pointf) byte {
			if vector != types.Ptf(200, 200) {
				t.Fatalf("direction vector = %+v", vector)
			}
			return 64
		},
		sinCos: func(direction byte) (float32, float32) {
			if direction != 64 {
				t.Fatalf("direction = %d", direction)
			}
			return 0.25, -0.5
		},
		round: func(value float32) int32 { return int32(math.RoundToEven(float64(value))) },
	}
	if got := finishRainOrb4B7310(source, 77, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	effect := move.UnionEffect()
	if effect.Field_108 != uint32(138)|uint32(125)<<16 || effect.Field_110 != 0x07aabbcc ||
		effect.Field_111 != 0xdd010109 || randomCalls != 2 || activated != 1 || deleted != 1 {
		t.Fatalf("move orb not initialized: effect=%+v random=%d activated=%d deleted=%d", effect, randomCalls, activated, deleted)
	}
}

func TestFinishBlueRainOrbDeletesSourceWhenSpawnFails4B7310(t *testing.T) {
	source := &client.Drawable{PosVec: image.Pt(30, 40), TypeIDVal: 78}
	source.UnionEffect().Field_108 = 30
	source.UnionEffect().Field_109 = 40
	requireNativeDrawableAddress4B7310(t, source)
	deleted := 0
	hooks := rainOrbMoveHooks4B7310{
		typeID: func(name string) int {
			if name != "BlueMoveOrb" {
				t.Fatalf("type name = %q", name)
			}
			return 89
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 89 || pos != image.Pt(30, 60) {
				t.Fatalf("spawn = type %d at %v", typ, pos)
			}
			return nil
		},
		random: func(minimum, maximum int) int {
			if minimum != 6 || maximum != 8 {
				t.Fatalf("unexpected random range (%d, %d)", minimum, maximum)
			}
			return 6
		},
		activate: func(*client.Drawable) { t.Fatal("nil move orb activated") },
		delete: func(got *client.Drawable) {
			deleted++
			if got != source {
				t.Fatalf("deleted %p, want %p", got, source)
			}
		},
		direction: func(types.Pointf) byte { return 0 },
		sinCos:    func(byte) (float32, float32) { return 1, 0 },
		round:     manaBombCancelFloatToInt48EA70,
	}
	finishRainOrb4B7310(source, 77, hooks)
	if deleted != 1 {
		t.Fatalf("delete calls = %d, want 1", deleted)
	}
}
