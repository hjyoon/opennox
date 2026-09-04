package opennox

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSpellGestureReset4FCAC0OrderPointersAndLiveIteration(t *testing.T) {
	const (
		allocator = uintptr(0x100001234)
		unit1     = uintptr(0x200002345)
		unit2     = uintptr(0x300003456)
		staleUnit = uintptr(0x400004567)
		update1   = uintptr(0x500005678)
		update2   = uintptr(0x600006789)
		caster    = uintptr(0x70000789a)
	)
	next := map[uintptr]uintptr{unit1: staleUnit, unit2: 0}
	updates := map[uintptr]uintptr{unit1: update1, unit2: update2}
	var events []string
	var casterGlobal uintptr

	got := spellGestureReset4FCAC0(-7, 9, spellGestureResetHooks4FCAC0[uintptr, uintptr, uintptr, uintptr]{
		resetDurations: func(value int32) {
			events = append(events, fmt.Sprintf("reset:%d", value))
		},
		loadMagicClass: func() uintptr {
			events = append(events, "load-magic")
			return allocator
		},
		freeAllMagicObjects: func(value uintptr) {
			events = append(events, fmt.Sprintf("free-magic:%#x", value))
		},
		clearMagicEntityHead: func() {
			events = append(events, "clear-head")
		},
		firstPlayerUnit: func() uintptr {
			events = append(events, "first-unit")
			return unit1
		},
		loadPlayerUpdate: func(unit uintptr) uintptr {
			events = append(events, fmt.Sprintf("load-update:%#x", unit))
			return updates[unit]
		},
		storeField47LowByte: func(update uintptr, value uint8) {
			events = append(events, fmt.Sprintf("field47:%#x:%d", update, value))
		},
		storeSpellCastStart: func(update uintptr, value uint32) {
			events = append(events, fmt.Sprintf("cast-start:%#x:%d", update, value))
		},
		storeTrapSpell: func(update uintptr, index int, value uint32) {
			events = append(events, fmt.Sprintf("trap:%#x:%d:%d", update, index, value))
		},
		storeTrapSpellCountLowByte: func(update uintptr, value uint8) {
			events = append(events, fmt.Sprintf("trap-count:%#x:%d", update, value))
			if update == update1 {
				next[unit1] = unit2
			}
		},
		nextPlayerUnit: func(unit uintptr) uintptr {
			events = append(events, fmt.Sprintf("next:%#x", unit))
			return next[unit]
		},
		newObjectByTypeID: func(name string) uintptr {
			events = append(events, "new-object:"+name)
			return caster
		},
		storeImaginaryCaster: func(value uintptr) {
			events = append(events, fmt.Sprintf("store-caster:%#x", value))
			casterGlobal = value
		},
		createObjectAt: func(object, owner uintptr, x, y float32) {
			events = append(events, fmt.Sprintf("create:%#x:%#x:%g:%g", object, owner, x, y))
			if casterGlobal != caster {
				t.Fatalf("create observed caster global %#x, want %#x", casterGlobal, caster)
			}
		},
	})

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{
		"reset:-7",
		"load-magic",
		"free-magic:0x100001234",
		"clear-head",
		"first-unit",
		"load-update:0x200002345",
		"field47:0x500005678:0",
		"cast-start:0x500005678:0",
		"trap:0x500005678:0:0",
		"trap:0x500005678:1:0",
		"trap:0x500005678:2:0",
		"trap:0x500005678:3:0",
		"trap:0x500005678:4:0",
		"trap-count:0x500005678:0",
		"next:0x200002345",
		"load-update:0x300003456",
		"field47:0x600006789:0",
		"cast-start:0x600006789:0",
		"trap:0x600006789:0:0",
		"trap:0x600006789:1:0",
		"trap:0x600006789:2:0",
		"trap:0x600006789:3:0",
		"trap:0x600006789:4:0",
		"trap-count:0x600006789:0",
		"next:0x300003456",
		"new-object:ImaginaryCaster",
		"store-caster:0x70000789a",
		"create:0x70000789a:0x0:2944:2944",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellGestureReset4FCAC0ForwardsNilAllocatorAndSkipsCaster(t *testing.T) {
	var events []string
	got := spellGestureReset4FCAC0(3, 0, spellGestureResetHooks4FCAC0[uintptr, uintptr, uintptr, uintptr]{
		resetDurations:       func(value int32) { events = append(events, fmt.Sprintf("reset:%d", value)) },
		loadMagicClass:       func() uintptr { events = append(events, "load-magic"); return 0 },
		freeAllMagicObjects:  func(value uintptr) { events = append(events, fmt.Sprintf("free-magic:%#x", value)) },
		clearMagicEntityHead: func() { events = append(events, "clear-head") },
		firstPlayerUnit:      func() uintptr { events = append(events, "first-unit"); return 0 },
		newObjectByTypeID: func(string) uintptr {
			t.Fatal("caster lookup must be skipped")
			return 0
		},
		storeImaginaryCaster: func(uintptr) { t.Fatal("caster store must be skipped") },
		createObjectAt:       func(uintptr, uintptr, float32, float32) { t.Fatal("create must be skipped") },
	})

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	want := []string{"reset:3", "load-magic", "free-magic:0x0", "clear-head", "first-unit"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellGestureReset4FCAC0StoresNilCasterBeforeFailure(t *testing.T) {
	casterGlobal := uintptr(0x100000001)
	var events []string
	got := spellGestureReset4FCAC0(0, 1, spellGestureResetHooks4FCAC0[uintptr, uintptr, uintptr, uintptr]{
		resetDurations:       func(int32) { events = append(events, "reset") },
		loadMagicClass:       func() uintptr { events = append(events, "load-magic"); return 0 },
		freeAllMagicObjects:  func(uintptr) { events = append(events, "free-magic") },
		clearMagicEntityHead: func() { events = append(events, "clear-head") },
		firstPlayerUnit:      func() uintptr { events = append(events, "first-unit"); return 0 },
		newObjectByTypeID: func(name string) uintptr {
			events = append(events, "new-object:"+name)
			return 0
		},
		storeImaginaryCaster: func(value uintptr) {
			events = append(events, fmt.Sprintf("store-caster:%#x", value))
			casterGlobal = value
		},
		createObjectAt: func(uintptr, uintptr, float32, float32) {
			t.Fatal("create must be skipped after a nil caster result")
		},
	})

	if got != 0 || casterGlobal != 0 {
		t.Fatalf("result/caster = (%d, %#x), want (0, 0)", got, casterGlobal)
	}
	want := []string{
		"reset", "load-magic", "free-magic", "clear-head", "first-unit",
		"new-object:ImaginaryCaster", "store-caster:0x0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestSpellGestureReset4FCAC0FaultPrefixes(t *testing.T) {
	allEvents := []string{
		"reset", "load-magic", "free-magic", "clear-head", "first-unit",
		"load-update", "field47", "cast-start",
		"trap-0", "trap-1", "trap-2", "trap-3", "trap-4", "trap-count", "next-unit",
		"new-object", "store-caster", "create",
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
				spellGestureReset4FCAC0(-11, 1, spellGestureResetHooks4FCAC0[uintptr, uintptr, uintptr, uintptr]{
					resetDurations: func(value int32) {
						if value != -11 {
							t.Fatalf("duration argument = %d", value)
						}
						observe("reset")
					},
					loadMagicClass: func() uintptr { observe("load-magic"); return 0x100000001 },
					freeAllMagicObjects: func(value uintptr) {
						if value != 0x100000001 {
							t.Fatalf("allocator = %#x", value)
						}
						observe("free-magic")
					},
					clearMagicEntityHead: func() { observe("clear-head") },
					firstPlayerUnit:      func() uintptr { observe("first-unit"); return 0x200000002 },
					loadPlayerUpdate: func(unit uintptr) uintptr {
						if unit != 0x200000002 {
							t.Fatalf("unit = %#x", unit)
						}
						observe("load-update")
						return 0x300000003
					},
					storeField47LowByte: func(update uintptr, value uint8) {
						if update != 0x300000003 || value != 0 {
							t.Fatalf("field47 args = (%#x, %d)", update, value)
						}
						observe("field47")
					},
					storeSpellCastStart: func(update uintptr, value uint32) {
						if update != 0x300000003 || value != 0 {
							t.Fatalf("cast-start args = (%#x, %d)", update, value)
						}
						observe("cast-start")
					},
					storeTrapSpell: func(update uintptr, index int, value uint32) {
						if update != 0x300000003 || value != 0 {
							t.Fatalf("trap args = (%#x, %d, %d)", update, index, value)
						}
						observe(fmt.Sprintf("trap-%d", index))
					},
					storeTrapSpellCountLowByte: func(update uintptr, value uint8) {
						if update != 0x300000003 || value != 0 {
							t.Fatalf("trap-count args = (%#x, %d)", update, value)
						}
						observe("trap-count")
					},
					nextPlayerUnit: func(unit uintptr) uintptr {
						if unit != 0x200000002 {
							t.Fatalf("next unit = %#x", unit)
						}
						observe("next-unit")
						return 0
					},
					newObjectByTypeID: func(name string) uintptr {
						if name != spellRuntimeCasterType4FC9B0 {
							t.Fatalf("object type = %q", name)
						}
						observe("new-object")
						return 0x400000004
					},
					storeImaginaryCaster: func(value uintptr) {
						if value != 0x400000004 {
							t.Fatalf("caster = %#x", value)
						}
						observe("store-caster")
					},
					createObjectAt: func(object, owner uintptr, x, y float32) {
						if object != 0x400000004 || owner != 0 || x != 2944 || y != 2944 {
							t.Fatalf("create args = (%#x, %#x, %g, %g)", object, owner, x, y)
						}
						observe("create")
					},
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
