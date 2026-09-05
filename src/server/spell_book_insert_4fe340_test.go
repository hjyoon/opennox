package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellBookInsertTestUnit4FE340      = uint64(0x100000101)
	spellBookInsertTestSequence4FE340  = uint64(0x200000202)
	spellBookInsertTestUpdate4FE340    = uint64(0x300000303)
	spellBookInsertTestPlayer4FE340    = uint64(0x400000404)
	spellBookInsertTestPlayerB4FE340   = uint64(0x500000505)
	spellBookInsertTestAllocator4FE340 = uint64(0x600000606)
	spellBookInsertTestEntityID4FE340  = uint64(0x700000707)
	spellBookInsertTestHeadA4FE340     = uint64(0x800000808)
	spellBookInsertTestHeadB4FE340     = uint64(0x900000909)
	spellBookInsertTestDefs4FE340      = uint64(0xa00000a0a)
)

type spellBookInsertTestSpellKey4FE340 struct {
	sequence uint64
	index    int32
}

type spellBookInsertTestEntity4FE340 struct {
	unit        uint64
	spells      [5]int32
	spellIndex  uint8
	glyphMode   uint8
	definitions uint64
	field36     uint8
	frame       uint32
	delay       uint32
	targetMode  uint32
	next        uint64
	prev        uint64
}

type spellBookInsertTestInform4FE340 struct {
	index, code uint8
	result      int32
}

type spellBookInsertTestAudio4FE340 struct {
	id   int32
	unit uint64
	kind int32
	code uint32
}

type spellBookInsertTestWorld4FE340 struct {
	events  []string
	faultAt int
	after   map[string]func()

	unit, sequence, update, trade uint64
	player, allocator, allocated  uint64
	definitions, head             uint64
	flags                         uint32
	class                         uint8
	spells                        map[spellBookInsertTestSpellKey4FE340]int32
	spellLoads                    map[spellBookInsertTestSpellKey4FE340]int
	known                         map[uint64]map[int32]uint32
	castStart                     uint32
	count                         int32
	mana, summoned                int32
	playerClass                   map[uint64]uint8
	playerIndex                   map[uint64]uint8
	slaves                        int32
	balance                       map[string]float64
	curTraps                      uint8
	precheck                      map[int32]int32
	castGate                      map[int32]int32
	frames                        []uint32
	frameLoads                    int
	targetMode, delay             int32

	state    int32
	casting  uint8
	entities map[uint64]*spellBookInsertTestEntity4FE340
	informs  []spellBookInsertTestInform4FE340
	audio    []spellBookInsertTestAudio4FE340
}

func newSpellBookInsertTestWorld4FE340() *spellBookInsertTestWorld4FE340 {
	w := &spellBookInsertTestWorld4FE340{
		after:       make(map[string]func()),
		unit:        spellBookInsertTestUnit4FE340,
		sequence:    spellBookInsertTestSequence4FE340,
		update:      spellBookInsertTestUpdate4FE340,
		player:      spellBookInsertTestPlayer4FE340,
		allocator:   spellBookInsertTestAllocator4FE340,
		allocated:   spellBookInsertTestEntityID4FE340,
		definitions: spellBookInsertTestDefs4FE340,
		class:       spellBookInsertPlayerClass4FE340,
		spells:      make(map[spellBookInsertTestSpellKey4FE340]int32),
		spellLoads:  make(map[spellBookInsertTestSpellKey4FE340]int),
		known:       make(map[uint64]map[int32]uint32),
		count:       1,
		mana:        1,
		summoned:    1,
		playerClass: map[uint64]uint8{spellBookInsertTestPlayer4FE340: 1, spellBookInsertTestPlayerB4FE340: 1},
		playerIndex: map[uint64]uint8{spellBookInsertTestPlayer4FE340: 7, spellBookInsertTestPlayerB4FE340: 9},
		balance:     map[string]float64{spellBookInsertMaxBomberKey4FE340: 8, spellBookInsertMaxTrapKey4FE340: 8},
		precheck:    make(map[int32]int32),
		castGate:    make(map[int32]int32),
		frames:      []uint32{100, 101},
		targetMode:  -2,
		delay:       -3,
		entities:    make(map[uint64]*spellBookInsertTestEntity4FE340),
	}
	w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, 0}] = 10
	w.known[w.player] = map[int32]uint32{10: 1}
	w.known[spellBookInsertTestPlayerB4FE340] = map[int32]uint32{10: 1}
	w.entities[w.allocated] = &spellBookInsertTestEntity4FE340{}
	return w
}

