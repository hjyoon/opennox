package legacy

import (
	"github.com/opennox/opennox/v1/server"
)

func projectileTrailUpdateCall53AEC0(projectile *server.Object) {
	if projectile == nil {
		return
	}
	server.ProjectileTrailUpdateRaw53AEC0(projectile, sparkTrailPushRawRuntime().Trail)
}
