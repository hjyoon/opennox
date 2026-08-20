package server

import (
	"fmt"
	"reflect"
	"testing"
)

type unitAdjustHPWorld4EE460 struct {
	flag         int32
	unit         string
	healthByUnit map[string]string
	current      map[string]uint16
	maximum      map[string]uint16
	delta        int32
	classLow     map[string]uint8
	events       []string
	faultAt      int
	afterEvent   map[string]func()
	setValues    []uint16
}

func newUnitAdjustHPWorld4EE460() *unitAdjustHPWorld4EE460 {
	return &unitAdjustHPWorld4EE460{
		healthByUnit: make(map[string]string),
		current:      make(map[string]uint16),
		maximum:      make(map[string]uint16),
		classLow:     make(map[string]uint8),
		afterEvent:   make(map[string]func()),
	}
}

func (w *unitAdjustHPWorld4EE460) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *unitAdjustHPWorld4EE460) hooks() unitAdjustHPHooks4EE460[string, string] {
	return unitAdjustHPHooks4EE460[string, string]{
		gameFlag: func(flag uint32) int32 {
			value := w.flag
			w.record(fmt.Sprintf("game-flag:%08x=%d", flag, value))
			return value
		},
		loadUnitArg: func() string {
			unit := w.unit
			w.record("load-unit=" + unit)
			return unit
		},
		loadHealth: func(unit string) string {
			health := w.healthByUnit[unit]
			w.record("load-health:" + unit + "=" + health)
			if unit == "" {
				panic("nil-unit-health")
			}
			return health
		},
		loadCurrent: func(health string) uint16 {
			value := w.current[health]
			w.record(fmt.Sprintf("load-current:%s=%d", health, value))
			return value
		},
		loadMaximum: func(health string) uint16 {
			value := w.maximum[health]
			w.record(fmt.Sprintf("load-maximum:%s=%d", health, value))
			return value
		},
		loadDeltaArg: func() int32 {
			value := w.delta
			w.record(fmt.Sprintf("load-delta=%d", value))
			return value
		},
		setHP: func(unit string, value uint16) {
			w.setValues = append(w.setValues, value)
			w.record(fmt.Sprintf("set-hp:%s=%d", unit, value))
		},
		loadClassLow: func(unit string) uint8 {
			value := w.classLow[unit]
			w.record(fmt.Sprintf("load-class:%s=%02x", unit, value))
			return value
		},
		informOwnerHP: func(unit string) {
			w.record("inform-owner:" + unit)
		},
	}
}

