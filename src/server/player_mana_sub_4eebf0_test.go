package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerManaSubTestPlayer4EEBF0 struct {
	name  string
	token uint32
}

type playerManaSubTestUpdate4EEBF0 struct {
	name     string
	current  uint16
	previous uint16
	player   *playerManaSubTestPlayer4EEBF0
}

type playerManaSubTestObject4EEBF0 struct {
	name   string
	class  uint8
	update *playerManaSubTestUpdate4EEBF0
}

type playerManaSubTestWorld4EEBF0 struct {
	unit              *playerManaSubTestObject4EEBF0
	unitResult        string
	engineGod         bool
	amount            int32
	protectValue      string
	events            []string
	faultAt           int
	after             map[string]func()
	afterCurrentStore func()
	protectedBy       int16
	protectedFor      uint32
}

func playerManaSubObjectName4EEBF0(obj *playerManaSubTestObject4EEBF0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerManaSubUpdateName4EEBF0(update *playerManaSubTestUpdate4EEBF0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func playerManaSubPlayerName4EEBF0(player *playerManaSubTestPlayer4EEBF0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *playerManaSubTestWorld4EEBF0) record(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerManaSubTestWorld4EEBF0) hooks() playerManaSubHooks4EEBF0[
	*playerManaSubTestObject4EEBF0,
	*playerManaSubTestUpdate4EEBF0,
	*playerManaSubTestPlayer4EEBF0,
	string,
] {
	return playerManaSubHooks4EEBF0[
		*playerManaSubTestObject4EEBF0,
		*playerManaSubTestUpdate4EEBF0,
		*playerManaSubTestPlayer4EEBF0,
		string,
	]{
		loadUnitArg: func() (*playerManaSubTestObject4EEBF0, string) {
			w.record("arg:%s:%s", playerManaSubObjectName4EEBF0(w.unit), w.unitResult)
			return w.unit, w.unitResult
		},
		loadClassLow: func(unit *playerManaSubTestObject4EEBF0) uint8 {
			w.record("class:%s:%02x", playerManaSubObjectName4EEBF0(unit), unit.class)
			return unit.class
		},
		loadEngineGodMode: func() bool {
			godMode := w.engineGod
			w.record("god-mode:%t", godMode)
			return godMode
		},
		loadUpdateData: func(unit *playerManaSubTestObject4EEBF0) (*playerManaSubTestUpdate4EEBF0, string) {
			update := unit.update
			result := "update-result:" + playerManaSubUpdateName4EEBF0(update)
			w.record("update:%s=%s", playerManaSubObjectName4EEBF0(unit), playerManaSubUpdateName4EEBF0(update))
			return update, result
		},
		loadCurrent: func(update *playerManaSubTestUpdate4EEBF0) uint16 {
			name := playerManaSubUpdateName4EEBF0(update)
			if update == nil {
				w.record("current:%s", name)
				panic("current:nil")
			}
			current := update.current
			w.record("current:%s=%d", name, current)
			return current
		},
		loadAmountArg: func() int32 {
			amount := w.amount
			w.record("amount:%d", amount)
			return amount
		},
		storePrevious: func(update *playerManaSubTestUpdate4EEBF0, value uint16) {
			w.record("store-previous:%s:%d", playerManaSubUpdateName4EEBF0(update), value)
			update.previous = value
		},
		storeCurrent: func(update *playerManaSubTestUpdate4EEBF0, value uint16) {
			w.record("store-current:%s:%d", playerManaSubUpdateName4EEBF0(update), value)
			update.current = value
			if w.afterCurrentStore != nil {
				w.afterCurrentStore()
			}
		},
		loadPlayer: func(update *playerManaSubTestUpdate4EEBF0) *playerManaSubTestPlayer4EEBF0 {
			name := playerManaSubUpdateName4EEBF0(update)
			if update == nil {
				w.record("player:%s", name)
				panic("player:nil-update")
			}
			player := update.player
			w.record("player:%s=%s", name, playerManaSubPlayerName4EEBF0(player))
			return player
		},
		loadProtection: func(player *playerManaSubTestPlayer4EEBF0) uint32 {
			name := playerManaSubPlayerName4EEBF0(player)
			if player == nil {
				w.record("protection:%s", name)
				panic("protection:nil")
			}
			token := player.token
			w.record("protection:%s=%08x", name, token)
			return token
		},
		protectMana: func(token uint32, delta int16) string {
			w.protectedFor, w.protectedBy = token, delta
			w.record("protect:%08x:%d", token, delta)
			return w.protectValue
		},
	}
}

func newPlayerManaSubTestWorld4EEBF0() *playerManaSubTestWorld4EEBF0 {
	player := &playerManaSubTestPlayer4EEBF0{name: "player-1", token: 0x12345678}
	update := &playerManaSubTestUpdate4EEBF0{
		name:     "update-1",
		current:  100,
		previous: 7,
		player:   player,
	}
	return &playerManaSubTestWorld4EEBF0{
		unit: &playerManaSubTestObject4EEBF0{
			name:   "unit",
			class:  0x84,
			update: update,
		},
		unitResult:   "unit-result",
		amount:       10,
		protectValue: "protect-result",
	}
}

func playerManaSubFullTrace4EEBF0() (*playerManaSubTestWorld4EEBF0, []string) {
	w := newPlayerManaSubTestWorld4EEBF0()
	update1 := w.unit.update
	update2 := &playerManaSubTestUpdate4EEBF0{
		name:     "update-2",
		current:  100,
		previous: 9,
		player:   update1.player,
	}
	player2 := &playerManaSubTestPlayer4EEBF0{name: "player-2", token: 0x89abcdef}
	w.after = map[string]func(){
		"god-mode:false": func() {
			w.unit.update = update2
		},
		"update:unit=update-2": func() {
			w.engineGod = true
		},
		"current:update-2=100": func() {
			if update2.previous == 9 {
				w.amount = 10
			}
		},
		"store-previous:update-2:100": func() {
			update2.current = 500
		},
		"current:update-2=5": func() {
			update2.player = update1.player
		},
		"player:update-2=player-1": func() {
			update1.player.token = 0xfedcba98
			update2.player = player2
		},
	}
	w.afterCurrentStore = func() {
		update2.current = 5
		update2.player = player2
	}
	return w, []string{
		"arg:unit:unit-result",
		"class:unit:84",
		"god-mode:false",
		"update:unit=update-2",
		"current:update-2=100",
		"amount:10",
		"store-previous:update-2:100",
		"store-current:update-2:90",
		"current:update-2=5",
		"player:update-2=player-1",
		"protection:player-1=fedcba98",
		"protect:fedcba98:-5",
	}
}

func TestPlayerManaSub4EEBF0ExactTraceCachingAndLiveReload(t *testing.T) {
	w, want := playerManaSubFullTrace4EEBF0()
	if got := playerManaSub4EEBF0(w.hooks()); got != "protect-result" {
		t.Fatalf("result = %q, want protect-result", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, want)
	}
	cachedUpdate := w.unit.update
	if cachedUpdate.previous != 100 || cachedUpdate.current != 5 {
		t.Fatalf("cached update previous/current = %d/%d, want 100/5", cachedUpdate.previous, cachedUpdate.current)
	}
	if w.protectedFor != 0xfedcba98 || w.protectedBy != -5 {
		t.Fatalf("protection = (%#x,%d), want (0xfedcba98,-5)", w.protectedFor, w.protectedBy)
	}
}

func TestPlayerManaSub4EEBF0AllFullPathFaultPrefixes(t *testing.T) {
	_, want := playerManaSubFullTrace4EEBF0()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w, _ := playerManaSubFullTrace4EEBF0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %q, want %q", w.events, want[:faultAt])
				}
			}()
			playerManaSub4EEBF0(w.hooks())
		})
	}
}