func (w *spellBookInsertTestWorld4FE340) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellBookInsertTestWorld4FE340) entity(id uint64) *spellBookInsertTestEntity4FE340 {
	entity := w.entities[id]
	if entity == nil {
		entity = &spellBookInsertTestEntity4FE340{}
		w.entities[id] = entity
	}
	return entity
}

func (w *spellBookInsertTestWorld4FE340) hooks() spellBookInsertHooks4FE340[
	uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
] {
	return spellBookInsertHooks4FE340[
		uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64,
	]{
		loadUnitArg: func() uint64 {
			value := w.unit
			w.observe("unit")
			return value
		},
		loadUnitFlags: func(uint64) uint32 {
			value := w.flags
			w.observe("flags")
			return value
		},
		loadSequenceArg: func() uint64 {
			value := w.sequence
			w.observe("sequence")
			return value
		},
		loadSpell: func(sequence uint64, index int32) int32 {
			key := spellBookInsertTestSpellKey4FE340{sequence, index}
			value := w.spells[key]
			w.spellLoads[key]++
			w.observe(fmt.Sprintf("spell:%d", index))
			return value
		},
		loadUnitClassLow: func(uint64) uint8 {
			value := w.class
			w.observe("class")
			return value
		},
		loadUpdate: func(uint64) uint64 {
			value := w.update
			w.observe("update")
			return value
		},
		loadTrade: func(uint64) uint64 {
			value := w.trade
			w.observe("trade")
			return value
		},
		loadPlayer: func(uint64) uint64 {
			value := w.player
			w.observe("player")
			return value
		},
		loadKnownSpell: func(player uint64, spellID int32) uint32 {
			value := w.known[player][spellID]
			w.observe(fmt.Sprintf("known:%d", spellID))
			return value
		},
		loadSpellCastStart: func(uint64) uint32 {
			value := w.castStart
			w.observe("cast-start")
			return value
		},
		loadCountArg: func() int32 {
			value := w.count
			w.observe("count")
			return value
		},
		manaPreflight: func(uint64, uint64, int32) int32 {
			value := w.mana
			w.observe("mana")
			return value
		},
		loadPlayerClass: func(player uint64) uint8 {
			value := w.playerClass[player]
			w.observe("player-class")
			return value
		},
		checkSummoned: func(uint64, int32) int32 {
			value := w.summoned
			w.observe("summoned")
			return value
		},
		countSlaves: func(uint64, uint32, uint32) int32 {
			value := w.slaves
			w.observe("slaves")
			return value
		},
		balanceFloat: func(key string) float64 {
			value := w.balance[key]
			w.observe("balance:" + key)
			return value
		},
		loadCurTrapsLow: func(uint64) uint8 {
			value := w.curTraps
			w.observe("traps")
			return value
		},
		spellPrecheck: func(_ uint64, spellID int32) int32 {
			value := w.precheck[spellID]
			w.observe(fmt.Sprintf("precheck:%d", spellID))
			return value
		},
		spellCastGate: func(_ uint64, spellID, mode int32) int32 {
			value := w.castGate[spellID]
			w.observe(fmt.Sprintf("cast:%d:%d", spellID, mode))
			return value
		},
		loadPlayerIndex: func(player uint64) uint8 {
			value := w.playerIndex[player]
			w.observe("player-index")
			return value
		},
		informResult: func(index, code uint8, result int32) {
			w.informs = append(w.informs, spellBookInsertTestInform4FE340{index: index, code: code, result: result})
			w.observe(fmt.Sprintf("inform:%d", result))
		},
		audioEvent: func(id int32, unit uint64, kind int32, code uint32) {
			w.audio = append(w.audio, spellBookInsertTestAudio4FE340{id: id, unit: unit, kind: kind, code: code})
			w.observe(fmt.Sprintf("audio:%d", id))
		},
		setPlayerState: func(_ uint64, state int32) {
			w.state = state
			w.observe(fmt.Sprintf("state:%d", state))
		},
		loadFrame: func() uint32 {
			index := w.frameLoads
			if index >= len(w.frames) {
				index = len(w.frames) - 1
			}
			value := w.frames[index]
			w.frameLoads++
			w.observe("frame")
			return value
		},
		storeCasting: func(_ uint64, value uint8) {
			w.casting = value
			w.observe(fmt.Sprintf("store-casting:%d", value))
		},
		storeSpellCastStart: func(_ uint64, value uint32) {
			w.castStart = value
			w.observe(fmt.Sprintf("store-start:%d", value))
		},
		loadAllocator: func() uint64 {
			value := w.allocator
			w.observe("allocator")
			return value
		},
		allocate: func(uint64) uint64 {
			value := w.allocated
			w.observe("allocate")
			return value
		},
		loadTargetModeArg: func() int32 {
			value := w.targetMode
			w.observe("target")
			return value
		},
		loadDelayArg: func() int32 {
			value := w.delay
			w.observe("delay")
			return value
		},
		storeEntityUnit: func(entity, unit uint64) {
			w.entity(entity).unit = unit
			w.observe("store-unit")
		},
		storeTargetMode: func(entity uint64, value uint32) {
			w.entity(entity).targetMode = value
			w.observe("store-target")
		},
		storeField36: func(entity uint64, value uint8) {
			w.entity(entity).field36 = value
			w.observe("store-field36")
		},
		storeEntityFrame: func(entity uint64, value uint32) {
			w.entity(entity).frame = value
			w.observe("store-frame")
		},
		storeDelay: func(entity uint64, value uint32) {
			w.entity(entity).delay = value
			w.observe("store-delay")
		},
		loadDefinitions: func() uint64 {
			value := w.definitions
			w.observe("definitions")
			return value
		},
		storeDefinitions: func(entity, definitions uint64) {
			w.entity(entity).definitions = definitions
			w.observe("store-definitions")
		},
		storeSpellIndex: func(entity uint64, value uint8) {
			w.entity(entity).spellIndex = value
			w.observe(fmt.Sprintf("store-index:%d", value))
		},
		storeGlyphMode: func(entity uint64, value uint8) {
			w.entity(entity).glyphMode = value
			w.observe(fmt.Sprintf("store-glyph:%d", value))
		},
		storeEntitySpell: func(entity uint64, index, value int32) {
			w.entity(entity).spells[index] = value
			w.observe(fmt.Sprintf("store-spell:%d", index))
		},
		loadHead: func() uint64 {
			value := w.head
			w.observe("head")
			return value
		},
		storeNext: func(entity, next uint64) {
			w.entity(entity).next = next
			w.observe("store-next")
		},
		storePrev: func(entity, prev uint64) {
			w.entity(entity).prev = prev
			w.observe("store-prev")
		},
		storeHead: func(entity uint64) {
			w.head = entity
			w.observe("store-head")
		},
	}
}

