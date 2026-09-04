package opennox

const (
	spellBookDeadMask4FCB80       = uint32(0x8020)
	spellBookPlayerClass4FCB80    = uint32(0x4)
	spellBookGlyph4FCB80          = int32(34)
	spellBookStartMessage4FCB80   = uint8(112)
	spellBookDuplicate4FCB80      = int32(6)
	spellBookNotEnoughMana4FCB80  = int32(11)
	spellBookManaEmptySound4FCB80 = int32(232)
)

// spellCastByBookHooks4FCB80 names every observable load, store, and callback
// boundary in GAME.EXE 004FCB80. Pointer-bearing values are generic native-
// width tokens; the PE32 spell IDs and counters retain their original widths.
type spellCastByBookHooks4FCB80[
	Node comparable,
	Object, Update, Player, Leaf, Settings, Phonemes, Allocator any,
	SpellArg comparable,
] struct {
	loadHead        func() Node
	loadObject      func(Node) Object
	loadObjectFlags func(Object) uint32
	loadNext        func(Node) Node
	loadPrev        func(Node) Node
	storePrev       func(Node, Node)
	storeHead       func(Node)
	storeNext       func(Node, Node)
	loadAllocator   func() Allocator
	freeFirst       func(Allocator, Node)

	loadFrame       func() uint32
	loadDeadline    func(Node) uint32
	loadObjectClass func(Object) uint32
	loadUpdate      func(Object) Update
	loadProgress    func(Node) uint8
	loadSpellIndex  func(Node) uint8
	loadSpellLow    func(Node, int) uint8
	loadSpell       func(Node, int) int32
	loadPlayer      func(Update) Player
	loadPlayerIndex func(Player) uint8
	reportStart     func(uint8, uint8, uint8)

	loadLeaf             func(Node) Leaf
	loadLeafSpell        func(Leaf) int32
	loadSettings         func() Settings
	loadPhonemeSequence  func(int32) Phonemes
	loadPhoneme          func(Phonemes, uint8) uint8
	loadGestureSuppress  func() uint32
	loadBroadcastGesture func(Settings) uint32
	broadcastPhoneme     func(Object, uint8)
	advanceLeaf          func(Leaf, uint8) Leaf
	storeLeaf            func(Node, Leaf)
	storePlayerLeaf      func(Update, Leaf)
	storeProgress        func(Node, uint8)

	loadGlyphMode   func(Node) uint8
	phonemeRoot     func() Leaf
	loadDelay       func(Node) uint32
	storeDeadline   func(Node, uint32)
	storeSpellIndex func(Node, uint8)

	loadTrapCount  func(Update) uint8
	loadTrapSpell  func(Update, int) int32
	informResult   func(uint8, uint8, int32)
	chargeMana     func(Object, int32, int32) int32
	audioEvent     func(int32, Object, int32, uint32)
	storeTrapSpell func(Update, int, int32)
	storeTrapCount func(Update, uint8)

	loadTargetMode    func(Node) uint32
	loadCursorX       func(Player) int32
	loadCursorY       func(Player) int32
	storeCastX        func(Update, int32)
	storeCastY        func(Update, int32)
	storePlayerTarget func(Player, Object)
	loadCursorObject  func(Update) Object
	playerSpell       func(Object)
	storeCastStart    func(Update, uint32)
	storeCasting      func(Update, uint8)
	castByUser        func(int32, Object, SpellArg)
}

