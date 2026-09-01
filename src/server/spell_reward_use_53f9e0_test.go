package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type spellRewardUseTestData53F9E0 struct {
	spell uint8
}

type spellRewardUseTestPlayer53F9E0 struct {
	name   string
	class  uint8
	levels [137]uint32
}

type spellRewardUseTestUpdate53F9E0 struct {
	player *spellRewardUseTestPlayer53F9E0
}

type spellRewardUseTestObject53F9E0 struct {
	name    string
	class   uint8
	netCode uint32
	update  *spellRewardUseTestUpdate53F9E0
	data    *spellRewardUseTestData53F9E0
}

type spellRewardUseTestWorld53F9E0 struct {
	owner             *spellRewardUseTestObject53F9E0
	item              *spellRewardUseTestObject53F9E0
	flags             int32
	classCheckResult  int32
	grantResult       int32
	events            []string
	faultAt           int
	afterSpell        func(int)
	spellLoads        int
	afterFlags        func()
	afterPrimary      func()
	observedCheckArgs [2]uint8
	observedGrantArgs [4]int32
}

func spellRewardUseTestName53F9E0(obj *spellRewardUseTestObject53F9E0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *spellRewardUseTestWorld53F9E0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *spellRewardUseTestWorld53F9E0) hooks() spellRewardUseHooks53F9E0[
	*spellRewardUseTestObject53F9E0,
	*spellRewardUseTestUpdate53F9E0,
	*spellRewardUseTestPlayer53F9E0,
	*spellRewardUseTestData53F9E0,
] {
	return spellRewardUseHooks53F9E0[
		*spellRewardUseTestObject53F9E0,
		*spellRewardUseTestUpdate53F9E0,
		*spellRewardUseTestPlayer53F9E0,
		*spellRewardUseTestData53F9E0,
	]{
		loadItemArg: func() *spellRewardUseTestObject53F9E0 {
			w.event("item:" + spellRewardUseTestName53F9E0(w.item))
			return w.item
		},
		loadOwnerArg: func() *spellRewardUseTestObject53F9E0 {
			w.event("owner:" + spellRewardUseTestName53F9E0(w.owner))
			return w.owner
		},
		loadUseData: func(item *spellRewardUseTestObject53F9E0) *spellRewardUseTestData53F9E0 {
			w.event("data:" + spellRewardUseTestName53F9E0(item))
			return item.data
		},
		loadClassLow: func(owner *spellRewardUseTestObject53F9E0) uint8 {
			w.event(fmt.Sprintf("class:%s=%02x", spellRewardUseTestName53F9E0(owner), owner.class))
			return owner.class
		},
		loadUpdateData: func(owner *spellRewardUseTestObject53F9E0) *spellRewardUseTestUpdate53F9E0 {
			w.event("update:" + spellRewardUseTestName53F9E0(owner))
			return owner.update
		},
		loadPlayer: func(update *spellRewardUseTestUpdate53F9E0) *spellRewardUseTestPlayer53F9E0 {
			w.event("player:" + update.player.name)
			return update.player
		},
		loadPlayerClass: func(player *spellRewardUseTestPlayer53F9E0) uint8 {
			w.event(fmt.Sprintf("player-class:%s=%02x", player.name, player.class))
			return player.class
		},
		loadSpell: func(data *spellRewardUseTestData53F9E0) uint8 {
			value := data.spell
			w.spellLoads++
			w.event(fmt.Sprintf("spell:%d=%02x", w.spellLoads, value))
			if w.afterSpell != nil {
				w.afterSpell(w.spellLoads)
			}
			return value
		},
		checkSpellClass: func(class, spell uint8) int32 {
			w.observedCheckArgs = [2]uint8{class, spell}
			w.event(fmt.Sprintf("check:%02x:%02x=%08x", class, spell, uint32(w.classCheckResult)))
			return w.classCheckResult
		},
		primaryMessage: func(owner *spellRewardUseTestObject53F9E0, message string, value uint8) {
			w.event(fmt.Sprintf("primary:%s:%s:%d", spellRewardUseTestName53F9E0(owner), message, value))
			if w.afterPrimary != nil {
				w.afterPrimary()
			}
		},
		loadNetCode: func(owner *spellRewardUseTestObject53F9E0) uint32 {
			w.event(fmt.Sprintf("netcode:%s=%08x", spellRewardUseTestName53F9E0(owner), owner.netCode))
			return owner.netCode
		},
		audit: func(sound int32, owner *spellRewardUseTestObject53F9E0, kind int32, code uint32) {
			w.event(fmt.Sprintf("audit:%d:%s:%d:%08x", sound, spellRewardUseTestName53F9E0(owner), kind, code))
		},
		gameFlagsCheck: func(mask uint32) int32 {
			w.event(fmt.Sprintf("flags:%08x=%08x", mask, uint32(w.flags)))
			if w.afterFlags != nil {
				w.afterFlags()
			}
			return w.flags
		},
		loadSpellLevel: func(player *spellRewardUseTestPlayer53F9E0, spell int32) uint32 {
			value := player.levels[spell]
			w.event(fmt.Sprintf("level:%s:%d=%08x", player.name, spell, value))
			return value
		},
		grantSpell: func(owner *spellRewardUseTestObject53F9E0, spell, notify, quest, override int32) int32 {
			w.observedGrantArgs = [4]int32{spell, notify, quest, override}
			w.event(fmt.Sprintf("grant:%s:%d:%d:%d:%d=%08x", spellRewardUseTestName53F9E0(owner), spell, notify, quest, override, uint32(w.grantResult)))
			return w.grantResult
		},
		delayedDeleteItem: func(item *spellRewardUseTestObject53F9E0) {
			w.event("delete:" + spellRewardUseTestName53F9E0(item))
		},
	}
}

