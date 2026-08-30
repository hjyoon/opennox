package legacy

import "bytes"

const (
	rewardMarkerXferCurrentVersion4F74D0 = uint16(63)
	rewardMarkerXferSpellCount4F74D0     = 137
	rewardMarkerXferAbilityCount4F74D0   = 6
	rewardMarkerXferGuideCount4F74D0     = 41
)

type rewardMarkerXferHeader4F74D0 uint8

const (
	rewardMarkerXferCategoryMask4F74D0 rewardMarkerXferHeader4F74D0 = iota
	rewardMarkerXferRewardFlags4F74D0
)

type rewardMarkerXferList4F74D0 uint8

const (
	rewardMarkerXferSpells4F74D0 rewardMarkerXferList4F74D0 = iota
	rewardMarkerXferAbilities4F74D0
	rewardMarkerXferGuides4F74D0
)

type rewardMarkerXferField4F74D0 uint8

const (
	// GAME.EXE transfers these records in address order 196, 192, 200,
	// 204, 208, followed by the version-gated 212 and low byte of 216.
	rewardMarkerXferField196_4F74D0 rewardMarkerXferField4F74D0 = iota
	rewardMarkerXferField192_4F74D0
	rewardMarkerXferField200_4F74D0
	rewardMarkerXferField204_4F74D0
	rewardMarkerXferField208_4F74D0
	rewardMarkerXferField212_4F74D0
	rewardMarkerXferField216Low_4F74D0
)

func rewardMarkerXferListSize4F74D0(list rewardMarkerXferList4F74D0) int {
	switch list {
	case rewardMarkerXferSpells4F74D0:
		return rewardMarkerXferSpellCount4F74D0
	case rewardMarkerXferAbilities4F74D0:
		return rewardMarkerXferAbilityCount4F74D0
	case rewardMarkerXferGuides4F74D0:
		return rewardMarkerXferGuideCount4F74D0
	default:
		panic("invalid RewardMarkerXfer list")
	}
}

// rewardMarkerXferDeps4F74D0 exposes every observable object, InitData,
// stream, name-table, mode, and inventory access in GAME.EXE 004F74D0.
// Object and InitData identities stay generic so the semantic contract cannot
// inherit PE32 pointer truncation.
type rewardMarkerXferDeps4F74D0[O, D any] struct {
	loadInitData func(O) D
	loadField34  func(O) uint32
	storeField34 func(O, uint32)

	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32
	rwHeader     func(D, rewardMarkerXferHeader4F74D0)

	loadListValue  func(D, rewardMarkerXferList4F74D0, int) uint8
	storeListValue func(D, rewardMarkerXferList4F74D0, int, uint8)
	rwCount        func(uint16) uint16
	readMode       func() int32
	rwNameLength   func(uint8) uint8
	rwNameBytes    func([]byte)
	resolveName    func(rewardMarkerXferList4F74D0, []byte) int
	loadName       func(rewardMarkerXferList4F74D0, int) []byte

	rwField           func(D, rewardMarkerXferField4F74D0)
	transferInventory func(uint16, O, int32) int32
}

func rewardMarkerXferNames4F74D0[O, D any](
	data D,
	list rewardMarkerXferList4F74D0,
	deps rewardMarkerXferDeps4F74D0[O, D],
) bool {
	size := rewardMarkerXferListSize4F74D0(list)
	var exactOneCount uint16
	for index := 0; index < size; index++ {
		if deps.loadListValue(data, list, index) == 1 {
			exactOneCount++
		}
	}
	count := deps.rwCount(exactOneCount)

	// Each list queries the stream mode independently. Any nonzero value is
	// the read branch here; the suffix inventory check is deliberately stricter.
	if deps.readMode() != 0 {
		for index := uint16(0); index < count; index++ {
			length := deps.rwNameLength(0)
			var name [256]byte
			deps.rwNameBytes(name[:length])
			cname := name[:length]
			if end := bytes.IndexByte(cname, 0); end >= 0 {
				cname = cname[:end]
			}
			id := deps.resolveName(list, cname)
			if id == 0 {
				return false
			}
			// GAME.EXE does not clear the list and does not validate the
			// resolved ID against the nominal list extent.
			deps.storeListValue(data, list, id, 1)
		}
		return true
	}

	for index := 0; index < size; index++ {
		if deps.loadListValue(data, list, index) == 0 {
			continue
		}
		// The name getter is called twice. Only the low byte of the first
		// name's length is serialized; the second pointer supplies bytes.
		firstName := deps.loadName(list, index)
		length := deps.rwNameLength(uint8(len(firstName)))
		secondName := deps.loadName(list, index)
		deps.rwNameBytes(secondName[:length])
	}
	return true
}

// rewardMarkerXfer4F74D0 preserves GAME.EXE 004F74D0's transfer order and
// failure prefixes while keeping object and InitData identities native-width.
// The two four-byte header records and suffix records retain their PE32 wire
// sizes; object references are handled by common serialization and inventory.
//
// Versions are gated as signed 16-bit values. List counts are not clamped,
// read-mode is sampled afresh for all three lists and the suffix, and stream
// return values are ignored. A failed name lookup or inventory transfer does
// not restore Field34. Only a successful or skipped suffix restores the entry
// value. There are deliberately no nil, callback, or resolved-ID guards.
func rewardMarkerXfer4F74D0[O, D any](
	object O,
	deps rewardMarkerXferDeps4F74D0[O, D],
) int32 {
	data := deps.loadInitData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(rewardMarkerXferCurrentVersion4F74D0)
	version := int16(versionWord)
	if version <= 0 || version > int16(rewardMarkerXferCurrentVersion4F74D0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	deps.rwHeader(data, rewardMarkerXferCategoryMask4F74D0)
	deps.rwHeader(data, rewardMarkerXferRewardFlags4F74D0)
	for _, list := range [...]rewardMarkerXferList4F74D0{
		rewardMarkerXferSpells4F74D0,
		rewardMarkerXferAbilities4F74D0,
		rewardMarkerXferGuides4F74D0,
	} {
		if !rewardMarkerXferNames4F74D0(data, list, deps) {
			return 0
		}
	}

	for _, field := range [...]rewardMarkerXferField4F74D0{
		rewardMarkerXferField196_4F74D0,
		rewardMarkerXferField192_4F74D0,
		rewardMarkerXferField200_4F74D0,
		rewardMarkerXferField204_4F74D0,
		rewardMarkerXferField208_4F74D0,
	} {
		deps.rwField(data, field)
	}
	if version >= 62 {
		deps.rwField(data, rewardMarkerXferField212_4F74D0)
	}
	if version >= 63 {
		deps.rwField(data, rewardMarkerXferField216Low_4F74D0)
	}

	liveField34 := deps.loadField34(object)
	if liveField34 != 0 && deps.readMode() == 1 {
		if deps.transferInventory(versionWord, object, int32(liveField34)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}
