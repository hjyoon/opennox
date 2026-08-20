package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitHPSetOnMaxTestHealth4EE6F0 struct {
	name    string
	current uint16
	field2  uint16
	maximum uint16
}

type unitHPSetOnMaxTestObject4EE6F0 struct {
	name   string
	class  uint8
	health *unitHPSetOnMaxTestHealth4EE6F0
}

type unitHPSetOnMaxTestWorld4EE6F0 struct {
	unit         *unitHPSetOnMaxTestObject4EE6F0
	events       []string
	afterMaximum func(*unitHPSetOnMaxTestWorld4EE6F0)
	afterSetHP   func(*unitHPSetOnMaxTestWorld4EE6F0)
	afterCurrent func(*unitHPSetOnMaxTestWorld4EE6F0)
	afterStore   func(*unitHPSetOnMaxTestWorld4EE6F0)
}

func unitHPSetOnMaxObjectName4EE6F0(obj *unitHPSetOnMaxTestObject4EE6F0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func unitHPSetOnMaxHealthName4EE6F0(health *unitHPSetOnMaxTestHealth4EE6F0) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func (w *unitHPSetOnMaxTestWorld4EE6F0) event(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *unitHPSetOnMaxTestWorld4EE6F0) hooks() unitHPSetOnMaxHooks4EE6F0[
	*unitHPSetOnMaxTestObject4EE6F0,
	*unitHPSetOnMaxTestHealth4EE6F0,
] {
	return unitHPSetOnMaxHooks4EE6F0[
		*unitHPSetOnMaxTestObject4EE6F0,
		*unitHPSetOnMaxTestHealth4EE6F0,
	]{
		loadUnitArg: func() *unitHPSetOnMaxTestObject4EE6F0 {
			w.event("arg:%s", unitHPSetOnMaxObjectName4EE6F0(w.unit))
			return w.unit
		},
		loadHealth: func(obj *unitHPSetOnMaxTestObject4EE6F0) *unitHPSetOnMaxTestHealth4EE6F0 {
			w.event("health:%s=%s", unitHPSetOnMaxObjectName4EE6F0(obj), unitHPSetOnMaxHealthName4EE6F0(obj.health))
			return obj.health
		},
		loadMaximum: func(health *unitHPSetOnMaxTestHealth4EE6F0) uint16 {
			value := health.maximum
			w.event("max:%s=%d", unitHPSetOnMaxHealthName4EE6F0(health), value)
			if w.afterMaximum != nil {
				w.afterMaximum(w)
			}
			return value
		},
		setHP: func(obj *unitHPSetOnMaxTestObject4EE6F0, value uint16) {
			w.event("set:%s=%d", unitHPSetOnMaxObjectName4EE6F0(obj), value)
			obj.health.current = value
			if w.afterSetHP != nil {
				w.afterSetHP(w)
			}
		},
		loadCurrent: func(health *unitHPSetOnMaxTestHealth4EE6F0) uint16 {
			w.event("cur:%s", unitHPSetOnMaxHealthName4EE6F0(health))
			value := health.current
			if w.afterCurrent != nil {
				w.afterCurrent(w)
			}
			return value
		},
		storeField2: func(health *unitHPSetOnMaxTestHealth4EE6F0, value uint16) {
			w.event("field2:%s<-%d", unitHPSetOnMaxHealthName4EE6F0(health), value)
			health.field2 = value
			if w.afterStore != nil {
				w.afterStore(w)
			}
		},
		loadClassLow: func(obj *unitHPSetOnMaxTestObject4EE6F0) uint8 {
			w.event("class:%s=%#x", unitHPSetOnMaxObjectName4EE6F0(obj), obj.class)
			return obj.class
		},
		informOwner: func(obj *unitHPSetOnMaxTestObject4EE6F0) {
			w.event("inform:%s", unitHPSetOnMaxObjectName4EE6F0(obj))
		},
	}
}

func newUnitHPSetOnMaxWorld4EE6F0() *unitHPSetOnMaxTestWorld4EE6F0 {
	health := &unitHPSetOnMaxTestHealth4EE6F0{name: "entry", current: 7, field2: 8, maximum: 100}
	return &unitHPSetOnMaxTestWorld4EE6F0{
		unit: &unitHPSetOnMaxTestObject4EE6F0{name: "unit", health: health},
	}
}

func assertUnitHPSetOnMaxEvents4EE6F0(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got: %q\nwant: %q", got, want)
	}
}

