package opennox

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellBookNodeA4FCB80     = uint64(0x100000101)
	spellBookNodeB4FCB80     = uint64(0x200000202)
	spellBookNodeC4FCB80     = uint64(0x300000303)
	spellBookObjectA4FCB80   = uint64(0x400000404)
	spellBookObjectB4FCB80   = uint64(0x500000505)
	spellBookObjectC4FCB80   = uint64(0x600000606)
	spellBookUpdateA4FCB80   = uint64(0x700000707)
	spellBookPlayerA4FCB80   = uint64(0x800000808)
	spellBookLeafA4FCB80     = uint64(0x900000909)
	spellBookLeafB4FCB80     = uint64(0xa00000a0a)
	spellBookRoot4FCB80      = uint64(0xb00000b0b)
	spellBookSettings4FCB80  = uint64(0xc00000c0c)
	spellBookPhonemes4FCB80  = uint64(0xd00000d0d)
	spellBookAllocator4FCB80 = uint64(0xe00000e0e)
	spellBookCursorObj4FCB80 = uint64(0xf00000f0f)
)

var spellBookFaultSentinel4FCB80 = &struct{}{}

type spellBookNodeState4FCB80 struct {
	object                  uint64
	next, prev              uint64
	spells                  map[int]int32
	index, mode, progress   uint8
	leaf                    uint64
	deadline, delay, target uint32
}

type spellBookObjectState4FCB80 struct {
	flags, class uint32
	update       uint64
}

type spellBookUpdateState4FCB80 struct {
	player       uint64
	leaf         uint64
	traps        map[int]int32
	trapCount    uint8
	castStart    uint32
	casting      uint8
	cursorObject uint64
	castX, castY int32
}

type spellBookPlayerState4FCB80 struct {
	index      uint8
	cursorX    int32
	cursorY    int32
	castTarget uint64
}

type spellBookLeafState4FCB80 struct {
	spell int32
	next  map[uint8]uint64
}

type spellBookReport4FCB80 struct {
	index, message, spell uint8
}

type spellBookInform4FCB80 struct {
	index, code uint8
	result      int32
}

type spellBookHarness4FCB80 struct {
	events []string
	failAt int

	head       uint64
	frame      uint32
	allocator  uint64
	root       uint64
	settings   uint64
	broadcast  uint32
	suppress   uint32
	charge     int32
	nodes      map[uint64]*spellBookNodeState4FCB80
	objects    map[uint64]*spellBookObjectState4FCB80
	updates    map[uint64]*spellBookUpdateState4FCB80
	players    map[uint64]*spellBookPlayerState4FCB80
	leaves     map[uint64]*spellBookLeafState4FCB80
	phonemeSeq map[int32]uint64
	phonemes   map[uint64][]uint8

	reports       []spellBookReport4FCB80
	broadcastObjs []uint64
	broadcastPh   []uint8
	informs       []spellBookInform4FCB80
	chargedObjs   []uint64
	audioObjs     []uint64
	playerCasts   []uint64
	userCasts     []struct {
		spell  int32
		object uint64
		arg    uint64
	}
	freed []struct {
		allocator uint64
		node      uint64
	}

	onBroadcast   func(*spellBookHarness4FCB80)
	onAdvance     func(*spellBookHarness4FCB80)
	onRoot        func(*spellBookHarness4FCB80)
	onInform      func(*spellBookHarness4FCB80)
	onCharge      func(*spellBookHarness4FCB80)
	onStoreTrap   func(*spellBookHarness4FCB80, int, int32)
	onStoreTarget func(*spellBookHarness4FCB80)
	onPlayerSpell func(*spellBookHarness4FCB80)
	onUserCast    func(*spellBookHarness4FCB80)
	onFree        func(*spellBookHarness4FCB80)
}

