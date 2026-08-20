package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerManaRefreshTestPlayer4EECF0 struct {
	name  string
	token uint32
}

type playerManaRefreshTestUpdate4EECF0 struct {
	name     string
	current  uint16
	previous uint16
	maximum  uint16
	player   *playerManaRefreshTestPlayer4EECF0
}

type playerManaRefreshTestObject4EECF0 struct {
	name   string
	class  uint32
	update *playerManaRefreshTestUpdate4EECF0
}

type playerManaRefreshTestWorld4EECF0 struct {
	unit          *playerManaRefreshTestObject4EECF0
	protectResult string
	events        []string
	faultAt       int
	after         map[string]func()
	protectedFor  uint32
	protectedBy   int16
}

func playerManaRefreshObjectName4EECF0(obj *playerManaRefreshTestObject4EECF0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerManaRefreshUpdateName4EECF0(update *playerManaRefreshTestUpdate4EECF0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func playerManaRefreshPlayerName4EECF0(player *playerManaRefreshTestPlayer4EECF0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *playerManaRefreshTestWorld4EECF0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerManaRefreshTestWorld4EECF0) finish(event string) {
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerManaRefreshTestWorld4EECF0) hooks() playerManaRefreshHooks4EECF0[
	*playerManaRefreshTestObject4EECF0,
	*playerManaRefreshTestUpdate4EECF0,
	*playerManaRefreshTestPlayer4EECF0,
	string,
] {
	return playerManaRefreshHooks4EECF0[
		*playerManaRefreshTestObject4EECF0,
		*playerManaRefreshTestUpdate4EECF0,
		*playerManaRefreshTestPlayer4EECF0,
		string,
	]{
		loadUnitArg: func() (*playerManaRefreshTestObject4EECF0, string) {
			unit := w.unit
			result := "unit-result:" + playerManaRefreshObjectName4EECF0(unit)
			event := "arg:" + playerManaRefreshObjectName4EECF0(unit)
			w.record(event)
			w.finish(event)
			return unit, result
		},
		loadClassLow: func(unit *playerManaRefreshTestObject4EECF0) uint8 {
			classLow := uint8(unit.class)
			event := fmt.Sprintf("class-low:%s=%#02x", playerManaRefreshObjectName4EECF0(unit), classLow)
			w.record(event)
			w.finish(event)
			return classLow
		},
		loadUpdateData: func(unit *playerManaRefreshTestObject4EECF0) *playerManaRefreshTestUpdate4EECF0 {
			update := unit.update
			event := "update:" + playerManaRefreshObjectName4EECF0(unit) + "=" + playerManaRefreshUpdateName4EECF0(update)
			w.record(event)
			w.finish(event)
			return update
		},
		loadCurrent: func(update *playerManaRefreshTestUpdate4EECF0) uint16 {
			name := playerManaRefreshUpdateName4EECF0(update)
			if update == nil {
				event := "current:" + name
				w.record(event)
				panic("current:nil")
			}
			current := update.current
			event := fmt.Sprintf("current:%s=%#04x", name, current)
			w.record(event)
			w.finish(event)
			return current
		},
		loadPlayer: func(update *playerManaRefreshTestUpdate4EECF0) *playerManaRefreshTestPlayer4EECF0 {
			name := playerManaRefreshUpdateName4EECF0(update)
			if update == nil {
				event := "player:" + name
				w.record(event)
				panic("player:nil-update")
			}
			player := update.player
			event := "player:" + name + "=" + playerManaRefreshPlayerName4EECF0(player)
			w.record(event)
			w.finish(event)
			return player
		},
		storePrevious: func(update *playerManaRefreshTestUpdate4EECF0, current uint16) {
			event := fmt.Sprintf("store-previous:%s=%#04x", playerManaRefreshUpdateName4EECF0(update), current)
			w.record(event)
			if update == nil {
				panic("store-previous:nil")
			}
			update.previous = current
			w.finish(event)
		},
		loadMaximum: func(update *playerManaRefreshTestUpdate4EECF0) uint16 {
			name := playerManaRefreshUpdateName4EECF0(update)
			if update == nil {
				event := "maximum:" + name
				w.record(event)
				panic("maximum:nil")
			}
			maximum := update.maximum
			event := fmt.Sprintf("maximum:%s=%#04x", name, maximum)
			w.record(event)
			w.finish(event)
			return maximum
		},
		storeCurrent: func(update *playerManaRefreshTestUpdate4EECF0, maximum uint16) {
			event := fmt.Sprintf("store-current:%s=%#04x", playerManaRefreshUpdateName4EECF0(update), maximum)
			w.record(event)
			if update == nil {
				panic("store-current:nil")
			}
			update.current = maximum
			w.finish(event)
		},
		loadProtection: func(player *playerManaRefreshTestPlayer4EECF0) uint32 {
			name := playerManaRefreshPlayerName4EECF0(player)
			if player == nil {
				event := "protection:" + name
				w.record(event)
				panic("protection:nil")
			}
			token := player.token
			event := fmt.Sprintf("protection:%s=%08x", name, token)
			w.record(event)
			w.finish(event)
			return token
		},
		protectMana: func(token uint32, maximum int16) string {
			result := w.protectResult
			event := fmt.Sprintf("protect:%08x:%d", token, maximum)
			w.record(event)
			w.protectedFor, w.protectedBy = token, maximum
			w.finish(event)
			return result
		},
	}
}

func newPlayerManaRefreshTestWorld4EECF0() *playerManaRefreshTestWorld4EECF0 {
	player := &playerManaRefreshTestPlayer4EECF0{name: "player", token: 0x12345678}
	update := &playerManaRefreshTestUpdate4EECF0{
		name:     "update",
		current:  0x1234,
		previous: 0x1111,
		maximum:  0x4321,
		player:   player,
	}
	return &playerManaRefreshTestWorld4EECF0{
		unit:          &playerManaRefreshTestObject4EECF0{name: "unit", class: 0x04, update: update},
		protectResult: "protect-result",
		after:         make(map[string]func()),
	}
}

func playerManaRefreshFullEvents4EECF0() []string {
	return []string{
		"arg:unit",
		"class-low:unit=0x04",
		"update:unit=update",
		"current:update=0x1234",
		"player:update=player",
		"store-previous:update=0x1234",
		"maximum:update=0x4321",
		"store-current:update=0x4321",
		"protection:player=12345678",
		"protect:12345678:17185",
	}
}

func TestPlayerManaRefresh4EECF0EntryGates(t *testing.T) {
	tests := []struct {
		name       string
		unit       *playerManaRefreshTestObject4EECF0
		wantResult string
		wantEvents []string
	}{
		{name: "nil unit", unit: nil, wantResult: "unit-result:nil", wantEvents: []string{"arg:nil"}},
		{name: "other", unit: &playerManaRefreshTestObject4EECF0{name: "unit", class: 0x80000000}, wantResult: "unit-result:unit", wantEvents: []string{"arg:unit", "class-low:unit=0x00"}},
		{name: "upper byte player bit", unit: &playerManaRefreshTestObject4EECF0{name: "unit", class: 0x00000400}, wantResult: "unit-result:unit", wantEvents: []string{"arg:unit", "class-low:unit=0x00"}},
		{name: "monster", unit: &playerManaRefreshTestObject4EECF0{name: "unit", class: 0x40000002}, wantResult: "unit-result:unit", wantEvents: []string{"arg:unit", "class-low:unit=0x02"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerManaRefreshTestWorld4EECF0()
			w.unit = test.unit
			if got := playerManaRefresh4EECF0(w.hooks()); got != test.wantResult {
				t.Fatalf("result = %q, want %q", got, test.wantResult)
			}
			if !reflect.DeepEqual(w.events, test.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, test.wantEvents)
			}
		})
	}
}

func TestPlayerManaRefresh4EECF0StoresAndReturnsProtectionResult(t *testing.T) {
	w := newPlayerManaRefreshTestWorld4EECF0()
	w.unit.class = 0xa5a50006
	update := w.unit.update
	if got := playerManaRefresh4EECF0(w.hooks()); got != "protect-result" {
		t.Fatalf("result = %q, want protection result", got)
	}
	if update.previous != 0x1234 || update.current != 0x4321 || update.maximum != 0x4321 {
		t.Fatalf("mana = previous %#04x current %#04x maximum %#04x", update.previous, update.current, update.maximum)
	}
	if w.protectedFor != 0x12345678 || w.protectedBy != 0x4321 {
		t.Fatalf("protection = token %#08x value %d", w.protectedFor, w.protectedBy)
	}
	want := playerManaRefreshFullEvents4EECF0()
	want[1] = "class-low:unit=0x06"
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestPlayerManaRefresh4EECF0PassesEveryMaximumWordAsSignedLowWord(t *testing.T) {
	for _, maximum := range []uint16{0, 1, 1000, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("%04x", maximum), func(t *testing.T) {
			w := newPlayerManaRefreshTestWorld4EECF0()
			w.unit.update.maximum = maximum
			playerManaRefresh4EECF0(w.hooks())
			if w.unit.update.current != maximum || w.protectedBy != int16(maximum) {
				t.Fatalf("current/protection = %#04x/%d, want %#04x/%d", w.unit.update.current, w.protectedBy, maximum, int16(maximum))
			}
		})
	}
}

func TestPlayerManaRefresh4EECF0CachesLoadedValuesInOriginalOrder(t *testing.T) {
	w := newPlayerManaRefreshTestWorld4EECF0()
	originalUnit := w.unit
	originalUpdate := originalUnit.update
	replacementPlayer := &playerManaRefreshTestPlayer4EECF0{name: "replacement-player", token: 0x11111111}
	finalPlayer := &playerManaRefreshTestPlayer4EECF0{name: "final-player", token: 0x22222222}
	replacementUpdate := &playerManaRefreshTestUpdate4EECF0{
		name:     "replacement",
		current:  0x2345,
		previous: 0x3456,
		maximum:  0x4567,
		player:   originalUpdate.player,
	}
	finalUpdate := &playerManaRefreshTestUpdate4EECF0{name: "final", current: 1, previous: 2, maximum: 3, player: finalPlayer}
	w.after["arg:unit"] = func() {
		w.unit = &playerManaRefreshTestObject4EECF0{name: "new-unit", class: 0, update: finalUpdate}
	}
	w.after["class-low:unit=0x04"] = func() {
		originalUnit.class = 0
		originalUnit.update = replacementUpdate
	}
	w.after["update:unit=replacement"] = func() {
		originalUnit.update = finalUpdate
	}
	w.after["current:replacement=0x2345"] = func() {
		replacementUpdate.current = 0xaaaa
		replacementUpdate.player = replacementPlayer
	}
	w.after["player:replacement=replacement-player"] = func() {
		replacementUpdate.player = finalPlayer
	}
	w.after["store-previous:replacement=0x2345"] = func() {
		replacementUpdate.maximum = 0x5678
	}
	w.after["maximum:replacement=0x5678"] = func() {
		replacementUpdate.maximum = 0x9999
	}
	w.after["store-current:replacement=0x5678"] = func() {
		replacementPlayer.token = 0xabcdef01
	}
	w.after["protection:replacement-player=abcdef01"] = func() {
		replacementPlayer.token = 0x33333333
	}

	if got := playerManaRefresh4EECF0(w.hooks()); got != "protect-result" {
		t.Fatalf("result = %q, want protection result", got)
	}
	if replacementUpdate.previous != 0x2345 || replacementUpdate.current != 0x5678 || replacementUpdate.maximum != 0x9999 {
		t.Fatalf("replacement mana = previous %#04x current %#04x maximum %#04x", replacementUpdate.previous, replacementUpdate.current, replacementUpdate.maximum)
	}
	if w.protectedFor != 0xabcdef01 || w.protectedBy != int16(0x5678) {
		t.Fatalf("protection = token %#08x value %d", w.protectedFor, w.protectedBy)
	}
	if originalUpdate.previous != 0x1111 || originalUpdate.current != 0x1234 || finalUpdate.current != 1 {
		t.Fatalf("uncached records changed: original=%+v final=%+v", *originalUpdate, *finalUpdate)
	}
	want := []string{
		"arg:unit",
		"class-low:unit=0x04",
		"update:unit=replacement",
		"current:replacement=0x2345",
		"player:replacement=replacement-player",
		"store-previous:replacement=0x2345",
		"maximum:replacement=0x5678",
		"store-current:replacement=0x5678",
		"protection:replacement-player=abcdef01",
		"protect:abcdef01:22136",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestPlayerManaRefresh4EECF0NilUpdateFaultsOnCurrent(t *testing.T) {
	w := newPlayerManaRefreshTestWorld4EECF0()
	w.unit.update = nil
	defer func() {
		if got := recover(); got != "current:nil" {
			t.Fatalf("panic = %v, want current:nil", got)
		}
		want := []string{"arg:unit", "class-low:unit=0x04", "update:unit=nil", "current:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	playerManaRefresh4EECF0(w.hooks())
}

func TestPlayerManaRefresh4EECF0NilPlayerFaultsAfterBothStores(t *testing.T) {
	w := newPlayerManaRefreshTestWorld4EECF0()
	update := w.unit.update
	update.player = nil
	defer func() {
		if got := recover(); got != "protection:nil" {
			t.Fatalf("panic = %v, want protection:nil", got)
		}
		want := []string{
			"arg:unit",
			"class-low:unit=0x04",
			"update:unit=update",
			"current:update=0x1234",
			"player:update=nil",
			"store-previous:update=0x1234",
			"maximum:update=0x4321",
			"store-current:update=0x4321",
			"protection:nil",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
		if update.previous != 0x1234 || update.current != 0x4321 {
			t.Fatalf("stores before fault = previous %#04x current %#04x", update.previous, update.current)
		}
	}()
	playerManaRefresh4EECF0(w.hooks())
}

func TestPlayerManaRefresh4EECF0AllFaultPrefixes(t *testing.T) {
	want := playerManaRefreshFullEvents4EECF0()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newPlayerManaRefreshTestWorld4EECF0()
			update := w.unit.update
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
				wantPrevious := uint16(0x1111)
				if faultAt > 6 {
					wantPrevious = 0x1234
				}
				wantCurrent := uint16(0x1234)
				if faultAt > 8 {
					wantCurrent = 0x4321
				}
				if update.previous != wantPrevious || update.current != wantCurrent {
					t.Fatalf("fault state = previous %#04x current %#04x, want %#04x/%#04x", update.previous, update.current, wantPrevious, wantCurrent)
				}
			}()
			playerManaRefresh4EECF0(w.hooks())
		})
	}
}
