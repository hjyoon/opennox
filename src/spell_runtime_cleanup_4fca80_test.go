package opennox

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSpellRuntimeCleanup4FCA80OrderPointersAndState(t *testing.T) {
	const (
		allocator = uintptr(0x100001234)
		caster    = uintptr(0x200005678)
	)
	magicGlobal := allocator
	casterGlobal := caster
	headPresent := true
	var events []string

	got := spellRuntimeCleanup4FCA80(spellRuntimeCleanupHooks4FCA80[uintptr, uintptr]{
		freeDurations: func() {
			events = append(events, "free-durations")
		},
		loadMagicClass: func() uintptr {
			events = append(events, "load-magic")
			value := magicGlobal
			magicGlobal = 0x300009abc
			return value
		},
		freeMagicClass: func(value uintptr) {
			events = append(events, fmt.Sprintf("free-magic:%#x", value))
			if value != allocator {
				t.Fatalf("magic allocator = %#x, want cached %#x", value, allocator)
			}
		},
		loadImaginaryCaster: func() uintptr {
			events = append(events, "load-caster")
			return casterGlobal
		},
		clearMagicEntityHead: func() {
			events = append(events, "clear-head")
			headPresent = false
		},
		delayedDelete: func(value uintptr) {
			events = append(events, fmt.Sprintf("delayed-delete:%#x", value))
			if value != caster {
				t.Fatalf("caster = %#x, want cached %#x", value, caster)
			}
			if headPresent {
				t.Fatal("delayed delete observed uncleared queue head")
			}
			if casterGlobal != caster {
				t.Fatalf("delayed delete observed caster global %#x, want %#x", casterGlobal, caster)
			}
			if magicGlobal != 0x300009abc {
				t.Fatalf("cleanup overwrote magic allocator global: %#x", magicGlobal)
			}
		},
		clearImaginaryCaster: func() {
			events = append(events, "clear-caster")
			casterGlobal = 0
		},
	})

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	if casterGlobal != 0 || headPresent {
		t.Fatalf("final caster/head = (%#x, %t), want (0, false)", casterGlobal, headPresent)
	}
	wantEvents := []string{
		"free-durations",
		"load-magic",
		"free-magic:0x100001234",
		"load-caster",
		"clear-head",
		"delayed-delete:0x200005678",
		"clear-caster",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
}

func TestSpellRuntimeCleanup4FCA80ForwardsZeroTokens(t *testing.T) {
	var events []string
	got := spellRuntimeCleanup4FCA80(spellRuntimeCleanupHooks4FCA80[uintptr, uintptr]{
		freeDurations: func() { events = append(events, "free-durations") },
		loadMagicClass: func() uintptr {
			events = append(events, "load-magic")
			return 0
		},
		freeMagicClass: func(value uintptr) {
			events = append(events, fmt.Sprintf("free-magic:%#x", value))
		},
		loadImaginaryCaster: func() uintptr {
			events = append(events, "load-caster")
			return 0
		},
		clearMagicEntityHead: func() { events = append(events, "clear-head") },
		delayedDelete: func(value uintptr) {
			events = append(events, fmt.Sprintf("delayed-delete:%#x", value))
		},
		clearImaginaryCaster: func() { events = append(events, "clear-caster") },
	})

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{
		"free-durations",
		"load-magic",
		"free-magic:0x0",
		"load-caster",
		"clear-head",
		"delayed-delete:0x0",
		"clear-caster",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellRuntimeCleanup4FCA80FaultPrefixes(t *testing.T) {
	allEvents := []string{
		"free-durations",
		"load-magic",
		"free-magic",
		"load-caster",
		"clear-head",
		"delayed-delete",
		"clear-caster",
	}
	stop := &struct{}{}

	for failAt := range allEvents {
		t.Run(fmt.Sprintf("fault-%02d-%s", failAt, allEvents[failAt]), func(t *testing.T) {
			events := make([]string, 0, len(allEvents))
			observe := func(event string) {
				if len(events) == failAt {
					panic(stop)
				}
				events = append(events, event)
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				spellRuntimeCleanup4FCA80(spellRuntimeCleanupHooks4FCA80[uintptr, uintptr]{
					freeDurations: func() { observe("free-durations") },
					loadMagicClass: func() uintptr {
						observe("load-magic")
						return 0x100000001
					},
					freeMagicClass: func(value uintptr) {
						if value != 0x100000001 {
							t.Fatalf("magic allocator = %#x", value)
						}
						observe("free-magic")
					},
					loadImaginaryCaster: func() uintptr {
						observe("load-caster")
						return 0x200000002
					},
					clearMagicEntityHead: func() { observe("clear-head") },
					delayedDelete: func(value uintptr) {
						if value != 0x200000002 {
							t.Fatalf("caster = %#x", value)
						}
						observe("delayed-delete")
					},
					clearImaginaryCaster: func() { observe("clear-caster") },
				})
			}()

			if recovered != stop {
				t.Fatalf("recovered = %#v, want sentinel", recovered)
			}
			if want := allEvents[:failAt]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %#v, want prefix %#v", events, want)
			}
		})
	}
}
