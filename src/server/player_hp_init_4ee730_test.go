package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerHPInitTestHealth4EE730 struct {
	name    string
	current uint16
}

type playerHPInitTestUpdate4EE730 struct {
	name          string
	samples       [playerHPInitSampleCount4EE730]uint16
	currentSample uint16
}

type playerHPInitTestObject4EE730 struct {
	name   string
	class  uint8
	health *playerHPInitTestHealth4EE730
	update *playerHPInitTestUpdate4EE730
}

type playerHPInitTestWorld4EE730 struct {
	unit    *playerHPInitTestObject4EE730
	events  []string
	faultAt int

	afterUpdate       func(*playerHPInitTestWorld4EE730)
	afterCurrent      func(*playerHPInitTestWorld4EE730, *playerHPInitTestHealth4EE730)
	afterSample       func(*playerHPInitTestWorld4EE730, int)
	afterCurrentStore func(*playerHPInitTestWorld4EE730)
}

func playerHPInitObjectName4EE730(obj *playerHPInitTestObject4EE730) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerHPInitHealthName4EE730(health *playerHPInitTestHealth4EE730) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func playerHPInitUpdateName4EE730(update *playerHPInitTestUpdate4EE730) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func (w *playerHPInitTestWorld4EE730) record(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerHPInitTestWorld4EE730) hooks() playerHPInitHooks4EE730[
	*playerHPInitTestObject4EE730,
	*playerHPInitTestHealth4EE730,
	*playerHPInitTestUpdate4EE730,
] {
	return playerHPInitHooks4EE730[
		*playerHPInitTestObject4EE730,
		*playerHPInitTestHealth4EE730,
		*playerHPInitTestUpdate4EE730,
	]{
		loadUnitArg: func() *playerHPInitTestObject4EE730 {
			w.record("arg:%s", playerHPInitObjectName4EE730(w.unit))
			return w.unit
		},
		loadClassLow: func(obj *playerHPInitTestObject4EE730) uint8 {
			w.record("class:%s=%#x", playerHPInitObjectName4EE730(obj), obj.class)
			return obj.class
		},
		loadHealth: func(obj *playerHPInitTestObject4EE730) *playerHPInitTestHealth4EE730 {
			health := obj.health
			w.record("health:%s=%s", playerHPInitObjectName4EE730(obj), playerHPInitHealthName4EE730(health))
			return health
		},
		loadUpdateData: func(obj *playerHPInitTestObject4EE730) *playerHPInitTestUpdate4EE730 {
			update := obj.update
			w.record("update:%s=%s", playerHPInitObjectName4EE730(obj), playerHPInitUpdateName4EE730(update))
			if w.afterUpdate != nil {
				w.afterUpdate(w)
			}
			return update
		},
		loadCurrent: func(health *playerHPInitTestHealth4EE730) uint16 {
			w.record("current:%s", playerHPInitHealthName4EE730(health))
			value := health.current
			if w.afterCurrent != nil {
				w.afterCurrent(w, health)
			}
			return value
		},
		storeSample: func(update *playerHPInitTestUpdate4EE730, index int, value uint16) {
			w.record("sample:%s[%d]=%d", playerHPInitUpdateName4EE730(update), index, value)
			update.samples[index] = value
			if w.afterSample != nil {
				w.afterSample(w, index)
			}
		},
		storeCurrentSample: func(update *playerHPInitTestUpdate4EE730, value uint16) {
			w.record("current-sample:%s=%d", playerHPInitUpdateName4EE730(update), value)
			update.currentSample = value
			if w.afterCurrentStore != nil {
				w.afterCurrentStore(w)
			}
		},
	}
}

func newPlayerHPInitWorld4EE730() *playerHPInitTestWorld4EE730 {
	return &playerHPInitTestWorld4EE730{
		unit: &playerHPInitTestObject4EE730{
			name:   "unit",
			class:  playerHPInitPlayerBit4EE730,
			health: &playerHPInitTestHealth4EE730{name: "entry", current: 17},
			update: &playerHPInitTestUpdate4EE730{name: "cached"},
		},
	}
}

func TestPlayerHPInit4EE730EntryGatesAndLoadOrder(t *testing.T) {
	tests := []struct {
		name string
		edit func(*playerHPInitTestWorld4EE730)
		want []string
	}{
		{
			name: "nil unit",
			edit: func(w *playerHPInitTestWorld4EE730) { w.unit = nil },
			want: []string{"arg:nil"},
		},
		{
			name: "non-player",
			edit: func(w *playerHPInitTestWorld4EE730) { w.unit.class = 0x82 },
			want: []string{"arg:unit", "class:unit=0x82"},
		},
		{
			name: "nil initial health still loads update",
			edit: func(w *playerHPInitTestWorld4EE730) {
				w.unit.health = nil
				w.unit.update = nil
			},
			want: []string{"arg:unit", "class:unit=0x4", "health:unit=nil", "update:unit=nil"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newPlayerHPInitWorld4EE730()
			test.edit(w)
			playerHPInit4EE730(w.hooks())
			if !reflect.DeepEqual(w.events, test.want) {
				t.Fatalf("events = %q, want %q", w.events, test.want)
			}
		})
	}
}