func runSpellBookInsertTest4FE340(w *spellBookInsertTestWorld4FE340) (result int32, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	result = spellBookInsert4FE340(w.hooks())
	return result, false
}

func TestSpellBookInsert4FE340NonGlyphTraceAndFaultPrefixes(t *testing.T) {
	w := newSpellBookInsertTestWorld4FE340()
	result, panicked := runSpellBookInsertTest4FE340(w)
	if panicked || result != 1 {
		t.Fatalf("result = (%d, panic %t), want (1, false)", result, panicked)
	}
	want := []string{
		"unit", "flags", "sequence",
		"spell:0", "spell:1", "spell:2", "spell:3", "spell:4",
		"class", "update", "trade", "player",
		"spell:0", "known:10", "spell:1", "known:0", "spell:2", "known:0", "spell:3", "known:0", "spell:4", "known:0",
		"cast-start", "count", "spell:0",
		"spell:0", "precheck:10", "spell:0", "cast:10:0",
		"state:2", "frame", "store-casting:1", "store-start:100", "allocator", "allocate",
		"target", "delay", "store-unit", "store-target", "store-field36", "frame", "store-frame", "store-delay",
		"definitions", "store-definitions", "store-index:0", "store-glyph:0",
		"count", "spell:0", "store-spell:0", "spell:0",
		"count", "store-spell:1", "count", "store-spell:2", "count", "store-spell:3", "count", "store-spell:4",
		"store-prev", "head", "store-next", "head", "store-head",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events:\n got %q\nwant %q", w.events, want)
	}
	entity := w.entity(w.allocated)
	if entity.unit != w.unit || entity.spells != [5]int32{10, 0, 0, 0, 0} || entity.definitions != w.definitions ||
		entity.targetMode != uint32(w.targetMode) || entity.delay != uint32(w.delay) || entity.frame != 101 ||
		w.state != spellBookInsertState4FE340 || w.casting != 1 || w.castStart != 100 || w.head != w.allocated {
		t.Fatalf("success state = entity %#v, state %d, casting %d, start %d, head %#x", entity, w.state, w.casting, w.castStart, w.head)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		faultWorld := newSpellBookInsertTestWorld4FE340()
		faultWorld.faultAt = faultAt
		_, didPanic := runSpellBookInsertTest4FE340(faultWorld)
		if !didPanic {
			t.Fatalf("fault %d did not panic", faultAt)
		}
		if !reflect.DeepEqual(faultWorld.events, want[:faultAt]) {
			t.Fatalf("fault %d events = %q, want prefix %q", faultAt, faultWorld.events, want[:faultAt])
		}
	}
}

