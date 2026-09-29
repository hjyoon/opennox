package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/opennox/v1/client"
)

const (
	gauntletHideStaticDrawable48EA70 = 0x0f
	gauntletGreenZap48EA70           = 0x10
)

type gauntletDrawableHooks48EA70 struct {
	connected    func() bool
	lookupStatic func(code uint16) *client.Drawable
	typeID       func() int
	spawn        func(typ int, pos image.Point) *client.Drawable
	decay        func(dr *client.Drawable, lifetime int)
}

// handleGauntletDrawableNative48EA70 handles the two MSG_GAUNTLET branches
// that store a drawable pointer in a PE32 int in the legacy decoder. Other
// subcommands deliberately remain in the legacy switch until they are ported.
func handleGauntletDrawableNative48EA70(data []byte, hooks gauntletDrawableHooks48EA70) (int, bool) {
	if len(data) < 2 {
		return -1, true
	}
	switch data[1] {
	case gauntletHideStaticDrawable48EA70:
		if len(data) < 4 {
			return -1, true
		}
		if hooks.connected() {
			if dr := hooks.lookupStatic(binary.LittleEndian.Uint16(data[2:4])); dr != nil {
				dr.UnionEffect().Field_108 &= 0xffffff00
			}
		}
		return 4, true
	case gauntletGreenZap48EA70:
		if len(data) < 12 {
			return -1, true
		}
		typ := hooks.typeID()
		if !hooks.connected() {
			return 12, true
		}
		state := greenBoltFXState48EA70{
			From: image.Pt(
				int(binary.LittleEndian.Uint16(data[2:4])),
				int(binary.LittleEndian.Uint16(data[4:6])),
			),
			To: image.Pt(
				int(binary.LittleEndian.Uint16(data[6:8])),
				int(binary.LittleEndian.Uint16(data[8:10])),
			),
			Duration: binary.LittleEndian.Uint16(data[10:12]),
		}
		if dr := hooks.spawn(typ, state.To); dr != nil {
			setGreenBoltFXPayloadNative48EA70(dr, state)
			hooks.decay(dr, int(state.Duration))
		}
		return 12, true
	default:
		return 0, false
	}
}

func (c *Client) handleGauntletDrawablePacketNative48EA70(data []byte) (int, bool) {
	return handleGauntletDrawableNative48EA70(data, gauntletDrawableHooks48EA70{
		connected:    nox_client_isConnected,
		lookupStatic: func(code uint16) *client.Drawable { return c.Objs.ByNetCodeStatic(int(code)) },
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxGauntletGreenZapType, "GreenZap", c.Things.IndByID,
			)
		},
		spawn: c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		decay: c.Objs.TransparentDecay,
	})
}
