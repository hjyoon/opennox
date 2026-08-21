package server

import (
	"reflect"
	"testing"
)

type skullInitTestData4F0450 struct {
	name   string
	typeID uint32
}

type skullInitTestObject4F0450 struct {
	initData   *DirectionInitData
	updateData *skullInitTestData4F0450
	direction1 uint16
	direction2 uint16
}

func skullInitTestHooks4F0450(events *[]string) skullInitHooks4F0450[
	*skullInitTestObject4F0450,
	*DirectionInitData,
	*skullInitTestData4F0450,
] {
	return skullInitHooks4F0450[*skullInitTestObject4F0450, *DirectionInitData, *skullInitTestData4F0450]{
		loadInitData: func(unit *skullInitTestObject4F0450) *DirectionInitData {
			*events = append(*events, "load-init")
			return unit.initData
		},
		loadUpdateData: func(unit *skullInitTestObject4F0450) *skullInitTestData4F0450 {
			*events = append(*events, "load-update")
			return unit.updateData
		},
		directionToAngle: func(data *DirectionInitData) uint32 {
			*events = append(*events, "direction-to-angle")
			return uint32(data.X) ^ uint32(data.Y)
		},
		storeDirection2: func(unit *skullInitTestObject4F0450, angle uint16) {
			*events = append(*events, "store-direction-2")
			unit.direction2 = angle
		},
		storeDirection1: func(unit *skullInitTestObject4F0450, angle uint16) {
			*events = append(*events, "store-direction-1")
			unit.direction1 = angle
		},
		resolveProjectileType: func(update *skullInitTestData4F0450) int32 {
			*events = append(*events, "resolve-projectile:"+update.name)
			return -2
		},
		storeProjectileType: func(update *skullInitTestData4F0450, value uint32) {
			*events = append(*events, "store-projectile-type")
			update.typeID = value
		},
	}
}

