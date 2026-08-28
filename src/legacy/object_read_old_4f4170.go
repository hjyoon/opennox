package legacy

const (
	objectReadOldFlagsMask4F4170     = uint32(0x11408162)
	objectReadOldFlagsKeepMask4F4170 = uint32(0xeebf7e9d)
	objectReadOldEnabledFlag4F4170   = uint32(0x01000000)
	objectReadOldPreserveFlag4F4170  = uint32(0x00000040)
	objectReadOldOwnedSkipFlag4F4170 = uint32(0x00000020)
	objectReadOldStatusMask4F4170    = uint32(0x0000005e)
	objectReadOldGameFlag22_4F4170   = uint32(0x00200000)
	objectReadOldGameFlag23_4F4170   = uint32(0x00400000)
)

// objectReadOldDeps4F4170 exposes every observable field access, transfer,
// allocation, and callback in GAME.EXE 004F4170. Object and ID identities are
// generic so the semantic contract never assumes a 32-bit pointer width.
type objectReadOldDeps4F4170[O comparable, P comparable] struct {
	readOnly func() int32

	storeField34 func(O, uint32)
	rwExtent     func(O)
	loadFlags    func(O) uint32
	rwU32        func(uint32) uint32
	storeFlags   func(O, uint32)
	setOn        func(O)
	setOff       func(O)

	rwPositionX       func(O)
	rwPositionY       func(O)
	rwOldPosition     func(int32, int32) (int32, int32)
	storePositionX    func(O, float32)
	storePositionY    func(O, float32)
	loadPositionX     func(O) float32
	loadPositionY     func(O) float32
	storeNewPositionX func(O, float32)
	storeNewPositionY func(O, float32)

	loadIDPointer  func(O) P
	stringLength   func(P) uintptr
	rwU8           func(uint8) uint8
	allocateID     func(uint16) P
	storeIDPointer func(O, P)
	rwIDBytes      func(P, uint8)
	terminateID    func(P, uint8)

	rwTeamID          func(O)
	loadInventoryHead func(O) O
	loadInventoryNext func(O) O

	rwScriptID    func(O)
	loadScriptID  func(O) int32
	storeScriptID func(O, int32)
	gameFlags     func(uint32) int32
	nextScriptID  func() int32

	loadField129     func(O) O
	loadTypeInd      func(O) uint16
	ownedTypeAllowed func(uint16) int32
	loadField128     func(O) O
	rwU16            func(uint16) uint16
	rwI32            func(int32) int32
	addPendingOwn    func(int32, int32)
	rwOwnedScriptID  func(O)

	loadField5  func(O) uint32
	unsetStatus func(O, uint32)
	setStatus   func(O, uint32)
}

func objectReadOldOwned4F4170[O comparable, P comparable](
	object O,
	deps objectReadOldDeps4F4170[O, P],
) bool {
	if deps.loadFlags(object)&objectReadOldOwnedSkipFlag4F4170 != 0 {
		return false
	}
	return deps.ownedTypeAllowed(deps.loadTypeInd(object)) != 0
}

