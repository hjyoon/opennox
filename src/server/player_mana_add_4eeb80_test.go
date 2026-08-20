package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerManaAddTestPlayer4EEB80 struct {
	name  string
	token uint32
}

type playerManaAddTestUpdate4EEB80 struct {
	name     string
	current  uint16
	previous uint16
	maximum  uint16
	player   *playerManaAddTestPlayer4EEB80
}

type playerManaAddTestObject4EEB80 struct {
	name   string
	class  uint8
	update *playerManaAddTestUpdate4EEB80
}

type playerManaAddTestWorld4EEB80 struct {
	unit          *playerManaAddTestObject4EEB80
	unitResult    uint16
	amount        int32
	repairResult  uint16
	events        []string
	stores        []uint16
	faultAt       int
	after         map[string]func()
	protectedWith uint32
	protectedBy   int16
	repairedWith  uint32
	repairedTo    uint16
}

func playerManaAddObjectName4EEB80(obj *playerManaAddTestObject4EEB80) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerManaAddUpdateName4EEB80(update *playerManaAddTestUpdate4EEB80) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func playerManaAddPlayerName4EEB80(player *playerManaAddTestPlayer4EEB80) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *playerManaAddTestWorld4EEB80) record(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerManaAddTestWorld4EEB80) hooks() playerManaAddHooks4EEB80[
	*playerManaAddTestObject4EEB80,
	*playerManaAddTestUpdate4EEB80,
	*playerManaAddTestPlayer4EEB80,
] {
	return playerManaAddHooks4EEB80[
		*playerManaAddTestObject4EEB80,
		*playerManaAddTestUpdate4EEB80,
		*playerManaAddTestPlayer4EEB80,
	]{
		loadUnitArg: func() (*playerManaAddTestObject4EEB80, uint16) {
			unit, result := w.unit, w.unitResult
			w.record("arg:%s:%04x", playerManaAddObjectName4EEB80(unit), result)
			return unit, result
		},
		loadClassLow: func(unit *playerManaAddTestObject4EEB80) uint8 {
			class := unit.class
			w.record("class:%s", playerManaAddObjectName4EEB80(unit))
			return class
		},
		loadUpdateData: func(unit *playerManaAddTestObject4EEB80) *playerManaAddTestUpdate4EEB80 {
			update := unit.update
			w.record("update:%s=%s", playerManaAddObjectName4EEB80(unit), playerManaAddUpdateName4EEB80(update))
			return update
		},
		loadAmountArg: func() int32 {
			amount := w.amount
			w.record("amount:%d", amount)
			return amount
		},
		loadCurrent: func(update *playerManaAddTestUpdate4EEB80) uint16 {
			name := playerManaAddUpdateName4EEB80(update)
			if update == nil {
				w.record("current:%s", name)
				panic("current:nil")
			}
			current := update.current
			w.record("current:%s=%d", name, current)
			return current
		},
		loadMaximum: func(update *playerManaAddTestUpdate4EEB80) uint16 {
			name := playerManaAddUpdateName4EEB80(update)
			if update == nil {
				w.record("maximum:%s", name)
				panic("maximum:nil")
			}
			maximum := update.maximum
			w.record("maximum:%s=%d", name, maximum)
			return maximum
		},
		storePrevious: func(update *playerManaAddTestUpdate4EEB80, value uint16) {
			w.record("store-previous:%s:%d", playerManaAddUpdateName4EEB80(update), value)
			update.previous = value
		},
		storeCurrent: func(update *playerManaAddTestUpdate4EEB80, value uint16) {
			w.record("store-current:%s:%d", playerManaAddUpdateName4EEB80(update), value)
			update.current = value
			w.stores = append(w.stores, value)
		},
		loadPlayer: func(update *playerManaAddTestUpdate4EEB80) *playerManaAddTestPlayer4EEB80 {
			name := playerManaAddUpdateName4EEB80(update)
			if update == nil {
				w.record("player:%s", name)
				panic("player:nil-update")
			}
			player := update.player
			w.record("player:%s=%s", name, playerManaAddPlayerName4EEB80(player))
			return player
		},
		loadProtection: func(player *playerManaAddTestPlayer4EEB80) uint32 {
			name := playerManaAddPlayerName4EEB80(player)
			if player == nil {
				w.record("protection:%s", name)
				panic("protection:nil")
			}
			token := player.token
			w.record("protection:%s=%08x", name, token)
			return token
		},
		protectMana: func(token uint32, amount int16) {
			w.protectedWith, w.protectedBy = token, amount
			w.record("protect-mana:%08x:%d", token, amount)
		},
		protectPlayerHPMana: func(token uint32, maximum uint16) uint16 {
			w.repairedWith, w.repairedTo = token, maximum
			w.record("repair-mana:%08x:%d", token, maximum)
			return w.repairResult
		},
	}
}

