//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func init() {
	legacy.ClientUpdateMagicMissile4CD9E0 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateMagicMissileDrawable4CD9E0(dr)
	}
	legacy.ClientUpdateTeleportWake4CD8D0 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateTeleportWakeDrawable4CD8D0(dr)
	}
	legacy.ClientUpdateSparkleTrail4CDBF0 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateSparkleTrailDrawable4CDBF0(dr)
	}
	legacy.ClientUpdateVortexSource4CC950 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateVortexSourceDrawable4CC950(dr)
	}
	legacy.ClientUpdateManaBombCharge4CCAC0 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return updateManaBombCharge4CCAC0(dr, noxClient.manaBombDrawableHooks4CCAC0())
	}
	legacy.ClientUpdateManaBombOrb4CA720 = func(_ *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateManaBombOrb4CA720(dr)
	}
	legacy.ClientUpdateCloud4CE1D0 = func(vp *noxrender.Viewport, dr *client.Drawable, radius int) int {
		return updateCloudDrawable4CE1D0(dr, radius, noxClient.cloudDrawableHooks4CE200())
	}
	legacy.ClientUpdateCloudRise4CE340 = func(_ *noxrender.Viewport, dr *client.Drawable) int {
		return updateCloudParticleRise4CE340(dr)
	}
	legacy.ClientUpdateMeteor4CCD00 = func(_ *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateMeteorDrawable4CCD00(dr)
	}
	legacy.ClientUpdateFist4CCDB0 = func(_ *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateFistDrawable4CCDB0(dr)
	}
	legacy.ClientUpdatePredictLinear4CA540 = func(vp *noxrender.Viewport, dr *client.Drawable) int {
		return noxClient.updateClientPredictLinear4CA540(vp, dr)
	}
	legacy.ClientUpdateMagic4CDD80 = func(_ *noxrender.Viewport, dr *client.Drawable) {
		noxClient.updateMagicDrawable4CDD80(dr)
	}
}