func TestPlayerManaSub4EEBF0EntryAndGodModeResults(t *testing.T) {
	w := newPlayerManaSubTestWorld4EEBF0()
	w.unit = nil
	if got := playerManaSub4EEBF0(w.hooks()); got != "unit-result" {
		t.Fatalf("nil result = %q, want unit-result", got)
	}
	if !reflect.DeepEqual(w.events, []string{"arg:nil:unit-result"}) {
		t.Fatalf("nil events = %q", w.events)
	}

	w = newPlayerManaSubTestWorld4EEBF0()
	w.unit.class = 0xfb
	if got := playerManaSub4EEBF0(w.hooks()); got != "unit-result" {
		t.Fatalf("non-Player result = %q, want unit-result", got)
	}
	if !reflect.DeepEqual(w.events, []string{"arg:unit:unit-result", "class:unit:fb"}) {
		t.Fatalf("non-Player events = %q", w.events)
	}

	w = newPlayerManaSubTestWorld4EEBF0()
	w.engineGod = true
	w.unit.update = nil
	if got := playerManaSub4EEBF0(w.hooks()); got != "update-result:nil" {
		t.Fatalf("GodMode result = %q, want update-result:nil", got)
	}
	want := []string{
		"arg:unit:unit-result",
		"class:unit:84",
		"god-mode:true",
		"update:unit=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("GodMode events = %q, want %q", w.events, want)
	}
}

