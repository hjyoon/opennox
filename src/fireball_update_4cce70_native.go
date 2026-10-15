//go:build !server

package opennox

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func init() {
	legacy.ClientUpdateFireball4CCE70 = func(_ *noxrender.Viewport, dr *client.Drawable, power int) int {
		return noxClient.updateFireballDrawable4CCE70(dr, power)
	}
}
