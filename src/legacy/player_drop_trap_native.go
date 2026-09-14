package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

// playerDropATrapNative10002030 keeps the original first-match inventory
// search, but follows native-width Object and Player links. The legacy test at
// Object+0xA selects the third byte of ObjClass, not the whole class mask.
func playerDropATrapNative10002030(
	owner *server.Object,
	drop func(*server.Object, *server.Object, *types.Pointf),
) bool {
	if owner == nil || owner.UpdateData == nil {
		return false
	}
	update := (*server.PlayerUpdateData)(owner.UpdateData)
	player := update.Player
	if player == nil {
		return false
	}
	pos := player.Pos3632()
	if player.Field3680&3 != 0 || update.State == server.PlayerState1 {
		return false
	}
	for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
		if byte(uint32(item.ObjClass)>>16) == 17 {
			drop(owner, item, &pos)
			return true
		}
	}
	return false
}
