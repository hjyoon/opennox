package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type spellGrantTestObject4FB550 struct {
	name   string
	class  uint8
	update *spellGrantTestUpdate4FB550
}

type spellGrantTestUpdate4FB550 struct {
	name   string
	player *spellGrantTestPlayer4FB550
	trade  *spellGrantTestTrade4FB550
}

type spellGrantTestPlayer4FB550 struct {
	name        string
	levels      [137]uint32
	protection  uint32
	notifyField uint32
	unit        *spellGrantTestObject4FB550
}

type spellGrantTestTrade4FB550 struct {
	name string
}

type spellGrantTestFlagKey4FB550 struct {
	spell int32
	mask  int32
}

type spellGrantTestWorld4FB550 struct {
	unitArg    *spellGrantTestObject4FB550
	players    []*spellGrantTestPlayer4FB550
	game       map[uint32][]int32
	gameCalls  map[uint32]int
	flags      map[spellGrantTestFlagKey4FB550][]int32
	flagCalls  map[spellGrantTestFlagKey4FB550]int
	valid      map[int32]int32
	events     []string
	faultAt    int
	afterEvent map[string]func()
}

func spellGrantObjectName4FB550(obj *spellGrantTestObject4FB550) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func spellGrantUpdateName4FB550(update *spellGrantTestUpdate4FB550) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func spellGrantPlayerName4FB550(player *spellGrantTestPlayer4FB550) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func spellGrantTradeName4FB550(trade *spellGrantTestTrade4FB550) string {
	if trade == nil {
		return "nil"
	}
	return trade.name
}

