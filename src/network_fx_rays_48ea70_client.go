//go:build !server

package opennox

import (
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/legacy"
)

func (c *Client) handleRayFXPacketNative48EA70(op netmsg.Op, data []byte) int {
	return handleRayFXNative48EA70(op, data, rayFXHooks48EA70{
		connected:       nox_client_isConnected,
		drawRay:         legacy.Nox_xxx_netDrawRays_49BDD0,
		lightningSparks: legacy.Nox_xxx_makeRayLightningParticles_49BDD0,
		plasmaEndSparks: legacy.Nox_xxx_makeRayPlasmaParticles_49BDD0,
	})
}
