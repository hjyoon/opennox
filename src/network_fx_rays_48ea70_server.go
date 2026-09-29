//go:build server

package opennox

import "github.com/opennox/libs/noxnet/netmsg"

func (c *Client) handleRayFXPacketNative48EA70(op netmsg.Op, data []byte) int {
	if _, ok := decodeRayFXState48EA70(op, data); !ok {
		return -1
	}
	return rayFXPacketSize48EA70
}