// objectReadOldVer4F4170 preserves the pre-v61 object record exactly. In
// particular, read-only is intentionally reloaded and compared inconsistently:
// some original branches require exact one while others accept any nonzero
// value. Do not cache it or add an entry nil guard.
func objectReadOldVer4F4170[O comparable, P comparable](
	object O,
	objectVersion, mapVersion int32,
	deps objectReadOldDeps4F4170[O, P],
) int32 {
	if deps.readOnly() == 1 {
		deps.storeField34(object, 0)
	}
	deps.rwExtent(object)

	originalFlags := deps.loadFlags(object)
	serializedFlags := deps.rwU32(originalFlags & objectReadOldFlagsMask4F4170)
	flags := deps.loadFlags(object) & objectReadOldFlagsKeepMask4F4170
	deps.storeFlags(object, flags)
	if originalFlags&objectReadOldPreserveFlag4F4170 != 0 {
		flags |= objectReadOldPreserveFlag4F4170
		deps.storeFlags(object, flags)
	}
	flags = deps.loadFlags(object)
	deps.storeFlags(object, serializedFlags|flags)
	if deps.readOnly() == 1 {
		if serializedFlags&objectReadOldEnabledFlag4F4170 != 0 {
			deps.setOn(object)
		} else {
			deps.setOff(object)
		}
	}

	if deps.readOnly() != 0 {
		if mapVersion < 40 || objectVersion < 4 {
			x, y := deps.rwOldPosition(0, 0)
			deps.storePositionX(object, float32(x))
			deps.storePositionY(object, float32(y))
		} else {
			deps.rwPositionX(object)
			deps.rwPositionY(object)
		}
		x := deps.loadPositionX(object)
		y := deps.loadPositionY(object)
		deps.storeNewPositionX(object, x)
		deps.storeNewPositionY(object, y)
	} else {
		deps.rwPositionX(object)
		deps.rwPositionY(object)
	}

	if mapVersion >= 10 {
		var zeroPointer P
		id := deps.loadIDPointer(object)
		var length uint8
		if id != zeroPointer {
			length = uint8(deps.stringLength(id))
		}
		length = deps.rwU8(length)
		if deps.readOnly() == 1 && length != 0 {
			id = deps.allocateID(uint16(length) + 1)
			deps.storeIDPointer(object, id)
			if id == zeroPointer {
				return 0
			}
		}
		id = deps.loadIDPointer(object)
		deps.rwIDBytes(id, length)
		id = deps.loadIDPointer(object)
		if id != zeroPointer {
			deps.terminateID(id, length)
		}
	}

	if mapVersion >= 20 {
		deps.rwTeamID(object)
	}

	if mapVersion >= 30 {
		var count uint8
		var zeroObject O
		for item := deps.loadInventoryHead(object); item != zeroObject; {
			next := deps.loadInventoryNext(item)
			count++
			item = next
		}
		count = deps.rwU8(count)
		if deps.readOnly() == 1 {
			deps.storeField34(object, uint32(count))
		}
	}

	if mapVersion < 40 {
		return 1
	}

	deps.rwScriptID(object)
	if deps.loadScriptID(object) == 0 &&
		deps.readOnly() == 1 &&
		deps.gameFlags(objectReadOldGameFlag22_4F4170) == 0 &&
		deps.gameFlags(objectReadOldGameFlag23_4F4170) == 0 {
		deps.storeScriptID(object, deps.nextScriptID())
	}

	if objectVersion >= 2 {
		var zeroObject O
		var ownedCount uint32
		for owned := deps.loadField129(object); owned != zeroObject; {
			if objectReadOldOwned4F4170(owned, deps) {
				ownedCount++
			}
			owned = deps.loadField128(owned)
		}

		var transferredCount uint16
		if objectVersion < 5 {
			serializedCount := deps.rwU32(uint32(uint16(ownedCount)))
			transferredCount = uint16(serializedCount)
		} else {
			transferredCount = deps.rwU16(uint16(ownedCount))
		}

		if deps.readOnly() != 0 {
			for i := uint32(0); i < uint32(transferredCount); i++ {
				ownedScriptID := deps.rwI32(mapVersion)
				if deps.gameFlags(objectReadOldGameFlag22_4F4170) == 0 &&
					deps.gameFlags(objectReadOldGameFlag23_4F4170) == 0 {
					deps.addPendingOwn(deps.loadScriptID(object), ownedScriptID)
				}
			}
		} else {
			for owned := deps.loadField129(object); owned != zeroObject; {
				if objectReadOldOwned4F4170(owned, deps) {
					deps.rwOwnedScriptID(owned)
				}
				owned = deps.loadField128(owned)
			}
		}
	}

	if objectVersion >= 3 {
		status := deps.rwU32(deps.loadField5(object) & objectReadOldStatusMask4F4170)
		deps.unsetStatus(object, objectReadOldStatusMask4F4170)
		deps.setStatus(object, status)
	}
	return 1
}
