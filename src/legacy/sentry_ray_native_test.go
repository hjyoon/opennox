//go:build !server

package legacy

import (
	"image"
	"testing"

	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestSentryRayRendererUsesNativeViewportCoordinates(t *testing.T) {
	vp := noxrender.Viewport{
		Screen: image.Rect(111, 222, 911, 822),
		World:  image.Rect(1000, 2000, 1800, 2600),
		Size:   image.Pt(800, 600),
	}
	from := image.Pt(1234, 2345)
	to := image.Pt(1567, 2456)
	gotFrom, gotTo := clientSentryRayScreenPosition4C5060(&vp, from, to)
	if want := vp.ToScreenPos(from); gotFrom != want {
		t.Fatalf("from screen position = %v, want %v", gotFrom, want)
	}
	if want := vp.ToScreenPos(to); gotTo != want {
		t.Fatalf("to screen position = %v, want %v", gotTo, want)
	}
}

func TestSentryRayQueuePreservesWireCoordinates(t *testing.T) {
	ClearSentryRays4C5050()
	t.Cleanup(ClearSentryRays4C5050)
	from := image.Pt(65535, 32768)
	to := image.Pt(0x1234, 0x5678)
	AddSentryRay4C5020(from, to)
	if got := SentryRayCount4C5020(); got != 1 {
		t.Fatalf("sentry-ray count = %d, want 1", got)
	}
	gotFrom, gotTo, ok := SentryRayAt4C5020(0)
	if !ok || gotFrom != from || gotTo != to {
		t.Fatalf("queued ray = %v -> %v, ok=%t; want %v -> %v", gotFrom, gotTo, ok, from, to)
	}
	if _, _, ok := SentryRayAt4C5020(1); ok {
		t.Fatal("out-of-range sentry ray was returned")
	}
}
