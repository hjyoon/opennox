package server

import (
	"reflect"
	"testing"
)

type frogInitTestData4F03B0 struct {
	delay uint8
	byte1 uint8
	byte2 uint8
	guard uint8
}

type frogInitTestObject4F03B0 struct {
	update     *frogInitTestData4F03B0
	direction2 uint16
}

func frogInitTestHooks4F03B0(events *[]string) frogInitHooks4F03B0[
	*frogInitTestObject4F03B0,
	*frogInitTestData4F03B0,
] {
	return frogInitHooks4F03B0[
		*frogInitTestObject4F03B0,
		*frogInitTestData4F03B0,
	]{
		loadUpdateData: func(unit *frogInitTestObject4F03B0) *frogInitTestData4F03B0 {
			*events = append(*events, "load-update")
			return unit.update
		},
		storeDelay: func(update *frogInitTestData4F03B0, value uint8) {
			*events = append(*events, "store-delay")
			update.delay = value
		},
		storeByte1: func(update *frogInitTestData4F03B0, value uint8) {
			*events = append(*events, "store-byte1")
			update.byte1 = value
		},
		storeByte2: func(update *frogInitTestData4F03B0, value uint8) {
			*events = append(*events, "store-byte2")
			update.byte2 = value
		},
		storeDirection: func(unit *frogInitTestObject4F03B0, value uint16) {
			*events = append(*events, "store-direction")
			unit.direction2 = value
		},
	}
}

func TestFrogInit4F03B0CachesUpdatePreservesOrderAndNarrowsResults(t *testing.T) {
	entry := &frogInitTestData4F03B0{delay: 0x11, byte1: 0x22, byte2: 0x33, guard: 0xa5}
	live := &frogInitTestData4F03B0{delay: 0x44, byte1: 0x55, byte2: 0x66, guard: 0x5a}
	unit := &frogInitTestObject4F03B0{update: entry, direction2: 0x7777}
	events := make([]string, 0, 7)
	hooks := frogInitTestHooks4F03B0(&events)
	randomCall := 0
	hooks.randomInt = func(minimum, maximum int32, path string, line int32) int32 {
		randomCall++
		switch randomCall {
		case 1:
			events = append(events, "random-delay")
			if minimum != frogInitDelayMinimum4F03B0 || maximum != frogInitDelayMaximum4F03B0 ||
				path != frogInitDelayRandomPath4F03B0 || line != frogInitDelayRandomLine4F03B0 {
				t.Fatalf("delay RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
			}
			unit.update = live
			return 0x1234563a
		case 2:
			events = append(events, "random-direction")
			if minimum != frogInitDirectionMinimum4F03B0 || maximum != frogInitDirectionMaximum4F03B0 ||
				path != frogInitDirectionRandomPath4F03B0 || line != frogInitDirectionRandomLine4F03B0 {
				t.Fatalf("direction RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
			}
			if entry.delay != 0x3a || entry.byte1 != 1 || entry.byte2 != 0 {
				t.Fatalf("second RNG observed entry bytes = %#x/%#x/%#x", entry.delay, entry.byte1, entry.byte2)
			}
			return 0x1234abcd
		default:
			t.Fatalf("unexpected RNG call %d", randomCall)
			return 0
		}
	}

	got := frogInit4F03B0(unit, hooks)
	if got != 0x1234abcd {
		t.Fatalf("return = %#x, want full second RNG result", got)
	}
	wantEvents := []string{
		"load-update",
		"random-delay",
		"store-delay",
		"store-byte1",
		"store-byte2",
		"random-direction",
		"store-direction",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if entry.delay != 0x3a || entry.byte1 != 1 || entry.byte2 != 0 || entry.guard != 0xa5 {
		t.Fatalf("entry update = %+v", *entry)
	}
	if *live != (frogInitTestData4F03B0{delay: 0x44, byte1: 0x55, byte2: 0x66, guard: 0x5a}) {
		t.Fatalf("live update was touched: %+v", *live)
	}
	if unit.direction2 != 0xabcd {
		t.Fatalf("Direction2 = %#x, want low AX 0xabcd", unit.direction2)
	}
}

func TestFrogInit4F03B0FaultPrefixes(t *testing.T) {
	tests := []struct {
		stage      string
		wantEvents []string
	}{
		{stage: "load-update", wantEvents: []string{"load-update"}},
		{stage: "random-delay", wantEvents: []string{"load-update", "random-delay"}},
		{stage: "store-delay", wantEvents: []string{"load-update", "random-delay", "store-delay"}},
		{stage: "store-byte1", wantEvents: []string{"load-update", "random-delay", "store-delay", "store-byte1"}},
		{stage: "store-byte2", wantEvents: []string{"load-update", "random-delay", "store-delay", "store-byte1", "store-byte2"}},
		{stage: "random-direction", wantEvents: []string{"load-update", "random-delay", "store-delay", "store-byte1", "store-byte2", "random-direction"}},
		{stage: "store-direction", wantEvents: []string{"load-update", "random-delay", "store-delay", "store-byte1", "store-byte2", "random-direction", "store-direction"}},
	}
	for _, tc := range tests {
		t.Run(tc.stage, func(t *testing.T) {
			unit := &frogInitTestObject4F03B0{update: &frogInitTestData4F03B0{}}
			events := make([]string, 0, 7)
			hooks := frogInitTestHooks4F03B0(&events)
			load := hooks.loadUpdateData
			hooks.loadUpdateData = func(unit *frogInitTestObject4F03B0) *frogInitTestData4F03B0 {
				update := load(unit)
				if tc.stage == "load-update" {
					panic(tc.stage)
				}
				return update
			}
			hooks.randomInt = func(minimum, maximum int32, _ string, _ int32) int32 {
				stage := "random-direction"
				if minimum == frogInitDelayMinimum4F03B0 && maximum == frogInitDelayMaximum4F03B0 {
					stage = "random-delay"
				}
				events = append(events, stage)
				if tc.stage == stage {
					panic(tc.stage)
				}
				return minimum
			}
			hooks.storeDelay = func(update *frogInitTestData4F03B0, value uint8) {
				events = append(events, "store-delay")
				if tc.stage == "store-delay" {
					panic(tc.stage)
				}
				update.delay = value
			}
			hooks.storeByte1 = func(update *frogInitTestData4F03B0, value uint8) {
				events = append(events, "store-byte1")
				if tc.stage == "store-byte1" {
					panic(tc.stage)
				}
				update.byte1 = value
			}
			hooks.storeByte2 = func(update *frogInitTestData4F03B0, value uint8) {
				events = append(events, "store-byte2")
				if tc.stage == "store-byte2" {
					panic(tc.stage)
				}
				update.byte2 = value
			}
			hooks.storeDirection = func(unit *frogInitTestObject4F03B0, value uint16) {
				events = append(events, "store-direction")
				if tc.stage == "store-direction" {
					panic(tc.stage)
				}
				unit.direction2 = value
			}

			func() {
				defer func() {
					if got := recover(); got != tc.stage {
						t.Fatalf("panic = %v, want %q", got, tc.stage)
					}
				}()
				frogInit4F03B0(unit, hooks)
			}()
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}