func newSpellBookHarness4FCB80() *spellBookHarness4FCB80 {
	return &spellBookHarness4FCB80{
		failAt:     -1,
		frame:      100,
		allocator:  spellBookAllocator4FCB80,
		root:       spellBookRoot4FCB80,
		settings:   spellBookSettings4FCB80,
		charge:     1,
		nodes:      make(map[uint64]*spellBookNodeState4FCB80),
		objects:    make(map[uint64]*spellBookObjectState4FCB80),
		updates:    make(map[uint64]*spellBookUpdateState4FCB80),
		players:    make(map[uint64]*spellBookPlayerState4FCB80),
		leaves:     make(map[uint64]*spellBookLeafState4FCB80),
		phonemeSeq: make(map[int32]uint64),
		phonemes:   make(map[uint64][]uint8),
	}
}

func (s *spellBookHarness4FCB80) observe(event string) {
	if s.failAt == len(s.events) {
		panic(spellBookFaultSentinel4FCB80)
	}
	s.events = append(s.events, event)
}

func (s *spellBookHarness4FCB80) hooks() spellCastByBookHooks4FCB80[
	uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
] {
	return spellCastByBookHooks4FCB80[
		uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
	]{
		loadHead: func() uint64 {
			s.observe("load-head")
			return s.head
		},
		loadObject: func(node uint64) uint64 {
			s.observe("load-object")
			return s.nodes[node].object
		},
		loadObjectFlags: func(object uint64) uint32 {
			s.observe("load-flags")
			return s.objects[object].flags
		},
		loadNext: func(node uint64) uint64 {
			s.observe("load-next")
			return s.nodes[node].next
		},
		loadPrev: func(node uint64) uint64 {
			s.observe("load-prev")
			return s.nodes[node].prev
		},
		storePrev: func(node, prev uint64) {
			s.observe("store-prev")
			s.nodes[node].prev = prev
		},
		storeHead: func(node uint64) {
			s.observe("store-head")
			s.head = node
		},
		storeNext: func(node, next uint64) {
			s.observe("store-next")
			s.nodes[node].next = next
		},
		loadAllocator: func() uint64 {
			s.observe("load-allocator")
			return s.allocator
		},
		freeFirst: func(allocator, node uint64) {
			s.observe("free-first")
			s.freed = append(s.freed, struct {
				allocator uint64
				node      uint64
			}{allocator, node})
			if s.onFree != nil {
				s.onFree(s)
			}
		},
		loadFrame: func() uint32 {
			s.observe("load-frame")
			return s.frame
		},
		loadDeadline: func(node uint64) uint32 {
			s.observe("load-deadline")
			return s.nodes[node].deadline
		},
		loadObjectClass: func(object uint64) uint32 {
			s.observe("load-class")
			return s.objects[object].class
		},
		loadUpdate: func(object uint64) uint64 {
			s.observe("load-update")
			return s.objects[object].update
		},
		loadProgress: func(node uint64) uint8 {
			s.observe("load-progress")
			return s.nodes[node].progress
		},
		loadSpellIndex: func(node uint64) uint8 {
			s.observe("load-index")
			return s.nodes[node].index
		},
		loadSpellLow: func(node uint64, index int) uint8 {
			s.observe(fmt.Sprintf("load-spell-low-%d", index))
			return uint8(s.nodes[node].spells[index])
		},
		loadSpell: func(node uint64, index int) int32 {
			s.observe(fmt.Sprintf("load-spell-%d", index))
			return s.nodes[node].spells[index]
		},
		loadPlayer: func(update uint64) uint64 {
			s.observe("load-player")
			return s.updates[update].player
		},
		loadPlayerIndex: func(player uint64) uint8 {
			s.observe("load-player-index")
			return s.players[player].index
		},
		reportStart: func(index, message, spell uint8) {
			s.observe("report-start")
			s.reports = append(s.reports, spellBookReport4FCB80{index, message, spell})
		},
		loadLeaf: func(node uint64) uint64 {
			s.observe("load-leaf")
			return s.nodes[node].leaf
		},
		loadLeafSpell: func(leaf uint64) int32 {
			s.observe("load-leaf-spell")
			return s.leaves[leaf].spell
		},
		loadSettings: func() uint64 {
			s.observe("load-settings")
			return s.settings
		},
		loadPhonemeSequence: func(spell int32) uint64 {
			s.observe("load-phoneme-sequence")
			return s.phonemeSeq[spell]
		},
		loadPhoneme: func(sequence uint64, progress uint8) uint8 {
			s.observe("load-phoneme")
			return s.phonemes[sequence][progress]
		},
		loadGestureSuppress: func() uint32 {
			s.observe("load-suppress")
			return s.suppress
		},
		loadBroadcastGesture: func(settings uint64) uint32 {
			s.observe("load-broadcast-setting")
			if settings != s.settings {
				panic("wrong settings token")
			}
			return s.broadcast
		},
		broadcastPhoneme: func(object uint64, phoneme uint8) {
			s.observe("broadcast")
			s.broadcastObjs = append(s.broadcastObjs, object)
			s.broadcastPh = append(s.broadcastPh, phoneme)
			if s.onBroadcast != nil {
				s.onBroadcast(s)
			}
		},
		advanceLeaf: func(leaf uint64, phoneme uint8) uint64 {
			s.observe("advance-leaf")
			result := s.leaves[leaf].next[phoneme]
			if s.onAdvance != nil {
				s.onAdvance(s)
			}
			return result
		},
		storeLeaf: func(node, leaf uint64) {
			s.observe("store-leaf")
			s.nodes[node].leaf = leaf
		},
		storePlayerLeaf: func(update, leaf uint64) {
			s.observe("store-player-leaf")
			s.updates[update].leaf = leaf
		},
		storeProgress: func(node uint64, progress uint8) {
			s.observe("store-progress")
			s.nodes[node].progress = progress
		},
		loadGlyphMode: func(node uint64) uint8 {
			s.observe("load-glyph-mode")
			return s.nodes[node].mode
		},
		phonemeRoot: func() uint64 {
			s.observe("phoneme-root")
			if s.onRoot != nil {
				s.onRoot(s)
			}
			return s.root
		},
		loadDelay: func(node uint64) uint32 {
			s.observe("load-delay")
			return s.nodes[node].delay
		},
		storeDeadline: func(node uint64, deadline uint32) {
			s.observe("store-deadline")
			s.nodes[node].deadline = deadline
		},
		storeSpellIndex: func(node uint64, index uint8) {
			s.observe("store-index")
			s.nodes[node].index = index
		},
		loadTrapCount: func(update uint64) uint8 {
			s.observe("load-trap-count")
			return s.updates[update].trapCount
		},
		loadTrapSpell: func(update uint64, index int) int32 {
			s.observe(fmt.Sprintf("load-trap-%d", index))
			return s.updates[update].traps[index]
		},
		informResult: func(index, code uint8, result int32) {
			s.observe("inform-result")
			s.informs = append(s.informs, spellBookInform4FCB80{index, code, result})
			if s.onInform != nil {
				s.onInform(s)
			}
		},
		chargeMana: func(object uint64, spell, mode int32) int32 {
			s.observe("charge-mana")
			if mode != 2 {
				panic("wrong mana mode")
			}
			s.chargedObjs = append(s.chargedObjs, object)
			if s.onCharge != nil {
				s.onCharge(s)
			}
			return s.charge
		},
		audioEvent: func(sound int32, object uint64, kind int32, code uint32) {
			s.observe("audio-event")
			if sound != spellBookManaEmptySound4FCB80 || kind != 0 || code != 0 {
				panic("wrong audio arguments")
			}
			s.audioObjs = append(s.audioObjs, object)
		},
		storeTrapSpell: func(update uint64, index int, spell int32) {
			s.observe(fmt.Sprintf("store-trap-%d", index))
			s.updates[update].traps[index] = spell
			if index == 5 {
				s.updates[update].trapCount = uint8(spell)
			}
			if s.onStoreTrap != nil {
				s.onStoreTrap(s, index, spell)
			}
		},
		storeTrapCount: func(update uint64, count uint8) {
			s.observe("store-trap-count")
			s.updates[update].trapCount = count
		},
		loadTargetMode: func(node uint64) uint32 {
			s.observe("load-target-mode")
			return s.nodes[node].target
		},
		loadCursorX: func(player uint64) int32 {
			s.observe("load-cursor-x")
			return s.players[player].cursorX
		},
		loadCursorY: func(player uint64) int32 {
			s.observe("load-cursor-y")
			return s.players[player].cursorY
		},
		storeCastX: func(update uint64, x int32) {
			s.observe("store-cast-x")
			s.updates[update].castX = x
		},
		storeCastY: func(update uint64, y int32) {
			s.observe("store-cast-y")
			s.updates[update].castY = y
		},
		storePlayerTarget: func(player, object uint64) {
			s.observe("store-player-target")
			s.players[player].castTarget = object
			if s.onStoreTarget != nil {
				s.onStoreTarget(s)
			}
		},
		loadCursorObject: func(update uint64) uint64 {
			s.observe("load-cursor-object")
			return s.updates[update].cursorObject
		},
		playerSpell: func(object uint64) {
			s.observe("player-spell")
			s.playerCasts = append(s.playerCasts, object)
			if s.onPlayerSpell != nil {
				s.onPlayerSpell(s)
			}
		},
		storeCastStart: func(update uint64, value uint32) {
			s.observe("store-cast-start")
			s.updates[update].castStart = value
		},
		storeCasting: func(update uint64, value uint8) {
			s.observe("store-casting")
			s.updates[update].casting = value
		},
		castByUser: func(spell int32, object, arg uint64) {
			s.observe("cast-by-user")
			s.userCasts = append(s.userCasts, struct {
				spell  int32
				object uint64
				arg    uint64
			}{spell, object, arg})
			if s.onUserCast != nil {
				s.onUserCast(s)
			}
		},
	}
}

