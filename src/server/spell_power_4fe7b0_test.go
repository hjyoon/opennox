package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellPowerTestCasterA4FE7B0 = uint64(0x100000101)
	spellPowerTestCasterB4FE7B0 = uint64(0x200000202)
	spellPowerTestUpdateA4FE7B0 = uint64(0x300000303)
	spellPowerTestUpdateB4FE7B0 = uint64(0x400000404)
	spellPowerTestPlayerA4FE7B0 = uint64(0x500000505)
	spellPowerTestPlayerB4FE7B0 = uint64(0x600000606)
)

type spellPowerTestLevelKey4FE7B0 struct {
	player uint64
	spell  int32
}

type spellPowerTestWorld4FE7B0 struct {
	events  []string
	faultAt int
	after   map[string]func()

	cache, lookup uint32
	caster        uint64
	types         map[uint64]uint16
	gameResult    int32
	classes       map[uint64]uint32
	updates       map[uint64]uint64
	spell         int32
	players       map[uint64]uint64
	levels        map[spellPowerTestLevelKey4FE7B0]int32
	monsterPower  map[uint64]int32
}

func newSpellPowerTestWorld4FE7B0() *spellPowerTestWorld4FE7B0 {
	return &spellPowerTestWorld4FE7B0{
		after:        make(map[string]func()),
		cache:        0,
		lookup:       0x12345,
		caster:       spellPowerTestCasterA4FE7B0,
		types:        map[uint64]uint16{spellPowerTestCasterA4FE7B0: 0x2345},
		classes:      map[uint64]uint32{spellPowerTestCasterA4FE7B0: 0xabcd0004},
		updates:      map[uint64]uint64{spellPowerTestCasterA4FE7B0: spellPowerTestUpdateA4FE7B0},
		spell:        -17,
		players:      map[uint64]uint64{spellPowerTestUpdateA4FE7B0: spellPowerTestPlayerA4FE7B0},
		levels:       map[spellPowerTestLevelKey4FE7B0]int32{{spellPowerTestPlayerA4FE7B0, -17}: -0x1234567},
		monsterPower: map[uint64]int32{spellPowerTestUpdateA4FE7B0: 0x76543210},
	}
}

func (w *spellPowerTestWorld4FE7B0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellPowerTestWorld4FE7B0) hooks() spellPowerHooks4FE7B0[uint64, uint64, uint64] {
	return spellPowerHooks4FE7B0[uint64, uint64, uint64]{
		loadImaginaryCasterType: func() uint32 {
			value := w.cache
			w.observe("cache")
			return value
		},
		lookupObjectType: func(name string) uint32 {
			value := w.lookup
			w.observe("lookup:" + name)
			return value
		},
		storeImaginaryCasterType: func(value uint32) {
			w.cache = value
			w.observe(fmt.Sprintf("store:%x", value))
		},
		loadCasterArg: func() uint64 {
			value := w.caster
			w.observe("caster")
			return value
		},
		loadCasterType: func(object uint64) uint16 {
			value := w.types[object]
			w.observe(fmt.Sprintf("type:%x", object))
			return value
		},
		hasGameFlag: func(mask uint32) int32 {
			value := w.gameResult
			w.observe(fmt.Sprintf("game:%x", mask))
			return value
		},
		loadClass: func(object uint64) uint32 {
			value := w.classes[object]
			w.observe(fmt.Sprintf("class:%x", object))
			return value
		},
		loadUpdate: func(object uint64) uint64 {
			value := w.updates[object]
			w.observe(fmt.Sprintf("update:%x", object))
			return value
		},
		loadSpellArg: func() int32 {
			value := w.spell
			w.observe("spell")
			return value
		},
		loadPlayer: func(update uint64) uint64 {
			value := w.players[update]
			w.observe(fmt.Sprintf("player:%x", update))
			return value
		},
		loadPlayerPower: func(player uint64, spell int32) int32 {
			value := w.levels[spellPowerTestLevelKey4FE7B0{player, spell}]
			w.observe(fmt.Sprintf("level:%x:%d", player, spell))
			return value
		},
		loadMonsterPower: func(update uint64) int32 {
			value := w.monsterPower[update]
			w.observe(fmt.Sprintf("monster:%x", update))
			return value
		},
	}
}

