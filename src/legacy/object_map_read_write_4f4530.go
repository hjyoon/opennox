package legacy

const (
	objectMapCurrentVersion4F4530 = uint16(64)
	objectMapModernVersion4F4530  = int32(61)
	objectMapScriptVersion4F4530  = int32(63)
	objectMapFrameVersion4F4530   = int32(64)
	objectMapVersionedMap4F4530   = int32(40)
	objectMapDecayFlag4F4530      = uint32(0x00400000)
	objectMapGameFlag22_4F4530    = uint32(0x00200000)
	objectMapGameFlag23_4F4530    = uint32(0x00400000)
	objectMapFlagsMask4F4530      = uint32(0x11408162)
	objectMapFlagsKeepMask4F4530  = uint32(0xeebf7e9d)
	objectMapEnabledFlag4F4530    = uint32(0x01000000)
	objectMapPreserveFlag4F4530   = uint32(0x00000040)
	objectMapOwnedSkipFlag4F4530  = uint32(0x00000020)
	objectMapStatusMask4F4530     = uint32(0x0000005e)
)

// objectMapReadWriteDeps4F4530 exposes every observable object access,
// transfer, global read, allocation, and callback in GAME.EXE 004F4530.
// Object and text identities stay generic so this contract cannot silently
// inherit the original PE32 pointer width.
type objectMapReadWriteDeps4F4530[O comparable, P comparable] struct {
	loadField34 func(O) uint32
	readOnly    func() int32
	rwU16       func(uint16) uint16
	readOld     func(O, int32, int32) int32

	storeField34  func(O, uint32)
	rwExtent      func(O)
	rwScriptID    func(O)
	loadScriptID  func(O) int32
	gameFlags     func(uint32) int32
	nextScriptID  func() int32
	storeScriptID func(O, int32)

	rwPositionX       func(O)
	rwPositionY       func(O)
	loadPositionX     func(O) float32
	loadPositionY     func(O) float32
	storeNewPositionX func(O, float32)
	storeNewPositionY func(O, float32)

	extendedAdmission func(O) int8
	rwU8              func(uint8) uint8
	loadFlags         func(O) uint32
	rwU32             func(uint32) uint32
	storeFlags        func(O, uint32)
	setOn             func(O)
	setOff            func(O)

	loadIDPointer  func(O) P
	stringLength   func(P) uintptr
	allocateID     func(uint16) P
	storeIDPointer func(O, P)
	rwIDBytes      func(P, uint8)
	terminateID    func(P, uint8)

	rwTeamID          func(O)
	loadInventoryHead func(O) O
	loadInventoryNext func(O) O

	loadField129      func(O) O
	loadTypeInd       func(O) uint16
	ownedTypeAllowed  func(uint16) int32
	loadField128      func(O) O
	rwI32             func(int32) int32
	readOwnedScriptID func() int32
	addPendingOwn     func(int32, int32)
	rwOwnedScriptID   func(O)

	loadField5  func(O) uint32
	unsetStatus func(O, uint32)
	setStatus   func(O, uint32)

	loadField189  func(O) P
	scriptHandler func(O, P) int32
	gameFrame     func() uint32
	storeField32  func(O, uint32)
}

func objectMapOwned4F4530[O comparable, P comparable](
	object O,
	deps objectMapReadWriteDeps4F4530[O, P],
) bool {
	if deps.loadFlags(object)&objectMapOwnedSkipFlag4F4530 != 0 {
		return false
	}
	return deps.ownedTypeAllowed(deps.loadTypeInd(object)) != 0
}