func (s *spellBookHarness4FCB80) run() {
	spellCastByBook4FCB80(s.hooks())
}

func setupSpellBookMismatch4FCB80() *spellBookHarness4FCB80 {
	s := newSpellBookHarness4FCB80()
	s.head = spellBookNodeA4FCB80
	s.broadcast = 1
	s.suppress = 1
	s.nodes[spellBookNodeA4FCB80] = &spellBookNodeState4FCB80{
		object: spellBookObjectA4FCB80,
		spells: map[int]int32{0: 0x109},
		leaf:   spellBookLeafA4FCB80,
		delay:  7,
	}
	s.objects[spellBookObjectA4FCB80] = &spellBookObjectState4FCB80{class: 4, update: spellBookUpdateA4FCB80}
	s.objects[spellBookObjectB4FCB80] = &spellBookObjectState4FCB80{class: 0}
	s.objects[spellBookObjectC4FCB80] = &spellBookObjectState4FCB80{class: 4, update: 0xdeadbeef}
	s.updates[spellBookUpdateA4FCB80] = &spellBookUpdateState4FCB80{player: spellBookPlayerA4FCB80, traps: make(map[int]int32)}
	s.players[spellBookPlayerA4FCB80] = &spellBookPlayerState4FCB80{index: 17}
	s.leaves[spellBookLeafA4FCB80] = &spellBookLeafState4FCB80{spell: 3, next: map[uint8]uint64{5: spellBookLeafB4FCB80}}
	s.leaves[spellBookLeafB4FCB80] = &spellBookLeafState4FCB80{spell: 0, next: make(map[uint8]uint64)}
	s.phonemeSeq[0x109] = spellBookPhonemes4FCB80
	s.phonemes[spellBookPhonemes4FCB80] = []uint8{5}
	s.onBroadcast = func(s *spellBookHarness4FCB80) {
		s.nodes[spellBookNodeA4FCB80].object = spellBookObjectB4FCB80
	}
	s.onAdvance = func(s *spellBookHarness4FCB80) {
		s.nodes[spellBookNodeA4FCB80].object = spellBookObjectC4FCB80
	}
	return s
}

