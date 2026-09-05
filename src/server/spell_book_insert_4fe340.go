package server

const (
	spellBookInsertBlockedFlags4FE340 = uint32(0x8022)
	spellBookInsertPlayerClass4FE340  = uint8(0x04)
	spellBookInsertSpellCount4FE340   = int32(137)
	spellBookInsertGlyph4FE340        = int32(34)
	spellBookInsertConjurer4FE340     = uint8(2)

	spellBookInsertState4FE340 = int32(2)

	spellBookInsertCreatureControlFailed4FE340 = int32(4)
	spellBookInsertTooManyGlyphs4FE340         = int32(5)
	spellBookInsertNotEnoughManaGlyph4FE340    = int32(12)

	spellBookInsertFizzleSound4FE340    = int32(231)
	spellBookInsertManaEmptySound4FE340 = int32(232)

	spellBookInsertMaxBomberKey4FE340 = "MaxBomberCount"
	spellBookInsertMaxTrapKey4FE340   = "MaxTrapCount"
)

// spellBookInsertHooks4FE340 exposes every observable load, store, and
// callback in GAME.EXE 004FE340. Pointer-bearing values are generic tokens so
// the semantic core cannot inherit the original PE32 pointer width. Spell IDs,
// counters, frames, and record scalar fields retain their exact source widths.
type spellBookInsertHooks4FE340[
	Unit, Sequence, Update, Player, Trade, Entity, Definitions comparable,
	Allocator any,
] struct {
	loadUnitArg        func() Unit
	loadUnitFlags      func(Unit) uint32
	loadSequenceArg    func() Sequence
	loadSpell          func(Sequence, int32) int32
	loadUnitClassLow   func(Unit) uint8
	loadUpdate         func(Unit) Update
	loadTrade          func(Update) Trade
	loadPlayer         func(Update) Player
	loadKnownSpell     func(Player, int32) uint32
	loadSpellCastStart func(Update) uint32
	loadCountArg       func() int32

	manaPreflight   func(Unit, Sequence, int32) int32
	loadPlayerClass func(Player) uint8
	checkSummoned   func(Unit, int32) int32
	countSlaves     func(Unit, uint32, uint32) int32
	balanceFloat    func(string) float64
	loadCurTrapsLow func(Update) uint8
	spellPrecheck   func(Unit, int32) int32
	spellCastGate   func(Unit, int32, int32) int32
	loadPlayerIndex func(Player) uint8
	informResult    func(uint8, uint8, int32)
	audioEvent      func(int32, Unit, int32, uint32)

	setPlayerState      func(Unit, int32)
	loadFrame           func() uint32
	storeCasting        func(Update, uint8)
	storeSpellCastStart func(Update, uint32)
	loadAllocator       func() Allocator
	allocate            func(Allocator) Entity
	loadTargetModeArg   func() int32
	loadDelayArg        func() int32
	storeEntityUnit     func(Entity, Unit)
	storeTargetMode     func(Entity, uint32)
	storeField36        func(Entity, uint8)
	storeEntityFrame    func(Entity, uint32)
	storeDelay          func(Entity, uint32)
	loadDefinitions     func() Definitions
	storeDefinitions    func(Entity, Definitions)
	storeSpellIndex     func(Entity, uint8)
	storeGlyphMode      func(Entity, uint8)
	storeEntitySpell    func(Entity, int32, int32)
	loadHead            func() Entity
	storeNext           func(Entity, Entity)
	storePrev           func(Entity, Entity)
	storeHead           func(Entity)
}

