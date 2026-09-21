//go:build server

package opennox

import (
	"errors"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func e2eClientVisualUnavailable() bool {
	e2eError(errors.New("client visual E2E assertion is unavailable in a server-only build"))
	return false
}

func e2eRebuildClientLightGrid(_ *noxrender.Viewport) bool {
	return e2eClientVisualUnavailable()
}

func e2eRunClientPolygonDrawColor() bool {
	return e2eClientVisualUnavailable()
}

func e2eClientPolygonMinimapZone(_ *client.Drawable) (int, bool) {
	return 0, e2eClientVisualUnavailable()
}