// objectMapReadWrite4F4530 preserves GAME.EXE's common object record. The
// inconsistent read-only comparisons, signed version checks, live reloads,
// admission byte, and wrapping frame delta are all observable and must not be
// normalized. There is deliberately no nil-object guard.
func objectMapReadWrite4F4530[O comparable, P comparable](
	object O,
	mapVersion int32,
	deps objectMapReadWriteDeps4F4530[O, P],
) int32 {
	originalField34 := deps.loadField34(object)

	var objectVersion int32
	if mapVersion >= objectMapVersionedMap4F4530 || deps.readOnly() == 0 {
		versionWord := deps.rwU16(objectMapCurrentVersion4F4530)
		version := int16(versionWord)
		if version > int16(objectMapCurrentVersion4F4530) {
			return 0
		}
		objectVersion = int32(version)
	}
	if mapVersion < objectMapVersionedMap4F4530 || objectVersion < objectMapModernVersion4F4530 {
		return deps.readOld(object, objectVersion, mapVersion)
	}

	if deps.readOnly() == 1 {
		deps.storeField34(object, 0)
	}
	deps.rwExtent(object)
	deps.rwScriptID(object)
	if deps.loadScriptID(object) == 0 &&
		deps.readOnly() == 1 &&
		deps.gameFlags(objectMapGameFlag22_4F4530) == 0 &&
		deps.gameFlags(objectMapGameFlag23_4F4530) == 0 {
		deps.storeScriptID(object, deps.nextScriptID())
	}

	if deps.readOnly() == 0 {
		deps.rwPositionX(object)
		deps.rwPositionY(object)
	} else {
		deps.rwPositionX(object)
		deps.rwPositionY(object)
		x := deps.loadPositionX(object)
		y := deps.loadPositionY(object)
		deps.storeNewPositionX(object, x)
		deps.storeNewPositionY(object, y)
	}

	hasExtendedData := deps.rwU8(uint8(deps.extendedAdmission(object)))
	if hasExtendedData == 0 {
		return 1
	}

	originalFlags := deps.loadFlags(object)
	serializedFlags := deps.rwU32(originalFlags & objectMapFlagsMask4F4530)
	flags := deps.loadFlags(object) & objectMapFlagsKeepMask4F4530
	deps.storeFlags(object, flags)
	if originalFlags&objectMapPreserveFlag4F4530 != 0 {
		flags |= objectMapPreserveFlag4F4530
		deps.storeFlags(object, flags)
	}
	flags = deps.loadFlags(object)
	deps.storeFlags(object, flags|serializedFlags)
	if deps.readOnly() == 1 {
		if serializedFlags&objectMapEnabledFlag4F4530 != 0 {
			deps.setOn(object)
		} else {
			deps.setOff(object)
		}
	}

	var zeroPointer P
	id := deps.loadIDPointer(object)
	var idLength uint8
	if id != zeroPointer {
		idLength = uint8(deps.stringLength(id))
	}
	idLength = deps.rwU8(idLength)
	if deps.readOnly() == 1 && idLength != 0 {
		id = deps.allocateID(uint16(idLength) + 1)
		deps.storeIDPointer(object, id)
		if id == zeroPointer {
			return 0
		}
	}
	id = deps.loadIDPointer(object)
	deps.rwIDBytes(id, idLength)
	id = deps.loadIDPointer(object)
	if id != zeroPointer {
		deps.terminateID(id, idLength)
	}

	deps.rwTeamID(object)

	var zeroObject O
	var inventoryCount uint8
	for item := deps.loadInventoryHead(object); item != zeroObject; {
		item = deps.loadInventoryNext(item)
		inventoryCount++
	}
	inventoryCount = deps.rwU8(inventoryCount)
	if deps.readOnly() == 1 {
		deps.storeField34(object, uint32(inventoryCount))
	}

	var ownedCount uint32
	for owned := deps.loadField129(object); owned != zeroObject; {
		if objectMapOwned4F4530(owned, deps) {
			ownedCount++
		}
		owned = deps.loadField128(owned)
	}
	transferredOwnedCount := deps.rwU16(uint16(ownedCount))
	if deps.readOnly() != 0 {
		for i := uint32(0); i < uint32(transferredOwnedCount); i++ {
			ownedScriptID := deps.readOwnedScriptID()
			if deps.gameFlags(objectMapGameFlag22_4F4530) == 0 &&
				deps.gameFlags(objectMapGameFlag23_4F4530) == 0 {
				deps.addPendingOwn(deps.loadScriptID(object), ownedScriptID)
			}
		}
	} else {
		for owned := deps.loadField129(object); owned != zeroObject; {
			if objectMapOwned4F4530(owned, deps) {
				deps.rwOwnedScriptID(owned)
			}
			owned = deps.loadField128(owned)
		}
	}

	status := deps.rwU32(deps.loadField5(object) & objectMapStatusMask4F4530)
	deps.unsetStatus(object, objectMapStatusMask4F4530)
	deps.setStatus(object, status)

	if objectVersion >= objectMapScriptVersion4F4530 {
		context := deps.loadField189(object)
		if deps.scriptHandler(object, context) == 0 {
			return 0
		}
	}
	if objectVersion >= objectMapFrameVersion4F4530 {
		frameDelta := int32(originalField34 - deps.gameFrame())
		frameDelta = deps.rwI32(frameDelta)
		if frameDelta > 0 &&
			deps.readOnly() == 1 &&
			deps.loadFlags(object)&objectMapDecayFlag4F4530 != 0 {
			deps.storeField32(object, uint32(frameDelta))
		}
	}
	return 1
}