func TestSpellPower4FE7B0ExactPlayerTrace(t *testing.T) {
	w := newSpellPowerTestWorld4FE7B0()
	if got := spellPower4FE7B0(w.hooks()); got != -0x1234567 {
		t.Fatalf("power = %#x, want %#x", got, int32(-0x1234567))
	}
	want := []string{
		"cache", "lookup:ImaginaryCaster", "store:12345", "caster",
		"type:100000101", "game:570", "class:100000101",
		"update:100000101", "spell", "player:300000303", "level:500000505:-17",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.cache != w.lookup {
		t.Fatalf("cache = %#x, want lookup %#x", w.cache, w.lookup)
	}
}

func TestSpellPower4FE7B0GatesAndClassPrecedence(t *testing.T) {
	t.Run("imaginary caster compares zero-extended type with full cache", func(t *testing.T) {
		w := newSpellPowerTestWorld4FE7B0()
		w.cache = 0x2345
		if got := spellPower4FE7B0(w.hooks()); got != 1 {
			t.Fatalf("power = %d, want 1", got)
		}
		if want := []string{"cache", "caster", "type:100000101"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}

		w = newSpellPowerTestWorld4FE7B0()
		w.cache = 0x12345
		w.classes[w.caster] = 0
		if got := spellPower4FE7B0(w.hooks()); got != 3 {
			t.Fatalf("full-width cache mismatch power = %d, want 3", got)
		}
	})

	t.Run("lookup result zero may match type zero", func(t *testing.T) {
		w := newSpellPowerTestWorld4FE7B0()
		w.lookup = 0
		w.types[w.caster] = 0
		if got := spellPower4FE7B0(w.hooks()); got != 1 {
			t.Fatalf("power = %d, want 1", got)
		}
		if want := []string{"cache", "lookup:ImaginaryCaster", "store:0", "caster", "type:100000101"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("every nonzero game result forces three", func(t *testing.T) {
		for _, result := range []int32{1, -1, 7} {
			w := newSpellPowerTestWorld4FE7B0()
			w.cache = 9
			w.gameResult = result
			if got := spellPower4FE7B0(w.hooks()); got != 3 {
				t.Fatalf("result %d power = %d, want 3", result, got)
			}
			if want := []string{"cache", "caster", "type:100000101", "game:570"}; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("result %d events = %q, want %q", result, w.events, want)
			}
		}
	})

	t.Run("nil gate follows type and game", func(t *testing.T) {
		w := newSpellPowerTestWorld4FE7B0()
		w.cache = 9
		w.caster = 0
		if got := spellPower4FE7B0(w.hooks()); got != 2 {
			t.Fatalf("power = %d, want 2", got)
		}
		if want := []string{"cache", "caster", "type:0", "game:570"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("player wins over monster and alone reads spell", func(t *testing.T) {
		w := newSpellPowerTestWorld4FE7B0()
		w.cache = 9
		w.classes[w.caster] = uint32(spellPowerPlayerClass4FE7B0 | spellPowerMonsterClass4FE7B0)
		if got := spellPower4FE7B0(w.hooks()); got != -0x1234567 {
			t.Fatalf("power = %#x", got)
		}
		if len(w.events) == 0 || w.events[len(w.events)-1] != "level:500000505:-17" {
			t.Fatalf("events = %q", w.events)
		}
	})

	t.Run("monster and fallback skip spell", func(t *testing.T) {
		w := newSpellPowerTestWorld4FE7B0()
		w.cache = 9
		w.classes[w.caster] = spellPowerMonsterClass4FE7B0
		if got := spellPower4FE7B0(w.hooks()); got != 0x76543210 {
			t.Fatalf("monster power = %#x", got)
		}
		if want := []string{"cache", "caster", "type:100000101", "game:570", "class:100000101", "update:100000101", "monster:300000303"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("monster events = %q, want %q", w.events, want)
		}

		w = newSpellPowerTestWorld4FE7B0()
		w.cache = 9
		w.classes[w.caster] = 0xffffff00
		if got := spellPower4FE7B0(w.hooks()); got != 3 {
			t.Fatalf("fallback power = %d, want 3", got)
		}
		if w.events[len(w.events)-1] != "class:100000101" {
			t.Fatalf("fallback events = %q", w.events)
		}
	})
}

func TestSpellPower4FE7B0CachesAndReloadsAtInstructionBoundaries(t *testing.T) {
	w := newSpellPowerTestWorld4FE7B0()
	w.types[spellPowerTestCasterB4FE7B0] = 0x4321
	w.classes[spellPowerTestCasterB4FE7B0] = 0
	w.updates[spellPowerTestCasterB4FE7B0] = spellPowerTestUpdateA4FE7B0
	w.after["lookup:ImaginaryCaster"] = func() { w.lookup = 0x99999 }
	w.after["store:12345"] = func() { w.caster = spellPowerTestCasterB4FE7B0 }
	w.after["type:200000202"] = func() { w.cache = 0x4321 }
	w.after["game:570"] = func() { w.classes[spellPowerTestCasterB4FE7B0] = spellPowerPlayerClass4FE7B0 }
	w.after["class:200000202"] = func() {
		w.classes[spellPowerTestCasterB4FE7B0] = spellPowerMonsterClass4FE7B0
	}
	w.after["update:200000202"] = func() {
		w.updates[spellPowerTestCasterB4FE7B0] = spellPowerTestUpdateB4FE7B0
		w.spell = 77
	}
	w.after["spell"] = func() { w.players[spellPowerTestUpdateA4FE7B0] = spellPowerTestPlayerB4FE7B0 }
	w.after["player:300000303"] = func() {
		w.players[spellPowerTestUpdateA4FE7B0] = spellPowerTestPlayerA4FE7B0
		w.levels[spellPowerTestLevelKey4FE7B0{spellPowerTestPlayerB4FE7B0, 77}] = 0x10203040
	}

	if got := spellPower4FE7B0(w.hooks()); got != 0x10203040 {
		t.Fatalf("power = %#x, want 0x10203040", got)
	}
	if w.cache != 0x4321 {
		t.Fatalf("post-callback cache = %#x, want mutation 0x4321", w.cache)
	}
	want := []string{
		"cache", "lookup:ImaginaryCaster", "store:12345", "caster",
		"type:200000202", "game:570", "class:200000202",
		"update:200000202", "spell", "player:300000303", "level:600000606:77",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestSpellPower4FE7B0EveryObservableFaultPrefix(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*spellPowerTestWorld4FE7B0)
	}{
		{name: "cache_miss_player", setup: func(*spellPowerTestWorld4FE7B0) {}},
		{name: "cache_hit_monster", setup: func(w *spellPowerTestWorld4FE7B0) {
			w.cache = 9
			w.classes[w.caster] = spellPowerMonsterClass4FE7B0
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			baseline := newSpellPowerTestWorld4FE7B0()
			tc.setup(baseline)
			_ = spellPower4FE7B0(baseline.hooks())
			want := append([]string(nil), baseline.events...)

			for faultAt := 1; faultAt <= len(want); faultAt++ {
				w := newSpellPowerTestWorld4FE7B0()
				tc.setup(w)
				w.faultAt = faultAt
				func() {
					defer func() {
						if recover() == nil {
							t.Fatalf("fault %d did not panic", faultAt)
						}
					}()
					_ = spellPower4FE7B0(w.hooks())
				}()
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("fault %d events = %q, want %q", faultAt, w.events, want[:faultAt])
				}
			}
		})
	}
}
