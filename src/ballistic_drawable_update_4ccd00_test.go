package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func setBallisticDrawableState4CCD00(dr *client.Drawable, frame uint32, height, velocity, bounce float32) {
	effect := dr.UnionEffect()
	effect.Field_108 = frame
	effect.Field_109 = math.Float32bits(height)
	effect.Field_110 = math.Float32bits(velocity)
	effect.Field_111 = math.Float32bits(bounce)
}

func TestUpdateMeteorDrawable4CCD00HighAddress(t *testing.T) {
	dr := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	setBallisticDrawableState4CCD00(dr, 100, 10, 3, 0)
	if got := updateMeteorDrawable4CCD00(dr, func() uint32 { return 104 }); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	effect := dr.UnionEffect()
	if effect.Field_108 != 104 || math.Float32frombits(effect.Field_109) != 16 ||
		math.Float32frombits(effect.Field_110) != -1 || dr.ZVal != 16 || dr.VelZ != -1 {
		t.Fatalf("meteor state = frame:%d height:%v velocity:%v Z:%d VelZ:%d",
			effect.Field_108, math.Float32frombits(effect.Field_109),
			math.Float32frombits(effect.Field_110), dr.ZVal, dr.VelZ)
	}
}

func TestUpdateMeteorDrawable4CCD00Landing(t *testing.T) {
	dr := &client.Drawable{ZVal: 99, VelZ: 99}
	setBallisticDrawableState4CCD00(dr, 20, 2, -3, 0)
	updateMeteorDrawable4CCD00(dr, func() uint32 { return 21 })
	effect := dr.UnionEffect()
	if effect.Field_109 != 0 || effect.Field_110 != 0 || dr.ZVal != 0 || dr.VelZ != 0 {
		t.Fatalf("landed meteor = height:%#x velocity:%#x Z:%d VelZ:%d",
			effect.Field_109, effect.Field_110, dr.ZVal, dr.VelZ)
	}
}

func TestUpdateFistDrawable4CCDB0HighAddressBounce(t *testing.T) {
	dr := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	frame := uint32(101)
	setBallisticDrawableState4CCD00(dr, 100, 1, -2, 20)
	if got := updateFistDrawable4CCDB0(dr, func() uint32 { return frame }); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	effect := dr.UnionEffect()
	if effect.Field_108 != 101 || math.Float32frombits(effect.Field_109) != 0 ||
		math.Float32frombits(effect.Field_110) != 4 || dr.ZVal != 0 || dr.VelZ != 4 {
		t.Fatalf("bounced fist = frame:%d height:%v velocity:%v Z:%d VelZ:%d",
			effect.Field_108, math.Float32frombits(effect.Field_109),
			math.Float32frombits(effect.Field_110), dr.ZVal, dr.VelZ)
	}

	frame = 102
	updateFistDrawable4CCDB0(dr, func() uint32 { return frame })
	if math.Float32frombits(effect.Field_109) != 4 ||
		math.Float32frombits(effect.Field_110) != 3.5 || dr.ZVal != 4 || dr.VelZ != 3 {
		t.Fatalf("rising fist = height:%v velocity:%v Z:%d VelZ:%d",
			math.Float32frombits(effect.Field_109), math.Float32frombits(effect.Field_110), dr.ZVal, dr.VelZ)
	}
}

func TestUpdateFistDrawable4CCDB0StopsWeakBounce(t *testing.T) {
	dr := &client.Drawable{ZVal: 99, VelZ: 99}
	setBallisticDrawableState4CCD00(dr, 30, 1, -2, 8)
	updateFistDrawable4CCDB0(dr, func() uint32 { return 31 })
	effect := dr.UnionEffect()
	if effect.Field_109 != 0 || effect.Field_110 != 0 || dr.ZVal != 0 || dr.VelZ != 0 {
		t.Fatalf("weak bounce = height:%#x velocity:%#x Z:%d VelZ:%d",
			effect.Field_109, effect.Field_110, dr.ZVal, dr.VelZ)
	}
}