func TestSpellCastByBook4FCB80EmptyHead(t *testing.T) {
	s := newSpellBookHarness4FCB80()
	s.run()
	if want := []string{"load-head"}; !reflect.DeepEqual(s.events, want) {
		t.Fatalf("events = %#v, want %#v", s.events, want)
	}
}

func TestSpellCastByBook4FCB80DeadUnlinkCachesNextBeforeFree(t *testing.T) {
	s := newSpellBookHarness4FCB80()
	s.head = spellBookNodeA4FCB80
	s.nodes[spellBookNodeA4FCB80] = &spellBookNodeState4FCB80{object: spellBookObjectA4FCB80, next: spellBookNodeB4FCB80}
	s.nodes[spellBookNodeB4FCB80] = &spellBookNodeState4FCB80{object: spellBookObjectB4FCB80, deadline: 101}
	s.nodes[spellBookNodeC4FCB80] = &spellBookNodeState4FCB80{object: spellBookObjectC4FCB80, deadline: 101}
	s.objects[spellBookObjectA4FCB80] = &spellBookObjectState4FCB80{flags: spellBookDeadMask4FCB80}
	s.objects[spellBookObjectB4FCB80] = &spellBookObjectState4FCB80{}
	s.objects[spellBookObjectC4FCB80] = &spellBookObjectState4FCB80{}
	s.onFree = func(s *spellBookHarness4FCB80) {
		s.nodes[spellBookNodeA4FCB80].next = spellBookNodeC4FCB80
	}
	s.run()

	want := []string{
		"load-head", "load-object", "load-flags",
		"load-next", "load-prev", "store-prev", "load-prev", "load-next", "store-head",
		"load-allocator", "load-next", "free-first",
		"load-object", "load-flags", "load-frame", "load-deadline", "load-next",
	}
	if !reflect.DeepEqual(s.events, want) {
		t.Fatalf("events = %#v, want %#v", s.events, want)
	}
	if s.head != spellBookNodeB4FCB80 || s.nodes[spellBookNodeB4FCB80].prev != 0 {
		t.Fatalf("head/next.prev = (%#x, %#x)", s.head, s.nodes[spellBookNodeB4FCB80].prev)
	}
	if got := s.freed; len(got) != 1 || got[0].allocator != spellBookAllocator4FCB80 || got[0].node != spellBookNodeA4FCB80 {
		t.Fatalf("freed = %#v", got)
	}
}

