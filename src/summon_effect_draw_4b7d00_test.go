//go:build !server

package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

func TestDrawSummonEffect4B7D00HighAddress(t *testing.T) {
	anim := &summonAnimateDrawData4B7D00{frameCount: 4, frameDelay: 1, animationKind: 2}
	parent := &client.Drawable{
		PosVec:         image.Pt(100, 200),
		NetCode32:      3,
		AnimFrameSlave: 17,
		AnimStart:      90,
		DrawData:       unsafe.Pointer(anim),
	}
	child := &client.Drawable{}
	parent.UnionSummon().Child = child
	parent.UnionSummon().Lifetime = 20
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(parent)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(child)) <= uintptr(^uint32(0))) {
		t.Skipf("allocator returned a low drawable address: parent=%p child=%p", parent, child)
	}

	animateCalls := 0
	var childAlpha byte
	got := drawSummonEffect4B7D00(parent, summonEffectDrawHooks4B7D00{
		frame: func() uint32 { return 100 },
		animate: func(dr *client.Drawable) int {
			animateCalls++
			if dr != parent {
				t.Fatalf("animated %p, want parent %p", dr, parent)
			}
			if animateCalls > 1 && anim.animationKind != 5 {
				t.Fatalf("satellite animation kind = %d", anim.animationKind)
			}
			return 1
		},
		drawChild: func(dr *client.Drawable, alpha byte) int {
			if dr != child {
				t.Fatalf("draw child = %p, want %p", dr, child)
			}
			childAlpha = alpha
			return 1
		},
		pointSpark: func(image.Point) { t.Fatal("non-final frame emitted sparks") },
		trig: func(i int) image.Point {
			return image.Pt(i%3-1, 1-i%3)
		},
		deleteChild: func(*client.Drawable) { t.Fatal("live effect deleted child") },
		deleteSelf:  func(*client.Drawable) { t.Fatal("live effect deleted itself") },
	})
	if got != 1 || animateCalls != 27 || childAlpha != 127 {
		t.Fatalf("result=%d animateCalls=%d childAlpha=%d", got, animateCalls, childAlpha)
	}
	if parent.PosVec != image.Pt(100, 200) || parent.AnimFrameSlave != 17 || anim.animationKind != 2 {
		t.Fatalf("restored parent = pos:%v frame:%d kind:%d", parent.PosVec, parent.AnimFrameSlave, anim.animationKind)
	}
	if parent.UnionSummon().Child != child {
		t.Fatalf("native child pointer changed: got %p want %p", parent.UnionSummon().Child, child)
	}
}

func TestDrawSummonEffect4B7D00FinalAndExpired(t *testing.T) {
	parent := &client.Drawable{PosVec: image.Pt(20, 30), AnimStart: 100}
	child := &client.Drawable{}
	parent.UnionSummon().Child = child
	parent.UnionSummon().Lifetime = 10
	sparks, childDraws := 0, 0
	hooks := summonEffectDrawHooks4B7D00{
		frame:   func() uint32 { return 109 },
		animate: func(*client.Drawable) int { return 1 },
		drawChild: func(got *client.Drawable, alpha byte) int {
			childDraws++
			if got != child || alpha != 229 {
				t.Fatalf("final child draw = %p alpha %d", got, alpha)
			}
			return 1
		},
		pointSpark: func(pos image.Point) {
			sparks++
			if pos != image.Pt(20, 30) {
				t.Fatalf("spark position = %v", pos)
			}
		},
		trig:        func(int) image.Point { return image.Point{} },
		deleteChild: func(*client.Drawable) { t.Fatal("final live frame deleted child") },
		deleteSelf:  func(*client.Drawable) { t.Fatal("final live frame deleted self") },
	}
	if got := drawSummonEffect4B7D00(parent, hooks); got != 1 || sparks != 1 || childDraws != 1 {
		t.Fatalf("final frame result=%d sparks=%d childDraws=%d", got, sparks, childDraws)
	}

	deletedChild, deletedSelf := 0, 0
	hooks.frame = func() uint32 { return 110 }
	hooks.pointSpark = func(image.Point) { t.Fatal("expired frame emitted sparks") }
	hooks.animate = func(*client.Drawable) int { t.Fatal("expired frame animated"); return 0 }
	hooks.drawChild = func(*client.Drawable, byte) int { t.Fatal("expired frame drew child"); return 0 }
	hooks.deleteChild = func(got *client.Drawable) {
		deletedChild++
		if got != child {
			t.Fatalf("deleted child = %p, want %p", got, child)
		}
	}
	hooks.deleteSelf = func(got *client.Drawable) {
		deletedSelf++
		if got != parent {
			t.Fatalf("deleted parent = %p, want %p", got, parent)
		}
	}
	if got := drawSummonEffect4B7D00(parent, hooks); got != 0 || deletedChild != 1 || deletedSelf != 1 || parent.UnionSummon().Child != nil {
		t.Fatalf("expired result=%d childDeletes=%d selfDeletes=%d child=%p", got, deletedChild, deletedSelf, parent.UnionSummon().Child)
	}
}

func TestSummonEffectDrawDispatch4B7D00(t *testing.T) {
	fn := legacy.Get_nox_thing_summon_effect_draw()
	if fn == nil {
		t.Fatal("summon-effect callback pointer is nil")
	}
	dr := &client.Drawable{DrawFuncPtr: fn}
	if dr.DrawFuncPtr != fn {
		t.Fatal("summon-effect callback pointer was not retained at native width")
	}
	dr.DrawFuncPtr = nil
	if got, ok := (*Client)(nil).callSummonEffectDraw4B7D00(dr, nil); ok || got != 0 {
		t.Fatalf("non-summon dispatch = (%d, %t)", got, ok)
	}
}
