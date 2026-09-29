package opennox

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const (
	objectXStatusPacketSize48EA70       = 7
	playerStatusPacketSize48EA70        = 5
	objectModifierPacketSize48EA70      = 7
	statModifierPacketSize48EA70        = 8
	npcReportPacketSize48EA70           = 21
	clientStatusReportPacketSize48EA70  = 7
	animationFramePacketSize48EA70      = 7
	clientStatusReportMask48EA70        = uint32(0x423)
	clientStatusReportClearMask48EA70   = uint32(1059)
	monsterGeneratorStartingBit48EA70   = uint32(0x400)
	monsterGeneratorTerminatedBit48EA70 = uint32(0x800)
)

type objectStatusPacketHooks48EA70 struct {
	connected        func() bool
	byNetCode        func(uint16) *client.Drawable
	localDrawable    func() *client.Drawable
	modifierByID     func(byte) *server.ModifierEff
	prepareGenerator func(*client.Drawable)
	localNetCode     func() uint32
	setStatModifier  func(byte, float32)
}

type npcReportPacketHooks48EA70 struct {
	connected func() bool
	byID      func(int) *server.NPC
	newNPC    func(int) *server.NPC
	resetNPC  func(*server.NPC, int)
	color     func(byte, byte, byte) uint32
}

type clientStatusPacketHooks48EA70 struct {
	playerByID        func(int) *server.Player
	host              func() bool
	unsetStatus       func(*server.Player, uint32)
	needStatus        func(*server.Player, uint32)
	renderingDisabled func() bool
	localNetCode      func() uint32
	onClientStatus    func(uint32)
	setPoisoned       func(int)
}

func objectStatusPacketSize48EA70(op netmsg.Op) (int, bool) {
	switch op {
	case netmsg.MSG_REPORT_X_STATUS:
		return objectXStatusPacketSize48EA70, true
	case netmsg.MSG_REPORT_PLAYER_STATUS:
		return playerStatusPacketSize48EA70, true
	case netmsg.MSG_REPORT_MODIFIER:
		return objectModifierPacketSize48EA70, true
	case netmsg.MSG_REPORT_STAT_MODIFIER:
		return statModifierPacketSize48EA70, true
	case netmsg.MSG_REPORT_ANIMATION_FRAME:
		return animationFramePacketSize48EA70, true
	default:
		return 0, false
	}
}

func drawableItemModifiers48EA70(dr *client.Drawable) *client.DrawableUnionItem {
	if dr == nil {
		return nil
	}
	return (*client.DrawableUnionItem)(unsafe.Pointer(&dr.Union))
}

// prepareMonsterGeneratorStatus48EA70 preserves sub_4BC720 without sending a
// native-width Drawable through the legacy PE32 callback boundary.
func prepareMonsterGeneratorStatus48EA70(dr *client.Drawable) {
	if dr == nil || dr.DrawData == nil {
		return
	}
	data := unsafe.Slice((*byte)(dr.DrawData), 33)
	dr.UnionEffect().Field_108 = uint32(data[27]) * (uint32(data[32]) + 1)
}

func setDrawableModifiers48EA70(dr *client.Drawable, modifiers [4]*server.ModifierEff) {
	item := drawableItemModifiers48EA70(dr)
	if item == nil {
		return
	}
	if modifiers[0] != nil {
		item.Field_108 = modifiers[0].C()
	} else {
		item.Field_108 = nil
	}
	if modifiers[1] != nil {
		item.Field_109 = modifiers[1].C()
	} else {
		item.Field_109 = nil
	}
	if modifiers[2] != nil {
		item.Field_110 = modifiers[2].C()
	} else {
		item.Field_110 = nil
	}
	if modifiers[3] != nil {
		item.Field_111 = modifiers[3].C()
	} else {
		item.Field_111 = nil
	}
	item.Field_112_0 = -1
	item.Field_112_2 = -1
}

