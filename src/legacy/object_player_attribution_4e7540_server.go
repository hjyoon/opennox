package legacy

import "github.com/opennox/opennox/v1/server"

func recordPlayerAttributionNative4E7540(source, target *server.Object, frame func() uint32) {
	recordPlayerAttribution4E7540(source, target, playerAttributionHooks4E7540[*server.Object, *server.PlayerUpdateData, *server.Player]{
		class: func(obj *server.Object) uint32 {
			return uint32(obj.ObjClass)
		},
		updateData: func(obj *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(obj.UpdateData)
		},
		player: func(update *server.PlayerUpdateData) *server.Player {
			return update.Player
		},
		playerIndex: func(player *server.Player) byte {
			return player.PlayerInd
		},
		setPlayerIndex: func(player *server.Player, index uint32) {
			player.SetLastAggressorPlayerIndex(index)
		},
		frame: frame,
		setFrame: func(player *server.Player, frame uint32) {
			player.SetLastAggressorFrame(frame)
		},
		setPending: func(player *server.Player, pending uint32) {
			player.SetLastAggressorPending(pending)
		},
	})
}
