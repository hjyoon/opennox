//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func init() {
	legacy.ClientUpdateDeathBallSparks4CDF80 = func(_ *noxrender.Viewport, dr *client.Drawable, count int) int {
		return updateDeathBallSparks4CDF80(dr, count, noxClient.deathBallDrawableHooks4CDF80())
	}
	legacy.ClientUpdateDeathBallCharge4CE0C0 = func(_ *noxrender.Viewport, dr *client.Drawable) int {
		return updateDeathBallCharge4CE0C0(dr, noxClient.deathBallDrawableHooks4CDF80())
	}
	legacy.ClientUpdateLinearOrb4CA650 = func(_ *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateLinearOrb4CA650(dr)
	}
}
