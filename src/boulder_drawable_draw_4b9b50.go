//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func init() {
	legacy.Nox_thing_boulder_draw = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.drawBoulder4B9B50(dr, vp)
	}
}

// boulderDrawFrame4B9B50 advances the 32-frame rolling animation and returns
// the image index. Keeping this state access in Go avoids interpreting the
// native Drawable through the original PE32 byte offsets.
func boulderDrawFrame4B9B50(dr *client.Drawable) int {
	effect := dr.UnionEffect()
	if effect.Field_108 == 0 && effect.Field_109 == 0 {
		effect.Field_108 = uint32(dr.PosVec.X)
		effect.Field_109 = uint32(dr.PosVec.Y)
	}

	dx := int32(uint32(dr.PosVec.X) - effect.Field_108)
	dy := int32(uint32(dr.PosVec.Y) - effect.Field_109)
	if int64(dx)*int64(dx)+int64(dy)*int64(dy) < 100 {
		return int(effect.Field_110 + effect.Field_111)
	}

	if dx <= 0 {
		if dy > 0 {
			effect.Field_111 = 0
			advanceBoulderFrame4B9B50(effect)
		} else {
			effect.Field_111 = 16
			retreatBoulderFrame4B9B50(effect)
		}
	} else if dy > 0 {
		effect.Field_111 = 16
		advanceBoulderFrame4B9B50(effect)
	} else {
		effect.Field_111 = 0
		retreatBoulderFrame4B9B50(effect)
	}

	effect.Field_108 = uint32(dr.PosVec.X)
	effect.Field_109 = uint32(dr.PosVec.Y)
	return int(effect.Field_110 + effect.Field_111)
}

func advanceBoulderFrame4B9B50(effect *client.DrawableUnionEffect) {
	effect.Field_110++
	if effect.Field_110 >= 16 {
		effect.Field_110 = 0
	}
}

func retreatBoulderFrame4B9B50(effect *client.DrawableUnionEffect) {
	if effect.Field_110 != 0 {
		effect.Field_110--
	} else {
		effect.Field_110 = 15
	}
}

func (c *Client) drawBoulder4B9B50(dr *client.Drawable, vp *noxrender.Viewport) int {
	if dr == nil {
		return 0
	}
	frame := boulderDrawFrame4B9B50(dr)
	if img, ok := staticRandomDrawImage(dr.DrawData, frame); ok {
		legacy.Nox_xxx_drawObject_4C4770_draw(vp, dr, img)
	}
	return 1
}

func (c *Client) callBoulderDraw4B9B50(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	if dr == nil || dr.DrawFuncPtr != legacy.Get_nox_thing_boulder_draw() {
		return 0, false
	}
	return c.drawBoulder4B9B50(dr, vp), true
}