func newPlayerManaAddTestWorld4EEB80() *playerManaAddTestWorld4EEB80 {
	player := &playerManaAddTestPlayer4EEB80{name: "player-1", token: 0x12345678}
	update := &playerManaAddTestUpdate4EEB80{
		name:     "update-1",
		current:  100,
		previous: 7,
		maximum:  120,
		player:   player,
	}
	return &playerManaAddTestWorld4EEB80{
		unit: &playerManaAddTestObject4EEB80{
			name:   "unit",
			class:  0x84,
			update: update,
		},
		unitResult:   0xabcd,
		amount:       30,
		repairResult: 0xbeef,
	}
}

func playerManaAddFullTrace4EEB80() (*playerManaAddTestWorld4EEB80, []string) {
	w := newPlayerManaAddTestWorld4EEB80()
	update := w.unit.update
	player2 := &playerManaAddTestPlayer4EEB80{name: "player-2", token: 0x89abcdef}
	w.after = map[string]func(){
		"amount:30": func() {
			w.unit.update = &playerManaAddTestUpdate4EEB80{name: "replacement"}
		},
		"store-previous:update-1:100": func() {
			update.maximum = 125
		},
		"protect-mana:12345678:30": func() {
			update.current = 91
			update.maximum = 90
			update.player = player2
		},
	}
	return w, []string{
		"arg:unit:abcd",
		"class:unit",
		"update:unit=update-1",
		"amount:30",
		"current:update-1=100",
		"maximum:update-1=120",
		"store-previous:update-1:100",
		"store-current:update-1:130",
		"store-current:update-1:120",
		"player:update-1=player-1",
		"protection:player-1=12345678",
		"protect-mana:12345678:30",
		"maximum:update-1=90",
		"current:update-1=91",
		"player:update-1=player-2",
		"protection:player-2=89abcdef",
		"repair-mana:89abcdef:90",
	}
}

func TestPlayerManaAdd4EEB80ExactTraceAndLiveReloads(t *testing.T) {
	w, want := playerManaAddFullTrace4EEB80()
	update := w.unit.update
	if got := playerManaAdd4EEB80(w.hooks()); got != 0xbeef {
		t.Fatalf("result = %#x, want 0xbeef", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, want)
	}
	if update.previous != 100 || update.current != 91 || update.maximum != 90 {
		t.Fatalf("cached update state = previous/current/maximum %d/%d/%d", update.previous, update.current, update.maximum)
	}
	if !reflect.DeepEqual(w.stores, []uint16{130, 120}) {
		t.Fatalf("initial stores = %v, want [130 120]", w.stores)
	}
	if w.protectedWith != 0x12345678 || w.protectedBy != 30 {
		t.Fatalf("protect args = (%#x,%d)", w.protectedWith, w.protectedBy)
	}
	if w.repairedWith != 0x89abcdef || w.repairedTo != 90 {
		t.Fatalf("repair args = (%#x,%d)", w.repairedWith, w.repairedTo)
	}
}

func TestPlayerManaAdd4EEB80AllFullPathFaultPrefixes(t *testing.T) {
	_, want := playerManaAddFullTrace4EEB80()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w, _ := playerManaAddFullTrace4EEB80()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %q, want %q", w.events, want[:faultAt])
				}
			}()
			playerManaAdd4EEB80(w.hooks())
		})
	}
}

func TestPlayerManaAdd4EEB80EntryGatesReturnOriginalAX(t *testing.T) {
	w := newPlayerManaAddTestWorld4EEB80()
	w.unit = nil
	w.unitResult = 0
	if got := playerManaAdd4EEB80(w.hooks()); got != 0 {
		t.Fatalf("nil result = %#x, want 0", got)
	}
	if !reflect.DeepEqual(w.events, []string{"arg:nil:0000"}) {
		t.Fatalf("nil events = %q", w.events)
	}

	w = newPlayerManaAddTestWorld4EEB80()
	w.unit.class = 0xfb
	w.unitResult = 0x5678
	if got := playerManaAdd4EEB80(w.hooks()); got != 0x5678 {
		t.Fatalf("non-Player result = %#x, want 0x5678", got)
	}
	if want := []string{"arg:unit:5678", "class:unit"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("non-Player events = %q, want %q", w.events, want)
	}
}

