package client

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestDrawableCallDrawRejectsInvalidCallback(t *testing.T) {
	marker := 0
	dr := &Drawable{DrawFuncPtr: unsafe.Pointer(&marker)}
	if drawableDrawFuncCallable(dr.DrawFuncPtr) {
		t.Fatal("unregistered pointer was accepted as a drawable callback")
	}
	if got := dr.CallDraw(&noxrender.Viewport{}); got != 0 {
		t.Fatalf("invalid drawable callback result = %d, want 0", got)
	}
	if got := (*Drawable)(nil).CallDraw(&noxrender.Viewport{}); got != 0 {
		t.Fatalf("nil drawable callback result = %d, want 0", got)
	}
	if got := dr.CallDraw(nil); got != 0 {
		t.Fatalf("nil viewport callback result = %d, want 0", got)
	}
}

func TestDrawableDrawCallableAcceptsRegistryAndDefault(t *testing.T) {
	registeredMarker, defaultMarker := 0, 0
	registered := unsafe.Pointer(&registeredMarker)
	defaultDraw := unsafe.Pointer(&defaultMarker)
	drawFuncPtrs[registered] = struct{}{}
	t.Cleanup(func() { delete(drawFuncPtrs, registered) })

	oldDefault := ThingDrawDefault
	ThingDrawDefault = defaultDraw
	t.Cleanup(func() { ThingDrawDefault = oldDefault })

	if !drawableDrawFuncCallable(registered) {
		t.Fatal("registered drawable callback was rejected")
	}
	if !drawableDrawFuncCallable(defaultDraw) {
		t.Fatal("default drawable callback was rejected")
	}
	if drawableDrawFuncCallable(nil) {
		t.Fatal("nil drawable callback was accepted")
	}
}
