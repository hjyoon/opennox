package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerSetMaxManaTestUpdate4EECD0 struct {
	name    string
	maximum uint16
}

type playerSetMaxManaTestObject4EECD0 struct {
	name   string
	class  uint32
	update *playerSetMaxManaTestUpdate4EECD0
}

type playerSetMaxManaTestWorld4EECD0 struct {
	unit       *playerSetMaxManaTestObject4EECD0
	maximumArg uint16
	events     []string
	faultAt    int

	afterUnit    func(*playerSetMaxManaTestWorld4EECD0, *playerSetMaxManaTestObject4EECD0)
	afterClass   func(*playerSetMaxManaTestWorld4EECD0, *playerSetMaxManaTestObject4EECD0, uint8)
	afterUpdate  func(*playerSetMaxManaTestWorld4EECD0, *playerSetMaxManaTestUpdate4EECD0)
	afterMaximum func(*playerSetMaxManaTestWorld4EECD0, uint16)
	afterStore   func(*playerSetMaxManaTestWorld4EECD0, *playerSetMaxManaTestUpdate4EECD0, uint16)
}

func playerSetMaxManaObjectName4EECD0(obj *playerSetMaxManaTestObject4EECD0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerSetMaxManaUpdateName4EECD0(update *playerSetMaxManaTestUpdate4EECD0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func (w *playerSetMaxManaTestWorld4EECD0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerSetMaxManaTestWorld4EECD0) hooks() playerSetMaxManaHooks4EECD0[
	*playerSetMaxManaTestObject4EECD0,
	*playerSetMaxManaTestUpdate4EECD0,
	string,
] {
	return playerSetMaxManaHooks4EECD0[
		*playerSetMaxManaTestObject4EECD0,
		*playerSetMaxManaTestUpdate4EECD0,
		string,
	]{
		loadUnitArg: func() (*playerSetMaxManaTestObject4EECD0, string) {
			unit := w.unit
			result := "unit-result:" + playerSetMaxManaObjectName4EECD0(unit)
			w.record("arg:" + playerSetMaxManaObjectName4EECD0(unit))
			if w.afterUnit != nil {
				w.afterUnit(w, unit)
			}
			return unit, result
		},
		loadClassLow: func(unit *playerSetMaxManaTestObject4EECD0) uint8 {
			classLow := uint8(unit.class)
			w.record(fmt.Sprintf("class-low:%s=%#02x", playerSetMaxManaObjectName4EECD0(unit), classLow))
			if w.afterClass != nil {
				w.afterClass(w, unit, classLow)
			}
			return classLow
		},
		loadUpdateData: func(unit *playerSetMaxManaTestObject4EECD0) (*playerSetMaxManaTestUpdate4EECD0, string) {
			update := unit.update
			result := "update-result:" + playerSetMaxManaUpdateName4EECD0(update)
			w.record("update:" + playerSetMaxManaObjectName4EECD0(unit) + "=" + playerSetMaxManaUpdateName4EECD0(update))
			if w.afterUpdate != nil {
				w.afterUpdate(w, update)
			}
			return update, result
		},
		loadMaximumArg: func() uint16 {
			maximum := w.maximumArg
			w.record(fmt.Sprintf("maximum:%#04x", maximum))
			if w.afterMaximum != nil {
				w.afterMaximum(w, maximum)
			}
			return maximum
		},
		storeMaximum: func(update *playerSetMaxManaTestUpdate4EECD0, maximum uint16) {
			w.record(fmt.Sprintf("store:%s=%#04x", playerSetMaxManaUpdateName4EECD0(update), maximum))
			if update == nil {
				panic("nil update dereference")
			}
			update.maximum = maximum
			if w.afterStore != nil {
				w.afterStore(w, update, maximum)
			}
		},
	}
}

func newPlayerSetMaxManaTestWorld4EECD0() *playerSetMaxManaTestWorld4EECD0 {
	return &playerSetMaxManaTestWorld4EECD0{
		unit: &playerSetMaxManaTestObject4EECD0{
			name:   "unit",
			class:  0x04,
			update: &playerSetMaxManaTestUpdate4EECD0{name: "update", maximum: 0x1111},
		},
		maximumArg: 0x4321,
	}
}

func TestPlayerSetMaxMana4EECD0EntryGates(t *testing.T) {
	tests := []struct {
		name       string
		unit       *playerSetMaxManaTestObject4EECD0
		wantResult string
		wantEvents []string
	}{
		{name: "nil unit", unit: nil, wantResult: "unit-result:nil", wantEvents: []string{"arg:nil"}},
		{name: "other", unit: &playerSetMaxManaTestObject4EECD0{name: "unit", class: 0x80000000}, wantResult: "unit-result:unit", wantEvents: []string{"arg:unit", "class-low:unit=0x00"}},
		{name: "upper byte player bit", unit: &playerSetMaxManaTestObject4EECD0{name: "unit", class: 0x00000400}, wantResult: "unit-result:unit", wantEvents: []string{"arg:unit", "class-low:unit=0x00"}},
		{name: "monster", unit: &playerSetMaxManaTestObject4EECD0{name: "unit", class: 0x40000002}, wantResult: "unit-result:unit", wantEvents: []string{"arg:unit", "class-low:unit=0x02"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerSetMaxManaTestWorld4EECD0()
			w.unit = test.unit
			if got := playerSetMaxMana4EECD0(w.hooks()); got != test.wantResult {
				t.Fatalf("result = %q, want %q", got, test.wantResult)
			}
			if !reflect.DeepEqual(w.events, test.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, test.wantEvents)
			}
		})
	}
}

func TestPlayerSetMaxMana4EECD0PreservesEveryMaximumWord(t *testing.T) {
	for _, maximum := range []uint16{0, 1, 999, 1000, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("%04x", maximum), func(t *testing.T) {
			w := newPlayerSetMaxManaTestWorld4EECD0()
			w.maximumArg = maximum
			update := w.unit.update
			if got := playerSetMaxMana4EECD0(w.hooks()); got != "update-result:update" {
				t.Fatalf("result = %q, want update identity", got)
			}
			if update.maximum != maximum {
				t.Fatalf("maximum = %#04x, want %#04x", update.maximum, maximum)
			}
		})
	}
}

