package opennox

import (
	"image"
	"testing"

	"github.com/opennox/opennox/v1/client/gui"
)

func TestSetNPCDialogWindowY479D30OnlyChangesOffsetY(t *testing.T) {
	win := &gui.Window{
		SizeVal: image.Pt(112, 24),
		Off:     image.Pt(391, 7),
		EndPos:  image.Pt(503, 31),
	}
	setNPCDialogWindowY479D30(win, 95)

	if got, want := win.Offs(), image.Pt(391, 95); got != want {
		t.Fatalf("offset = %v, want %v", got, want)
	}
	if got, want := win.Size(), image.Pt(112, 24); got != want {
		t.Fatalf("size = %v, want %v", got, want)
	}
	if got, want := win.End(), image.Pt(503, 31); got != want {
		t.Fatalf("end = %v, want %v", got, want)
	}
}
