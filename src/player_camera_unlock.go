package opennox

import "github.com/opennox/opennox/v1/server"

// playerCameraUnlock_4E6040 follows the original object -> player update data
// -> player chain without adding class or nil guards, then clears the camera.
func playerCameraUnlock_4E6040(obj *server.Object) {
	ud := (*server.PlayerUpdateData)(obj.UpdateData)
	pl := ud.Player
	pl.CameraFollowObj = nil
}