func TestUnitAdjustHP4EE460FlagPrecedesUnitAndNullFault(t *testing.T) {
	w := newUnitAdjustHPWorld4EE460()
	w.flag = -1
	unitAdjustHP4EE460(w.hooks())
	want := []string{"game-flag:04000000=-1"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	w = newUnitAdjustHPWorld4EE460()
	defer func() {
		if got := recover(); got != "nil-unit-health" {
			t.Fatalf("panic = %v, want nil-unit-health", got)
		}
		want := []string{"game-flag:04000000=0", "load-unit=", "load-health:="}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	unitAdjustHP4EE460(w.hooks())
}

func TestUnitAdjustHP4EE460EarlyHealthBranchesDelayDelta(t *testing.T) {
	for _, tc := range []struct {
		name    string
		health  string
		current uint16
		maximum uint16
		want    []string
	}{
		{
			name:   "null-health",
			health: "",
			want: []string{
				"game-flag:04000000=0", "load-unit=unit", "load-health:unit=",
			},
		},
		{
			name:   "equal",
			health: "health", current: 20, maximum: 20,
			want: []string{
				"game-flag:04000000=0", "load-unit=unit", "load-health:unit=health",
				"load-current:health=20", "load-maximum:health=20",
			},
		},
		{
			name:   "above",
			health: "health", current: 21, maximum: 20,
			want: []string{
				"game-flag:04000000=0", "load-unit=unit", "load-health:unit=health",
				"load-current:health=21", "load-maximum:health=20",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newUnitAdjustHPWorld4EE460()
			w.unit = "unit"
			w.healthByUnit["unit"] = tc.health
			w.current[tc.health] = tc.current
			w.maximum[tc.health] = tc.maximum
			w.delta = 99
			unitAdjustHP4EE460(w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
			if len(w.setValues) != 0 {
				t.Fatalf("set values = %v, want none", w.setValues)
			}
		})
	}
}

func TestUnitAdjustHP4EE460WrapClampAndLowWord(t *testing.T) {
	for _, tc := range []struct {
		name    string
		current uint16
		maximum uint16
		delta   int32
		want    uint16
	}{
		{name: "normal", current: 10, maximum: 20, delta: 3, want: 13},
		{name: "exact-max", current: 10, maximum: 20, delta: 10, want: 20},
		{name: "above-max", current: 10, maximum: 20, delta: 100, want: 20},
		{name: "negative-low-word", current: 10, maximum: 20, delta: -11, want: 0xffff},
		{name: "signed-overflow", current: 1, maximum: 20, delta: int32(^uint32(0) >> 1), want: 0},
		{name: "minimum", current: 1, maximum: 20, delta: -1 << 31, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newUnitAdjustHPWorld4EE460()
			w.unit = "unit"
			w.healthByUnit["unit"] = "health"
			w.current["health"] = tc.current
			w.maximum["health"] = tc.maximum
			w.delta = tc.delta
			unitAdjustHP4EE460(w.hooks())
			if !reflect.DeepEqual(w.setValues, []uint16{tc.want}) {
				t.Fatalf("set values = %v, want [%d]", w.setValues, tc.want)
			}
		})
	}
}

func TestUnitAdjustHP4EE460CachesHealthAndLoadsLaterValuesLive(t *testing.T) {
	w := newUnitAdjustHPWorld4EE460()
	w.unit = "unit"
	w.healthByUnit["unit"] = "health-1"
	w.current["health-1"] = 10
	w.maximum["health-1"] = 20
	w.maximum["health-2"] = 99
	w.delta = 3
	w.classLow["unit"] = 0
	w.afterEvent["load-health:unit=health-1"] = func() {
		w.healthByUnit["unit"] = "health-2"
	}
	w.afterEvent["load-current:health-1=10"] = func() {
		w.current["health-1"] = 19
		w.maximum["health-1"] = 15
	}
	w.afterEvent["load-maximum:health-1=15"] = func() {
		w.delta = 4
	}
	w.afterEvent["load-delta=4"] = func() {
		w.delta = 100
	}
	w.afterEvent["set-hp:unit=14"] = func() {
		w.classLow["unit"] = unitAdjustHPMonsterClass4EE460
	}
	w.afterEvent["load-class:unit=02"] = func() {
		w.classLow["unit"] = 0
	}

	unitAdjustHP4EE460(w.hooks())
	want := []string{
		"game-flag:04000000=0",
		"load-unit=unit",
		"load-health:unit=health-1",
		"load-current:health-1=10",
		"load-maximum:health-1=15",
		"load-delta=4",
		"set-hp:unit=14",
		"load-class:unit=02",
		"inform-owner:unit",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestUnitAdjustHP4EE460AllFaultPrefixes(t *testing.T) {
	want := []string{
		"game-flag:04000000=0",
		"load-unit=unit",
		"load-health:unit=health",
		"load-current:health=10",
		"load-maximum:health=20",
		"load-delta=3",
		"set-hp:unit=13",
		"load-class:unit=02",
		"inform-owner:unit",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newUnitAdjustHPWorld4EE460()
			w.unit = "unit"
			w.healthByUnit["unit"] = "health"
			w.current["health"] = 10
			w.maximum["health"] = 20
			w.delta = 3
			w.classLow["unit"] = unitAdjustHPMonsterClass4EE460
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			unitAdjustHP4EE460(w.hooks())
		})
	}
}

type mobInformOwnerHPWorld4EE4C0 struct {
	object        string
	ownerByObject map[string]string
	classLow      map[string]uint8
	updateByOwner map[string]string
	playerByData  map[string]string
	indexByPlayer map[string]uint8
	events        []string
	faultAt       int
	afterEvent    map[string]func()
}

func newMobInformOwnerHPWorld4EE4C0() *mobInformOwnerHPWorld4EE4C0 {
	return &mobInformOwnerHPWorld4EE4C0{
		ownerByObject: make(map[string]string),
		classLow:      make(map[string]uint8),
		updateByOwner: make(map[string]string),
		playerByData:  make(map[string]string),
		indexByPlayer: make(map[string]uint8),
		afterEvent:    make(map[string]func()),
	}
}

func (w *mobInformOwnerHPWorld4EE4C0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *mobInformOwnerHPWorld4EE4C0) hooks() mobInformOwnerHPHooks4EE4C0[string, string, string] {
	return mobInformOwnerHPHooks4EE4C0[string, string, string]{
		loadObjectArg: func() string {
			obj := w.object
			w.record("load-object=" + obj)
			return obj
		},
		loadOwner: func(obj string) string {
			owner := w.ownerByObject[obj]
			w.record("load-owner:" + obj + "=" + owner)
			return owner
		},
		loadClassLow: func(owner string) uint8 {
			value := w.classLow[owner]
			w.record(fmt.Sprintf("load-class:%s=%02x", owner, value))
			return value
		},
		loadUpdateData: func(owner string) string {
			update := w.updateByOwner[owner]
			w.record("load-update:" + owner + "=" + update)
			return update
		},
		loadPlayer: func(update string) string {
			player := w.playerByData[update]
			w.record("load-player:" + update + "=" + player)
			if update == "" {
				panic("nil-update-player")
			}
			return player
		},
		loadPlayerInd: func(player string) uint8 {
			value := w.indexByPlayer[player]
			w.record(fmt.Sprintf("load-index:%s=%d", player, value))
			if player == "" {
				panic("nil-player-index")
			}
			return value
		},
		reportHP: func(index uint8, obj string) {
			w.record(fmt.Sprintf("report:%d:%s", index, obj))
		},
	}
}

func TestMobInformOwnerHP4EE4C0GuardOrder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		object string
		owner  string
		class  uint8
		want   []string
	}{
		{name: "null-object", want: []string{"load-object="}},
		{name: "null-owner", object: "object", want: []string{"load-object=object", "load-owner:object="}},
		{
			name: "non-player", object: "object", owner: "owner", class: 2,
			want: []string{"load-object=object", "load-owner:object=owner", "load-class:owner=02"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newMobInformOwnerHPWorld4EE4C0()
			w.object = tc.object
			w.ownerByObject[tc.object] = tc.owner
			w.classLow[tc.owner] = tc.class
			mobInformOwnerHP4EE4C0(w.hooks())
			if !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("events = %v, want %v", w.events, tc.want)
			}
		})
	}
}