func TestSpellBookInsert4FE340EntryGatesAndOrderedPlayerLoad(t *testing.T) {
	t.Run("blocked flags", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.flags = spellBookInsertBlockedFlags4FE340
		if got, _ := runSpellBookInsertTest4FE340(w); got != 0 || !reflect.DeepEqual(w.events, []string{"unit", "flags"}) {
			t.Fatalf("result/events = %d/%q", got, w.events)
		}
	})
	t.Run("signed spell bounds", func(t *testing.T) {
		for _, bad := range []int32{-1, spellBookInsertSpellCount4FE340} {
			w := newSpellBookInsertTestWorld4FE340()
			w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, 2}] = bad
			if got, _ := runSpellBookInsertTest4FE340(w); got != 0 {
				t.Fatalf("spell %d result = %d", bad, got)
			}
			if want := []string{"unit", "flags", "sequence", "spell:0", "spell:1", "spell:2"}; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("spell %d events = %q, want %q", bad, w.events, want)
			}
		}
	})
	t.Run("trade caches player before branch", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.trade = 0x100000001
		if got, _ := runSpellBookInsertTest4FE340(w); got != 0 {
			t.Fatalf("result = %d", got)
		}
		last := w.events[len(w.events)-2:]
		if want := []string{"trade", "player"}; !reflect.DeepEqual(last, want) {
			t.Fatalf("trade tail = %q", last)
		}
	})
	t.Run("zero spell may be unknown", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, 0}] = 0
		if got, _ := runSpellBookInsertTest4FE340(w); got != 1 {
			t.Fatalf("result = %d", got)
		}
	})
	t.Run("nonzero unknown rejects", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.known[w.player][10] = 0
		if got, _ := runSpellBookInsertTest4FE340(w); got != 0 || w.events[len(w.events)-1] != "known:10" {
			t.Fatalf("result/events = %d/%q", got, w.events)
		}
	})
}

func TestSpellBookInsert4FE340CountExtentAndLiveRecordLoads(t *testing.T) {
	t.Run("negative count stores five zero slots", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.count = -1
		if got, _ := runSpellBookInsertTest4FE340(w); got != 1 {
			t.Fatalf("result = %d", got)
		}
		if got := w.entity(w.allocated).spells; got != [5]int32{} {
			t.Fatalf("record spells = %v", got)
		}
		if got := w.spellLoads[spellBookInsertTestSpellKey4FE340{w.sequence, 0}]; got != 4 {
			t.Fatalf("spell[0] loads = %d, want validation + known + two non-Glyph loads", got)
		}
	})

	t.Run("positive count walks beyond five without widening the record", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.count = 7
		for i, id := range []int32{10, 11, 12, 13, 14, 15, spellBookInsertGlyph4FE340} {
			w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, int32(i)}] = id
			if i < 5 {
				w.known[w.player][id] = 1
			}
		}
		if got, _ := runSpellBookInsertTest4FE340(w); got != 1 {
			t.Fatalf("result = %d", got)
		}
		entity := w.entity(w.allocated)
		if entity.spells != [5]int32{10, 11, 12, 13, 14} || entity.glyphMode != 0 {
			t.Fatalf("record = spells %v, glyph %d", entity.spells, entity.glyphMode)
		}
		for index := int32(0); index < 7; index++ {
			if w.spellLoads[spellBookInsertTestSpellKey4FE340{w.sequence, index}] == 0 {
				t.Fatalf("spell[%d] was not read", index)
			}
		}
	})

	t.Run("copy and glyph test are independent live reads", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.after["store-spell:0"] = func() {
			w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, 0}] = spellBookInsertGlyph4FE340
		}
		if got, _ := runSpellBookInsertTest4FE340(w); got != 1 {
			t.Fatalf("result = %d", got)
		}
		entity := w.entity(w.allocated)
		if entity.spells[0] != 10 || entity.glyphMode != 1 {
			t.Fatalf("record = spell %d, glyph %d", entity.spells[0], entity.glyphMode)
		}
	})

	t.Run("two head reads may observe different nodes", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		w.head = spellBookInsertTestHeadA4FE340
		w.entity(spellBookInsertTestHeadA4FE340)
		w.entity(spellBookInsertTestHeadB4FE340)
		w.after["store-next"] = func() { w.head = spellBookInsertTestHeadB4FE340 }
		if got, _ := runSpellBookInsertTest4FE340(w); got != 1 {
			t.Fatalf("result = %d", got)
		}
		if w.entity(w.allocated).next != spellBookInsertTestHeadA4FE340 ||
			w.entity(spellBookInsertTestHeadA4FE340).prev != 0 ||
			w.entity(spellBookInsertTestHeadB4FE340).prev != w.allocated || w.head != w.allocated {
			t.Fatalf("links = new %#v, A %#v, B %#v, head %#x", w.entity(w.allocated), w.entity(spellBookInsertTestHeadA4FE340), w.entity(spellBookInsertTestHeadB4FE340), w.head)
		}
	})
}

