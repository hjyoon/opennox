package legacy

import "math"

const (
	doorXferCurrentVersion4F4CB0  = uint16(60)
	doorXferTargetVersion4F4CB0   = int16(41)
	doorXferGridInverseBits4F4CB0 = uint32(0x3d321643)
)

// doorXferDeps4F4CB0 exposes every observable object/update-data access,
// transfer, global-mode read, direction-table read, and external call in
// GAME.EXE 004F4CB0. Object and update-data identities remain generic so the
// contract cannot inherit PE32 pointer width.
type doorXferDeps4F4CB0[O comparable, U any] struct {
	loadField34           func(O) uint32
	loadUpdateData        func(O) U
	rwVersion             func(uint16) uint16
	mapReadWrite          func(O, int32) int32
	readOnly              func() int32
	loadCurrentDirection  func(U) int32
	loadLockCode          func(U) uint8
	loadTargetDirection   func(U) int32
	rwDirection           func(int32) int32
	rwLockCode            func(int32) int32
	rwTargetDirection     func(int32) int32
	storeCurrentDirection func(U, int32)
	storeFractionalDir    func(U, int16)
	storeTargetDirection  func(U, int32)
	storeSyncedDirection  func(U, int32)
	loadDirectionX        func(int32) int32
	loadPositionX         func(O) float32
	loadDirectionY        func(int32) int32
	loadPositionY         func(O) float32
	truncQwordLow         func(float64) int32
	attachWall            func(O, int32, int32)
	storeTileX            func(U, int32)
	storeTileY            func(U, int32)
	storeLockCode         func(U, uint8)
	transferInventory     func(uint16, O, int32) int32
	storeField34          func(O, uint32)
}

// Keep the original x87 53-bit precision boundaries explicit and prevent a
// target compiler from contracting the addition and multiplication.
//
//go:noinline
func doorXferAdd64_4F4CB0(left, right float64) float64 { return left + right }

//go:noinline
func doorXferMul64_4F4CB0(left, right float64) float64 { return left * right }

func doorXferTileCoordinate4F4CB0(
	offset int32,
	position float32,
	truncQwordLow func(float64) int32,
) int32 {
	sum := doorXferAdd64_4F4CB0(float64(offset), float64(position))
	scaled := doorXferMul64_4F4CB0(
		sum,
		float64(math.Float32frombits(doorXferGridInverseBits4F4CB0)),
	)
	return truncQwordLow(scaled)
}

// doorXferTruncSignedQwordLow4F4CB0 models the low dword returned by the
// original helper at 00566DCC. Invalid x87 conversions produce integer
// indefinite 0x8000000000000000, whose low dword is zero.
func doorXferTruncSignedQwordLow4F4CB0(value float64) int32 {
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return 0
	}
	return int32(int64(math.Trunc(value)))
}

func doorXferFractionalDirection4F4CB0(direction int32) int16 {
	shifted := int32(uint32(direction) << 8)
	return int16(shifted / 32)
}

// doorXfer4F4CB0 preserves the original entry-time Field34 and UpdateData
// caches, signed version branches, exact-zero write cache, exact-one read
// application, 32-bit wrapping direction arithmetic, x87 tile conversion,
// live inventory gate, and failure prefixes. There are deliberately no
// object, update-data, or direction-index guards.
func doorXfer4F4CB0[O comparable, U any](
	object O,
	deps doorXferDeps4F4CB0[O, U],
) int32 {
	originalField34 := deps.loadField34(object)
	update := deps.loadUpdateData(object)

	versionWord := deps.rwVersion(doorXferCurrentVersion4F4CB0)
	version := int16(versionWord)
	if version > int16(doorXferCurrentVersion4F4CB0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	var direction int32
	var lockCode int32
	var targetDirection int32
	if deps.readOnly() == 0 {
		direction = deps.loadCurrentDirection(update)
		lockCode = int32(deps.loadLockCode(update))
		targetDirection = deps.loadTargetDirection(update)
	}
	direction = deps.rwDirection(direction)
	lockCode = deps.rwLockCode(lockCode)
	if version < doorXferTargetVersion4F4CB0 {
		targetDirection = direction
	} else {
		targetDirection = deps.rwTargetDirection(targetDirection)
	}

	if deps.readOnly() == 1 {
		deps.storeCurrentDirection(update, direction)
		deps.storeFractionalDir(update, doorXferFractionalDirection4F4CB0(direction))
		deps.storeTargetDirection(update, targetDirection)
		deps.storeSyncedDirection(update, direction)

		offsetX := deps.loadDirectionX(targetDirection) / 2
		positionX := deps.loadPositionX(object)
		tileX := doorXferTileCoordinate4F4CB0(offsetX, positionX, deps.truncQwordLow)
		offsetY := deps.loadDirectionY(targetDirection) / 2
		positionY := deps.loadPositionY(object)
		tileY := doorXferTileCoordinate4F4CB0(offsetY, positionY, deps.truncQwordLow)

		deps.attachWall(object, tileX, tileY)
		deps.storeTileX(update, tileX)
		deps.storeTileY(update, tileY)
		deps.storeLockCode(update, uint8(lockCode))
	}

	inventoryCount := deps.loadField34(object)
	if inventoryCount != 0 && deps.readOnly() == 1 {
		if deps.transferInventory(versionWord, object, int32(inventoryCount)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}