func TestSpellCastByBook4FCB80DeadUnlinkStoresPreviousNext(t *testing.T) {
	s := newSpellBookHarness4FCB80()
	s.head = spellBookNodeA4FCB80
	s.nodes[spellBookNodeA4FCB80] = &spellBookNodeState4FCB80{object: spellBookObjectA4FCB80, prev: spellBookNodeB4FCB80}
	s.nodes[spellBookNodeB4FCB80] = &spellBookNodeState4FCB80{next: spellBookNodeA4FCB80}
	s.objects[spellBookObjectA4FCB80] = &spellBookObjectState4FCB80{flags: spellBookDeadMask4FCB80}
	s.run()

	want := []string{
		"load-head", "load-object", "load-flags", "load-next", "load-prev", "load-next",
		"store-next", "load-allocator", "load-next", "free-first",
	}
	if !reflect.DeepEqual(s.events, want) || s.nodes[spellBookNodeB4FCB80].next != 0 {
		t.Fatalf("events/previous.next = (%#v, %#x), want (%#v, 0)", s.events, s.nodes[spellBookNodeB4FCB80].next, want)
	}
}

func TestSpellCastByBook4FCB80MismatchOrderAndLiveReloads(t *testing.T) {
	s := setupSpellBookMismatch4FCB80()
	s.run()

	want := []string{
		"load-head", "load-object", "load-flags", "load-frame", "load-deadline", "load-class", "load-update",
		"load-progress", "load-index", "load-spell-low-0", "load-player", "load-player-index", "report-start",
		"load-leaf", "load-index", "load-leaf-spell", "load-spell-0",
		"load-settings", "load-index", "load-spell-0", "load-phoneme-sequence", "load-progress", "load-phoneme",
		"load-suppress", "load-broadcast-setting", "load-object", "broadcast",
		"load-leaf", "advance-leaf", "load-object", "store-leaf", "load-class", "load-leaf", "store-player-leaf",
		"load-progress", "store-progress", "load-delay", "load-frame", "store-deadline", "load-next",
	}
	if !reflect.DeepEqual(s.events, want) {
		t.Fatalf("events = %#v, want %#v", s.events, want)
	}
	if got, wantReport := s.reports, []spellBookReport4FCB80{{17, 112, 9}}; !reflect.DeepEqual(got, wantReport) {
		t.Fatalf("reports = %#v, want %#v", got, wantReport)
	}
	if !reflect.DeepEqual(s.broadcastObjs, []uint64{spellBookObjectA4FCB80}) || !reflect.DeepEqual(s.broadcastPh, []uint8{5}) {
		t.Fatalf("broadcasts = %#v/%#v", s.broadcastObjs, s.broadcastPh)
	}
	if s.nodes[spellBookNodeA4FCB80].leaf != spellBookLeafB4FCB80 || s.updates[spellBookUpdateA4FCB80].leaf != spellBookLeafB4FCB80 {
		t.Fatalf("node/update leaves = (%#x, %#x)", s.nodes[spellBookNodeA4FCB80].leaf, s.updates[spellBookUpdateA4FCB80].leaf)
	}
	if s.nodes[spellBookNodeA4FCB80].progress != 1 || s.nodes[spellBookNodeA4FCB80].deadline != 107 {
		t.Fatalf("progress/deadline = (%d, %d)", s.nodes[spellBookNodeA4FCB80].progress, s.nodes[spellBookNodeA4FCB80].deadline)
	}
}

