package opennox

import (
	"encoding/binary"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type simpleObjectVisualState519410 struct {
	Code, TypeID uint16
	X, Y         uint16
}

type complexObjectVisualState519410 struct {
	simpleObjectVisualState519410
	Direction byte
	Animation byte
	Frame     byte
}

type objectEnchantVisualState48EA70 struct {
	Code  uint16
	Buffs uint32
}

func decodeSimpleObjectVisualState519410(data []byte) (simpleObjectVisualState519410, bool) {
	if len(data) < 9 {
		return simpleObjectVisualState519410{}, false
	}
	return simpleObjectVisualState519410{
		Code:   binary.LittleEndian.Uint16(data[1:]),
		TypeID: binary.LittleEndian.Uint16(data[3:]),
		X:      binary.LittleEndian.Uint16(data[5:]),
		Y:      binary.LittleEndian.Uint16(data[7:]),
	}, true
}

func decodeComplexObjectVisualState519410(data []byte) (complexObjectVisualState519410, bool) {
	base, ok := decodeSimpleObjectVisualState519410(data)
	if !ok || len(data) < 11 {
		return complexObjectVisualState519410{}, false
	}
	direction := (data[9] >> 4) & 7
	if direction > 3 {
		direction++
	}
	return complexObjectVisualState519410{
		simpleObjectVisualState519410: base,
		Direction:                     direction,
		Animation:                     data[9] & 0x0f,
		Frame:                         data[10],
	}, true
}

func applyComplexObjectVisualState519410(dr *client.Drawable, state complexObjectVisualState519410, frame uint32) {
	dr.Field_72 = frame
	dr.SetFrameMB(int(state.Frame))
	dr.AnimDir = state.Direction
	if dr.AnimInd != uint32(state.Animation) {
		dr.AnimInd = uint32(state.Animation)
		dr.AnimStart = frame
	}
}

func decodeObjectEnchantVisualState48EA70(data []byte) (objectEnchantVisualState48EA70, bool) {
	if len(data) < 7 {
		return objectEnchantVisualState48EA70{}, false
	}
	return objectEnchantVisualState48EA70{
		Code:  binary.LittleEndian.Uint16(data[1:]),
		Buffs: binary.LittleEndian.Uint32(data[3:]),
	}, true
}

func applyObjectEnchantVisualState48EA70(dr *client.Drawable, buffs uint32, local *client.Drawable, itemLightOverride byte, defaultLight float32) bool {
	hadLight := dr.HasEnchant(server.ENCHANT_LIGHT)
	dr.Buffs = buffs
	if dr == local {
		*memmap.PtrUint32(0x5D4594, 1062540) = buffs
	}
	if hadLight && !dr.HasEnchant(server.ENCHANT_LIGHT) && (dr != local || itemLightOverride&8 == 0) {
		dr.SetLightIntensity(defaultLight)
		return true
	}
	return false
}

func (c *Client) handleSimpleObjectPacketNative519410(data []byte) int {
	state, ok := decodeSimpleObjectVisualState519410(data)
	if !ok {
		return -1
	}
	if nox_client_isConnected() {
		if dr := c.Nox_xxx_spriteCreate_48E970(int(state.TypeID), state.Code, int(state.X), int(state.Y)); dr != nil {
			dr.Field_72 = c.srv.Frame()
		}
	}
	return 9
}

func (c *Client) handleComplexObjectPacketNative519410(data []byte) int {
	state, ok := decodeComplexObjectVisualState519410(data)
	if !ok {
		return -1
	}
	if !nox_client_isConnected() {
		return 11
	}
	if state.Code != 0 || state.TypeID != 0 {
		if dr := c.Nox_xxx_spriteCreate_48E970(int(state.TypeID), state.Code, int(state.X), int(state.Y)); dr != nil {
			applyComplexObjectVisualState519410(dr, state, c.srv.Frame())
		}
		if nox_xxx_netClearHighBit_578B30(state.Code) == uint16(legacy.ClientPlayerNetCode()) && sub_416120(9) {
			nox_xxx_cliUpdateCameraPos_435600(int(state.X), int(state.Y))
		}
		return 11
	}
	nox_xxx_cliUpdateCameraPos_435600(int(state.X), int(state.Y))
	inputSetKeyTimeoutLegacy(9)
	return 11
}

func (c *Client) handleObjectEnchantPacketNative48EA70(data []byte) int {
	state, ok := decodeObjectEnchantVisualState48EA70(data)
	if !ok {
		return -1
	}
	if !nox_client_isConnected() {
		return 7
	}
	dr := c.Objs.ByNetCode(state.Code)
	if dr == nil {
		return 7
	}
	defaultLight := float32(0)
	if typ := c.Things.TypeByInd(int(dr.TypeIDVal)); typ != nil {
		defaultLight = typ.LightIntensity
	}
	applyObjectEnchantVisualState48EA70(dr, state.Buffs, c.ClientPlayerUnit(), sub_467430(), defaultLight)
	return 7
}
