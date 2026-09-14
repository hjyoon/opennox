package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

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