// spellCastByBook4FCB80 preserves the original queue traversal's live reloads
// and low-byte stores. It intentionally adds no nil, bounds, or capacity
// guards: invalid PE32 state faults at the corresponding hook boundary and
// suppresses every later observation.
func spellCastByBook4FCB80[
	Node comparable,
	Object, Update, Player, Leaf, Settings, Phonemes, Allocator any,
	SpellArg comparable,
](h spellCastByBookHooks4FCB80[
	Node, Object, Update, Player, Leaf, Settings, Phonemes, Allocator, SpellArg,
]) {
	var nilNode Node

	unlink := func(node Node) Node {
		next := h.loadNext(node)
		if next != nilNode {
			prev := h.loadPrev(node)
			h.storePrev(next, prev)
		}
		prev := h.loadPrev(node)
		if prev == nilNode {
			next = h.loadNext(node)
			h.storeHead(next)
		} else {
			next = h.loadNext(node)
			h.storeNext(prev, next)
		}
		allocator := h.loadAllocator()
		next = h.loadNext(node)
		h.freeFirst(allocator, node)
		return next
	}

	advanceSpell := func(node Node) Node {
		h.storeProgress(node, 0)
		root := h.phonemeRoot()
		h.storeLeaf(node, root)
		delay := h.loadDelay(node)
		frame := h.loadFrame()
		h.storeDeadline(node, frame+delay)
		index := h.loadSpellIndex(node)
		h.storeSpellIndex(node, index+1)
		return h.loadNext(node)
	}

	for node := h.loadHead(); node != nilNode; {
		object := h.loadObject(node)
		if h.loadObjectFlags(object)&spellBookDeadMask4FCB80 != 0 {
			node = unlink(node)
			continue
		}

		frame := h.loadFrame()
		deadline := h.loadDeadline(node)
		if frame < deadline {
			node = h.loadNext(node)
			continue
		}

		var update Update
		if h.loadObjectClass(object)&spellBookPlayerClass4FCB80 != 0 {
			update = h.loadUpdate(object)
		}

		if h.loadProgress(node) == 0 {
			index := h.loadSpellIndex(node)
			spellLow := h.loadSpellLow(node, int(index))
			player := h.loadPlayer(update)
			playerIndex := h.loadPlayerIndex(player)
			h.reportStart(playerIndex, spellBookStartMessage4FCB80, spellLow)
		}

		leaf := h.loadLeaf(node)
		index := h.loadSpellIndex(node)
		leafSpell := h.loadLeafSpell(leaf)
		currentSpell := h.loadSpell(node, int(index))
		if leafSpell != currentSpell {
			settings := h.loadSettings()
			index = h.loadSpellIndex(node)
			currentSpell = h.loadSpell(node, int(index))
			phonemes := h.loadPhonemeSequence(currentSpell)
			progress := h.loadProgress(node)
			phoneme := h.loadPhoneme(phonemes, progress)
			suppress := h.loadGestureSuppress()
			if suppress == 0 || h.loadBroadcastGesture(settings) != 0 {
				object = h.loadObject(node)
				h.broadcastPhoneme(object, phoneme)
			}

			leaf = h.loadLeaf(node)
			leaf = h.advanceLeaf(leaf, phoneme)
			object = h.loadObject(node)
			h.storeLeaf(node, leaf)
			if h.loadObjectClass(object)&spellBookPlayerClass4FCB80 != 0 {
				leaf = h.loadLeaf(node)
				h.storePlayerLeaf(update, leaf)
			}
			progress = h.loadProgress(node)
			h.storeProgress(node, progress+1)
			delay := h.loadDelay(node)
			frame = h.loadFrame()
			h.storeDeadline(node, frame+delay)
			node = h.loadNext(node)
			continue
		}

		nextSpell := h.loadSpell(node, int(index)+1)
		glyphMode := h.loadGlyphMode(node)
		if glyphMode == 0 {
			if currentSpell != spellBookGlyph4FCB80 && nextSpell != 0 {
				node = advanceSpell(node)
				continue
			}
		} else if currentSpell != spellBookGlyph4FCB80 {
			object = h.loadObject(node)
			if h.loadObjectClass(object)&spellBookPlayerClass4FCB80 != 0 {
				duplicate := false
				count := h.loadTrapCount(update)
				if count != 0 {
					for i := 0; ; i++ {
						if h.loadTrapSpell(update, i) == currentSpell {
							player := h.loadPlayer(update)
							playerIndex := h.loadPlayerIndex(player)
							h.informResult(playerIndex, 0, spellBookDuplicate4FCB80)
							duplicate = true
						}
						count = h.loadTrapCount(update)
						if i+1 >= int(count) {
							break
						}
					}
				}
				if !duplicate {
					object = h.loadObject(node)
					if h.chargeMana(object, currentSpell, 2) < 0 {
						player := h.loadPlayer(update)
						playerIndex := h.loadPlayerIndex(player)
						h.informResult(playerIndex, 0, spellBookNotEnoughMana4FCB80)
						object = h.loadObject(node)
						h.audioEvent(spellBookManaEmptySound4FCB80, object, 0, 0)
					} else {
						count = h.loadTrapCount(update)
						h.storeTrapSpell(update, int(count), currentSpell)
						count = h.loadTrapCount(update)
						h.storeTrapCount(update, count+1)
					}
				}
			}
			if currentSpell != spellBookGlyph4FCB80 && nextSpell != 0 {
				node = advanceSpell(node)
				continue
			}
		}

		object = h.loadObject(node)
		if h.loadObjectClass(object)&spellBookPlayerClass4FCB80 != 0 {
			player := h.loadPlayer(update)
			x := h.loadCursorX(player)
			h.storeCastX(update, x)
			y := h.loadCursorY(player)
			h.storeCastY(update, y)
			if h.loadTargetMode(node) != 0 {
				object = h.loadObject(node)
				h.storePlayerTarget(player, object)
			} else {
				object = h.loadCursorObject(update)
				h.storePlayerTarget(player, object)
			}
			object = h.loadObject(node)
			h.playerSpell(object)
			h.storeCastStart(update, 0)
			h.storeCasting(update, 0)
			h.storeTrapCount(update, 0)
		} else {
			var nilArg SpellArg
			h.castByUser(currentSpell, object, nilArg)
		}
		node = unlink(node)
	}
}

func spellBookReact4FCB70(castByBook, castDurations func()) {
	castByBook()
	castDurations()
}