func TestPlayerManaAdd4EEB80WrappingClampAndSignedProtection(t *testing.T) {
	tests := []struct {
		name       string
		current    uint16
		maximum    uint16
		amount     int32
		want       uint16
		wantStores []uint16
		wantDelta  int16
	}{
		{name: "below maximum", current: 100, maximum: 200, amount: 30, want: 130, wantStores: []uint16{130}, wantDelta: 30},
		{name: "unsigned clamp", current: 190, maximum: 200, amount: 20, want: 200, wantStores: []uint16{210, 200}, wantDelta: 20},
		{name: "low word wraps", current: 65530, maximum: 100, amount: 10, want: 4, wantStores: []uint16{4}, wantDelta: 10},
		{name: "negative wraps high", current: 1, maximum: 65535, amount: -2, want: 65535, wantStores: []uint16{65535}, wantDelta: -2},
		{name: "whole slot only affects low word", current: 7, maximum: 100, amount: 0x10001, want: 8, wantStores: []uint16{8}, wantDelta: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerManaAddTestWorld4EEB80()
			update := w.unit.update
			update.current, update.maximum = test.current, test.maximum
			w.amount = test.amount
			if got := playerManaAdd4EEB80(w.hooks()); got != test.maximum {
				t.Fatalf("result = %d, want live maximum %d", got, test.maximum)
			}
			if update.previous != test.current || update.current != test.want {
				t.Fatalf("previous/current = %d/%d, want %d/%d", update.previous, update.current, test.current, test.want)
			}
			if !reflect.DeepEqual(w.stores, test.wantStores) {
				t.Fatalf("stores = %v, want %v", w.stores, test.wantStores)
			}
			if w.protectedBy != test.wantDelta || w.protectedWith != update.player.token {
				t.Fatalf("protect = (%#x,%d), want (%#x,%d)", w.protectedWith, w.protectedBy, update.player.token, test.wantDelta)
			}
			if w.repairedWith != 0 {
				t.Fatalf("unexpected repair token %#x", w.repairedWith)
			}
		})
	}
}

func TestPlayerManaAdd4EEB80ReturnsLiveMaximumWithoutRepair(t *testing.T) {
	for _, current := range []uint16{69, 70} {
		t.Run(fmt.Sprintf("current-%d", current), func(t *testing.T) {
			w := newPlayerManaAddTestWorld4EEB80()
			update := w.unit.update
			w.after = map[string]func(){
				"protect-mana:12345678:30": func() {
					update.maximum = 70
					update.current = current
				},
			}
			if got := playerManaAdd4EEB80(w.hooks()); got != 70 {
				t.Fatalf("result = %d, want 70", got)
			}
			if w.repairedWith != 0 {
				t.Fatalf("repair called with token %#x", w.repairedWith)
			}
			wantTail := []string{"maximum:update-1=70", fmt.Sprintf("current:update-1=%d", current)}
			if got := w.events[len(w.events)-2:]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("tail = %q, want %q", got, wantTail)
			}
		})
	}
}

func TestPlayerManaAdd4EEB80NilPointersFaultOnlyAtOriginalDereferences(t *testing.T) {
	t.Run("nil update still loads amount", func(t *testing.T) {
		w := newPlayerManaAddTestWorld4EEB80()
		w.unit.update = nil
		defer func() {
			if got := recover(); got != "current:nil" {
				t.Fatalf("panic = %v, want current:nil", got)
			}
			want := []string{"arg:unit:abcd", "class:unit", "update:unit=nil", "amount:30", "current:nil"}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		playerManaAdd4EEB80(w.hooks())
	})

	t.Run("nil player faults at protection token", func(t *testing.T) {
		w := newPlayerManaAddTestWorld4EEB80()
		w.unit.update.player = nil
		defer func() {
			if got := recover(); got != "protection:nil" {
				t.Fatalf("panic = %v, want protection:nil", got)
			}
			wantTail := []string{"player:update-1=nil", "protection:nil"}
			if got := w.events[len(w.events)-2:]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("tail = %q, want %q", got, wantTail)
			}
		}()
		playerManaAdd4EEB80(w.hooks())
	})
}
