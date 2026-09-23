package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
)

type fixedObjectStateHooks48EA70 struct {
	connected func() bool
	byNetCode func(uint16) *client.Drawable
	frame     func() uint32
}

func handleFixedObjectStateNative48EA70(op netmsg.Op, data []byte, hooks fixedObjectStateHooks48EA70) int {
	if len(data) < 4 {
		return -1
	}
	if !hooks.connected() {
		return 4
	}
	dr := hooks.byNetCode(binary.LittleEndian.Uint16(data[1:3]))
	if dr == nil {
		return 4
	}
	value := data[3]
	switch op {
	case netmsg.MSG_DOOR_ANGLE:
		dr.Field_74_4 = value
	case netmsg.MSG_OBELISK_CHARGE:
		dr.SetActive()
		dr.SetLightIntensity(float32(16 * int(value) / 10))
		dr.SetFrameMB(8 * int(value) / 50)
		if dr.AnimFrameSlave == 8 {
			dr.AnimFrameSlave = 7
		}
	case netmsg.MSG_PENTAGRAM_ACTIVATE:
		if value != 0 {
			dr.ObjClass |= object.ClassLight
			dr.SetLightIntensity(41.958)
		} else {
			dr.ObjClass &^= object.ClassLight
			dr.SetLightIntensity(0)
		}
		dr.SetFrameMB(int(value))
		dr.Field_72 = hooks.frame()
	default:
		return -1
	}
	return 4
}

func (c *Client) handleFixedObjectStatePacketNative48EA70(op netmsg.Op, data []byte) int {
	return handleFixedObjectStateNative48EA70(op, data, fixedObjectStateHooks48EA70{
		connected: nox_client_isConnected,
		byNetCode: c.Objs.ByNetCode,
		frame:     c.Server.Frame,
	})
}