// handleObjectStatusNative48EA70 implements the Drawable-bearing subset of
// cdecode.c opcodes 101..107 using typed native-width fields.
func handleObjectStatusNative48EA70(op netmsg.Op, data []byte, hooks objectStatusPacketHooks48EA70) int {
	size, ok := objectStatusPacketSize48EA70(op)
	if !ok || len(data) < size {
		return -1
	}
	if hooks.connected == nil || !hooks.connected() {
		return size
	}

	switch op {
	case netmsg.MSG_REPORT_X_STATUS:
		if hooks.byNetCode == nil {
			return size
		}
		dr := hooks.byNetCode(binary.LittleEndian.Uint16(data[1:3]))
		if dr == nil {
			return size
		}
		old := dr.Flags70Val
		dr.Flags70Val = binary.LittleEndian.Uint32(data[3:7])
		if dr.ObjClass.Has(object.ClassMonsterGenerator) {
			if old&monsterGeneratorStartingBit48EA70 == 0 && dr.Flags70Val&monsterGeneratorStartingBit48EA70 != 0 && hooks.prepareGenerator != nil {
				hooks.prepareGenerator(dr)
			}
			if dr.Flags70Val&monsterGeneratorTerminatedBit48EA70 != 0 {
				dr.ObjClass &^= object.ClassLight
				dr.ObjFlags &^= object.FlagFlicker
			}
		}
	case netmsg.MSG_REPORT_PLAYER_STATUS:
		if hooks.localDrawable != nil {
			if dr := hooks.localDrawable(); dr != nil {
				dr.ObjFlags = object.Flags(binary.LittleEndian.Uint32(data[1:5]))
			}
		}
	case netmsg.MSG_REPORT_MODIFIER:
		if hooks.byNetCode == nil {
			return size
		}
		dr := hooks.byNetCode(binary.LittleEndian.Uint16(data[1:3]))
		if dr == nil {
			return size
		}
		var modifiers [4]*server.ModifierEff
		if hooks.modifierByID != nil {
			for i := range modifiers {
				modifiers[i] = hooks.modifierByID(data[3+i])
			}
		}
		setDrawableModifiers48EA70(dr, modifiers)
	case netmsg.MSG_REPORT_STAT_MODIFIER:
		code := uint32(binary.LittleEndian.Uint16(data[1:3]) & 0x7fff)
		if hooks.localNetCode != nil && code == hooks.localNetCode() && hooks.setStatModifier != nil {
			hooks.setStatModifier(data[7], math.Float32frombits(binary.LittleEndian.Uint32(data[3:7])))
		}
	case netmsg.MSG_REPORT_ANIMATION_FRAME:
		if hooks.byNetCode != nil {
			if dr := hooks.byNetCode(binary.LittleEndian.Uint16(data[1:3])); dr != nil {
				dr.SetFrameMB(int(binary.LittleEndian.Uint32(data[3:7])))
			}
		}
	}
	return size
}

func handleNPCReportNative48EA70(data []byte, hooks npcReportPacketHooks48EA70) int {
	if len(data) < npcReportPacketSize48EA70 {
		return -1
	}
	if hooks.connected == nil || !hooks.connected() {
		return npcReportPacketSize48EA70
	}
	rawCode := binary.LittleEndian.Uint16(data[1:3])
	id := int(rawCode & 0x7fff)
	var npc *server.NPC
	if hooks.byID != nil {
		npc = hooks.byID(id)
	}
	if npc != nil {
		if hooks.resetNPC != nil {
			hooks.resetNPC(npc, id)
		}
	} else if hooks.newNPC != nil {
		npc = hooks.newNPC(id)
	}
	if npc == nil {
		return npcReportPacketSize48EA70
	}
	for i := range npc.Color8 {
		off := 3 + 3*i
		if hooks.color != nil {
			npc.Color8[i] = hooks.color(data[off], data[off+1], data[off+2])
		} else {
			npc.Color8[i] = 0
		}
	}
	npc.Field1312 = uint32(rawCode >> 15)
	return npcReportPacketSize48EA70
}

