package legacy

import "github.com/opennox/opennox/v1/server"

// Nox_xxx_unitPostCreateNotify_4E7F10 exposes the native-width object and
// player path used by object creation. Runtime dependencies remain callbacks
// so the two initial stores happen before the global player list is touched.
func Nox_xxx_unitPostCreateNotify_4E7F10(obj *server.Object) *server.Player {
	return unitPostCreateNotifyNative4E7F10(obj, unitPostCreateNotifyNativeDeps4E7F10{
		firstPlayer: func() *server.Player {
			return GetServer().S().Players.First()
		},
		nextPlayer: func(player *server.Player) *server.Player {
			return GetServer().S().Players.Next(player)
		},
		isHostile: func(unit, created *server.Object) int32 {
			if GetServer().S().IsHostileMimicXxx(unit, created) {
				return 1
			}
			return 0
		},
	})
}
