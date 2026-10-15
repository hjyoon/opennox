package legacy

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

// ClientUpdateFireball4CCE70 is shared by the five registered Fireball
// update entries. Power remains the original scalar 5, 4, 3, 2, or 1.
var ClientUpdateFireball4CCE70 func(*noxrender.Viewport, *client.Drawable, int) int