// spellBookInsert4FE340 preserves GAME.EXE 004FE340's signed comparisons,
// five-slot validation, callback order, live argument reloads, and intrusive
// list insertion. It intentionally adds no nil, sequence-length, count-range,
// or allocator guards: malformed state reaches the same hook boundary that
// the PE32 executable would dereference or call.
func spellBookInsert4FE340[
	Unit, Sequence, Update, Player, Trade, Entity, Definitions comparable,
	Allocator any,
](h spellBookInsertHooks4FE340[
	Unit, Sequence, Update, Player, Trade, Entity, Definitions, Allocator,
]) int32 {
	unit := h.loadUnitArg()
	if h.loadUnitFlags(unit)&spellBookInsertBlockedFlags4FE340 != 0 {
		return 0
	}

	sequence := h.loadSequenceArg()
	for index := int32(0); index < 5; index++ {
		spellID := h.loadSpell(sequence, index)
		if spellID < 0 || spellID >= spellBookInsertSpellCount4FE340 {
			return 0
		}
	}
	if h.loadUnitClassLow(unit)&spellBookInsertPlayerClass4FE340 == 0 {
		return 0
	}

	update := h.loadUpdate(unit)
	trade := h.loadTrade(update)
	player := h.loadPlayer(update)
	var nilTrade Trade
	if trade != nilTrade {
		return 0
	}
	for index := int32(0); index < 5; index++ {
		spellID := h.loadSpell(sequence, index)
		if h.loadKnownSpell(player, spellID) == 0 && spellID != 0 {
			return 0
		}
	}
	if h.loadSpellCastStart(update) != 0 {
		return 0
	}

	count := h.loadCountArg()
	hasGlyph := int32(0)
	if count > 0 {
		for index, remaining := int32(0), count; remaining != 0; index, remaining = index+1, remaining-1 {
			if h.loadSpell(sequence, index) == spellBookInsertGlyph4FE340 {
				hasGlyph = 1
			}
		}
	}

	fail := func(result, sound int32) int32 {
		currentPlayer := h.loadPlayer(update)
		playerIndex := h.loadPlayerIndex(currentPlayer)
		h.informResult(playerIndex, 0, result)
		h.audioEvent(sound, unit, 0, 0)
		return 0
	}

	if hasGlyph != 0 {
		if h.manaPreflight(unit, sequence, count) == 0 {
			return fail(spellBookInsertNotEnoughManaGlyph4FE340, spellBookInsertManaEmptySound4FE340)
		}

		currentPlayer := h.loadPlayer(update)
		if h.loadPlayerClass(currentPlayer) == spellBookInsertConjurer4FE340 {
			if h.checkSummoned(unit, 5) == 0 {
				return fail(spellBookInsertCreatureControlFailed4FE340, spellBookInsertFizzleSound4FE340)
			}
			current := h.countSlaves(unit, 2, 0x2000)
			limit := x87TruncSignedQwordLow566DCC(h.balanceFloat(spellBookInsertMaxBomberKey4FE340))
			if current >= limit {
				return fail(spellBookInsertTooManyGlyphs4FE340, spellBookInsertFizzleSound4FE340)
			}
		} else {
			limit := x87TruncSignedQwordLow566DCC(h.balanceFloat(spellBookInsertMaxTrapKey4FE340))
			current := int32(h.loadCurTrapsLow(update))
			if current >= limit {
				return fail(spellBookInsertTooManyGlyphs4FE340, spellBookInsertFizzleSound4FE340)
			}
		}

		sequence = h.loadSequenceArg()
		for index := int32(0); ; index++ {
			spellID := h.loadSpell(sequence, index)
			if result := h.spellPrecheck(unit, spellID); result != 0 {
				return fail(result, spellBookInsertFizzleSound4FE340)
			}
			spellID = h.loadSpell(sequence, index)
			if result := h.spellCastGate(unit, spellID, hasGlyph); result != 0 {
				return fail(result, spellBookInsertFizzleSound4FE340)
			}
			if index+1 >= h.loadCountArg() {
				break
			}
		}
		sequence = h.loadSequenceArg()
	} else {
		spellID := h.loadSpell(sequence, 0)
		if result := h.spellPrecheck(unit, spellID); result != 0 {
			return fail(result, spellBookInsertFizzleSound4FE340)
		}
		spellID = h.loadSpell(sequence, 0)
		if result := h.spellCastGate(unit, spellID, 0); result != 0 {
			return fail(result, spellBookInsertFizzleSound4FE340)
		}
	}

	h.setPlayerState(unit, spellBookInsertState4FE340)
	frame := h.loadFrame()
	h.storeCasting(update, 1)
	h.storeSpellCastStart(update, frame)
	allocator := h.loadAllocator()
	entity := h.allocate(allocator)
	var nilEntity Entity
	if entity == nilEntity {
		return 0
	}

	targetMode := h.loadTargetModeArg()
	delay := h.loadDelayArg()
	h.storeEntityUnit(entity, unit)
	h.storeTargetMode(entity, uint32(targetMode))
	h.storeField36(entity, 0)
	frame = h.loadFrame()
	h.storeEntityFrame(entity, frame)
	h.storeDelay(entity, uint32(delay))
	definitions := h.loadDefinitions()
	h.storeDefinitions(entity, definitions)
	h.storeSpellIndex(entity, 0)
	h.storeGlyphMode(entity, 0)
	for index := int32(0); index < 5; index++ {
		if index >= h.loadCountArg() {
			h.storeEntitySpell(entity, index, 0)
			continue
		}
		spellID := h.loadSpell(sequence, index)
		h.storeEntitySpell(entity, index, spellID)
		if h.loadSpell(sequence, index) == spellBookInsertGlyph4FE340 {
			h.storeGlyphMode(entity, 1)
		}
	}

	h.storePrev(entity, nilEntity)
	firstHead := h.loadHead()
	h.storeNext(entity, firstHead)
	secondHead := h.loadHead()
	if secondHead != nilEntity {
		h.storePrev(secondHead, entity)
	}
	h.storeHead(entity)
	return 1
}
