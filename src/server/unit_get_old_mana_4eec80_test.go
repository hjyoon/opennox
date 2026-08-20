package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitGetOldManaTestUpdate4EEC80 struct {
	name    string
	current uint16
}

type unitGetOldManaTestObject4EEC80 struct {
	name   string
	class  uint32
	update *unitGetOldManaTestUpdate4EEC80
}

type unitGetOldManaTestWorld4EEC80 struct {
	unit    *unitGetOldManaTestObject4EEC80
	events  []string
	faultAt int

	afterClass   func(*unitGetOldManaTestWorld4EEC80, uint32)
	afterUpdate  func(*unitGetOldManaTestWorld4EEC80, *unitGetOldManaTestUpdate4EEC80)
	afterCurrent func(*unitGetOldManaTestWorld4EEC80, *unitGetOldManaTestUpdate4EEC80, uint16)
}

func unitGetOldManaObjectName4EEC80(obj *unitGetOldManaTestObject4EEC80) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func unitGetOldManaUpdateName4EEC80(update *unitGetOldManaTestUpdate4EEC80) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func (w *unitGetOldManaTestWorld4EEC80) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *unitGetOldManaTestWorld4EEC80) hooks() unitGetOldManaHooks4EEC80[
	*unitGetOldManaTestObject4EEC80,
	*unitGetOldManaTestUpdate4EEC80,
] {
	return unitGetOldManaHooks4EEC80[
		*unitGetOldManaTestObject4EEC80,
		*unitGetOldManaTestUpdate4EEC80,
	]{
		loadUnitArg: func() *unitGetOldManaTestObject4EEC80 {
			w.record("arg:" + unitGetOldManaObjectName4EEC80(w.unit))
			return w.unit
		},
		loadClass: func(unit *unitGetOldManaTestObject4EEC80) uint32 {
			class := unit.class
			w.record(fmt.Sprintf("class:%s=%#08x", unitGetOldManaObjectName4EEC80(unit), class))
			if w.afterClass != nil {
				w.afterClass(w, class)
			}
			return class
		},
		loadUpdateData: func(unit *unitGetOldManaTestObject4EEC80) *unitGetOldManaTestUpdate4EEC80 {
			update := unit.update
			w.record("update:" + unitGetOldManaObjectName4EEC80(unit) + "=" + unitGetOldManaUpdateName4EEC80(update))
			if w.afterUpdate != nil {
				w.afterUpdate(w, update)
			}
			return update
		},
		loadCurrent: func(update *unitGetOldManaTestUpdate4EEC80) uint16 {
			w.record("current:" + unitGetOldManaUpdateName4EEC80(update))
			if update == nil {
				panic("nil update dereference")
			}
			current := update.current
			if w.afterCurrent != nil {
				w.afterCurrent(w, update, current)
			}
			return current
		},
	}
}

func newUnitGetOldManaTestWorld4EEC80() *unitGetOldManaTestWorld4EEC80 {
	return &unitGetOldManaTestWorld4EEC80{
		unit: &unitGetOldManaTestObject4EEC80{
			name:   "unit",
			class:  0x04,
			update: &unitGetOldManaTestUpdate4EEC80{name: "update", current: 0x4321},
		},
	}
}

func TestUnitGetOldMana4EEC80ClassGates(t *testing.T) {
	tests := []struct {
		name  string
		unit  *unitGetOldManaTestObject4EEC80
		want  uint16
		event []string
	}{
		{name: "nil unit", unit: nil, want: 0, event: []string{"arg:nil"}},
		{name: "other", unit: &unitGetOldManaTestObject4EEC80{name: "unit", class: 0x80000400}, want: 0, event: []string{"arg:unit", "class:unit=0x80000400"}},
		{name: "monster", unit: &unitGetOldManaTestObject4EEC80{name: "unit", class: 0x40000002}, want: 1000, event: []string{"arg:unit", "class:unit=0x40000002"}},
		{name: "player", unit: &unitGetOldManaTestObject4EEC80{name: "unit", class: 0x20000004, update: &unitGetOldManaTestUpdate4EEC80{name: "update", current: 0x1234}}, want: 0x1234, event: []string{"arg:unit", "class:unit=0x20000004", "update:unit=update", "current:update"}},
		{name: "player precedes monster", unit: &unitGetOldManaTestObject4EEC80{name: "unit", class: 0x10000006, update: &unitGetOldManaTestUpdate4EEC80{name: "update", current: 0xabcd}}, want: 0xabcd, event: []string{"arg:unit", "class:unit=0x10000006", "update:unit=update", "current:update"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newUnitGetOldManaTestWorld4EEC80()
			w.unit = test.unit
			if got := unitGetOldMana4EEC80(w.hooks()); got != test.want {
				t.Fatalf("result = %#04x, want %#04x", got, test.want)
			}
			if !reflect.DeepEqual(w.events, test.event) {
				t.Fatalf("events = %q, want %q", w.events, test.event)
			}
		})
	}
}

