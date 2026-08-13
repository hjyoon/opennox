package legacy

import "github.com/opennox/opennox/v1/server"

type unitPostCreateNotifyNativeDeps4E7F10 struct {
	firstPlayer func() *server.Player
	nextPlayer  func(*server.Player) *server.Player
	isHostile   func(*server.Object, *server.Object) int32
}

func unitPostCreateNotifyNative4E7F10(
	obj *server.Object,
	deps unitPostCreateNotifyNativeDeps4E7F10,
) *server.Player {
	return unitPostCreateNotify4E7F10(obj, unitPostCreateNotifyHooks4E7F10[*server.Object, *server.Player]{
		storeField35: func(obj *server.Object, value uint32) {
			obj.Field35 = value
		},
		storeField36: func(obj *server.Object, value uint32) {
			obj.Field36 = value
		},
		firstPlayer: deps.firstPlayer,
		loadPlayerInd: func(player *server.Player) uint8 {
			return player.PlayerInd
		},
		loadPlayerUnit: func(player *server.Player) *server.Object {
			return player.PlayerUnit
		},
		isHostile: deps.isHostile,
		loadField35: func(obj *server.Object) uint32 {
			return obj.Field35
		},
		loadField36: func(obj *server.Object) uint32 {
			return obj.Field36
		},
		nextPlayer: deps.nextPlayer,
	})
}
