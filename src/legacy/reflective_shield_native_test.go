package legacy

import (
	"image"
	"testing"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestReflectiveShieldPreservesSignedDirectionalOffsets(t *testing.T) {
	loadOriginalRenderTable(t, 161776, 9*8)
	shield, free := alloc.New(client.Drawable{})
	defer free()
	owner := &client.Drawable{PosVec: image.Pt(300, 400), ZVal: 7}
	view := &noxrender.Viewport{}
	// Signed coordinate pairs at GAME.EXE 005AE7F0.
	offsets := [...]image.Point{
		{X: -18, Y: -18}, {Y: -18}, {X: 18, Y: -18},
		{X: -18}, {}, {X: 18},
		{X: -18, Y: 18}, {Y: 18}, {X: 18, Y: 18},
	}
	for direction, offset := range offsets {
		if direction == 4 {
			continue
		}
		slot := reflectiveShieldSlot(direction)
		saved := *slot
		*slot = (*nox_drawable)(shield.C())
		owner.AnimDir = byte(direction)
		Nox_xxx_drawShield_499810(view, owner)
		*slot = saved
		want := owner.Pos().Add(offset).Add(image.Pt(0, int(owner.ZVal)))
		if got := shield.Pos(); got != want {
			t.Errorf("direction %d draws shield at %v, want %v", direction, got, want)
		}
	}
}
