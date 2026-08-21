package server

import (
	"reflect"
	"testing"
)

type boulderInitTestObject4F0420 struct {
	sourceX uint32
	sourceY uint32
	targetX uint32
	targetY uint32
	guard   uint32
}

func boulderInitTestHooks4F0420(events *[]string) boulderInitHooks4F0420[
	*boulderInitTestObject4F0420,
	uint32,
] {
	return boulderInitHooks4F0420[*boulderInitTestObject4F0420, uint32]{
		loadSourceX: func(unit *boulderInitTestObject4F0420) uint32 {
			*events = append(*events, "load-source-x")
			return unit.sourceX
		},
		loadSourceY: func(unit *boulderInitTestObject4F0420) uint32 {
			*events = append(*events, "load-source-y")
			return unit.sourceY
		},
		storeTargetX: func(unit *boulderInitTestObject4F0420, value uint32) {
			*events = append(*events, "store-target-x")
			unit.targetX = value
		},
		storeTargetY: func(unit *boulderInitTestObject4F0420, value uint32) {
			*events = append(*events, "store-target-y")
			unit.targetY = value
		},
	}
}

func TestBoulderInit4F0420LoadsBothCoordinatesBeforeStoresAndReturnsUnit(t *testing.T) {
	unit := &boulderInitTestObject4F0420{
		sourceX: 0x7fa12345,
		sourceY: 0x80000000,
		targetX: 0x11111111,
		targetY: 0x22222222,
		guard:   0xa5a5a5a5,
	}
	events := make([]string, 0, 4)
	hooks := boulderInitTestHooks4F0420(&events)
	storeX := hooks.storeTargetX
	hooks.storeTargetX = func(got *boulderInitTestObject4F0420, value uint32) {
		got.sourceX = 0x33333333
		got.sourceY = 0x44444444
		storeX(got, value)
	}

	got := boulderInit4F0420(unit, hooks)
	if got != unit {
		t.Fatalf("return = %p, want entry unit %p", got, unit)
	}
	wantEvents := []string{"load-source-x", "load-source-y", "store-target-x", "store-target-y"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if unit.targetX != 0x7fa12345 || unit.targetY != 0x80000000 {
		t.Fatalf("target bits = %#x/%#x, want entry source bits", unit.targetX, unit.targetY)
	}
	if unit.sourceX != 0x33333333 || unit.sourceY != 0x44444444 {
		t.Fatalf("source mutation was lost: %#x/%#x", unit.sourceX, unit.sourceY)
	}
	if unit.guard != 0xa5a5a5a5 {
		t.Fatalf("guard = %#x", unit.guard)
	}
}

func TestBoulderInit4F0420FaultPrefixes(t *testing.T) {
	tests := []struct {
		stage       string
		wantEvents  []string
		wantTargetX uint32
		wantTargetY uint32
	}{
		{stage: "load-source-x", wantEvents: []string{"load-source-x"}, wantTargetX: 0x11111111, wantTargetY: 0x22222222},
		{stage: "load-source-y", wantEvents: []string{"load-source-x", "load-source-y"}, wantTargetX: 0x11111111, wantTargetY: 0x22222222},
		{stage: "store-target-x", wantEvents: []string{"load-source-x", "load-source-y", "store-target-x"}, wantTargetX: 0x11111111, wantTargetY: 0x22222222},
		{stage: "store-target-y", wantEvents: []string{"load-source-x", "load-source-y", "store-target-x", "store-target-y"}, wantTargetX: 0xaaaa5555, wantTargetY: 0x22222222},
	}
	for _, tc := range tests {
		t.Run(tc.stage, func(t *testing.T) {
			unit := &boulderInitTestObject4F0420{
				sourceX: 0xaaaa5555,
				sourceY: 0x5555aaaa,
				targetX: 0x11111111,
				targetY: 0x22222222,
			}
			events := make([]string, 0, 4)
			hooks := boulderInitTestHooks4F0420(&events)
			loadX := hooks.loadSourceX
			hooks.loadSourceX = func(got *boulderInitTestObject4F0420) uint32 {
				value := loadX(got)
				if tc.stage == "load-source-x" {
					panic(tc.stage)
				}
				return value
			}
			loadY := hooks.loadSourceY
			hooks.loadSourceY = func(got *boulderInitTestObject4F0420) uint32 {
				value := loadY(got)
				if tc.stage == "load-source-y" {
					panic(tc.stage)
				}
				return value
			}
			hooks.storeTargetX = func(got *boulderInitTestObject4F0420, value uint32) {
				events = append(events, "store-target-x")
				if tc.stage == "store-target-x" {
					panic(tc.stage)
				}
				got.targetX = value
			}
			hooks.storeTargetY = func(got *boulderInitTestObject4F0420, value uint32) {
				events = append(events, "store-target-y")
				if tc.stage == "store-target-y" {
					panic(tc.stage)
				}
				got.targetY = value
			}

			func() {
				defer func() {
					if got := recover(); got != tc.stage {
						t.Fatalf("panic = %v, want %q", got, tc.stage)
					}
				}()
				boulderInit4F0420(unit, hooks)
			}()
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
			if unit.targetX != tc.wantTargetX || unit.targetY != tc.wantTargetY {
				t.Fatalf("targets = %#x/%#x, want %#x/%#x", unit.targetX, unit.targetY, tc.wantTargetX, tc.wantTargetY)
			}
		})
	}
}