func newSpellRewardUseTestWorld53F9E0() *spellRewardUseTestWorld53F9E0 {
	player := &spellRewardUseTestPlayer53F9E0{
		name:  "player",
		class: spellRewardUseWarrior53F9E0,
	}
	owner := &spellRewardUseTestObject53F9E0{
		name:    "owner",
		class:   0xf4,
		netCode: 0x89abcdef,
		update:  &spellRewardUseTestUpdate53F9E0{player: player},
	}
	item := &spellRewardUseTestObject53F9E0{
		name: "item",
		data: &spellRewardUseTestData53F9E0{spell: 2},
	}
	return &spellRewardUseTestWorld53F9E0{
		owner:       owner,
		item:        item,
		flags:       math.MinInt32,
		grantResult: math.MinInt32,
	}
}

func spellRewardUseSuccessTrace53F9E0() []string {
	return []string{
		"item:item",
		"owner:owner",
		"data:item",
		"class:owner=f4",
		"update:owner",
		"player:player",
		"player-class:player=01",
		"spell:1=02",
		"check:01:02=00000000",
		"flags:00001800=80000000",
		"player:player",
		"spell:2=05",
		"level:player:5=00000000",
		"spell:3=07",
		"grant:owner:7:1:1:0=80000000",
		"delete:item",
	}
}

func TestUseSpellReward53F9E0ExactSuccessTraceLiveSpellAndFaultPrefixes(t *testing.T) {
	want := spellRewardUseSuccessTrace53F9E0()
	build := func() *spellRewardUseTestWorld53F9E0 {
		w := newSpellRewardUseTestWorld53F9E0()
		w.afterSpell = func(load int) {
			switch load {
			case 1:
				w.item.data.spell = 5
			case 2:
				w.item.data.spell = 7
			}
		}
		return w
	}

	w := build()
	if got := useSpellReward53F9E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.observedCheckArgs != [2]uint8{1, 2} {
		t.Fatalf("check args = %v, want [1 2]", w.observedCheckArgs)
	}
	if w.observedGrantArgs != [4]int32{7, 1, 1, 0} {
		t.Fatalf("grant args = %v, want [7 1 1 0]", w.observedGrantArgs)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			useSpellReward53F9E0(w.hooks())
		})
	}
}

