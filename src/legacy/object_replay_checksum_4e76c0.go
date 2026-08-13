package legacy

type objectReplayChecksumPassHooks4E76C0[O comparable] struct {
	first    func() O
	noop     func(O)
	checksum func(O) uint32
	next     func(O) O
}

// objectReplayChecksumPass4E76C0 is the pointer-width-independent traversal
// contract for GAME.EXE 004E76C0. The checksum accumulator is intentionally
// unobservable, but every callback and the post-checksum next lookup remain
// observable just as they are in the original routine.
func objectReplayChecksumPass4E76C0[O comparable](h objectReplayChecksumPassHooks4E76C0[O]) O {
	var zero O
	obj := h.first()
	var checksum uint32
	for obj != zero {
		h.noop(obj)
		checksum ^= h.checksum(obj)
		obj = h.next(obj)
	}
	_ = checksum
	return obj
}

type objectReplayChecksumInput4E7700 struct {
	teamID       uint8
	typeInd      uint16
	scriptID     int32
	posXBits     uint32
	extent       uint32
	netCode      uint32
	field5       uint32
	objFlags     uint32
	posYBits     uint32
	newPosXBits  uint32
	newPosYBits  uint32
	prevPosXBits uint32
	prevPosYBits uint32
	velXBits     uint32
	velYBits     uint32
	forceXBits   uint32
	forceYBits   uint32
	pos24XBits   uint32
	pos24YBits   uint32
	zBits        uint32
	field27Bits  uint32
	direction1   int16
	direction2   int16
	field38      uint32
	field37      uint32
	field34      uint32
	field33      uint32
	field32      uint32
	massBits     uint32
	buffs        uint32
	field62      uint32

	healthPresent bool
	healthMax     uint16
	healthField2  uint16
	healthCur     uint16
}

// objectReplayChecksum4E7700 reproduces GAME.EXE 004E7700 using explicitly
// fixed-width inputs. Direction words are sign-extended; all float fields are
// supplied as their raw binary32 bit patterns. Health words follow the exact
// original +4, +2, +0 read order after the pointer gate.
func objectReplayChecksum4E7700(v objectReplayChecksumInput4E7700) uint32 {
	checksum := uint32(v.teamID)
	checksum ^= uint32(v.typeInd)
	checksum ^= uint32(v.scriptID)
	checksum ^= v.posXBits
	checksum ^= v.extent
	checksum ^= v.netCode
	checksum ^= v.field5
	checksum ^= v.objFlags
	checksum ^= v.posYBits
	checksum ^= v.newPosXBits
	checksum ^= v.newPosYBits
	checksum ^= v.prevPosXBits
	checksum ^= v.prevPosYBits
	checksum ^= v.velXBits
	checksum ^= v.velYBits
	checksum ^= v.forceXBits
	checksum ^= v.forceYBits
	checksum ^= v.pos24XBits
	checksum ^= v.pos24YBits
	checksum ^= v.zBits
	checksum ^= v.field27Bits
	checksum ^= uint32(int32(v.direction1))
	checksum ^= uint32(int32(v.direction2))
	checksum ^= v.field38
	checksum ^= v.field37
	checksum ^= v.field34
	checksum ^= v.field33
	checksum ^= v.field32
	checksum ^= v.massBits
	checksum ^= v.buffs
	checksum ^= v.field62
	if v.healthPresent {
		checksum ^= uint32(v.healthMax)
		checksum ^= uint32(v.healthField2)
		checksum ^= uint32(v.healthCur)
	}
	return checksum
}