func TestSpellCastByBook4FCB80SuppressionZeroSkipsSettingsField(t *testing.T) {
	s := setupSpellBookMismatch4FCB80()
	s.suppress = 0
	s.broadcast = 0
	s.run()
	for _, event := range s.events {
		if event == "load-broadcast-setting" {
			t.Fatal("broadcast setting was observed after a zero suppression global")
		}
	}
	if len(s.broadcastObjs) != 1 {
		t.Fatalf("broadcast count = %d, want 1", len(s.broadcastObjs))
	}
}

func TestSpellCastByBook4FCB80EqualTransitionUsesUnguardedNextSlot(t *testing.T) {
	s := newSpellBookHarness4FCB80()
	s.head = spellBookNodeA4FCB80
	s.nodes[spellBookNodeA4FCB80] = &spellBookNodeState4FCB80{
		object:   spellBookObjectA4FCB80,
		spells:   map[int]int32{4: 9, 5: 0x00010004},
		index:    4,
		leaf:     spellBookLeafA4FCB80,
		progress: 3,
		delay:    0xffffffff,
	}
	s.objects[spellBookObjectA4FCB80] = &spellBookObjectState4FCB80{}
	s.leaves[spellBookLeafA4FCB80] = &spellBookLeafState4FCB80{spell: 9}
	s.run()

	want := []string{
		"load-head", "load-object", "load-flags", "load-frame", "load-deadline", "load-class", "load-progress",
		"load-leaf", "load-index", "load-leaf-spell", "load-spell-4", "load-spell-5", "load-glyph-mode",
		"store-progress", "phoneme-root", "store-leaf", "load-delay", "load-frame", "store-deadline",
		"load-index", "store-index", "load-next",
	}
	if !reflect.DeepEqual(s.events, want) {
		t.Fatalf("events = %#v, want %#v", s.events, want)
	}
	if n := s.nodes[spellBookNodeA4FCB80]; n.progress != 0 || n.leaf != spellBookRoot4FCB80 || n.index != 5 || n.deadline != 99 {
		t.Fatalf("transition state = %#v", n)
	}
}

