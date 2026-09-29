package opennox

import (
	"encoding/binary"
	"fmt"
	"image"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	gauntletHideStaticDrawable48EA70 = 0x0f
	gauntletGreenZap48EA70           = 0x10
	gauntletQuestMessage48EA70       = 0x21
	gauntletQuestMessageSize48EA70   = 52
	gauntletQuestMessageKeyEnd48EA70 = 51
)

const gauntletQuestMessageSource48EA70 = `C:\NoxPost\src\Client\Network\cdecode.c`

var gauntletQuestItemKeys48EA70 = [...]string{
	"objcoll.c:NullKey",
	"objcoll.c:SilverKey",
	"objcoll.c:GoldKey",
	"objcoll.c:RubyKey",
	"objcoll.c:SapphireKey",
}

type gauntletPacketHooks48EA70 struct {
	connected    func() bool
	lookupStatic func(code uint16) *client.Drawable
	typeID       func() int
	spawn        func(typ int, pos image.Point) *client.Drawable
	decay        func(dr *client.Drawable, lifetime int)
	loadString   func(key, source string) string
	print        func(text string)
}

// handleGauntletPacketNative48EA70 handles MSG_GAUNTLET branches that retain
// pointers in PE32-sized locals in the legacy decoder. Other subcommands
// deliberately remain in the legacy switch until they are ported.
func handleGauntletPacketNative48EA70(data []byte, hooks gauntletPacketHooks48EA70) (int, bool) {
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
	case gauntletQuestMessage48EA70:
		if len(data) < gauntletQuestMessageSize48EA70 {
			return -1, true
		}
		if hooks.connected() {
			item := int(data[gauntletQuestMessageKeyEnd48EA70])
			if item >= len(gauntletQuestItemKeys48EA70) {
				return -1, true
			}
			itemName := hooks.loadString(gauntletQuestItemKeys48EA70[item], gauntletQuestMessageSource48EA70)
			formatKey := alloc.GoStringS(data[2:gauntletQuestMessageKeyEnd48EA70])
			format := hooks.loadString(formatKey, gauntletQuestMessageSource48EA70)
			hooks.print(fmt.Sprintf(format, itemName))
		}
		return gauntletQuestMessageSize48EA70, true
	default:
		return 0, false
	}
}

func (c *Client) handleGauntletPacketNative48EA70(data []byte) (int, bool) {
	return handleGauntletPacketNative48EA70(data, gauntletPacketHooks48EA70{
		connected:    nox_client_isConnected,
		lookupStatic: func(code uint16) *client.Drawable { return c.Objs.ByNetCodeStatic(int(code)) },
		typeID: func() int {
			return resolvePointSpriteFXTypeNative48EA70(
				&c.fxGauntletGreenZapType, "GreenZap", c.Things.IndByID,
			)
		},
		spawn: c.Nox_xxx_spriteLoadAdd_45A360_drawable,
		decay: c.Objs.TransparentDecay,
		loadString: func(key, source string) string {
			return c.Strings().GetStringInFile(strman.ID(key), source)
		},
		print: nox_xxx_printCentered_445490,
	})
}