func TestMobInformOwnerHP4EE4C0CachesEarlierPointersAndLoadsLaterOnesLive(t *testing.T) {
	w := newMobInformOwnerHPWorld4EE4C0()
	w.object = "object"
	w.ownerByObject["object"] = "owner-1"
	w.classLow["owner-1"] = mobInformOwnerPlayerClass4EE4C0
	w.updateByOwner["owner-1"] = "update-1"
	w.playerByData["update-2"] = "player-1"
	w.indexByPlayer["player-2"] = 37
	w.afterEvent["load-owner:object=owner-1"] = func() {
		w.ownerByObject["object"] = "owner-2"
	}
	w.afterEvent["load-class:owner-1=04"] = func() {
		w.updateByOwner["owner-1"] = "update-2"
	}
	w.afterEvent["load-update:owner-1=update-2"] = func() {
		w.updateByOwner["owner-1"] = "update-3"
		w.playerByData["update-2"] = "player-2"
	}
	w.afterEvent["load-player:update-2=player-2"] = func() {
		w.playerByData["update-2"] = "player-3"
		w.indexByPlayer["player-2"] = 41
	}
	w.afterEvent["load-index:player-2=41"] = func() {
		w.indexByPlayer["player-2"] = 99
	}

	mobInformOwnerHP4EE4C0(w.hooks())
	want := []string{
		"load-object=object",
		"load-owner:object=owner-1",
		"load-class:owner-1=04",
		"load-update:owner-1=update-2",
		"load-player:update-2=player-2",
		"load-index:player-2=41",
		"report:41:object",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestMobInformOwnerHP4EE4C0PreservesIntermediateNullFaults(t *testing.T) {
	for _, tc := range []struct {
		name       string
		updateData string
		player     string
		wantPanic  string
		want       []string
	}{
		{
			name: "null-update", wantPanic: "nil-update-player",
			want: []string{
				"load-object=object", "load-owner:object=owner", "load-class:owner=04",
				"load-update:owner=", "load-player:=",
			},
		},
		{
			name: "null-player", updateData: "update", wantPanic: "nil-player-index",
			want: []string{
				"load-object=object", "load-owner:object=owner", "load-class:owner=04",
				"load-update:owner=update", "load-player:update=", "load-index:=0",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newMobInformOwnerHPWorld4EE4C0()
			w.object = "object"
			w.ownerByObject["object"] = "owner"
			w.classLow["owner"] = mobInformOwnerPlayerClass4EE4C0
			w.updateByOwner["owner"] = tc.updateData
			w.playerByData[tc.updateData] = tc.player
			defer func() {
				if got := recover(); got != tc.wantPanic {
					t.Fatalf("panic = %v, want %q", got, tc.wantPanic)
				}
				if !reflect.DeepEqual(w.events, tc.want) {
					t.Fatalf("events = %v, want %v", w.events, tc.want)
				}
			}()
			mobInformOwnerHP4EE4C0(w.hooks())
		})
	}
}

func TestMobInformOwnerHP4EE4C0AllFaultPrefixes(t *testing.T) {
	want := []string{
		"load-object=object",
		"load-owner:object=owner",
		"load-class:owner=04",
		"load-update:owner=update",
		"load-player:update=player",
		"load-index:player=37",
		"report:37:object",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event-%d", faultAt), func(t *testing.T) {
			w := newMobInformOwnerHPWorld4EE4C0()
			w.object = "object"
			w.ownerByObject["object"] = "owner"
			w.classLow["owner"] = mobInformOwnerPlayerClass4EE4C0
			w.updateByOwner["owner"] = "update"
			w.playerByData["update"] = "player"
			w.indexByPlayer["player"] = 37
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
					t.Fatalf("events = %v, want %v", w.events, prefix)
				}
			}()
			mobInformOwnerHP4EE4C0(w.hooks())
		})
	}
}
