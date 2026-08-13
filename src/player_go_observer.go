package opennox

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/server"
)

type playerGoObserverHooks_4E6860 struct {
	abilityActive     func(*server.Object) int
	isMonsterBot      func(*server.Object) bool
	gameFlag          func(noxflags.GameFlag) bool
	ensureCrownID     func()
	ensureGameBallID  func()
	crownID           func() uint32
	gameBallID        func() uint32
	dropCrown         func(*server.Object, *server.Object, *types.Pointf)
	clearOwner        func(*server.Object)
	gameBallDropped   func()
	dropFlag          func(*server.Object, *server.Object, *types.Pointf)
	getPossess        func(*server.Object) *server.Object
	clearObserve      func(*server.Object)
	needTimestamp     func(*server.Player)
	anyPlayers        func() int
	resetState        func()
	forceLessons      func()
	resetTeams        func()
	finishReset       func()
	inform            func(ntype.PlayerInd, byte, uint32)
	applyInvisible    func(*server.Object)
	unlockCamera      func(*server.Object)
	leaveObserver     func(*server.Player)
	removeSpawned     func(*server.Object)
	setObserverUpdate func(*server.Object)
	resetCamping      func(*server.Object)
}

// playerGoObserver_4E6860 keeps GAME.EXE's cached unit/update-data pointers and
// the deliberate PlayerUnit reloads around callbacks. notify is narrowed only
// at the original 32-bit message boundary; keep is tested solely for zero.
func playerGoObserver_4E6860(
	pl *server.Player,
	notify, keep int,
	h playerGoObserverHooks_4E6860,
) int {
	if pl == nil {
		return 1
	}
	unit := pl.PlayerUnit
	if unit == nil {
		return 1
	}
	// GAME.EXE reads +748 before either early-return callback and uses this
	// exact pointer to clear CurTraps near the end of the function.
	ud := (*server.PlayerUpdateData)(unit.UpdateData)
	if keep == 0 && h.abilityActive(unit) == 1 {
		return 0
	}
	if h.isMonsterBot(unit) {
		return 0
	}

	if h.gameFlag(noxflags.GameModeKOTR | noxflags.GameModeCTF | noxflags.GameModeFlagBall) {
		h.ensureCrownID()
		h.ensureGameBallID()
		// The initial owner and the owner passed to each drop callback are
		// separate PlayerUnit reads in the original function.
		for it := pl.PlayerUnit.FirstOwned516(); it != nil; {
			crown := h.crownID()
			typ := uint32(it.TypeInd)
			if typ == crown {
				owner := pl.PlayerUnit
				h.dropCrown(owner, it, &owner.PosVec)
			} else if typ == h.gameBallID() {
				flags := it.ObjFlags &^ object.FlagNoCollide
				it.Obj130 = nil
				it.ObjFlags = flags
				h.clearOwner(it)
				h.gameBallDropped()
			} else if it.ObjClass&object.ClassFlag != 0 {
				owner := pl.PlayerUnit
				h.dropFlag(owner, it, &owner.PosVec)
			}
			// Drop and owner callbacks may rewrite the owned list.
			it = it.Field128
		}
	}

	if h.getPossess(pl.PlayerUnit) != nil {
		h.clearObserve(pl.PlayerUnit)
	}
	h.needTimestamp(pl)
	if h.anyPlayers() == 0 && !h.gameFlag(noxflags.GameModeQuest) {
		h.resetState()
		h.forceLessons()
		h.resetTeams()
		h.finishReset()
	}
	h.inform(ntype.PlayerInd(pl.PlayerInd), 12, uint32(notify))
	h.applyInvisible(unit)
	unit.ObjFlags |= object.FlagNoCollide
	pl.Pos3632Vec.X = unit.PosVec.X
	pl.Pos3632Vec.Y = unit.PosVec.Y
	h.unlockCamera(unit)
	if h.gameFlag(noxflags.GameModeCoop) {
		pl.Field3672 = 1
		pl.CameraFollowObj = nil
	} else if h.gameFlag(noxflags.GameModeFlagBall) && keep == 0 {
		h.leaveObserver(pl)
	}
	h.removeSpawned(unit)
	// 004E6A78 is a byte store even though CurTraps occupies a uint32 slot.
	*(*uint8)(unsafe.Pointer(&ud.CurTraps)) = 0
	h.setObserverUpdate(unit)
	h.resetCamping(unit)
	return 1
}
