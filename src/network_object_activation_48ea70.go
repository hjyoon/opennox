package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
)

const (
	objectActivationPacketSize48EA70 = 3
	objectDrawFramePacketSize48EA70  = 4
)

type objectActivationState48EA70 struct {
	Code uint16
}

type objectDrawFrameState48EA70 struct {
	Code  uint16
	Frame byte
}

type objectActivationHooks48EA70 struct {
	connected func() bool
	byNetCode func(uint16) *client.Drawable
}

func decodeObjectActivationState48EA70(data []byte) (objectActivationState48EA70, bool) {
	if len(data) < objectActivationPacketSize48EA70 {
		return objectActivationState48EA70{}, false
	}
	return objectActivationState48EA70{
		Code: binary.LittleEndian.Uint16(data[1:3]),
	}, true
}

func decodeObjectDrawFrameState48EA70(data []byte) (objectDrawFrameState48EA70, bool) {
	if len(data) < objectDrawFramePacketSize48EA70 {
		return objectDrawFrameState48EA70{}, false
	}
	return objectDrawFrameState48EA70{
		Code:  binary.LittleEndian.Uint16(data[1:3]),
		Frame: data[3],
	}, true
}

func handleObjectActivationNative48EA70(data []byte, enabled bool, hooks objectActivationHooks48EA70) int {
	state, ok := decodeObjectActivationState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return objectActivationPacketSize48EA70
	}
	dr := hooks.byNetCode(state.Code)
	if dr == nil {
		return objectActivationPacketSize48EA70
	}
	if enabled {
		dr.ObjFlags |= object.FlagEnabled
		return objectActivationPacketSize48EA70
	}
	// GAME.EXE clears the draw callback for class bit 0x40000 before it
	// disables the object. Keep the native pointer field intact for every
	// other class instead of addressing Drawable through PE32 byte offsets.
	if dr.Class().Has(object.ClassReadable) {
		dr.DrawFuncPtr = nil
	}
	dr.ObjFlags &^= object.FlagEnabled
	return objectActivationPacketSize48EA70
}

func handleObjectDrawFrameNative48EA70(data []byte, hooks objectActivationHooks48EA70) int {
	state, ok := decodeObjectDrawFrameState48EA70(data)
	if !ok {
		return -1
	}
	if !hooks.connected() {
		return objectDrawFramePacketSize48EA70
	}
	dr := hooks.byNetCode(state.Code)
	if dr != nil {
		dr.SetFrameMB(int(state.Frame))
	}
	return objectDrawFramePacketSize48EA70
}

func (c *Client) objectActivationHooksNative48EA70() objectActivationHooks48EA70 {
	return objectActivationHooks48EA70{
		connected: nox_client_isConnected,
		byNetCode: c.Objs.ByNetCode,
	}
}

func (c *Client) handleObjectActivationPacketNative48EA70(op netmsg.Op, data []byte) int {
	hooks := c.objectActivationHooksNative48EA70()
	switch op {
	case netmsg.MSG_ENABLE_OBJECT:
		return handleObjectActivationNative48EA70(data, true, hooks)
	case netmsg.MSG_DISABLE_OBJECT:
		return handleObjectActivationNative48EA70(data, false, hooks)
	case netmsg.MSG_DRAW_FRAME:
		return handleObjectDrawFrameNative48EA70(data, hooks)
	default:
		return -1
	}
}