func TestSpellCastByBook4FCB80TrapDuplicateReportsEveryMatch(t *testing.T) {
	s := setupSpellBookTrapFinal4FCB80()
	s.updates[spellBookUpdateA4FCB80].trapCount = 3
	s.updates[spellBookUpdateA4FCB80].traps[0] = 9
	s.updates[spellBookUpdateA4FCB80].traps[1] = 2
	s.updates[spellBookUpdateA4FCB80].traps[2] = 9
	s.run()

	wantInform := []spellBookInform4FCB80{{23, 0, 6}, {23, 0, 6}}
	if !reflect.DeepEqual(s.informs, wantInform) {
		t.Fatalf("informs = %#v, want %#v", s.informs, wantInform)
	}
	if len(s.chargedObjs) != 0 {
		t.Fatalf("charged duplicate spell through objects %#v", s.chargedObjs)
	}
	if !reflect.DeepEqual(s.playerCasts, []uint64{spellBookObjectB4FCB80}) {
		t.Fatalf("player casts = %#v", s.playerCasts)
	}
	if got := s.players[spellBookPlayerA4FCB80].castTarget; got != spellBookObjectA4FCB80 {
		t.Fatalf("cast target = %#x, want pre-store live object %#x", got, spellBookObjectA4FCB80)
	}
	update := s.updates[spellBookUpdateA4FCB80]
	if update.castX != -123 || update.castY != 456 || update.castStart != 0 || update.casting != 0 || update.trapCount != 0 {
		t.Fatalf("final update state = %#v", update)
	}
}

func TestSpellCastByBook4FCB80TrapSuccessReloadsAndOverwritesCountSlot(t *testing.T) {
	s := setupSpellBookTrapFinal4FCB80()
	node := s.nodes[spellBookNodeA4FCB80]
	node.spells[1] = 8
	node.delay = 4
	update := s.updates[spellBookUpdateA4FCB80]
	update.trapCount = 0
	s.onCharge = func(s *spellBookHarness4FCB80) {
		s.updates[spellBookUpdateA4FCB80].trapCount = 5
	}
	s.run()

	if !reflect.DeepEqual(s.chargedObjs, []uint64{spellBookObjectA4FCB80}) {
		t.Fatalf("charged objects = %#v", s.chargedObjs)
	}
	if update.traps[5] != 9 || update.trapCount != 10 {
		t.Fatalf("overflow trap/count = (%d, %d), want (9, 10)", update.traps[5], update.trapCount)
	}
	if node.index != 1 || node.progress != 0 || node.leaf != spellBookRoot4FCB80 || node.deadline != 104 {
		t.Fatalf("post-trap transition = %#v", node)
	}
	if len(s.playerCasts) != 0 {
		t.Fatalf("unexpected final cast %#v", s.playerCasts)
	}
}

func TestSpellCastByBook4FCB80ManaFailureReloadsObjectForAudio(t *testing.T) {
	s := setupSpellBookTrapFinal4FCB80()
	s.charge = -1
	s.onInform = func(s *spellBookHarness4FCB80) {
		s.nodes[spellBookNodeA4FCB80].object = spellBookObjectB4FCB80
	}
	s.run()

	if want := []spellBookInform4FCB80{{23, 0, 11}}; !reflect.DeepEqual(s.informs, want) {
		t.Fatalf("informs = %#v, want %#v", s.informs, want)
	}
	if !reflect.DeepEqual(s.audioObjs, []uint64{spellBookObjectB4FCB80}) {
		t.Fatalf("audio objects = %#v", s.audioObjs)
	}
}