func TestSpellBookInsert4FE340ReloadsCountAndSequenceAtOriginalSites(t *testing.T) {
	t.Run("Glyph loop reloads count after each cast gate", func(t *testing.T) {
		w := newSpellBookInsertTestWorld4FE340()
		for i, id := range []int32{spellBookInsertGlyph4FE340, 10, 11} {
			w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, int32(i)}] = id
			w.known[w.player][id] = 1
		}
		w.after["cast:34:1"] = func() { w.count = 3 }
		if got, _ := runSpellBookInsertTest4FE340(w); got != 1 {
			t.Fatalf("result = %d", got)
		}
		var prechecks []string
		for _, event := range w.events {
			if len(event) >= len("precheck:") && event[:len("precheck:")] == "precheck:" {
				prechecks = append(prechecks, event)
			}
		}
		if want := []string{"precheck:34", "precheck:10", "precheck:11"}; !reflect.DeepEqual(prechecks, want) {
			t.Fatalf("prechecks = %q, want %q", prechecks, want)
		}
	})

	t.Run("Glyph validation loop and record use three sequence snapshots", func(t *testing.T) {
		const (
			loopSequence   = uint64(0xb00000b0b)
			recordSequence = uint64(0xc00000c0c)
		)
		w := newSpellBookInsertTestWorld4FE340()
		w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, 0}] = spellBookInsertGlyph4FE340
		w.known[w.player][spellBookInsertGlyph4FE340] = 1
		w.spells[spellBookInsertTestSpellKey4FE340{loopSequence, 0}] = 20
		w.spells[spellBookInsertTestSpellKey4FE340{recordSequence, 0}] = 30
		w.after["mana"] = func() { w.sequence = loopSequence }
		w.after["cast:20:1"] = func() { w.sequence = recordSequence }
		if got, _ := runSpellBookInsertTest4FE340(w); got != 1 {
			t.Fatalf("result = %d", got)
		}
		if w.spellLoads[spellBookInsertTestSpellKey4FE340{loopSequence, 0}] != 2 {
			t.Fatalf("loop sequence loads = %d", w.spellLoads[spellBookInsertTestSpellKey4FE340{loopSequence, 0}])
		}
		entity := w.entity(w.allocated)
		if entity.spells[0] != 30 || entity.glyphMode != 0 {
			t.Fatalf("record = spell %d, glyph %d", entity.spells[0], entity.glyphMode)
		}
	})
}

