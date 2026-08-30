package opennox

const (
	playerSpellQuestFlag4FB2A0   = uint32(0x1000)
	playerSpellOffensive4FB2A0   = uint32(0x20)
	playerSpellNoReport4FB2A0    = uint32(0x100000)
	playerSpellGlyph4FB2A0       = int32(34)
	playerSpellFizzleSound4FB2A0 = int32(231)
	playerSpellNoManaSound4FB2A0 = int32(232)
	playerSpellNoMana4FB2A0      = int32(11)
	playerSpellCastFailed4FB2A0  = int32(8)
	playerSpellReadyState4FB2A0  = uint8(2)
	playerSpellIdleState4FB2A0   = uint8(13)
)

// playerSpellArg4FB2A0 is the layout-independent form of the 12-byte PE32
// spell argument assembled on the stack by GAME.EXE. Native Object pointers
// remain full-width until the legacy spell implementation is called.
type playerSpellArg4FB2A0[O comparable] struct {
	target O
	posX   float32
	posY   float32
}

// playerSpellHooks4FB2A0 exposes the observable loads and calls in GAME.EXE
// 004FB2A0. The original caches PlayerUpdateData, but deliberately reloads its
// Player and SpellPhonemeLeaf pointers after callbacks. Keeping those loads in
// the model prevents stale PE32-era aliases from being reintroduced.
type playerSpellHooks4FB2A0[O, U, P, L, M comparable] struct {
	loadUpdateData func(O) U
	loadLeaf       func(U) L
	isRootLeaf     func(L) bool
	loadSpellID    func(L) int32
	hasGameFlag    func(uint32) bool
	loadCursorObj  func(U) O
	hasSpellFlags  func(int32, uint32) bool
	isEnemy        func(O, O) bool
	loadPlayer     func(U) P
	loadSpellLevel func(P, int32) uint32
	precheck       func(O, int32) int32
	checkCantCast  func(O, int32, int32) int32
	loadPlayerInd  func(P) uint8
	informResult   func(uint8, uint8, int32)
	informSpell    func(uint8, uint8, L)
	audioEvent     func(int32, O, int32, int32)
	chargeMana     func(O, int32, int32) int32
	loadCastTarget func(P) O
	loadCursorPos  func(P) (int32, int32)
	castSpell      func(int32, O, playerSpellArg4FB2A0[O]) bool
	refundMana     func(O, int32)
	loadState      func(U) uint8
	setState       func(O, uint8)
	unknownMessage func() M
	lineMessage    func(O, M)
	reportSpell    func(uint8, int32, uint8)
}

// playerSpell4FB2A0 preserves GAME.EXE 004FB2A0's branch behavior and reload
// boundaries. A friendly offensive target outside Quest returns immediately;
// every other path performs the state transition before reporting the result.
func playerSpell4FB2A0[O, U, P, L, M comparable](
	unit O,
	h playerSpellHooks4FB2A0[O, U, P, L, M],
) {
	update := h.loadUpdateData(unit)
	result := int32(1)
	unknown := true

	leaf := h.loadLeaf(update)
	if h.isRootLeaf(leaf) {
		unknown = false
	} else {
		leaf = h.loadLeaf(update)
		var nilLeaf L
		if leaf != nilLeaf && h.loadSpellID(leaf) != 0 {
			if !h.hasGameFlag(playerSpellQuestFlag4FB2A0) {
				leaf = h.loadLeaf(update)
				target := h.loadCursorObj(update)
				spellID := h.loadSpellID(leaf)
				if h.hasSpellFlags(spellID, playerSpellOffensive4FB2A0) {
					var nilObject O
					if target != nilObject && !h.isEnemy(unit, target) {
						return
					}
				}
			}

			leaf = h.loadLeaf(update)
			player := h.loadPlayer(update)
			spellID := h.loadSpellID(leaf)
			if h.loadSpellLevel(player, spellID) != 0 || spellID == playerSpellGlyph4FB2A0 {
				unknown = false
				result = h.precheck(unit, spellID)
				if result == 0 {
					leaf = h.loadLeaf(update)
					result = h.checkCantCast(unit, h.loadSpellID(leaf), 0)
				}

				if result != 0 {
					player = h.loadPlayer(update)
					h.informResult(h.loadPlayerInd(player), 0, result)
					h.audioEvent(playerSpellFizzleSound4FB2A0, unit, 0, 0)
				} else {
					leaf = h.loadLeaf(update)
					mana := h.chargeMana(unit, h.loadSpellID(leaf), 1)
					if mana < 0 {
						result = playerSpellNoMana4FB2A0
						player = h.loadPlayer(update)
						h.informResult(h.loadPlayerInd(player), 0, result)
						h.audioEvent(playerSpellNoManaSound4FB2A0, unit, 0, 0)
					} else {
						player = h.loadPlayer(update)
						arg := playerSpellArg4FB2A0[O]{target: h.loadCastTarget(player)}
						if h.hasGameFlag(playerSpellQuestFlag4FB2A0) {
							leaf = h.loadLeaf(update)
							spellID = h.loadSpellID(leaf)
							if h.hasSpellFlags(spellID, playerSpellOffensive4FB2A0) {
								player = h.loadPlayer(update)
								target := h.loadCastTarget(player)
								var nilObject O
								if target != nilObject && !h.isEnemy(unit, target) {
									arg.target = nilObject
								}
							}
						}

						player = h.loadPlayer(update)
						x, y := h.loadCursorPos(player)
						arg.posX = float32(x)
						arg.posY = float32(y)
						leaf = h.loadLeaf(update)
						if h.castSpell(h.loadSpellID(leaf), unit, arg) {
							player = h.loadPlayer(update)
							leaf = h.loadLeaf(update)
							h.informSpell(h.loadPlayerInd(player), 1, leaf)
						} else {
							h.refundMana(unit, mana)
							result = playerSpellCastFailed4FB2A0
						}
					}
				}
			}
		}
	}

	if h.loadState(update) == playerSpellReadyState4FB2A0 {
		h.setState(unit, playerSpellIdleState4FB2A0)
	}
	if unknown {
		h.lineMessage(unit, h.unknownMessage())
		return
	}
	if result != 0 {
		leaf = h.loadLeaf(update)
		player := h.loadPlayer(update)
		spellID := h.loadSpellID(leaf)
		playerInd := h.loadPlayerInd(player)
		h.reportSpell(playerInd, spellID, 0)
		return
	}

	leaf = h.loadLeaf(update)
	if h.hasSpellFlags(h.loadSpellID(leaf), playerSpellNoReport4FB2A0) {
		return
	}
	leaf = h.loadLeaf(update)
	player := h.loadPlayer(update)
	spellID := h.loadSpellID(leaf)
	playerInd := h.loadPlayerInd(player)
	h.reportSpell(playerInd, spellID, 15)
}
