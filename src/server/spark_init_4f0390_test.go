package server

import (
	"reflect"
	"testing"
)

type sparkInitTestData4F0390 struct {
	initial   uint32
	remaining uint32
	guard     uint32
}

type sparkInitTestObject4F0390 struct {
	update *sparkInitTestData4F0390
}

func defaultSparkInitHooks4F0390(events *[]string) sparkInitHooks4F0390[
	*sparkInitTestObject4F0390,
	*sparkInitTestData4F0390,
] {
	return sparkInitHooks4F0390[
		*sparkInitTestObject4F0390,
		*sparkInitTestData4F0390,
	]{
		loadUpdateData: func(unit *sparkInitTestObject4F0390) *sparkInitTestData4F0390 {
			*events = append(*events, "load-update")
			return unit.update
		},
		storeLifetimeRemaining: func(update *sparkInitTestData4F0390, value uint32) {
			*events = append(*events, "store-remaining")
			update.remaining = value
		},
		storeLifetimeInitial: func(update *sparkInitTestData4F0390, value uint32) {
			*events = append(*events, "store-initial")
			update.initial = value
		},
	}
}

func TestSparkInit4F0390CachesUpdateStoresRemainingFirstAndReturnsIt(t *testing.T) {
	entry := &sparkInitTestData4F0390{
		initial:   0x11111111,
		remaining: 0x22222222,
		guard:     0xa5a5a5a5,
	}
	live := &sparkInitTestData4F0390{
		initial:   0x33333333,
		remaining: 0x44444444,
		guard:     0x5a5a5a5a,
	}
	unit := &sparkInitTestObject4F0390{update: entry}
	events := make([]string, 0, 3)
	hooks := defaultSparkInitHooks4F0390(&events)
	load := hooks.loadUpdateData
	hooks.loadUpdateData = func(got *sparkInitTestObject4F0390) *sparkInitTestData4F0390 {
		update := load(got)
		got.update = live
		return update
	}

	got := sparkInit4F0390(unit, hooks)
	if got != entry {
		t.Fatalf("return = %p, want cached %p", got, entry)
	}
	if !reflect.DeepEqual(events, []string{"load-update", "store-remaining", "store-initial"}) {
		t.Fatalf("events = %v", events)
	}
	if entry.initial != sparkInitLifetime4F0390 || entry.remaining != sparkInitLifetime4F0390 {
		t.Fatalf("entry lifetimes = %d/%d", entry.initial, entry.remaining)
	}
	if entry.guard != 0xa5a5a5a5 {
		t.Fatalf("entry guard = %#x", entry.guard)
	}
	if live.initial != 0x33333333 || live.remaining != 0x44444444 || live.guard != 0x5a5a5a5a {
		t.Fatalf("live update was touched: %+v", *live)
	}
}

func TestSparkInit4F0390FaultPrefixes(t *testing.T) {
	tests := []struct {
		name       string
		panicStage string
		wantEvents []string
	}{
		{name: "load update", panicStage: "load-update", wantEvents: []string{"load-update"}},
		{name: "store remaining", panicStage: "store-remaining", wantEvents: []string{"load-update", "store-remaining"}},
		{name: "store initial", panicStage: "store-initial", wantEvents: []string{"load-update", "store-remaining", "store-initial"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := &sparkInitTestObject4F0390{update: &sparkInitTestData4F0390{}}
			events := make([]string, 0, 3)
			hooks := defaultSparkInitHooks4F0390(&events)
			load := hooks.loadUpdateData
			hooks.loadUpdateData = func(got *sparkInitTestObject4F0390) *sparkInitTestData4F0390 {
				update := load(got)
				if tc.panicStage == "load-update" {
					panic(tc.panicStage)
				}
				return update
			}
			hooks.storeLifetimeRemaining = func(update *sparkInitTestData4F0390, value uint32) {
				events = append(events, "store-remaining")
				if tc.panicStage == "store-remaining" {
					panic(tc.panicStage)
				}
				update.remaining = value
			}
			hooks.storeLifetimeInitial = func(update *sparkInitTestData4F0390, value uint32) {
				events = append(events, "store-initial")
				if tc.panicStage == "store-initial" {
					panic(tc.panicStage)
				}
				update.initial = value
			}

			func() {
				defer func() {
					if got := recover(); got != tc.panicStage {
						t.Fatalf("panic = %v, want %q", got, tc.panicStage)
					}
				}()
				sparkInit4F0390(unit, hooks)
			}()
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}