func TestPlayerSetMaxMana4EECD0PlayerBitWinsAndUpperClassBytesAreIgnored(t *testing.T) {
	w := newPlayerSetMaxManaTestWorld4EECD0()
	w.unit.class = 0xa5a50006
	update := w.unit.update
	if got := playerSetMaxMana4EECD0(w.hooks()); got != "update-result:update" {
		t.Fatalf("result = %q, want update identity", got)
	}
	if update.maximum != 0x4321 {
		t.Fatalf("maximum = %#04x, want 0x4321", update.maximum)
	}
	want := []string{"arg:unit", "class-low:unit=0x06", "update:unit=update", "maximum:0x4321", "store:update=0x4321"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestPlayerSetMaxMana4EECD0CachesEachLoadedValue(t *testing.T) {
	w := newPlayerSetMaxManaTestWorld4EECD0()
	originalUnit := w.unit
	originalUpdate := originalUnit.update
	replacementUpdate := &playerSetMaxManaTestUpdate4EECD0{name: "replacement", maximum: 0x2222}
	finalUpdate := &playerSetMaxManaTestUpdate4EECD0{name: "final", maximum: 0x3333}
	w.afterUnit = func(w *playerSetMaxManaTestWorld4EECD0, _ *playerSetMaxManaTestObject4EECD0) {
		w.unit = &playerSetMaxManaTestObject4EECD0{name: "new-unit", class: 0, update: finalUpdate}
	}
	w.afterClass = func(_ *playerSetMaxManaTestWorld4EECD0, unit *playerSetMaxManaTestObject4EECD0, _ uint8) {
		unit.class = 0
		unit.update = replacementUpdate
	}
	w.afterUpdate = func(_ *playerSetMaxManaTestWorld4EECD0, _ *playerSetMaxManaTestUpdate4EECD0) {
		originalUnit.update = finalUpdate
	}
	w.afterMaximum = func(w *playerSetMaxManaTestWorld4EECD0, _ uint16) {
		w.maximumArg = 0x9876
	}

	if got := playerSetMaxMana4EECD0(w.hooks()); got != "update-result:replacement" {
		t.Fatalf("result = %q, want cached replacement identity", got)
	}
	if replacementUpdate.maximum != 0x4321 {
		t.Fatalf("cached update maximum = %#04x, want 0x4321", replacementUpdate.maximum)
	}
	if originalUpdate.maximum != 0x1111 || finalUpdate.maximum != 0x3333 || w.maximumArg != 0x9876 {
		t.Fatalf("uncached state changed: original=%#04x final=%#04x argument=%#04x", originalUpdate.maximum, finalUpdate.maximum, w.maximumArg)
	}
	want := []string{"arg:unit", "class-low:unit=0x04", "update:unit=replacement", "maximum:0x4321", "store:replacement=0x4321"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestPlayerSetMaxMana4EECD0NilUpdateLoadsAmountBeforeFault(t *testing.T) {
	w := newPlayerSetMaxManaTestWorld4EECD0()
	w.unit.update = nil
	defer func() {
		if got := recover(); got != "nil update dereference" {
			t.Fatalf("panic = %v, want nil update dereference", got)
		}
		want := []string{"arg:unit", "class-low:unit=0x04", "update:unit=nil", "maximum:0x4321", "store:nil=0x4321"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	playerSetMaxMana4EECD0(w.hooks())
}

func TestPlayerSetMaxMana4EECD0AllFaultPrefixes(t *testing.T) {
	want := []string{"arg:unit", "class-low:unit=0x04", "update:unit=update", "maximum:0x4321", "store:update=0x4321"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newPlayerSetMaxManaTestWorld4EECD0()
			update := w.unit.update
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
				if update.maximum != 0x1111 {
					t.Fatalf("fault prefix stored maximum = %#04x", update.maximum)
				}
			}()
			playerSetMaxMana4EECD0(w.hooks())
		})
	}
}