func TestPlayerManaSub4EEBF0SignedAmountAndLowWordRules(t *testing.T) {
	tests := []struct {
		name        string
		current     uint16
		amount      int32
		wantCurrent uint16
		wantDelta   int16
	}{
		{name: "ordinary", current: 100, amount: 10, wantCurrent: 90, wantDelta: -10},
		{name: "new below amount", current: 15, amount: 10, wantCurrent: 5, wantDelta: -5},
		{name: "equal", current: 10, amount: 10, wantCurrent: 0, wantDelta: 0},
		{name: "insufficient", current: 5, amount: 10, wantCurrent: 0, wantDelta: 0},
		{name: "negative adds", current: 100, amount: -1, wantCurrent: 101, wantDelta: 1},
		{name: "negative low word zero", current: 100, amount: -65536, wantCurrent: 100, wantDelta: 0},
		{name: "above word", current: 65535, amount: 65536, wantCurrent: 0, wantDelta: 0},
		{name: "signed minimum wraps", current: 65535, amount: math.MinInt32, wantCurrent: 65535, wantDelta: 0},
		{name: "signed maximum saturates", current: 65535, amount: math.MaxInt32, wantCurrent: 0, wantDelta: 0},
		{name: "second comparison uses new mana", current: 65535, amount: 32768, wantCurrent: 32767, wantDelta: -32767},
		{name: "zero amount and maximum mana", current: 65535, amount: 0, wantCurrent: 65535, wantDelta: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newPlayerManaSubTestWorld4EEBF0()
			w.unit.update.current = tc.current
			w.amount = tc.amount
			if got := playerManaSub4EEBF0(w.hooks()); got != "protect-result" {
				t.Fatalf("result = %q", got)
			}
			if w.unit.update.previous != tc.current || w.unit.update.current != tc.wantCurrent {
				t.Fatalf("previous/current = %d/%d, want %d/%d", w.unit.update.previous, w.unit.update.current, tc.current, tc.wantCurrent)
			}
			if w.protectedBy != tc.wantDelta {
				t.Fatalf("delta = %d, want %d", w.protectedBy, tc.wantDelta)
			}
		})
	}
}

func TestPlayerManaSub4EEBF0HasNoUpdateOrPlayerNilGuards(t *testing.T) {
	t.Run("nil update", func(t *testing.T) {
		w := newPlayerManaSubTestWorld4EEBF0()
		w.unit.update = nil
		defer func() {
			if got := recover(); got != "current:nil" {
				t.Fatalf("panic = %v, want current:nil", got)
			}
			want := []string{
				"arg:unit:unit-result", "class:unit:84", "god-mode:false",
				"update:unit=nil", "current:nil",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		playerManaSub4EEBF0(w.hooks())
	})

	t.Run("nil player after stores", func(t *testing.T) {
		w := newPlayerManaSubTestWorld4EEBF0()
		w.unit.update.current = 15
		w.unit.update.player = nil
		defer func() {
			if got := recover(); got != "protection:nil" {
				t.Fatalf("panic = %v, want protection:nil", got)
			}
			if w.unit.update.previous != 15 || w.unit.update.current != 5 {
				t.Fatalf("mana before fault = %d/%d, want 15/5", w.unit.update.previous, w.unit.update.current)
			}
		}()
		playerManaSub4EEBF0(w.hooks())
	})
}