func TestUseSpellReward53F9E0LoadsCachedPointersBeforeNonPlayerReturn(t *testing.T) {
	w := newSpellRewardUseTestWorld53F9E0()
	w.owner.class = 0xf0
	if got := useSpellReward53F9E0(w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"item:item", "owner:owner", "data:item", "class:owner=f0", "update:owner",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestUseSpellReward53F9E0RejectsClassAndKnownSpellWithLiveNetCode(t *testing.T) {
	tests := []struct {
		name        string
		playerClass uint8
		checkResult int32
		wantPrefix  []string
	}{
		{
			name:        "class",
			playerClass: 0x81,
			wantPrefix:  []string{"player-class:player=81"},
		},
		{
			name:        "known-spell",
			playerClass: spellRewardUseConjurer53F9E0,
			checkResult: math.MinInt32,
			wantPrefix: []string{
				"player-class:player=02",
				"spell:1=02",
				"check:02:02=80000000",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellRewardUseTestWorld53F9E0()
			w.owner.update.player.class = tc.playerClass
			w.classCheckResult = tc.checkResult
			w.afterPrimary = func() { w.owner.netCode = 0x10203040 }
			if got := useSpellReward53F9E0(w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			wantTail := append(append([]string{}, tc.wantPrefix...),
				"primary:owner:use.c:SpellRewardClassFail:0",
				"netcode:owner=10203040",
				"audit:925:owner:2:10203040",
			)
			if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("tail = %v, want %v", got, wantTail)
			}
		})
	}
}

func TestUseSpellReward53F9E0QuestAndGrantBranches(t *testing.T) {
	t.Run("zero-flags-skips-level-and-audits-zero-grant", func(t *testing.T) {
		w := newSpellRewardUseTestWorld53F9E0()
		w.flags = 0
		w.grantResult = 0
		w.owner.netCode = math.MaxUint32
		if got := useSpellReward53F9E0(w.hooks()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.observedGrantArgs != [4]int32{2, 1, 0, 0} {
			t.Fatalf("grant args = %v, want [2 1 0 0]", w.observedGrantArgs)
		}
		wantTail := []string{
			"flags:00001800=00000000",
			"spell:2=02",
			"grant:owner:2:1:0:0=00000000",
			"netcode:owner=ffffffff",
			"audit:925:owner:2:ffffffff",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	})

	t.Run("owned-level-clears-quest-argument", func(t *testing.T) {
		w := newSpellRewardUseTestWorld53F9E0()
		w.owner.update.player.levels[2] = math.MaxUint32
		if got := useSpellReward53F9E0(w.hooks()); got != 1 {
			t.Fatalf("result = %d, want 1", got)
		}
		if w.observedGrantArgs != [4]int32{2, 1, 0, 0} {
			t.Fatalf("grant args = %v, want [2 1 0 0]", w.observedGrantArgs)
		}
	})
}

func TestUseSpellReward53F9E0ReloadsPlayerButKeepsCachedDataAndUpdate(t *testing.T) {
	w := newSpellRewardUseTestWorld53F9E0()
	cachedUpdate := w.owner.update
	cachedData := w.item.data
	replacement := &spellRewardUseTestPlayer53F9E0{
		name:  "replacement",
		class: spellRewardUseConjurer53F9E0,
	}
	replacement.levels[2] = 9
	w.afterFlags = func() {
		cachedUpdate.player = replacement
		w.owner.update = &spellRewardUseTestUpdate53F9E0{
			player: &spellRewardUseTestPlayer53F9E0{name: "uncached"},
		}
		w.item.data = &spellRewardUseTestData53F9E0{spell: 99}
		cachedData.spell = 2
	}
	if got := useSpellReward53F9E0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if w.observedGrantArgs != [4]int32{2, 1, 0, 0} {
		t.Fatalf("grant args = %v, want [2 1 0 0]", w.observedGrantArgs)
	}
	want := []string{
		"player:replacement",
		"spell:2=02",
		"level:replacement:2=00000009",
		"spell:3=02",
	}
	for i := range w.events {
		if i+len(want) <= len(w.events) && reflect.DeepEqual(w.events[i:i+len(want)], want) {
			return
		}
	}
	t.Fatalf("events = %v, missing %v", w.events, want)
}
