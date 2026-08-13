package server

const (
	doorCollideFeedbackCooldown4E8AC0 = uint64(1500)
	doorCollideTileScale4E8AC0        = int32(23)
	doorCollideHalfExtent4E8AC0       = float64(34)

	doorCollideSoundUnlock4E8AC0 = uint32(234)
	doorCollideSoundDoor4E8AC0   = uint32(240)
	doorCollideSoundGate4E8AC0   = uint32(244)

	doorCollideGateSubclassByte4E8AC0 = uint8(0x04)
	doorCollidePlayerClassByte4E8AC0  = uint8(0x04)

	doorCollideGateLockedMagic4E8AC0     = "objcoll.c:GateLockedMagic"
	doorCollideDoorLockedMagic4E8AC0     = "objcoll.c:DoorLockedMagic"
	doorCollideGateLockedMechanism4E8AC0 = "objcoll.c:GateLockedMechanism"
	doorCollideDoorLockedMechanism4E8AC0 = "objcoll.c:DoorLockedMechanism"
	doorCollideKeyShared4E8AC0           = "GeneralPrint:KeyShared1"
	doorCollideGateLockedKey4E8AC0       = "objcoll.c:GateLockedKey"
	doorCollideDoorLockedKey4E8AC0       = "objcoll.c:DoorLockedKey"
)

type doorCollideRect4E8AC0 struct {
	MinX float32
	MinY float32
	MaxX float32
	MaxY float32
}

type doorCollideHooks4E8AC0[O comparable, U any] struct {
	loadUpdateData       func(O) U
	loadCurrentDirection func(U) int32
	loadTargetDirection  func(U) int32
	loadOwner            func(O) O
	loadOwnerExpiryFrame func(O) uint32
	frame                func() uint32
	storeOwner           func(O, O)
	ticks                func() uint64
	loadFeedbackTicks    func() uint64
	storeFeedbackTicks   func(uint64)
	loadSubclassByte     func(O) uint8
	audio                func(uint32, O)
	priorityMessage      func(O, string)
	loadLockCode         func(U) uint8
	findKey              func(O, O) O
	keyMessage           func(O, string, uint8)
	loadTileX            func(U) int32
	loadTileY            func(U) int32
	storeLockCode        func(U, uint8)
	questMode            func() bool
	questSync            func(O) int32
	storeQuestFrame      func(uint32)
	eachObjectInRect     func(doorCollideRect4E8AC0, DoorTilePoint)
	loadInventoryHolder  func(O) O
	loadClassByte        func(O) uint8
	questKeyState        func() int32
	delayedDelete        func(O)
}

// doorCollide4E8AC0 preserves GAME.EXE 004E8AC0. Door update data is loaded
// before the nil-unit branch, owner expiry uses an unsigned frame comparison,
// and the registered third collision argument is ignored. All callback-visible
// field reloads retain their original order, including the key lock byte after
// audio and the key holder immediately before the shared-key message.
func doorCollide4E8AC0[O comparable, U, C any](
	door, unit O,
	collision C,
	hooks doorCollideHooks4E8AC0[O, U],
) {
	_ = collision
	update := hooks.loadUpdateData(door)

	var zero O
	if unit == zero {
		return
	}
	if hooks.loadCurrentDirection(update) != hooks.loadTargetDirection(update) {
		return
	}

	owner := hooks.loadOwner(door)
	if owner != zero {
		if hooks.loadOwnerExpiryFrame(door) <= hooks.frame() {
			hooks.storeOwner(door, zero)
		} else if owner != unit {
			doorCollideFeedback4E8AC0(
				door, unit, update,
				doorCollideGateLockedMagic4E8AC0,
				doorCollideDoorLockedMagic4E8AC0,
				false,
				hooks,
			)
			return
		}
	}

	lockCode := hooks.loadLockCode(update)
	if lockCode == 0 {
		return
	}
	if lockCode == 5 {
		doorCollideFeedback4E8AC0(
			door, unit, update,
			doorCollideGateLockedMechanism4E8AC0,
			doorCollideDoorLockedMechanism4E8AC0,
			false,
			hooks,
		)
		return
	}

	key := hooks.findKey(unit, door)
	if key == zero {
		doorCollideFeedback4E8AC0(
			door, unit, update,
			doorCollideGateLockedKey4E8AC0,
			doorCollideDoorLockedKey4E8AC0,
			true,
			hooks,
		)
		return
	}

	tileX := hooks.loadTileX(update)
	tileY := hooks.loadTileY(update)
	hooks.storeLockCode(update, 0)
	direction := hooks.loadTargetDirection(update)
	rect, target := doorCollideGeometry4E8AC0(tileX, tileY, direction)

	if hooks.questMode() {
		_ = hooks.questSync(door)
		hooks.storeQuestFrame(hooks.frame())
	}
	hooks.eachObjectInRect(rect, target)
	hooks.audio(doorCollideSoundUnlock4E8AC0, door)

	holder := hooks.loadInventoryHolder(key)
	if holder != zero && holder != unit &&
		hooks.loadClassByte(holder)&doorCollidePlayerClassByte4E8AC0 != 0 &&
		hooks.questMode() && hooks.questKeyState() == 1 {
		hooks.priorityMessage(hooks.loadInventoryHolder(key), doorCollideKeyShared4E8AC0)
	}
	hooks.delayedDelete(key)
}

func doorCollideFeedback4E8AC0[O comparable, U any](
	door, unit O,
	update U,
	gateMessage, doorMessage string,
	keyed bool,
	hooks doorCollideHooks4E8AC0[O, U],
) {
	if hooks.ticks()-hooks.loadFeedbackTicks() <= doorCollideFeedbackCooldown4E8AC0 {
		return
	}

	message := doorMessage
	sound := doorCollideSoundDoor4E8AC0
	if hooks.loadSubclassByte(door)&doorCollideGateSubclassByte4E8AC0 != 0 {
		message = gateMessage
		sound = doorCollideSoundGate4E8AC0
	}
	hooks.audio(sound, door)
	if keyed {
		hooks.keyMessage(unit, message, hooks.loadLockCode(update))
	} else {
		hooks.priorityMessage(unit, message)
	}
	hooks.storeFeedbackTicks(hooks.ticks())
}

func doorCollideGeometry4E8AC0(tileX, tileY, direction int32) (doorCollideRect4E8AC0, DoorTilePoint) {
	centerX := tileX * doorCollideTileScale4E8AC0
	centerY := tileY * doorCollideTileScale4E8AC0
	centerY32 := float32(centerY)
	rect := doorCollideRect4E8AC0{
		MinX: float32(float64(centerX) - doorCollideHalfExtent4E8AC0),
		MinY: float32(float64(centerY) - doorCollideHalfExtent4E8AC0),
		MaxX: float32(float64(centerX) + doorCollideHalfExtent4E8AC0),
		// GAME.EXE spills only the Y center to binary32 before this addition.
		MaxY: float32(float64(centerY32) + doorCollideHalfExtent4E8AC0),
	}

	// Production DoorUpdate records use only these four direction values. The
	// original leaves the stack target indeterminate for every other value;
	// native-width ports use a deterministic zero target instead of propagating
	// uninitialized memory.
	var target DoorTilePoint
	switch direction {
	case 0:
		target = DoorTilePoint{X: tileX - 1, Y: tileY - 1}
	case 8:
		target = DoorTilePoint{X: tileX + 1, Y: tileY - 1}
	case 16:
		target = DoorTilePoint{X: tileX + 1, Y: tileY + 1}
	case 24:
		target = DoorTilePoint{X: tileX - 1, Y: tileY + 1}
	}
	return rect, target
}