func TestUnitGetOldMana4EEC80PreservesEveryCurrentWord(t *testing.T) {
	for _, current := range []uint16{0, 1, 999, 1000, 0x7fff, 0x8000, 0xffff} {
		t.Run(fmt.Sprintf("%04x", current), func(t *testing.T) {
			w := newUnitGetOldManaTestWorld4EEC80()
			w.unit.update.current = current
			if got := unitGetOldMana4EEC80(w.hooks()); got != current {
				t.Fatalf("result = %#04x, want %#04x", got, current)
			}
		})
	}
}

func TestUnitGetOldMana4EEC80CachesClassUpdateAndCurrent(t *testing.T) {
	w := newUnitGetOldManaTestWorld4EEC80()
	original := w.unit.update
	replacement := &unitGetOldManaTestUpdate4EEC80{name: "replacement", current: 0xabcd}
	w.afterClass = func(w *unitGetOldManaTestWorld4EEC80, _ uint32) {
		w.unit.class = 0x02
	}
	w.afterUpdate = func(w *unitGetOldManaTestWorld4EEC80, _ *unitGetOldManaTestUpdate4EEC80) {
		w.unit.update = replacement
	}
	w.afterCurrent = func(_ *unitGetOldManaTestWorld4EEC80, update *unitGetOldManaTestUpdate4EEC80, _ uint16) {
		update.current = 0x9876
	}

	if got := unitGetOldMana4EEC80(w.hooks()); got != 0x4321 {
		t.Fatalf("result = %#04x, want cached 0x4321", got)
	}
	if w.unit.class != 0x02 || w.unit.update != replacement || original.current != 0x9876 {
		t.Fatalf("mutations were not observed: class=%#x update=%p original=%#04x", w.unit.class, w.unit.update, original.current)
	}
	want := []string{"arg:unit", "class:unit=0x00000004", "update:unit=update", "current:update"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestUnitGetOldMana4EEC80CachesClassForMonsterDecision(t *testing.T) {
	w := newUnitGetOldManaTestWorld4EEC80()
	w.unit.class = 0x02
	w.afterClass = func(w *unitGetOldManaTestWorld4EEC80, _ uint32) {
		w.unit.class = 0x04
	}
	if got := unitGetOldMana4EEC80(w.hooks()); got != 1000 {
		t.Fatalf("result = %d, want cached-class monster value 1000", got)
	}
	want := []string{"arg:unit", "class:unit=0x00000002"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestUnitGetOldMana4EEC80HasNoUpdateNilGuard(t *testing.T) {
	w := newUnitGetOldManaTestWorld4EEC80()
	w.unit.update = nil
	defer func() {
		if got := recover(); got != "nil update dereference" {
			t.Fatalf("panic = %v, want nil update dereference", got)
		}
		want := []string{"arg:unit", "class:unit=0x00000004", "update:unit=nil", "current:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	}()
	unitGetOldMana4EEC80(w.hooks())
}

func TestUnitGetOldMana4EEC80AllFaultPrefixes(t *testing.T) {
	want := []string{"arg:unit", "class:unit=0x00000004", "update:unit=update", "current:update"}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newUnitGetOldManaTestWorld4EEC80()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %q, want %q", w.events, prefix)
				}
			}()
			unitGetOldMana4EEC80(w.hooks())
		})
	}
}
