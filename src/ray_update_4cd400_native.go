//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func init() {
	legacy.ClientUpdateCharm4CD400 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return updateCharmRay4CD400(vp, dr, noxClient.drainHealRayHooks4CD450())
	}
	legacy.ClientUpdateHeal4CD450 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return updateDrainHealRay4CD450(vp, dr, "HealOrb", noxClient.drainHealRayHooks4CD450())
	}
	legacy.ClientUpdateDrainMana4CD690 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return updateDrainHealRay4CD450(vp, dr, "DrainManaOrb", noxClient.drainHealRayHooks4CD450())
	}
}
