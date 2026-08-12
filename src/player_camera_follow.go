package opennox

import "github.com/opennox/opennox/v1/server"

// playerCameraFollow_4E6060 follows the original object -> player update data
// -> player chain without adding class or nil guards. Selecting the current
// target delegates to the original unlock path; any other target is stored.
func playerCameraFollow_4E6060(obj, target *server.Object) {
	ud := (*server.PlayerUpdateData)(obj.UpdateData)
	pl := ud.Player
	if pl.CameraFollowObj == target {
		playerCameraUnlock_4E6040(obj)
		return
	}
	pl.CameraFollowObj = target
}
