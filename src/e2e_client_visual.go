//go:build !server

package opennox

import (
	"image"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func e2eRebuildClientLightGrid(vp *noxrender.Viewport) bool {
	noxClient.sub_468F80(vp)
	return true
}

func e2eRunClientPolygonDrawColor() bool {
	legacy.Nox_xxx_polygonDrawColor_421B80()
	return true
}

func e2eClientPolygonMinimapZone(drawable *client.Drawable) (int, bool) {
	return legacy.Sub_472540(drawable), true
}

func e2eClientHUDMeter(index int) (e2eHUDMeterState, bool) {
	state, ok := legacy.Nox_client_guiHealthManaState(index)
	if !ok {
		return e2eHUDMeterState{}, false
	}
	return e2eHUDMeterState{
		Current:         state.Current,
		Maximum:         state.Maximum,
		PrimaryColor:    state.PrimaryColor,
		SecondaryColor:  state.SecondaryColor,
		Poisoned:        state.Poisoned,
		MeterPos:        image.Pt(state.MeterX, state.MeterY),
		RootPos:         image.Pt(state.RootX, state.RootY),
		PoisonTubeReady: state.PoisonTubeReady,
	}, true
}
