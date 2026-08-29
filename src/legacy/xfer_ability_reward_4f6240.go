package legacy

import "bytes"

const (
	abilityRewardXferCurrentVersion4F6240 = uint16(61)
	abilityRewardXferNameLimit4F6240      = 128
)

// abilityRewardXferDeps4F6240 exposes every observable object, use-data,
// stream, ability-table, mode, and inventory access in GAME.EXE 004F6240.
// Object and use-data identities remain generic so the contract cannot
// inherit PE32 pointer truncation.
type abilityRewardXferDeps4F6240[O, D any] struct {
	loadField34  func(O) uint32
	loadUseData  func(O) D
	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32

	loadAbility  func(D) uint8
	abilityName  func(uint8) string
	rwByte       func(uint8) uint8
	rwBytes      func([]byte)
	abilityID    func(string) int32
	storeAbility func(D, uint8)

	readMode          func() int32
	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// abilityRewardXfer4F6240 preserves the PE32 cache order, signed version
// gate, unconditional name round trip, one-byte length gate, low-byte ability
// store, live Field34 reload, exact-one read mode, zero-extended inventory
// version, and failure prefixes.
//
// UseData deliberately has no nil guard. The original caches the pointer on
// entry but first dereferences it only after common object serialization has
// succeeded. Registered ability names fit the 128-byte local buffer. A longer
// runtime name crossed an undefined strcpy boundary in PE32, so the native
// contract faults instead of inventing a wire encoding or rollback result.
func abilityRewardXfer4F6240[O, D any](
	object O,
	deps abilityRewardXferDeps4F6240[O, D],
) int32 {
	originalField34 := deps.loadField34(object)
	data := deps.loadUseData(object)

	versionWord := deps.rwVersion(abilityRewardXferCurrentVersion4F6240)
	version := int16(versionWord)
	if version > int16(abilityRewardXferCurrentVersion4F6240) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	name := deps.abilityName(deps.loadAbility(data))
	if len(name) >= abilityRewardXferNameLimit4F6240 {
		panic("AbilityRewardXfer ability name exceeds the PE32 stack buffer")
	}
	var buf [abilityRewardXferNameLimit4F6240]byte
	copy(buf[:], name)
	size := deps.rwByte(uint8(len(name)))
	if size >= abilityRewardXferNameLimit4F6240 {
		return 0
	}
	deps.rwBytes(buf[:int(size)])
	buf[size] = 0
	cname := buf[:size]
	if end := bytes.IndexByte(cname, 0); end >= 0 {
		cname = cname[:end]
	}
	deps.storeAbility(data, uint8(deps.abilityID(string(cname))))

	liveField34 := deps.loadField34(object)
	if liveField34 != 0 && deps.readMode() == 1 {
		if deps.transferInventory(versionWord, object, int32(liveField34)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}
