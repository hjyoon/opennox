package legacy

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func objectPlayerGoldNative4FA620(obj *server.Object) *server.Player {
	// GAME.EXE checks the low ObjClass byte for ClassPlayer before following
	// UpdateData. Keep that scalar-width behavior while retaining native-width
	// Go pointers throughout the traversal.
	if obj == nil || uint8(obj.ObjClass)&uint8(object.ClassPlayer) == 0 || obj.UpdateData == nil {
		return nil
	}
	update := (*server.PlayerUpdateData)(obj.UpdateData)
	return update.Player
}

func objectGetGoldNative4FA6D0(obj *server.Object) int {
	player := objectPlayerGoldNative4FA620(obj)
	if player == nil {
		return 0
	}
	return int(player.GoldVal)
}

func objectSetGoldNative4FA620(
	obj *server.Object,
	delta int32,
	protect func(token uint32, delta int32),
	reset func(token uint32, value int32),
) {
	player := objectPlayerGoldNative4FA620(obj)
	if player == nil {
		return
	}

	amount := uint32(0) - uint32(delta)
	if delta >= 0 || player.GoldVal >= amount {
		player.GoldVal += uint32(delta)
		protect(player.ProtPlayerGold, delta)
		return
	}

	player.GoldVal = 0
	reset(player.ProtPlayerGold, 0)
}
