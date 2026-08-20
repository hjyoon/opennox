package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerGetMaxManaTestUpdate4EECB0 struct {
	name    string
	maximum uint16
}

type playerGetMaxManaTestObject4EECB0 struct {
	name   string
	class  uint32
	update *playerGetMaxManaTestUpdate4EECB0
}

type playerGetMaxManaTestWorld4EECB0 struct {
	unit    *playerGetMaxManaTestObject4EECB0
	events  []string
	faultAt int

	afterClassLow func(*playerGetMaxManaTestWorld4EECB0, uint8)
	afterUpdate   func(*playerGetMaxManaTestWorld4EECB0, *playerGetMaxManaTestUpdate4EECB0)
	afterMaximum  func(*playerGetMaxManaTestWorld4EECB0, *playerGetMaxManaTestUpdate4EECB0, uint16)
}

func playerGetMaxManaObjectName4EECB0(obj *playerGetMaxManaTestObject4EECB0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerGetMaxManaUpdateName4EECB0(update *playerGetMaxManaTestUpdate4EECB0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func (w *playerGetMaxManaTestWorld4EECB0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerGetMaxManaTestWorld4EECB0) hooks() playerGetMaxManaHooks4EECB0[
	*playerGetMaxManaTestObject4EECB0,
	*playerGetMaxManaTestUpdate4EECB0,
] {
	return playerGetMaxManaHooks4EECB0[
		*playerGetMaxManaTestObject4EECB0,
		*playerGetMaxManaTestUpdate4EECB0,
	]{
		loadUnitArg: func() *playerGetMaxManaTestObject4EECB0 {
			w.record("arg:" + playerGetMaxManaObjectName4EECB0(w.unit))
			return w.unit
		},
		loadClassLow: func(unit *playerGetMaxManaTestObject4EECB0) uint8 {
			classLow := uint8(unit.class)
			w.record(fmt.Sprintf("class-low:%s=%#02x", playerGetMaxManaObjectName4EECB0(unit), classLow))
			if w.afterClassLow != nil {
				w.afterClassLow(w, classLow)
			}
			return classLow
		},
		loadUpdateData: func(unit *playerGetMaxManaTestObject4EECB0) *playerGetMaxManaTestUpdate4EECB0 {
			update := unit.update
			w.record("update:" + playerGetMaxManaObjectName4EECB0(unit) + "=" + playerGetMaxManaUpdateName4EECB0(update))
			if w.afterUpdate != nil {
				w.afterUpdate(w, update)
			}
			return update
		},
		loadMaximum: func(update *playerGetMaxManaTestUpdate4EECB0) uint16 {
			w.record("maximum:" + playerGetMaxManaUpdateName4EECB0(update))
			if update == nil {
				panic("nil update dereference")
			}
			maximum := update.maximum
			if w.afterMaximum != nil {
				w.afterMaximum(w, update, maximum)
			}
			return maximum
		},
	}
}

func newPlayerGetMaxManaTestWorld4EECB0() *playerGetMaxManaTestWorld4EECB0 {
	return &playerGetMaxManaTestWorld4EECB0{
		unit: &playerGetMaxManaTestObject4EECB0{
			name:   "unit",
			class:  0x04,
			update: &playerGetMaxManaTestUpdate4EECB0{name: "update", maximum: 0x4321},
		},
	}
}

func TestPlayerGetMaxMana4EECB0ClassGates(t *testing.T) {
	tests := []struct {
		name  string
		unit  *playerGetMaxManaTestObject4EECB0
		want  uint16
		event []string
	}{
		{name: "nil unit", unit: nil, want: 0, event: []string{"arg:nil"}},
		{name: "other", unit: &playerGetMaxManaTestObject4EECB0{name: "unit", class: 0x80000000}, want: 0, event: []string{"arg:unit", "class-low:unit=0x00"}},
		{name: "upper byte player bit", unit: &playerGetMaxManaTestObject4EECB0{name: "unit", class: 0x00000400}, want: 0, event: []string{"arg:unit", "class-low:unit=0x00"}},
		{name: "monster", unit: &playerGetMaxManaTestObject4EECB0{name: "unit", class: 0x40000002}, want: 0, event: []string{"arg:unit", "class-low:unit=0x02"}},
		{name: "player", unit: &playerGetMaxManaTestObject4EECB0{name: "unit", class: 0x20000004, update: &playerGetMaxManaTestUpdate4EECB0{name: "update", maximum: 0x1234}}, want: 0x1234, event: []string{"arg:unit", "class-low:unit=0x04", "update:unit=update", "maximum:update"}},
		{name: "player and monster", unit: &playerGetMaxManaTestObject4EECB0{name: "unit", class: 0x10000006, update: &playerGetMaxManaTestUpdate4EECB0{name: "update", maximum: 0xabcd}}, want: 0xabcd, event: []string{"arg:unit", "class-low:unit=0x06", "update:unit=update", "maximum:update"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerGetMaxManaTestWorld4EECB0()
			w.unit = test.unit
			if got := playerGetMaxMana4EECB0(w.hooks()); got != test.want {
				t.Fatalf("result = %#04x, want %#04x", got, test.want)
			}
			if !reflect.DeepEqual(w.events, test.event) {
				t.Fatalf("events = %q, want %q", w.events, test.event)
			}
		})
	}
}

func TestPlayerGetMaxMana4EECB0PreservesEveryMaximumWord(t *testing.T) {
	for _, maximum := range []uint16{0, 1, 999, 1000, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("%04x", maximum), func(t *testing.T) {
			w := newPlayerGetMaxManaTestWorld4EECB0()
			w.unit.update.maximum = maximum
			if got := playerGetMaxMana4EECB0(w.hooks()); got != maximum {
				t.Fatalf("result = %#04x, want %#04x", got, maximum)
			}
		})
	}
}

func TestPlayerGetMaxMana4EECB0CachesClassUpdateAndMaximum(t *testing.T) {
	w := newPlayerGetMaxManaTestWorld4EECB0()
	original := w.unit.update
	replacement := &playerGetMaxManaTestUpdate4EECB0{name: "replacement", maximum: 0xabcd}
	w.afterClassLow = func(w *playerGetMaxManaTestWorld4EECB0, _ uint8) {
		w.unit.class = 0
	}
	w.afterUpdate = func(w *playerGetMaxManaTestWorld4EECB0, _ *playerGetMaxManaTestUpdate4EECB0) {
		w.unit.update = replacement
	}
	w.afterMaximum = func(_ *playerGetMaxManaTestWorld4EECB0, update *playerGetMaxManaTestUpdate4EECB0, _ uint16) {
		update.maximum = 0x9876
	}

	if got := playerGetMaxMana4EECB0(w.hooks()); got != 0x4321 {
		t.Fatalf("result = %#04x, want cached 0x4321", got)
	}
	if w.unit.class != 0 || w.unit.update != replacement || original.maximum != 0x9876 {
		t.Fatalf("mutations were not observed: class=%#x update=%p original=%#04x", w.unit.class, w.unit.update, original.maximum)
	}
	want := []string{"arg:unit", "class-low:unit=0x04", "update:unit=update", "maximum:update"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestPlayerGetMaxMana4EECB0HasNoUpdateNilGuard(t *testing.T) {
	w := newPlayerGetMaxManaTestWorld4EECB0()
	w.unit.update = nil
	defer func() {
		if got := recover(); got != "nil update dereference" {
			t.Fatalf("panic = %v, want nil update dereference", got)
		}
		want := []string{"arg:unit", "class-low:unit=0x04", "update:unit=nil", "maximum:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	playerGetMaxMana4EECB0(w.hooks())
}

func TestPlayerGetMaxMana4EECB0AllFaultPrefixes(t *testing.T) {
	want := []string{"arg:unit", "class-low:unit=0x04", "update:unit=update", "maximum:update"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newPlayerGetMaxManaTestWorld4EECB0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
			}()
			playerGetMaxMana4EECB0(w.hooks())
		})
	}
}
