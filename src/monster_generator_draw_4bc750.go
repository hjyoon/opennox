//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

// Both normal world dispatch and direct C callbacks now execute the same
// complete native body, so Quest previews cannot fall back to the PE32 drawer.
func (c *Client) drawMonsterGenerator4BC750(dr *client.Drawable, vp *noxrender.Viewport) int {
	return int(legacy.DrawMonsterGenerator4BC750(vp, dr))
}

func (c *Client) callMonsterGeneratorDraw4BC750(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	if dr == nil || dr.DrawFuncPtr != legacy.Get_nox_thing_monster_gen_draw() {
		return 0, false
	}
	return c.drawMonsterGenerator4BC750(dr, vp), true
}
