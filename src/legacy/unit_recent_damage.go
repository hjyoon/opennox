package legacy

import "github.com/opennox/opennox/v1/server"

// unitWasDamagedRecently_4E6BD0 preserves GAME.EXE's health-pointer gate and
// uint32 frame subtraction. The frame callback deliberately runs before the
// timestamp reload.
func unitWasDamagedRecently_4E6BD0(unit *server.Object, currentFrame func() uint32) bool {
	if unit.HealthData == nil {
		return false
	}
	frame := currentFrame()
	return frame-unit.Frame134 <= 1
}
