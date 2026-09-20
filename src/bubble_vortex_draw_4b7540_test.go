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

func TestBubbleVortexDrawKindFor4B7540(t *testing.T) {
	tests := []struct {
		name string
		fn   unsafe.Pointer
		want bubbleVortexDrawKind4B7540
	}{
		{"bubble", legacy.Get_nox_thing_bubble_draw(), bubbleVortexDrawBubble4B7540},
		{"vortex", legacy.Get_nox_thing_vortex_draw(), bubbleVortexDrawVortex4B9F50},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fn == nil {
				t.Fatal("callback pointer is nil")
			}
			if got := bubbleVortexDrawKindFor4B7540(tc.fn); got != tc.want {
				t.Fatalf("kind = %d, want %d", got, tc.want)
			}
		})
	}
	if got := bubbleVortexDrawKindFor4B7540(nil); got != bubbleVortexDrawUnknown4B7540 {
		t.Fatalf("nil kind = %d", got)
	}
}

func TestBubbleDrawNativeState4B7540HighAddress(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(150, 260), ZVal: 12, Deadline: 100}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	effect := dr.UnionEffect()
	effect.Field_108 = 0x11112222
	effect.Field_109 = 0x33334444
	effect.Field_110 = 4 | 1<<8 | 1<<16 | 7<<24
	effect.Field_111 = 3 | 1<<8 | uint32(byte(0xfe))<<16
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(100, 200, 740, 680),
	}
	visual := bubbleDrawVisualFor4B7540(dr, vp)
	if visual.point != image.Pt(60, 68) || visual.baseRadius != 4 ||
		visual.glowColor != effect.Field_108 || visual.centerColor != effect.Field_109 {
		t.Fatalf("visual = %+v", visual)
	}
	startDecay, remove := prepareBubbleDraw4B7540(dr, 100)
	if !startDecay || remove || packedEffectByte4B7540(effect.Field_110, 8) != 3 ||
		packedEffectByte4B7540(effect.Field_110, 16) != 4 || packedEffectByte4B7540(effect.Field_110, 24) != 4 {
		t.Fatalf("prepare = decay %t remove %t state %#x", startDecay, remove, effect.Field_110)
	}
}

func TestAdvanceBubbleDraw4B7540StateMachine(t *testing.T) {
	t.Run("grow and reverse", func(t *testing.T) {
		dr := &client.Drawable{ZVal: 10}
		effect := dr.UnionEffect()
		effect.Field_110 = 4 | 1<<8 | 1<<16 | 7<<24
		if advanceBubbleDraw4B7540(dr, 4) {
			t.Fatal("growing bubble was removed")
		}
		if got := effect.Field_110; got != 5|2<<8|7<<16|7<<24 {
			t.Fatalf("grow state = %#x", got)
		}
	})
	t.Run("shrink and reverse", func(t *testing.T) {
		dr := &client.Drawable{ZVal: 10}
		effect := dr.UnionEffect()
		effect.Field_110 = 1 | 2<<8 | 1<<16 | 6<<24
		if advanceBubbleDraw4B7540(dr, 4) {
			t.Fatal("oscillating bubble was removed")
		}
		if got := effect.Field_110; got != 0|1<<8|6<<16|6<<24 {
			t.Fatalf("shrink state = %#x", got)
		}
	})
	t.Run("decay removal", func(t *testing.T) {
		dr := &client.Drawable{ZVal: 10}
		dr.UnionEffect().Field_110 = 1 | 3<<8 | 1<<16 | 4<<24
		if !advanceBubbleDraw4B7540(dr, 4) {
			t.Fatal("zero-radius decaying bubble survived")
		}
	})
	t.Run("vertical oscillation and signed altitude", func(t *testing.T) {
		dr := &client.Drawable{ZVal: 1}
		effect := dr.UnionEffect()
		effect.Field_111 = 3 | 1<<8 | uint32(byte(0xff))<<16
		if advanceBubbleDraw4B7540(dr, 1) {
			t.Fatal("zero-altitude bubble was removed")
		}
		if dr.ZVal != 0 || packedEffectByte4B7540(effect.Field_111, 8) != 3 ||
			packedEffectByte4B7540(effect.Field_111, 16) != 1 {
			t.Fatalf("vertical state = z %d packed %#x", dr.ZVal, effect.Field_111)
		}
		effect.Field_111 = setPackedEffectByte4B7540(effect.Field_111, 16, 0xff)
		if !advanceBubbleDraw4B7540(dr, 1) || int16(dr.ZVal) != -1 {
			t.Fatalf("negative altitude = %d", int16(dr.ZVal))
		}
	})
}

func TestVortexDrawStateFor4B9F50HighAddress(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(100, 200), ZVal: 8}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	effect := dr.UnionEffect()
	effect.Field_108 = 0x11112222
	effect.Field_109 = 0x33334444
	effect.Field_110 = 100
	effect.Field_111 = 200
	effect.Field_112 = 0 | 2<<8 | 50<<16 | 1<<24
	vp := &noxrender.Viewport{
		Screen: image.Rect(10, 20, 650, 500),
		World:  image.Rect(50, 100, 690, 580),
	}
	state := vortexDrawStateFor4B9F50(dr, vp)
	wantPoint := vp.ToScreenPos(image.Pt(150, 200)).Add(image.Pt(0, -8))
	tailDir := sincosTable16[252]
	wantTail := vp.ToScreenPos(image.Pt(100+50*tailDir.X/16, 200+50*tailDir.Y/16)).Add(image.Pt(0, -8))
	if !state.inside || !state.alive || state.point != wantPoint || state.tail != wantTail ||
		state.pointColor != effect.Field_108 || state.tailColor != vortexGray4B9F50 ||
		state.glowSize != 3 || state.pointRadius != 3 || state.nextAngle != 2 ||
		state.nextRadius != 49 || state.nextZ != 9 {
		t.Fatalf("vortex state = %+v, want point %v tail %v", state, wantPoint, wantTail)
	}
	applyVortexDrawState4B9F50(dr, state)
	if dr.ZVal != 9 || effect.Field_112 != 2|2<<8|49<<16|1<<24 {
		t.Fatalf("applied vortex = z %d motion %#x", dr.ZVal, effect.Field_112)
	}
}

func TestVortexDrawStateFor4B9F50BoundsAndExpiry(t *testing.T) {
	dr := &client.Drawable{ZVal: 399}
	effect := dr.UnionEffect()
	effect.Field_110 = 100
	effect.Field_111 = 100
	effect.Field_112 = 0 | uint32(byte(0xfe))<<8 | 50<<16 | 1<<24
	vp := &noxrender.Viewport{
		Screen: image.Rect(-1000, -1000, 1000, 1000),
		World:  image.Rect(-1000, -1000, 1000, 1000),
	}
	state := vortexDrawStateFor4B9F50(dr, vp)
	if !state.inside || state.alive || state.nextAngle != 254 || state.nextZ != 400 {
		t.Fatalf("expiry state = %+v", state)
	}
	applyVortexDrawState4B9F50(dr, state)
	if effect.Field_112 != 254|uint32(byte(0xfe))<<8|50<<16|1<<24 || dr.ZVal != 400 {
		t.Fatalf("expired apply = motion %#x z %d", effect.Field_112, dr.ZVal)
	}

	vp.Screen = image.Rect(150, 0, 300, 300)
	vp.World = image.Rect(0, 0, 150, 300)
	dr.ZVal = 0
	state = vortexDrawStateFor4B9F50(dr, vp)
	if state.inside {
		t.Fatalf("boundary point unexpectedly inside: %+v", state)
	}
}