func (w *spellGrantTestWorld4FB550) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *spellGrantTestWorld4FB550) finish(event string) {
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func spellGrantQueuedResult4FB550[K comparable](values map[K][]int32, calls map[K]int, key K) int32 {
	index := calls[key]
	calls[key] = index + 1
	queue := values[key]
	if len(queue) == 0 {
		return 0
	}
	if index >= len(queue) {
		return queue[len(queue)-1]
	}
	return queue[index]
}

func (w *spellGrantTestWorld4FB550) hooks() spellGrantHooks4FB550[
	*spellGrantTestObject4FB550,
	*spellGrantTestUpdate4FB550,
	*spellGrantTestPlayer4FB550,
	*spellGrantTestTrade4FB550,
	string,
] {
	return spellGrantHooks4FB550[
		*spellGrantTestObject4FB550,
		*spellGrantTestUpdate4FB550,
		*spellGrantTestPlayer4FB550,
		*spellGrantTestTrade4FB550,
		string,
	]{
		loadUnitArg: func() *spellGrantTestObject4FB550 {
			unit := w.unitArg
			event := "arg:" + spellGrantObjectName4FB550(unit)
			w.record(event)
			w.finish(event)
			return unit
		},
		loadClassLow: func(unit *spellGrantTestObject4FB550) uint8 {
			if unit == nil {
				event := "class:nil"
				w.record(event)
				panic(event)
			}
			class := unit.class
			event := fmt.Sprintf("class:%s=%02x", unit.name, class)
			w.record(event)
			w.finish(event)
			return class
		},
		loadUpdateData: func(unit *spellGrantTestObject4FB550) *spellGrantTestUpdate4FB550 {
			if unit == nil {
				event := "update:nil"
				w.record(event)
				panic(event)
			}
			update := unit.update
			event := "update:" + unit.name + "=" + spellGrantUpdateName4FB550(update)
			w.record(event)
			w.finish(event)
			return update
		},
		loadPlayer: func(update *spellGrantTestUpdate4FB550) *spellGrantTestPlayer4FB550 {
			if update == nil {
				event := "player:nil-update"
				w.record(event)
				panic(event)
			}
			player := update.player
			event := "player:" + update.name + "=" + spellGrantPlayerName4FB550(player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadSpellLevel: func(player *spellGrantTestPlayer4FB550, spellID int32) uint32 {
			if player == nil {
				event := fmt.Sprintf("level:nil:%d", spellID)
				w.record(event)
				panic(event)
			}
			level := player.levels[spellID]
			event := fmt.Sprintf("level:%s:%d=%08x", player.name, spellID, level)
			w.record(event)
			w.finish(event)
			return level
		},
		storeSpellLevel: func(player *spellGrantTestPlayer4FB550, spellID int32, level uint32) {
			event := fmt.Sprintf("store:%s:%d=%08x", spellGrantPlayerName4FB550(player), spellID, level)
			w.record(event)
			if player == nil {
				panic(event)
			}
			player.levels[spellID] = level
			w.finish(event)
		},
		loadProtection: func(player *spellGrantTestPlayer4FB550) uint32 {
			if player == nil {
				event := "token:nil"
				w.record(event)
				panic(event)
			}
			protection := player.protection
			event := fmt.Sprintf("token:%s=%08x", player.name, protection)
			w.record(event)
			w.finish(event)
			return protection
		},
		gameFlagsCheck: func(mask uint32) int32 {
			result := spellGrantQueuedResult4FB550(w.game, w.gameCalls, mask)
			event := fmt.Sprintf("game:%08x=%d", mask, result)
			w.record(event)
			w.finish(event)
			return result
		},
		loadString: func(key, path string, line int) string {
			message := fmt.Sprintf("%s@%d", key, line)
			event := fmt.Sprintf("string:%s:%s:%d=%s", key, path, line, message)
			w.record(event)
			w.finish(event)
			return message
		},
		sendLineMessage: func(unit *spellGrantTestObject4FB550, message string) {
			event := "line:" + spellGrantObjectName4FB550(unit) + ":" + message
			w.record(event)
			w.finish(event)
		},
		awardProtection: func(protection uint32, spellID, level int32) {
			event := fmt.Sprintf("award:%08x:%d:%d", protection, spellID, level)
			w.record(event)
			w.finish(event)
		},
		spellHasFlags: func(spellID, mask int32) int32 {
			key := spellGrantTestFlagKey4FB550{spell: spellID, mask: mask}
			result := spellGrantQueuedResult4FB550(w.flags, w.flagCalls, key)
			event := fmt.Sprintf("flags:%d:%08x=%d", spellID, uint32(mask), result)
			w.record(event)
			w.finish(event)
			return result
		},
		spellIsValid: func(spellID int32) int32 {
			result := w.valid[spellID]
			event := fmt.Sprintf("valid:%d=%d", spellID, result)
			w.record(event)
			w.finish(event)
			return result
		},
		audio: func(id uint32, unit *spellGrantTestObject4FB550, kind int32, code uint32) {
			event := fmt.Sprintf("audio:%d:%s:%d:%d", id, spellGrantObjectName4FB550(unit), kind, code)
			w.record(event)
			w.finish(event)
		},
		loadNotifyField: func(player *spellGrantTestPlayer4FB550) uint32 {
			if player == nil {
				event := "notify-field:nil"
				w.record(event)
				panic(event)
			}
			value := player.notifyField
			event := fmt.Sprintf("notify-field:%s=%08x", player.name, value)
			w.record(event)
			w.finish(event)
			return value
		},
		rewardNotify: func(recipient *spellGrantTestObject4FB550, kind int32, source *spellGrantTestObject4FB550, spellID int32) {
			event := fmt.Sprintf("reward:%s:%d:%s:%d", spellGrantObjectName4FB550(recipient), kind, spellGrantObjectName4FB550(source), spellID)
			w.record(event)
			w.finish(event)
		},
		checkPlayerState: func(unit *spellGrantTestObject4FB550) int32 {
			event := "check:" + spellGrantObjectName4FB550(unit) + "=0"
			w.record(event)
			w.finish(event)
			return 0
		},
		firstPlayer: func() *spellGrantTestPlayer4FB550 {
			var player *spellGrantTestPlayer4FB550
			if len(w.players) != 0 {
				player = w.players[0]
			}
			event := "first:" + spellGrantPlayerName4FB550(player)
			w.record(event)
			w.finish(event)
			return player
		},
		nextPlayer: func(player *spellGrantTestPlayer4FB550) *spellGrantTestPlayer4FB550 {
			var next *spellGrantTestPlayer4FB550
			for index, candidate := range w.players {
				if candidate == player && index+1 < len(w.players) {
					next = w.players[index+1]
					break
				}
			}
			event := "next:" + spellGrantPlayerName4FB550(player) + "=" + spellGrantPlayerName4FB550(next)
			w.record(event)
			w.finish(event)
			return next
		},
		loadPlayerUnit: func(player *spellGrantTestPlayer4FB550) *spellGrantTestObject4FB550 {
			if player == nil {
				event := "unit:nil-player"
				w.record(event)
				panic(event)
			}
			unit := player.unit
			event := "unit:" + player.name + "=" + spellGrantObjectName4FB550(unit)
			w.record(event)
			w.finish(event)
			return unit
		},
		loadTrade: func(update *spellGrantTestUpdate4FB550) *spellGrantTestTrade4FB550 {
			if update == nil {
				event := "trade:nil-update"
				w.record(event)
				panic(event)
			}
			trade := update.trade
			event := "trade:" + update.name + "=" + spellGrantTradeName4FB550(trade)
			w.record(event)
			w.finish(event)
			return trade
		},
		shopExit: func(trade *spellGrantTestTrade4FB550) {
			event := "shop-exit:" + spellGrantTradeName4FB550(trade)
			w.record(event)
			w.finish(event)
		},
		reportSpellAward: func(unit *spellGrantTestObject4FB550, spellID, notify, shop int32) {
			event := fmt.Sprintf("report:%s:%d:%d:%d", spellGrantObjectName4FB550(unit), spellID, notify, shop)
			w.record(event)
			w.finish(event)
		},
	}
}

func newSpellGrantTestWorld4FB550() *spellGrantTestWorld4FB550 {
	player := &spellGrantTestPlayer4FB550{
		name:        "selected-player",
		protection:  0x12345678,
		notifyField: 1,
	}
	update := &spellGrantTestUpdate4FB550{
		name:   "selected-update",
		player: player,
		trade:  &spellGrantTestTrade4FB550{name: "trade"},
	}
	unit := &spellGrantTestObject4FB550{
		name:   "selected-unit",
		class:  spellGrantPlayerClass4FB550,
		update: update,
	}
	player.unit = unit
	return &spellGrantTestWorld4FB550{
		unitArg:    unit,
		players:    []*spellGrantTestPlayer4FB550{player},
		game:       make(map[uint32][]int32),
		gameCalls:  make(map[uint32]int),
		flags:      make(map[spellGrantTestFlagKey4FB550][]int32),
		flagCalls:  make(map[spellGrantTestFlagKey4FB550]int),
		valid:      make(map[int32]int32),
		afterEvent: make(map[string]func()),
	}
}

func TestSpellGrantToPlayer4FB550ClassAndErrorGates(t *testing.T) {
	t.Run("class gate", func(t *testing.T) {
		world := newSpellGrantTestWorld4FB550()
		world.unitArg.class = 0x80
		if got := spellGrantToPlayer4FB550(10, 1, 1, 0, world.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		want := []string{"arg:selected-unit", "class:selected-unit=80"}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("events = %v, want %v", world.events, want)
		}
	})

	for _, spellID := range []int32{0, -1, 137, math.MaxInt32} {
		t.Run(fmt.Sprintf("invalid %d", spellID), func(t *testing.T) {
			world := newSpellGrantTestWorld4FB550()
			if got := spellGrantToPlayer4FB550(spellID, 0, 0, 0, world.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			want := []string{
				"arg:selected-unit",
				"class:selected-unit=04",
				fmt.Sprintf("string:AwardSpellError:%s:339=AwardSpellError@339", spellGrantMessagePath4FB550),
				"line:selected-unit:AwardSpellError@339",
			}
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("events = %v, want %v", world.events, want)
			}
		})
	}

	for _, test := range []struct {
		name       string
		coopResult int32
		level      uint32
		wantLine   int
	}{
		{name: "coop level three", coopResult: -1, level: 3, wantLine: 351},
		{name: "global level five", coopResult: 0, level: 5, wantLine: 351},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := newSpellGrantTestWorld4FB550()
			world.game[spellGrantCoopQuestFlag4FB550] = []int32{test.coopResult}
			world.unitArg.update.player.levels[10] = test.level
			if got := spellGrantToPlayer4FB550(10, 0, 0, 0, world.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			wantTail := []string{
				fmt.Sprintf("string:MaxSpellLevel:%s:%d=MaxSpellLevel@%d", spellGrantMessagePath4FB550, test.wantLine, test.wantLine),
				fmt.Sprintf("line:selected-unit:MaxSpellLevel@%d", test.wantLine),
			}
			if got := world.events[len(world.events)-2:]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("tail = %v, want %v", got, wantTail)
			}
		})
	}

	t.Run("quest single level", func(t *testing.T) {
		world := newSpellGrantTestWorld4FB550()
		world.game[spellGrantQuestFlag4FB550] = []int32{1}
		world.unitArg.update.player.levels[34] = 2
		if got := spellGrantToPlayer4FB550(34, 0, 0, 0, world.hooks()); got != 0 {
			t.Fatalf("result = %d, want 0", got)
		}
		wantTail := []string{
			"game:00001000=1",
			"player:selected-update=selected-player",
			"level:selected-player:34=00000002",
			fmt.Sprintf("string:MaxSpellLevel:%s:386=MaxSpellLevel@386", spellGrantMessagePath4FB550),
			"line:selected-unit:MaxSpellLevel@386",
		}
		if got := world.events[len(world.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %v, want %v", got, wantTail)
		}
	})
}

func TestSpellGrantToPlayer4FB550SignedClampsAndOverride(t *testing.T) {
	for _, test := range []struct {
		name     string
		initial  uint32
		override int32
		quest    []int32
		want     uint32
	}{
		{name: "wrap", initial: math.MaxUint32, want: 0},
		{name: "signed high bit", initial: math.MaxInt32, want: math.MaxInt32 + 1},
		{name: "signed positive clamp", initial: 6, want: 5},
		{name: "quest clamp", initial: 4, quest: []int32{0, 1}, want: 3},
		{name: "override after clamp", initial: 6, override: -1, quest: []int32{0, 1}, want: math.MaxUint32},
	} {
		t.Run(test.name, func(t *testing.T) {
			world := newSpellGrantTestWorld4FB550()
			world.unitArg.update.player.levels[10] = test.initial
			world.game[spellGrantQuestFlag4FB550] = test.quest
			if got := spellGrantToPlayer4FB550(10, 0, 0, test.override, world.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if got := world.unitArg.update.player.levels[10]; got != test.want {
				t.Fatalf("level = %#x, want %#x", got, test.want)
			}
			wantReport := fmt.Sprintf("report:selected-unit:10:0:0")
			if got := world.events[len(world.events)-1]; got != wantReport {
				t.Fatalf("last event = %q, want %q", got, wantReport)
			}
		})
	}
}

func TestSpellGrantToPlayer4FB550FamilyPreservesOriginalAsymmetry(t *testing.T) {
	world := newSpellGrantTestWorld4FB550()
	selected := world.unitArg.update.player
	selected.levels[10] = 3
	selected.levels[20] = 1
	world.flags[spellGrantTestFlagKey4FB550{spell: 10, mask: spellGrantFamilySourceA4FB550}] = []int32{7}
	world.flags[spellGrantTestFlagKey4FB550{spell: 10, mask: spellGrantFamilyMemberA4FB550}] = []int32{1}
	world.flags[spellGrantTestFlagKey4FB550{spell: 20, mask: spellGrantFamilyMemberA4FB550}] = []int32{1}
	world.valid[10] = 1
	world.valid[20] = 1
	world.game[spellGrantQuestFlag4FB550] = []int32{0, 0, 1, 1}
	replacement := &spellGrantTestObject4FB550{name: "replacement", class: 4}
	world.afterEvent["award:12345678:10:2"] = func() {
		world.unitArg = replacement
	}

	if got := spellGrantToPlayer4FB550(10, 0, 0, 0, world.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := selected.levels[10]; got != 3 {
		t.Fatalf("selected level = %d, want Quest-clamped 3", got)
	}
	if got := selected.levels[20]; got != 2 {
		t.Fatalf("member level = %d, want 2", got)
	}

	awards := make([]string, 0, 3)
	for _, event := range world.events {
		if len(event) >= len("award:") && event[:len("award:")] == "award:" {
			awards = append(awards, event)
		}
		if event == "flags:10:00004000=0" || event == "flags:10:00010000=0" {
			t.Fatalf("lower-priority family probe was not short-circuited: %q", event)
		}
	}
	wantAwards := []string{
		"award:12345678:10:4",
		"award:12345678:10:3",
		"award:12345678:10:2",
	}
	if !reflect.DeepEqual(awards, wantAwards) {
		t.Fatalf("awards = %v, want %v", awards, wantAwards)
	}
	if got := world.gameCalls[spellGrantQuestFlag4FB550]; got != 4 {
		t.Fatalf("Quest checks = %d, want single-level gate, selected, and two valid members", got)
	}
	wantTail := []string{
		"arg:replacement",
		"game:00000800=0",
		"report:replacement:10:0:0",
	}
	if got := world.events[len(world.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail = %v, want %v", got, wantTail)
	}
}

func TestSpellGrantToPlayer4FB550NotifyBroadcastUsesExactPredicates(t *testing.T) {
	world := newSpellGrantTestWorld4FB550()
	selected := world.unitArg.update.player
	otherUnit := &spellGrantTestObject4FB550{name: "other-unit", class: 4}
	other := &spellGrantTestPlayer4FB550{name: "other-player", unit: otherUnit}
	nilUnit := &spellGrantTestPlayer4FB550{name: "nil-player"}
	world.players = []*spellGrantTestPlayer4FB550{selected, other, nilUnit}
	world.game[spellGrantSoloFlag4FB550] = []int32{-7, 1}
	world.flags[spellGrantTestFlagKey4FB550{spell: 10, mask: spellGrantSoloSuppressMask4FB550}] = []int32{2}

	if got := spellGrantToPlayer4FB550(10, 2, 1, 0, world.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	rewards := make([]string, 0, 2)
	for _, event := range world.events {
		if len(event) >= len("reward:") && event[:len("reward:")] == "reward:" {
			rewards = append(rewards, event)
		}
	}
	wantRewards := []string{
		"reward:selected-unit:0:selected-unit:10",
		"reward:other-unit:0:selected-unit:10",
	}
	if !reflect.DeepEqual(rewards, wantRewards) {
		t.Fatalf("rewards = %v, want %v", rewards, wantRewards)
	}
	if got := world.events[len(world.events)-1]; got != "report:selected-unit:10:2:1" {
		t.Fatalf("report = %q", got)
	}
	for _, event := range world.events {
		if event == "trade:selected-update=trade" {
			t.Fatalf("notify=2 incorrectly satisfied exact shop gate")
		}
	}
}

func TestSpellGrantToPlayer4FB550GlyphQuestAndShopGates(t *testing.T) {
	world := newSpellGrantTestWorld4FB550()
	world.game[spellGrantQuestFlag4FB550] = []int32{0, 0, 1}
	world.game[spellGrantSoloFlag4FB550] = []int32{1, 1}
	world.unitArg.update.player.notifyField = 0

	if got := spellGrantToPlayer4FB550(34, 1, 1, 0, world.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	for _, event := range world.events {
		if event == "flags:34:00015000=0" {
			t.Fatalf("Glyph should short-circuit the Solo suppress-mask probe")
		}
		if len(event) >= len("reward:") && event[:len("reward:")] == "reward:" {
			t.Fatalf("Quest notify field zero emitted %q", event)
		}
	}
	wantTail := []string{
		"game:00000800=1",
		"trade:selected-update=trade",
		"shop-exit:trade",
		"report:selected-unit:34:1:1",
	}
	if got := world.events[len(world.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail = %v, want %v", got, wantTail)
	}
}

func TestSpellGrantToPlayer4FB550LivePlayerReloads(t *testing.T) {
	world := newSpellGrantTestWorld4FB550()
	first := world.unitArg.update.player
	second := &spellGrantTestPlayer4FB550{
		name:        "replacement-player",
		protection:  0x87654321,
		notifyField: 1,
		unit:        world.unitArg,
	}
	first.levels[10] = 1
	second.levels[10] = 4
	world.afterEvent["level:selected-player:10=00000001"] = func() {
		world.unitArg.update.player = second
	}

	if got := spellGrantToPlayer4FB550(10, 0, 0, 0, world.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := first.levels[10]; got != 1 {
		t.Fatalf("stale player was written: level = %d", got)
	}
	if got := second.levels[10]; got != 5 {
		t.Fatalf("replacement level = %d, want 5", got)
	}
	if !containsSpellGrantEvent4FB550(world.events, "award:87654321:10:5") {
		t.Fatalf("live protection award missing: %v", world.events)
	}
}

func containsSpellGrantEvent4FB550(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func newSpellGrantFullPathWorld4FB550() *spellGrantTestWorld4FB550 {
	world := newSpellGrantTestWorld4FB550()
	world.unitArg.update.player.levels[10] = 1
	world.unitArg.update.player.levels[20] = 1
	world.flags[spellGrantTestFlagKey4FB550{spell: 10, mask: spellGrantFamilySourceA4FB550}] = []int32{1}
	world.flags[spellGrantTestFlagKey4FB550{spell: 20, mask: spellGrantFamilyMemberA4FB550}] = []int32{1}
	world.valid[20] = 1
	world.game[spellGrantSoloFlag4FB550] = []int32{0, 1}
	otherUnit := &spellGrantTestObject4FB550{name: "other-unit", class: 4}
	world.players = append(world.players, &spellGrantTestPlayer4FB550{name: "other-player", unit: otherUnit})
	return world
}

func TestSpellGrantToPlayer4FB550EveryObservableFaultPrefix(t *testing.T) {
	complete := newSpellGrantFullPathWorld4FB550()
	if got := spellGrantToPlayer4FB550(10, 1, 1, 0, complete.hooks()); got != 1 {
		t.Fatalf("complete result = %d, want 1", got)
	}
	want := append([]string(nil), complete.events...)

	for fault := 1; fault <= len(want); fault++ {
		t.Run(fmt.Sprintf("event-%03d", fault), func(t *testing.T) {
			world := newSpellGrantFullPathWorld4FB550()
			world.faultAt = fault
			defer func() {
				if got := recover(); got != want[fault-1] {
					t.Fatalf("panic = %v, want %q", got, want[fault-1])
				}
				if prefix := want[:fault]; !reflect.DeepEqual(world.events, prefix) {
					t.Fatalf("events = %v, want prefix %v", world.events, prefix)
				}
			}()
			spellGrantToPlayer4FB550(10, 1, 1, 0, world.hooks())
		})
	}
}
