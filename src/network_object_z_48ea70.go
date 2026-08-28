package opennox

import (
	"encoding/binary"

	"github.com/opennox/opennox/v1/client"
)

type objectZState48EA70 struct {
	Code      uint16
	Magnitude byte
}

type objectZHooks48EA70 struct {
	connected func() bool
	byNetCode func(uint16) *client.Drawable
}

func decodeObjectZState48EA70(data []byte) (objectZState48EA70, bool) {
	if len(data) < 4 {
		return objectZState48EA70{}, false
	}
	return objectZState48EA70{
		Code:      binary.LittleEndian.Uint16(data[1:3]),
		Magnitude: data[3],
	}, true
}

func handleObjectZNative48EA70(data []byte, negative bool, hooks objectZHooks48EA70) int {
	state, ok := decodeObjectZState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return 4
	}
	dr := hooks.byNetCode(state.Code)
	if dr == nil {
		return 4
	}
	z := int32(state.Magnitude)
	if negative {
		z = -z
	}
	// GAME.EXE stores the low 16 bits at Drawable+104. Preserve that
	// representation for negative heights without involving pointer width.
	dr.ZVal = uint16(z)
	return 4
}

func (c *Client) handleObjectZPacketNative48EA70(data []byte, negative bool) int {
	return handleObjectZNative48EA70(data, negative, objectZHooks48EA70{
		connected: nox_client_isConnected,
		byNetCode: c.Objs.ByNetCode,
	})
}
