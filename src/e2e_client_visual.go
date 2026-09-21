//go:build !server

package opennox

import (
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
