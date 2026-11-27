package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func pointerChainAttach50F370(session *server.TradeSession, owner *server.Object) *server.TradeSession {
	session.Field0 = 1
	if owner.ObjClass&4 != 0 {
		update := (*server.PlayerUpdateData)(owner.UpdateData)
		storeForeignPointer(unsafe.Pointer(&update.Trade70), unsafe.Pointer(session))
	}
	return session
}

func pointerChainDetach50F490(session *server.TradeSession, owner *server.Object) unsafe.Pointer {
	session.Field0 = 0
	if owner.ObjClass&4 == 0 {
		return unsafe.Pointer(owner)
	}
	update := (*server.PlayerUpdateData)(owner.UpdateData)
	if update.Trade70 == session {
		storeForeignPointer(unsafe.Pointer(&update.Trade70), nil)
	}
	return unsafe.Pointer(update)
}

func pointerChainStat4D60E0(owner *server.Object) unsafe.Pointer {
	if owner.ObjFlags&32 != 0 {
		return unsafe.Pointer(owner)
	}
	player := (*server.PlayerUpdateData)(owner.UpdateData).Player
	player.RecordGuardedStat4D60E0()
	return unsafe.Pointer(player)
}

func pointerChainKill4D6170(owner *server.Object) unsafe.Pointer {
	if owner == nil || owner.ObjFlags&32 != 0 {
		return unsafe.Pointer(owner)
	}
	player := (*server.PlayerUpdateData)(owner.UpdateData).Player
	player.RecordMonsterKilled()
	return unsafe.Pointer(player)
}

func pointerChainLookup510DE0(owner *server.Object, code uint32) *server.Object {
	update := (*server.PlayerUpdateData)(owner.UpdateData)
	if update.Trade70 == nil {
		return nil
	}
	if node := game4TradeFindItem50FFE0(update.Trade70.Field20, code); node != nil {
		return node.Item0
	}
	return nil
}

func pointerChainLessons(owner *server.Object, delta int32, subtract bool) unsafe.Pointer {
	if owner.ObjClass&4 == 0 {
		return unsafe.Pointer(owner)
	}
	player := (*server.PlayerUpdateData)(owner.UpdateData).Player
	if subtract {
		player.Lessons -= delta
	} else {
		player.Lessons += delta
	}
	return unsafe.Pointer(player)
}
