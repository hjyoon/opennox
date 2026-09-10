package server

const (
	scriptMoveMonsterClassLow5123C0 = uint8(0x02)
	scriptMoveBlockedFlag5123C0     = uint32(0x00008000)
	scriptMoveFarAction5123C0       = uint32(8)
	scriptMoveRoamAction5123C0      = uint32(10)
	scriptMoveReportAction5123C0    = uint32(32)
	scriptMoveReportArg5123C0       = uint32(8)
)

// scriptMoveHooks5123C0 exposes every observable load, call, and store made
// by GAME.EXE 005123C0. Object, waypoint, update-data, and action handles stay
// at native width; only fields that were fixed 32-bit values in the original
// executable are represented as uint32.
type scriptMoveHooks5123C0[O comparable, W, MU, DU any, A comparable] struct {
	loadFlags             func(O) uint32
	loadClassLow          func(O) uint8
	loadMonsterUpdate     func(O) MU
	clearActionStack      func(O)
	pushAction            func(O, uint32) A
	storeActionArgU32     func(A, int, uint32)
	loadWaypointPoints    func(W) uint8
	storeActionWaypoint   func(A, int, W)
	loadRoamFlagsLow      func(MU) uint8
	storeActionArgLow     func(A, int, uint8)
	loadWaypointXBits     func(W) uint32
	loadWaypointYBits     func(W) uint32
	loadMoverType         func() uint32
	loadTypeInd           func(O) uint16
	moverGoTo             func(O, W)
	firstObject           func() O
	loadMoverUpdate       func(O) DU
	loadExtent            func(O) uint32
	loadMoverTargetExtent func(DU) uint32
	nextObject            func(O) O
}

// scriptMoveTo5123C0 implements the original script Move dispatch. It keeps
// the monster UpdateData handle cached across action-stack callbacks and
// reloads the mover type for every scanned object exactly as GAME.EXE does.
// The original has no null source, waypoint, or UpdateData guard, so this
// semantic core intentionally adds none.
func scriptMoveTo5123C0[O comparable, W, MU, DU any, A comparable](
	unit O,
	waypoint W,
	hooks scriptMoveHooks5123C0[O, W, MU, DU, A],
) {
	if hooks.loadFlags(unit)&scriptMoveBlockedFlag5123C0 != 0 {
		return
	}
	if hooks.loadClassLow(unit)&scriptMoveMonsterClassLow5123C0 != 0 {
		update := hooks.loadMonsterUpdate(unit)
		hooks.clearActionStack(unit)

		report := hooks.pushAction(unit, scriptMoveReportAction5123C0)
		var nilAction A
		if report != nilAction {
			hooks.storeActionArgU32(report, 0, scriptMoveReportArg5123C0)
		}

		if hooks.loadWaypointPoints(waypoint) != 0 {
			roam := hooks.pushAction(unit, scriptMoveRoamAction5123C0)
			if roam != nilAction {
				hooks.storeActionWaypoint(roam, 0, waypoint)
				flags := hooks.loadRoamFlagsLow(update)
				hooks.storeActionArgLow(roam, 2, flags)
			}
		}

		move := hooks.pushAction(unit, scriptMoveFarAction5123C0)
		if move == nilAction {
			return
		}
		x := hooks.loadWaypointXBits(waypoint)
		hooks.storeActionArgU32(move, 0, x)
		y := hooks.loadWaypointYBits(waypoint)
		hooks.storeActionArgU32(move, 1, y)
		hooks.storeActionArgU32(move, 2, 0)
		return
	}

	moverType := hooks.loadMoverType()
	if uint32(hooks.loadTypeInd(unit)) == moverType {
		hooks.moverGoTo(unit, waypoint)
		return
	}
	var nilObject O
	for obj := hooks.firstObject(); obj != nilObject; obj = hooks.nextObject(obj) {
		moverType = hooks.loadMoverType()
		if uint32(hooks.loadTypeInd(obj)) != moverType {
			continue
		}
		update := hooks.loadMoverUpdate(obj)
		extent := hooks.loadExtent(unit)
		if hooks.loadMoverTargetExtent(update) == extent {
			hooks.moverGoTo(obj, waypoint)
		}
	}
}

// moverGoToHooks5124C0 exposes GAME.EXE 005124C0. The cached update handle is
// separate from the live object because SetOn may invoke arbitrary callbacks.
type moverGoToHooks5124C0[O, W, U any] struct {
	loadUpdate         func(O) U
	setOn              func(O)
	storeVelocityXBits func(O, uint32)
	storeVelocityYBits func(O, uint32)
	storeStateLow      func(U, uint8)
	loadWaypointIndex  func(W) uint32
	storeWaypointIndex func(U, uint32)
	addToUpdatable     func(O)
}

func moverGoTo5124C0[O, W, U any](unit O, waypoint W, hooks moverGoToHooks5124C0[O, W, U]) {
	update := hooks.loadUpdate(unit)
	hooks.setOn(unit)
	hooks.storeVelocityXBits(unit, 0)
	hooks.storeVelocityYBits(unit, 0)
	hooks.storeStateLow(update, 0)
	index := hooks.loadWaypointIndex(waypoint)
	hooks.storeWaypointIndex(update, index)
	hooks.addToUpdatable(unit)
}