func TestDirectionToAngle509E00LoadsYBeforeXAndUsesSignedCenteredIndex(t *testing.T) {
	data := &DirectionInitData{X: -1, Y: 1}
	events := make([]string, 0, 3)
	got := directionToAngle509E00(data, directionToAngleHooks509E00[*DirectionInitData]{
		loadY: func(got *DirectionInitData) int32 {
			events = append(events, "load-y")
			return got.Y
		},
		loadX: func(got *DirectionInitData) int32 {
			events = append(events, "load-x")
			return got.X
		},
		loadTable: func(index int32) uint32 {
			events = append(events, "load-table")
			if index != 2 {
				t.Fatalf("table index = %d, want 2", index)
			}
			return 0x12345678
		},
	})
	if got != 0x12345678 {
		t.Fatalf("angle = %#x", got)
	}
	if want := []string{"load-y", "load-x", "load-table"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDirectionToAngle509E00FaultPrefixes(t *testing.T) {
	for _, stage := range []string{"load-y", "load-x", "load-table"} {
		t.Run(stage, func(t *testing.T) {
			events := make([]string, 0, 3)
			func() {
				defer func() {
					if got := recover(); got != stage {
						t.Fatalf("panic = %v, want %q", got, stage)
					}
				}()
				directionToAngle509E00(struct{}{}, directionToAngleHooks509E00[struct{}]{
					loadY: func(struct{}) int32 {
						events = append(events, "load-y")
						if stage == "load-y" {
							panic(stage)
						}
						return 1
					},
					loadX: func(struct{}) int32 {
						events = append(events, "load-x")
						if stage == "load-x" {
							panic(stage)
						}
						return -1
					},
					loadTable: func(int32) uint32 {
						events = append(events, "load-table")
						if stage == "load-table" {
							panic(stage)
						}
						return 0
					},
				})
			}()
			wantByStage := map[string][]string{
				"load-y":     {"load-y"},
				"load-x":     {"load-y", "load-x"},
				"load-table": {"load-y", "load-x", "load-table"},
			}
			if want := wantByStage[stage]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestSkullInit4F0450CachesPointersAndPreservesStoreOrder(t *testing.T) {
	initialInit := &DirectionInitData{X: int32(0x7fff1234), Y: -1}
	replacementInit := &DirectionInitData{X: 1, Y: 1}
	initialUpdate := &skullInitTestData4F0450{name: "InitialProjectile", typeID: 0x11111111}
	replacementUpdate := &skullInitTestData4F0450{name: "ReplacementProjectile", typeID: 0x22222222}
	unit := &skullInitTestObject4F0450{
		initData: initialInit, updateData: initialUpdate,
		direction1: 0xaaaa, direction2: 0xbbbb,
	}
	events := make([]string, 0, 7)
	hooks := skullInitTestHooks4F0450(&events)
	direction := hooks.directionToAngle
	hooks.directionToAngle = func(data *DirectionInitData) uint32 {
		unit.initData = replacementInit
		unit.updateData = replacementUpdate
		return direction(data)
	}
	resolve := hooks.resolveProjectileType
	hooks.resolveProjectileType = func(update *skullInitTestData4F0450) int32 {
		unit.updateData = replacementUpdate
		return resolve(update)
	}

	got := skullInit4F0450(unit, hooks)
	if got != -2 {
		t.Fatalf("result = %d, want -2", got)
	}
	wantEvents := []string{
		"load-init", "load-update", "direction-to-angle",
		"store-direction-2", "store-direction-1",
		"resolve-projectile:InitialProjectile", "store-projectile-type",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	wantAngle := uint16(uint32(initialInit.X) ^ uint32(initialInit.Y))
	if unit.direction1 != wantAngle || unit.direction2 != wantAngle {
		t.Fatalf("directions = %#x/%#x, want %#x", unit.direction1, unit.direction2, wantAngle)
	}
	if initialUpdate.typeID != 0xfffffffe {
		t.Fatalf("cached update type = %#x, want %#x", initialUpdate.typeID, uint32(0xfffffffe))
	}
	if replacementUpdate.typeID != 0x22222222 {
		t.Fatalf("replacement update changed to %#x", replacementUpdate.typeID)
	}
}

func TestSkullInit4F0450FaultPrefixes(t *testing.T) {
	stages := []string{
		"load-init", "load-update", "direction-to-angle", "store-direction-2",
		"store-direction-1", "resolve-projectile", "store-projectile-type",
	}
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			initData := &DirectionInitData{X: -1, Y: 1}
			update := &skullInitTestData4F0450{name: "Arrow", typeID: 0x11111111}
			unit := &skullInitTestObject4F0450{
				initData: initData, updateData: update,
				direction1: 0xaaaa, direction2: 0xbbbb,
			}
			events := make([]string, 0, 7)
			hooks := skullInitTestHooks4F0450(&events)
			panicAt := func(name string) {
				if stage == name {
					panic(stage)
				}
			}
			loadInit := hooks.loadInitData
			hooks.loadInitData = func(got *skullInitTestObject4F0450) *DirectionInitData {
				value := loadInit(got)
				panicAt("load-init")
				return value
			}
			loadUpdate := hooks.loadUpdateData
			hooks.loadUpdateData = func(got *skullInitTestObject4F0450) *skullInitTestData4F0450 {
				value := loadUpdate(got)
				panicAt("load-update")
				return value
			}
			direction := hooks.directionToAngle
			hooks.directionToAngle = func(got *DirectionInitData) uint32 {
				value := direction(got)
				panicAt("direction-to-angle")
				return value
			}
			hooks.storeDirection2 = func(got *skullInitTestObject4F0450, value uint16) {
				events = append(events, "store-direction-2")
				panicAt("store-direction-2")
				got.direction2 = value
			}
			hooks.storeDirection1 = func(got *skullInitTestObject4F0450, value uint16) {
				events = append(events, "store-direction-1")
				panicAt("store-direction-1")
				got.direction1 = value
			}
			resolve := hooks.resolveProjectileType
			hooks.resolveProjectileType = func(got *skullInitTestData4F0450) int32 {
				value := resolve(got)
				panicAt("resolve-projectile")
				return value
			}
			hooks.storeProjectileType = func(got *skullInitTestData4F0450, value uint32) {
				events = append(events, "store-projectile-type")
				panicAt("store-projectile-type")
				got.typeID = value
			}

			func() {
				defer func() {
					if got := recover(); got != stage {
						t.Fatalf("panic = %v, want %q", got, stage)
					}
				}()
				skullInit4F0450(unit, hooks)
			}()
			stageIndex := map[string]int{
				"load-init": 1, "load-update": 2, "direction-to-angle": 3,
				"store-direction-2": 4, "store-direction-1": 5,
				"resolve-projectile": 6, "store-projectile-type": 7,
			}[stage]
			allEvents := []string{
				"load-init", "load-update", "direction-to-angle",
				"store-direction-2", "store-direction-1",
				"resolve-projectile:Arrow", "store-projectile-type",
			}
			if want := allEvents[:stageIndex]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
			wantDirection2, wantDirection1 := uint16(0xbbbb), uint16(0xaaaa)
			if stageIndex > 4 {
				wantDirection2 = 0xfffe
			}
			if stageIndex > 5 {
				wantDirection1 = 0xfffe
			}
			if unit.direction1 != wantDirection1 || unit.direction2 != wantDirection2 {
				t.Fatalf("directions = %#x/%#x, want %#x/%#x", unit.direction1, unit.direction2, wantDirection1, wantDirection2)
			}
			if update.typeID != 0x11111111 {
				t.Fatalf("projectile type changed before completed store: %#x", update.typeID)
			}
		})
	}
}
