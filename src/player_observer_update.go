package opennox

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

// playerObserverPrepare_4E62F0 preserves the original four-slot cleanup and
// the otherwise redundant replay camera copy. Keeping the two camera reads is
// important because the replay path performs both of them in GAME.EXE.
func playerObserverPrepare_4E62F0(
	markers *[4]*server.Object,
	needSync func(),
	replayRead bool,
	cameraTarget func() *server.Object,
	setCameraPos func(types.Pointf),
) {
	for i := 0; i < 4; i++ {
		if obj := markers[i]; obj != nil && uint8(obj.ObjFlags)&0x20 != 0 {
			markers[i] = nil
		}
	}
	needSync()
	if replayRead {
		playerObserverCopyCameraPos_4E62F0(cameraTarget, setCameraPos)
	}
	playerObserverCopyCameraPos_4E62F0(cameraTarget, setCameraPos)
}

func playerObserverCopyCameraPos_4E62F0(cameraTarget func() *server.Object, setCameraPos func(types.Pointf)) {
	if target := cameraTarget(); target != nil {
		setCameraPos(target.PosVec)
	}
}

// playerObserverClosestOwnedMonster_4E6800 implements the callback passed by
// 004E62F0 to the rectangle iterator. The x87 comparison tests only C0, so an
// unordered distance replaces the current candidate just like a smaller one.
func playerObserverClosestOwnedMonster_4E6800(
	owner *server.Object,
	center *types.Pointf,
	each func(types.Rectf, func(*server.Object) bool),
) *server.Object {
	var found *server.Object
	best := float32(1e8)
	rect := types.RectFromPointsf(
		center.Sub(types.Pointf{X: 100, Y: 100}),
		center.Add(types.Pointf{X: 100, Y: 100}),
	)
	each(rect, func(candidate *server.Object) bool {
		if candidate.ObjClass&object.ClassMonster == 0 || candidate.ObjOwner == nil || candidate.ObjOwner != owner {
			return true
		}
		// The callback receives a pointer to Player.Pos3632Vec in GAME.EXE,
		// so it reloads the live center for every visited object.
		dx := float64(candidate.PosVec.X) - float64(center.X)
		dy := float64(candidate.PosVec.Y) - float64(center.Y)
		dist := dx*dx + dy*dy
		if !(dist >= float64(best)) {
			found = candidate
			best = float32(dist)
		}
		return true
	})
	return found
}

// playerObserverPanCamera_4E62F0 reproduces the x87 pan calculation. Cursor
// coordinates are explicitly int32 because the original fields are signed
// 32-bit integers even when the native Go int is 64 bits.
func playerObserverPanCamera_4E62F0(
	pos types.Pointf,
	cursorX, cursorY int32,
	valid func(types.Pointf) bool,
) (types.Pointf, bool) {
	const max = float64(30)
	scale := float64(float32(0.1))
	dx := float64(pos.X) - float64(cursorX)
	// GAME.EXE spills the Y delta through its float32 stack slot before it
	// compares or scales it. X remains in the x87 register stack.
	dy := float64(float32(float64(pos.Y) - float64(cursorY)))
	next := pos
	if dx > max {
		next.X = float32(float64(pos.X) - (dx-max)*scale)
	} else if dx < -max || math.IsNaN(dx) {
		// The second x87 comparison checks C0 only, so unordered follows the
		// same arithmetic path as a value below -max.
		next.X = float32(float64(pos.X) - (max+dx)*scale)
	}
	if dy > max {
		next.Y = float32(float64(pos.Y) - (dy-max)*scale)
	} else if dy < -max {
		next.Y = float32(float64(pos.Y) - (max+dy)*scale)
	}
	if !valid(next) {
		return pos, false
	}
	return next, true
}

// playerObserverMustLeaveForGameState_4E62F0 preserves the short-circuit
// order at 004E65C8..004E6614. anyPlayers must not run unless both earlier
// checks fail and status bit 0x100 is set.
func playerObserverMustLeaveForGameState_4E62F0(
	pl *server.Player,
	stateRestricted func() bool,
	flag16 func() bool,
	anyPlayers func() bool,
	allow func(*server.Player) bool,
) bool {
	if stateRestricted() || flag16() || pl.Field3680&0x100 != 0 && anyPlayers() {
		return !allow(pl)
	}
	return false
}

type playerObserverQuestHooks_4E62F0 struct {
	loading func(*server.Object)
	join    func() uint32
	joined  func(*server.Object)
	full    func(*server.Object)
	field79 func(*server.Object)
	leave   func(*server.Player)
}

// playerObserverHandleQuest_4E62F0 returns true when the current control event
// is fully handled. PlayerUpdateData.Player is deliberately reloaded around
// callbacks, while leave receives the player cached by the outer function.
func playerObserverHandleQuest_4E62F0(
	unit *server.Object,
	ud *server.PlayerUpdateData,
	cachedPlayer *server.Player,
	event *server.PlayerCtrl,
	h playerObserverQuestHooks_4E62F0,
) bool {
	if ud.Player.Field4792 == 0 {
		if ud.Field138 == 1 {
			h.loading(unit)
		} else {
			value := h.join()
			ud.Player.Field4792 = value
			if ud.Player.Field4792 == 1 {
				h.joined(unit)
			} else {
				h.full(unit)
			}
		}
	}
	if ud.QuestWarpGate != nil {
		h.field79(unit)
		return true
	}
	if ud.QuestExit != nil {
		h.leave(cachedPlayer)
		event.Active = false
		return true
	}
	if ud.Player.Field4792 == 0 {
		h.leave(cachedPlayer)
		event.Active = false
		return true
	}
	return false
}
