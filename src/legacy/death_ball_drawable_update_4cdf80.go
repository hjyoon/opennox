package legacy

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

var (
	ClientUpdateDeathBallSparks4CDF80 func(*noxrender.Viewport, *client.Drawable, int) int
	ClientUpdateDeathBallCharge4CE0C0 func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateLinearOrb4CA650       func(*noxrender.Viewport, *client.Drawable) int
)
