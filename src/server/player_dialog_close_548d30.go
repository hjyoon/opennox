package server

import "github.com/opennox/libs/noxnet/netmsg"

type PlayerDialogCloseRuntime548D30 struct {
	Unfreeze func(player *Object, force uint32)
	Send     func(recipient byte, packet [2]byte)
	CallEnd  func(index int32, caller, trigger *Object)
}

// PlayerDialogClose548D30 restores GAME.EXE 00548D30. The caller guarantees
// a player object and initialized update data, just as the original network
// dispatcher did. In particular, player update data is cached before
// Unfreeze, while the NPC end-script index is reloaded after the close packet
// and result stores.
func (s *Server) PlayerDialogClose548D30(player *Object, answer byte, runtime PlayerDialogCloseRuntime548D30) {
	update := player.UpdateDataPlayer()
	runtime.Unfreeze(player, 0)

	npc := update.DialogWith
	if npc == nil {
		return
	}
	npcUpdate := npc.UpdateDataMonster()
	if npcUpdate.DialogStartFunc == -1 || npcUpdate.DialogEndFunc == -1 {
		return
	}

	packet := [2]byte{byte(netmsg.MSG_DIALOG), 4}
	runtime.Send(update.Player.PlayerInd, packet)
	update.DialogWith = nil
	if npcUpdate.DialogFlags == 1 {
		npcUpdate.DialogResult = answer
	} else {
		npcUpdate.DialogResult = 0
	}
	runtime.CallEnd(npcUpdate.DialogEndFunc, player, npc)
}
