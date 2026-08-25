package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

func TestSpriteQueuePredicatesNative4756E0(t *testing.T) {
	static := legacy.Get_nox_thing_static_draw()
	if static == nil {
		t.Fatal("static draw pointer is nil")
	}
	dr := &client.Drawable{DrawFuncPtr: static, ObjFlags: object.Flags(1 | 0x40)}
	if !spriteBackWallQueueNative4756E0(dr) || spriteFrontQueueNative475740(dr) {
		t.Fatal("static wall drawable was not assigned exclusively to the back-wall queue")
	}
	dummyDraw := byte(0)
	dr.DrawFuncPtr = unsafe.Pointer(&dummyDraw)
	if spriteBackWallQueueNative4756E0(dr) || !spriteFrontQueueNative475740(dr) {
		t.Fatal("non-static front drawable was assigned to the wrong queue")
	}
	dr.ObjFlags = object.Flags(0x4000 | 0x40)
	if !spriteOverlayQueueNative4757A0(dr) {
		t.Fatal("overlay drawable was not recognized before the world queue")
	}
	dr.ObjFlags = 0
	if !spriteWorldQueueNative4757D0(dr) {
		t.Fatal("ordinary drawable was not assigned to the world queue")
	}
	dr.ObjFlags = object.Flags(0x1000)
	if spriteBackWallQueueNative4756E0(dr) || spriteFrontQueueNative475740(dr) || spriteOverlayQueueNative4757A0(dr) || spriteWorldQueueNative4757D0(dr) {
		t.Fatal("hidden drawable entered a render queue")
	}
}