func TestSpellBookInsert4FE340GlyphFailures(t *testing.T) {
	test := func(t *testing.T, configure func(*spellBookInsertTestWorld4FE340), wantResult, wantSound int32, wantTail []string) {
		t.Helper()
		w := newSpellBookInsertTestWorld4FE340()
		w.spells[spellBookInsertTestSpellKey4FE340{w.sequence, 0}] = spellBookInsertGlyph4FE340
		w.known[w.player][spellBookInsertGlyph4FE340] = 1
		w.known[spellBookInsertTestPlayerB4FE340][spellBookInsertGlyph4FE340] = 1
		configure(w)
		if got, _ := runSpellBookInsertTest4FE340(w); got != 0 {
			t.Fatalf("result = %d", got)
		}
		if got := w.informs; !reflect.DeepEqual(got, []spellBookInsertTestInform4FE340{{index: w.playerIndex[w.player], result: wantResult}}) {
			t.Fatalf("informs = %#v", got)
		}
		if got := w.audio; !reflect.DeepEqual(got, []spellBookInsertTestAudio4FE340{{id: wantSound, unit: w.unit}}) {
			t.Fatalf("audio = %#v", got)
		}
		if len(wantTail) != 0 {
			gotTail := w.events[len(w.events)-len(wantTail):]
			if !reflect.DeepEqual(gotTail, wantTail) {
				t.Fatalf("tail = %q, want %q", gotTail, wantTail)
			}
		}
	}

	t.Run("mana", func(t *testing.T) {
		test(t, func(w *spellBookInsertTestWorld4FE340) {
			w.mana = 0
			w.after["mana"] = func() { w.player = spellBookInsertTestPlayerB4FE340 }
		}, spellBookInsertNotEnoughManaGlyph4FE340, spellBookInsertManaEmptySound4FE340,
			[]string{"mana", "player", "player-index", "inform:12", "audio:232"})
	})
	t.Run("summon limit", func(t *testing.T) {
		test(t, func(w *spellBookInsertTestWorld4FE340) {
			w.playerClass[w.player] = spellBookInsertConjurer4FE340
			w.summoned = 0
		}, spellBookInsertCreatureControlFailed4FE340, spellBookInsertFizzleSound4FE340,
			[]string{"summoned", "player", "player-index", "inform:4", "audio:231"})
	})
	t.Run("bomber limit truncates", func(t *testing.T) {
		test(t, func(w *spellBookInsertTestWorld4FE340) {
			w.playerClass[w.player] = spellBookInsertConjurer4FE340
			w.slaves = 2
			w.balance[spellBookInsertMaxBomberKey4FE340] = 2.9
		}, spellBookInsertTooManyGlyphs4FE340, spellBookInsertFizzleSound4FE340,
			[]string{"slaves", "balance:MaxBomberCount", "player", "player-index", "inform:5", "audio:231"})
	})
	t.Run("trap limit loads balance before byte", func(t *testing.T) {
		test(t, func(w *spellBookInsertTestWorld4FE340) {
			w.curTraps = 2
			w.balance[spellBookInsertMaxTrapKey4FE340] = 2.9
		}, spellBookInsertTooManyGlyphs4FE340, spellBookInsertFizzleSound4FE340,
			[]string{"balance:MaxTrapCount", "traps", "player", "player-index", "inform:5", "audio:231"})
	})
	t.Run("precheck", func(t *testing.T) {
		test(t, func(w *spellBookInsertTestWorld4FE340) {
			w.precheck[spellBookInsertGlyph4FE340] = 9
		}, 9, spellBookInsertFizzleSound4FE340,
			[]string{"precheck:34", "player", "player-index", "inform:9", "audio:231"})
	})
	t.Run("cast gate", func(t *testing.T) {
		test(t, func(w *spellBookInsertTestWorld4FE340) {
			w.castGate[spellBookInsertGlyph4FE340] = 10
		}, 10, spellBookInsertFizzleSound4FE340,
			[]string{"cast:34:1", "player", "player-index", "inform:10", "audio:231"})
	})
}

func TestSpellBookInsert4FE340AllocationFailureKeepsStartedState(t *testing.T) {
	w := newSpellBookInsertTestWorld4FE340()
	w.allocated = 0
	if got, _ := runSpellBookInsertTest4FE340(w); got != 0 {
		t.Fatalf("result = %d", got)
	}
	if w.state != spellBookInsertState4FE340 || w.casting != 1 || w.castStart != 100 {
		t.Fatalf("started state = state %d, casting %d, start %d", w.state, w.casting, w.castStart)
	}
	if got := w.events[len(w.events)-2:]; !reflect.DeepEqual(got, []string{"allocator", "allocate"}) {
		t.Fatalf("tail = %q", got)
	}
	for _, event := range w.events {
		if event == "target" || event == "delay" {
			t.Fatalf("allocation failure observed later argument %q", event)
		}
	}
}
