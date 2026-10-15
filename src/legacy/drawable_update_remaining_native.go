package legacy

import (
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

// The renderer supplies the existing native implementations. The registered
// C names remain stable; these hooks carry native pointers, never PE32 handles.
var (
	ClientUpdateMagicMissile4CD9E0   func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateTeleportWake4CD8D0   func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateSparkleTrail4CDBF0   func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateVortexSource4CC950   func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateManaBombCharge4CCAC0 func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateManaBombOrb4CA720    func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateCloud4CE1D0          func(*noxrender.Viewport, *client.Drawable, int) int
	ClientUpdateCloudRise4CE340      func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateMeteor4CCD00         func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateFist4CCDB0           func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdatePredictLinear4CA540  func(*noxrender.Viewport, *client.Drawable) int
	ClientUpdateMagic4CDD80          func(*noxrender.Viewport, *client.Drawable)
)