func TestUnitHPSetOnMax4EE6F0EntryGatesAndZeroMaximum(t *testing.T) {
	tests := []struct {
		name string
		edit func(*unitHPSetOnMaxTestWorld4EE6F0)
		want []string
	}{
		{
			name: "nil unit",
			edit: func(w *unitHPSetOnMaxTestWorld4EE6F0) { w.unit = nil },
			want: []string{"arg:nil"},
		},
		{
			name: "nil health",
			edit: func(w *unitHPSetOnMaxTestWorld4EE6F0) { w.unit.health = nil },
			want: []string{"arg:unit", "health:unit=nil"},
		},
		{
			name: "zero maximum still restores",
			edit: func(w *unitHPSetOnMaxTestWorld4EE6F0) { w.unit.health.maximum = 0 },
			want: []string{
				"arg:unit", "health:unit=entry", "max:entry=0", "set:unit=0",
				"health:unit=entry", "cur:entry", "field2:entry<-0", "class:unit=0x0",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newUnitHPSetOnMaxWorld4EE6F0()
			tc.edit(w)
			unitHPSetOnMax4EE6F0(w.hooks())
			assertUnitHPSetOnMaxEvents4EE6F0(t, w.events, tc.want)
		})
	}
}

func TestUnitHPSetOnMax4EE6F0CachesMaximumAndReloadsHealth(t *testing.T) {
	w := newUnitHPSetOnMaxWorld4EE6F0()
	entry := w.unit.health
	live := &unitHPSetOnMaxTestHealth4EE6F0{name: "live", current: 37, field2: 9, maximum: 200}
	w.afterMaximum = func(w *unitHPSetOnMaxTestWorld4EE6F0) {
		entry.maximum = 65535
	}
	w.afterSetHP = func(w *unitHPSetOnMaxTestWorld4EE6F0) {
		w.unit.health = live
	}
	unitHPSetOnMax4EE6F0(w.hooks())
	assertUnitHPSetOnMaxEvents4EE6F0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "set:unit=100",
		"health:unit=live", "cur:live", "field2:live<-37", "class:unit=0x0",
	})
	if entry.current != 100 || entry.field2 != 8 || live.field2 != 37 {
		t.Fatalf("entry=%+v live=%+v", *entry, *live)
	}
}

func TestUnitHPSetOnMax4EE6F0CachesCurrentBeforeStore(t *testing.T) {
	w := newUnitHPSetOnMaxWorld4EE6F0()
	w.afterCurrent = func(w *unitHPSetOnMaxTestWorld4EE6F0) {
		w.unit.health.current = 12
	}
	w.afterStore = func(w *unitHPSetOnMaxTestWorld4EE6F0) {
		w.unit.class = unitHPSetOnMaxMonsterBit4EE6F0
	}
	unitHPSetOnMax4EE6F0(w.hooks())
	assertUnitHPSetOnMaxEvents4EE6F0(t, w.events, []string{
		"arg:unit", "health:unit=entry", "max:entry=100", "set:unit=100",
		"health:unit=entry", "cur:entry", "field2:entry<-100", "class:unit=0x2", "inform:unit",
	})
	if w.unit.health.current != 12 || w.unit.health.field2 != 100 {
		t.Fatalf("health = %+v, want current 12 and cached Field2 100", *w.unit.health)
	}
}

func TestUnitHPSetOnMax4EE6F0SetterMutationControlsFinalClass(t *testing.T) {
	w := newUnitHPSetOnMaxWorld4EE6F0()
	w.afterSetHP = func(w *unitHPSetOnMaxTestWorld4EE6F0) {
		w.unit.class = unitHPSetOnMaxMonsterBit4EE6F0
	}
	unitHPSetOnMax4EE6F0(w.hooks())
	if got := w.events[len(w.events)-2:]; !reflect.DeepEqual(got, []string{"class:unit=0x2", "inform:unit"}) {
		t.Fatalf("final events = %q", got)
	}
}

func TestUnitHPSetOnMax4EE6F0ReloadedNilHealthFaultPrefix(t *testing.T) {
	w := newUnitHPSetOnMaxWorld4EE6F0()
	w.afterSetHP = func(w *unitHPSetOnMaxTestWorld4EE6F0) {
		w.unit.health = nil
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil live HealthData did not preserve the original fault")
		}
		assertUnitHPSetOnMaxEvents4EE6F0(t, w.events, []string{
			"arg:unit", "health:unit=entry", "max:entry=100", "set:unit=100",
			"health:unit=nil", "cur:nil",
		})
	}()
	unitHPSetOnMax4EE6F0(w.hooks())
}