func setupSpellBookTrapFinal4FCB80() *spellBookHarness4FCB80 {
	s := newSpellBookHarness4FCB80()
	s.head = spellBookNodeA4FCB80
	s.nodes[spellBookNodeA4FCB80] = &spellBookNodeState4FCB80{
		object:   spellBookObjectA4FCB80,
		spells:   map[int]int32{0: 9, 1: 0},
		mode:     1,
		progress: 2,
		leaf:     spellBookLeafA4FCB80,
		target:   1,
	}
	s.objects[spellBookObjectA4FCB80] = &spellBookObjectState4FCB80{class: 4, update: spellBookUpdateA4FCB80}
	s.objects[spellBookObjectB4FCB80] = &spellBookObjectState4FCB80{class: 4, update: 0xdeadbeef}
	s.updates[spellBookUpdateA4FCB80] = &spellBookUpdateState4FCB80{
		player:       spellBookPlayerA4FCB80,
		traps:        make(map[int]int32),
		trapCount:    0,
		castStart:    0xaabbccdd,
		casting:      0xee,
		cursorObject: spellBookCursorObj4FCB80,
	}
	s.players[spellBookPlayerA4FCB80] = &spellBookPlayerState4FCB80{index: 23, cursorX: -123, cursorY: 456}
	s.leaves[spellBookLeafA4FCB80] = &spellBookLeafState4FCB80{spell: 9}
	s.onStoreTarget = func(s *spellBookHarness4FCB80) {
		s.nodes[spellBookNodeA4FCB80].object = spellBookObjectB4FCB80
	}
	s.onPlayerSpell = func(s *spellBookHarness4FCB80) {
		update := s.updates[spellBookUpdateA4FCB80]
		update.castStart = 0x11223344
		update.casting = 0x55
		update.trapCount = 0x66
	}
	return s
}

func TestSpellCastByBook4FCB80NonPlayerCastUsesNilArgument(t *testing.T) {
	s := newSpellBookHarness4FCB80()
	s.head = spellBookNodeA4FCB80
	s.nodes[spellBookNodeA4FCB80] = &spellBookNodeState4FCB80{
		object:   spellBookObjectA4FCB80,
		spells:   map[int]int32{0: 34, 1: 0},
		progress: 1,
		leaf:     spellBookLeafA4FCB80,
	}
	s.objects[spellBookObjectA4FCB80] = &spellBookObjectState4FCB80{}
	s.leaves[spellBookLeafA4FCB80] = &spellBookLeafState4FCB80{spell: 34}
	s.run()

	if len(s.userCasts) != 1 || s.userCasts[0].spell != 34 || s.userCasts[0].object != spellBookObjectA4FCB80 || s.userCasts[0].arg != 0 {
		t.Fatalf("user casts = %#v", s.userCasts)
	}
}

func TestSpellCastByBook4FCB80FaultPrefixes(t *testing.T) {
	setups := []struct {
		name  string
		setup func() *spellBookHarness4FCB80
	}{
		{"mismatch", setupSpellBookMismatch4FCB80},
		{"trap-final", setupSpellBookTrapFinal4FCB80},
	}
	for _, tc := range setups {
		t.Run(tc.name, func(t *testing.T) {
			baseline := tc.setup()
			baseline.run()
			for failAt := range baseline.events {
				t.Run(fmt.Sprintf("fault-%02d-%s", failAt, baseline.events[failAt]), func(t *testing.T) {
					s := tc.setup()
					s.failAt = failAt
					var recovered any
					func() {
						defer func() { recovered = recover() }()
						s.run()
					}()
					if recovered != spellBookFaultSentinel4FCB80 {
						t.Fatalf("recovered = %#v, want sentinel", recovered)
					}
					want := baseline.events[:failAt]
					if len(s.events) != len(want) {
						t.Fatalf("events = %#v, want prefix %#v", s.events, want)
					}
					for i := range want {
						if s.events[i] != want[i] {
							t.Fatalf("events = %#v, want prefix %#v", s.events, want)
						}
					}
				})
			}
		})
	}
}

func TestSpellBookReact4FCB70OrderAndFault(t *testing.T) {
	var events []string
	spellBookReact4FCB70(
		func() { events = append(events, "book") },
		func() { events = append(events, "durations") },
	)
	if want := []string{"book", "durations"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}

	events = nil
	stop := &struct{}{}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		spellBookReact4FCB70(
			func() { events = append(events, "book"); panic(stop) },
			func() { events = append(events, "durations") },
		)
	}()
	if recovered != stop || !reflect.DeepEqual(events, []string{"book"}) {
		t.Fatalf("recovered/events = (%#v, %#v)", recovered, events)
	}
}