func TestPlayerHPInit4EE730ReloadsEverySampleAndCachesUpdate(t *testing.T) {
	w := newPlayerHPInitWorld4EE730()
	cached := w.unit.update
	other := &playerHPInitTestUpdate4EE730{name: "other"}
	health := make([]*playerHPInitTestHealth4EE730, playerHPInitSampleCount4EE730+1)
	for i := range health {
		health[i] = &playerHPInitTestHealth4EE730{
			name:    fmt.Sprintf("live-%d", i),
			current: uint16(100 + i),
		}
	}
	w.afterUpdate = func(w *playerHPInitTestWorld4EE730) {
		w.unit.update = other
		w.unit.health = health[0]
	}
	w.afterSample = func(w *playerHPInitTestWorld4EE730, index int) {
		w.unit.health = health[index+1]
	}

	playerHPInit4EE730(w.hooks())
	for i, got := range cached.samples {
		if want := uint16(100 + i); got != want {
			t.Fatalf("cached sample %d = %d, want %d", i, got, want)
		}
	}
	if cached.currentSample != 132 {
		t.Fatalf("cached current sample = %d, want 132", cached.currentSample)
	}
	if other != w.unit.update || other.currentSample != 0 {
		t.Fatalf("live update was used: cached=%+v other=%+v", *cached, *other)
	}

	healthEvents := 0
	for _, event := range w.events {
		if len(event) >= len("health:") && event[:len("health:")] == "health:" {
			healthEvents++
		}
	}
	if healthEvents != playerHPInitSampleCount4EE730+2 {
		t.Fatalf("HealthData loads = %d, want %d", healthEvents, playerHPInitSampleCount4EE730+2)
	}
}

func TestPlayerHPInit4EE730CachesCurrentBeforeEachStore(t *testing.T) {
	w := newPlayerHPInitWorld4EE730()
	w.afterCurrent = func(_ *playerHPInitTestWorld4EE730, health *playerHPInitTestHealth4EE730) {
		health.current++
	}
	playerHPInit4EE730(w.hooks())

	for i, got := range w.unit.update.samples {
		if want := uint16(17 + i); got != want {
			t.Fatalf("sample %d = %d, want cached %d", i, got, want)
		}
	}
	if w.unit.update.currentSample != 49 || w.unit.health.current != 50 {
		t.Fatalf("trailing/current = %d/%d, want 49/50", w.unit.update.currentSample, w.unit.health.current)
	}
}

func TestPlayerHPInit4EE730UnguardedLivePointers(t *testing.T) {
	t.Run("nil live health", func(t *testing.T) {
		w := newPlayerHPInitWorld4EE730()
		w.afterUpdate = func(w *playerHPInitTestWorld4EE730) { w.unit.health = nil }
		defer func() {
			if recover() == nil {
				t.Fatal("nil live HealthData did not preserve the original fault")
			}
			want := []string{
				"arg:unit", "class:unit=0x4", "health:unit=entry", "update:unit=cached",
				"health:unit=nil", "current:nil",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		playerHPInit4EE730(w.hooks())
	})

	t.Run("nil cached update", func(t *testing.T) {
		w := newPlayerHPInitWorld4EE730()
		w.unit.update = nil
		defer func() {
			if recover() == nil {
				t.Fatal("nil cached UpdateData did not preserve the original fault")
			}
			want := []string{
				"arg:unit", "class:unit=0x4", "health:unit=entry", "update:unit=nil",
				"health:unit=entry", "current:entry", "sample:nil[0]=17",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		playerHPInit4EE730(w.hooks())
	})
}

func playerHPInitExpectedEvents4EE730() []string {
	events := []string{
		"arg:unit", "class:unit=0x4", "health:unit=entry", "update:unit=cached",
	}
	for i := 0; i < playerHPInitSampleCount4EE730; i++ {
		events = append(events,
			"health:unit=entry",
			"current:entry",
			fmt.Sprintf("sample:cached[%d]=17", i),
		)
	}
	return append(events, "health:unit=entry", "current:entry", "current-sample:cached=17")
}

func TestPlayerHPInit4EE730AllFaultPrefixes(t *testing.T) {
	wantEvents := playerHPInitExpectedEvents4EE730()
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("event-%03d", faultAt), func(t *testing.T) {
			w := newPlayerHPInitWorld4EE730()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want %q", w.events, want)
				}
			}()
			playerHPInit4EE730(w.hooks())
		})
	}
}
