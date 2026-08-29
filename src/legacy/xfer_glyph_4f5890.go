package legacy

const glyphXferCurrentVersion4F5890 = uint16(60)

// glyphXferDeps4F5890 exposes every observable object/glyph-data access,
// transfer, mode read, spell lookup, and external call in GAME.EXE 004F5890.
// Object, glyph-data, and spell-name identities remain generic so this
// contract cannot inherit the original PE32 pointer width.
type glyphXferDeps4F5890[O comparable, D, N any] struct {
	loadGlyphData func(O) D
	loadField34   func(O) uint32
	rwVersion     func(uint16) uint16
	mapReadWrite  func(O, int32) int32

	rwLegacyDword func()
	rwDirection1  func(O)
	rwTargetX     func(D)
	rwTargetY     func(D)
	rwSpellCount  func(D)

	readOnly       func() int32
	loadSpellCount func(D) uint8
	rwLegacySpells func(D)

	rwNameLength func(uint8) uint8
	rwNameBytes  func([]byte)
	spellID      func(string) uint32
	storeSpell   func(D, int, uint32)

	loadSpell        func(D, int) uint32
	spellName        func(uint32) N
	spellNameLength  func(N) uint8
	rwSpellNameBytes func(N, uint8)

	copyDirection          func(O)
	clearSpellTargetObject func(D)
	transferInventory      func(uint16, O, int32) int32
	storeField34           func(O, uint32)
}

// glyphXfer4F5890 preserves the entry-time GlyphInitData and Field34 caches,
// signed version thresholds, low-byte count semantics, repeated live spell
// count/ID loads, exact-one mode gates, and the original inventory-failure
// prefix. There are deliberately no object, glyph-data, spell-count, or spell
// identity guards.
func glyphXfer4F5890[O comparable, D, N any](
	object O,
	deps glyphXferDeps4F5890[O, D, N],
) int32 {
	data := deps.loadGlyphData(object)
	originalField34 := deps.loadField34(object)

	versionWord := deps.rwVersion(glyphXferCurrentVersion4F5890)
	version := int16(versionWord)
	if version > int16(glyphXferCurrentVersion4F5890) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}
	if version < 41 {
		deps.rwLegacyDword()
	}

	deps.rwDirection1(object)
	deps.rwTargetX(data)
	deps.rwTargetY(data)
	deps.rwSpellCount(data)

	postRead := false
	if deps.readOnly() == 1 {
		if version < 31 {
			deps.rwLegacySpells(data)
			postRead = deps.readOnly() == 1
		} else if deps.loadSpellCount(data) == 0 {
			// 004F5969 jumps directly to the read post-processing block.
			// Unlike every other completed spell path, it does not reload
			// the mode at 004F5A27.
			postRead = true
		} else {
			for index := 0; ; index++ {
				length := deps.rwNameLength(0)
				name := make([]byte, int(length))
				deps.rwNameBytes(name)
				nameEnd := len(name)
				for i, ch := range name {
					if ch == 0 {
						nameEnd = i
						break
					}
				}
				deps.storeSpell(data, index, deps.spellID(string(name[:nameEnd])))
				if index+1 >= int(deps.loadSpellCount(data)) {
					break
				}
			}
			postRead = deps.readOnly() == 1
		}
	} else {
		if deps.loadSpellCount(data) != 0 {
			for index := 0; ; index++ {
				spell := deps.loadSpell(data, index)
				name := deps.spellName(spell)
				length := deps.rwNameLength(deps.spellNameLength(name))

				// The original reloads both the spell ID and its name after
				// transferring the length byte. Do not reuse either cache.
				spell = deps.loadSpell(data, index)
				name = deps.spellName(spell)
				deps.rwSpellNameBytes(name, length)
				if index+1 >= int(deps.loadSpellCount(data)) {
					break
				}
			}
		}
		postRead = deps.readOnly() == 1
	}

	if postRead {
		deps.copyDirection(object)
		deps.clearSpellTargetObject(data)
	}

	liveField34 := deps.loadField34(object)
	if liveField34 != 0 && deps.readOnly() == 1 {
		if deps.transferInventory(versionWord, object, int32(liveField34)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}
