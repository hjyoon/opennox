package opennox

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

type objectLightHooks48EA70 struct {
	connected func() bool
	byNetCode func(uint16) *client.Drawable
}

func handleObjectLightNative48EA70(op netmsg.Op, data []byte, hooks objectLightHooks48EA70) int {
	var size int
	switch op {
	case netmsg.MSG_REPORT_LIGHT_COLOR:
		size = 6
	case netmsg.MSG_REPORT_LIGHT_INTENSITY:
		size = 7
	default:
		return -1
	}
	if len(data) < size {
		return -1
	}
	if !hooks.connected() {
		return size
	}
	dr := hooks.byNetCode(binary.LittleEndian.Uint16(data[1:3]))
	if dr == nil {
		return size
	}
	switch op {
	case netmsg.MSG_REPORT_LIGHT_COLOR:
		dr.SetLightColor(data[3], data[4], data[5])
	case netmsg.MSG_REPORT_LIGHT_INTENSITY:
		dr.SetLightIntensity(math.Float32frombits(binary.LittleEndian.Uint32(data[3:7])))
	}
	return size
}

func (c *Client) handleObjectLightPacketNative48EA70(op netmsg.Op, data []byte) int {
	return handleObjectLightNative48EA70(op, data, objectLightHooks48EA70{
		connected: nox_client_isConnected,
		byNetCode: c.Objs.ByNetCode,
	})
}
