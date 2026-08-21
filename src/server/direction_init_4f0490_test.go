package server

import (
	"reflect"
	"testing"
)

type directionInitTestObject4F0490 struct {
	initData   *DirectionInitData
	direction1 uint16
	direction2 uint16
	guard      uint32
}

func directionInitTestHooks4F0490(events *[]string) directionInitHooks4F0490[
	*directionInitTestObject4F0490,
	*DirectionInitData,
] {
	return directionInitHooks4F0490[*directionInitTestObject4F0490, *DirectionInitData]{
		loadInitData: func(unit *directionInitTestObject4F0490) *DirectionInitData {
			*events = append(*events, "load-init")
			return unit.initData
		},
		directionToAngle: func(data *DirectionInitData) uint32 {
			*events = append(*events, "direction-to-angle")
			return uint32(data.X) ^ uint32(data.Y)
		},
		storeDirection2: func(unit *directionInitTestObject4F0490, angle uint16) {
			*events = append(*events, "store-direction-2")
			unit.direction2 = angle
		},
		storeDirection1: func(unit *directionInitTestObject4F0490, angle uint16) {
			*events = append(*events, "store-direction-1")
			unit.direction1 = angle
		},
	}
}

func TestDirectionInit4F0490CachesInitDataAndPreservesFullReturn(t *testing.T) {
	initial := &DirectionInitData{X: int32(0x7fff1234), Y: -1}
	replacement := &DirectionInitData{X: 1, Y: 1}
	unit := &directionInitTestObject4F0490{
		initData: initial, direction1: 0xaaaa, direction2: 0xbbbb,
		guard: 0xa5a5a5a5,
	}
	events := make([]string, 0, 4)
	hooks := directionInitTestHooks4F0490(&events)
	direction := hooks.directionToAngle
	hooks.directionToAngle = func(data *DirectionInitData) uint32 {
		unit.initData = replacement
		unit.direction1 = 0x1111
		unit.direction2 = 0x2222
		return direction(data)
	}

	got := directionInit4F0490(unit, hooks)
	wantBits := uint32(initial.X) ^ uint32(initial.Y)
	if uint32(got) != wantBits {
		t.Fatalf("result bits = %#x, want %#x", uint32(got), wantBits)
	}
	if want := []string{"load-init", "direction-to-angle", "store-direction-2", "store-direction-1"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if unit.initData != replacement {
		t.Fatal("helper mutation of live InitData was lost")
	}
	if unit.direction1 != uint16(wantBits) || unit.direction2 != uint16(wantBits) {
		t.Fatalf("directions = %#x/%#x, want %#x", unit.direction1, unit.direction2, uint16(wantBits))
	}
	if unit.guard != 0xa5a5a5a5 {
		t.Fatalf("guard = %#x", unit.guard)
	}
}

func TestDirectionInit4F0490FaultPrefixes(t *testing.T) {
	stages := []string{"load-init", "direction-to-angle", "store-direction-2", "store-direction-1"}
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			unit := &directionInitTestObject4F0490{
				initData:   &DirectionInitData{X: -1, Y: 1},
				direction1: 0xaaaa,
				direction2: 0xbbbb,
			}
			events := make([]string, 0, 4)
			hooks := directionInitTestHooks4F0490(&events)
			panicAt := func(name string) {
				if stage == name {
					panic(stage)
				}
			}
			loadInit := hooks.loadInitData
			hooks.loadInitData = func(got *directionInitTestObject4F0490) *DirectionInitData {
				value := loadInit(got)
				panicAt("load-init")
				return value
			}
			direction := hooks.directionToAngle
			hooks.directionToAngle = func(got *DirectionInitData) uint32 {
				value := direction(got)
				panicAt("direction-to-angle")
				return value
			}
			hooks.storeDirection2 = func(got *directionInitTestObject4F0490, value uint16) {
				events = append(events, "store-direction-2")
				panicAt("store-direction-2")
				got.direction2 = value
			}
			hooks.storeDirection1 = func(got *directionInitTestObject4F0490, value uint16) {
				events = append(events, "store-direction-1")
				panicAt("store-direction-1")
				got.direction1 = value
			}

			func() {
				defer func() {
					if got := recover(); got != stage {
						t.Fatalf("panic = %v, want %q", got, stage)
					}
				}()
				directionInit4F0490(unit, hooks)
			}()
			stageIndex := map[string]int{
				"load-init": 1, "direction-to-angle": 2,
				"store-direction-2": 3, "store-direction-1": 4,
			}[stage]
			allEvents := []string{"load-init", "direction-to-angle", "store-direction-2", "store-direction-1"}
			if want := allEvents[:stageIndex]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
			wantDirection2, wantDirection1 := uint16(0xbbbb), uint16(0xaaaa)
			if stageIndex > 3 {
				wantDirection2 = 0xfffe
			}
			if unit.direction1 != wantDirection1 || unit.direction2 != wantDirection2 {
				t.Fatalf("directions = %#x/%#x, want %#x/%#x", unit.direction1, unit.direction2, wantDirection1, wantDirection2)
			}
		})
	}
}