func handleClientStatusNative48EA70(data []byte, hooks clientStatusPacketHooks48EA70) int {
	if len(data) < clientStatusReportPacketSize48EA70 {
		return -1
	}
	code := binary.LittleEndian.Uint16(data[1:3])
	if hooks.playerByID == nil {
		return clientStatusReportPacketSize48EA70
	}
	player := hooks.playerByID(int(code))
	if player == nil {
		return clientStatusReportPacketSize48EA70
	}
	if hooks.host == nil || !hooks.host() {
		if hooks.unsetStatus != nil {
			hooks.unsetStatus(player, clientStatusReportClearMask48EA70)
		}
		if hooks.needStatus != nil {
			hooks.needStatus(player, binary.LittleEndian.Uint32(data[3:7])&clientStatusReportMask48EA70)
		}
	}
	if hooks.renderingDisabled != nil && hooks.renderingDisabled() {
		return clientStatusReportPacketSize48EA70
	}
	if hooks.localNetCode == nil || uint32(code) != hooks.localNetCode() {
		return clientStatusReportPacketSize48EA70
	}
	if hooks.onClientStatus != nil {
		hooks.onClientStatus(player.Field3680)
	}
	if hooks.setPoisoned != nil {
		hooks.setPoisoned(int((player.Field3680 >> 10) & 1))
	}
	return clientStatusReportPacketSize48EA70
}

func (c *Client) objectStatusHooksNative48EA70() objectStatusPacketHooks48EA70 {
	return objectStatusPacketHooks48EA70{
		connected:        nox_client_isConnected,
		byNetCode:        c.Objs.ByNetCode,
		localDrawable:    c.ClientPlayerUnit,
		prepareGenerator: prepareMonsterGeneratorStatus48EA70,
		modifierByID: func(id byte) *server.ModifierEff {
			return c.srv.Modif.Nox_xxx_modifGetDescById413330(int(id))
		},
		localNetCode: func() uint32 {
			return uint32(legacy.ClientPlayerNetCode())
		},
		setStatModifier: func(index byte, value float32) {
			*memmap.PtrFloat32(0x5D4594, 1063100+uintptr(index)*4) = value
		},
	}
}

func (c *Client) handleObjectStatePacketNative48EA70(op netmsg.Op, data []byte) int {
	switch op {
	case netmsg.MSG_REPORT_X_STATUS,
		netmsg.MSG_REPORT_PLAYER_STATUS,
		netmsg.MSG_REPORT_MODIFIER,
		netmsg.MSG_REPORT_STAT_MODIFIER,
		netmsg.MSG_REPORT_ANIMATION_FRAME:
		return handleObjectStatusNative48EA70(op, data, c.objectStatusHooksNative48EA70())
	case netmsg.MSG_REPORT_NPC:
		return handleNPCReportNative48EA70(data, npcReportPacketHooks48EA70{
			connected: nox_client_isConnected,
			byID:      c.srv.NPCs.ByID,
			newNPC:    c.srv.NPCs.New,
			resetNPC:  c.srv.NPCs.Set,
			color: func(r, g, b byte) uint32 {
				return nox_color_rgb_4344A0(int(r), int(g), int(b))
			},
		})
	case netmsg.MSG_REPORT_CLIENT_STATUS:
		return handleClientStatusNative48EA70(data, clientStatusPacketHooks48EA70{
			playerByID: c.srv.Players.ByID,
			host: func() bool {
				return noxflags.HasGame(noxflags.GameHost)
			},
			unsetStatus: func(player *server.Player, mask uint32) {
				c.srv.UnsetPlayerStatus417530(player, mask, server.PlayerUnsetStatusRuntime417530{})
			},
			needStatus: func(player *server.Player, mask uint32) {
				c.srv.NeedPlayerStatus4174F0(player, mask)
			},
			renderingDisabled: func() bool {
				return noxflags.HasEngine(noxflags.EngineNoRendering)
			},
			localNetCode: func() uint32 {
				return uint32(legacy.ClientPlayerNetCode())
			},
			onClientStatus: func(status uint32) {
				nox_client_onClientStatusA(int(status))
			},
			setPoisoned: func(poisoned int) {
				legacy.Sub_470C40(poisoned)
			},
		})
	default:
		return -1
	}
}
