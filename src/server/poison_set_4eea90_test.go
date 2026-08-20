package server

import (
	"fmt"
	"reflect"
	"testing"
)

type poisonSetTestObject4EEA90 struct {
	name     string
	poison   uint8
	timer    uint16
	health   *poisonSetTestHealth4EEA90
	class    uint32
	subClass uint32
	update   *poisonSetTestUpdate4EEA90
	owner    *poisonSetTestObject4EEA90
}

type poisonSetTestHealth4EEA90 struct {
	name  string
	frame uint32
}

type poisonSetTestUpdate4EEA90 struct {
	name   string
	player *poisonSetTestPlayer4EEA90
}

type poisonSetTestPlayer4EEA90 struct{ name string }

type poisonSetTestInfo4EEA90 struct {
	name string
	unit *poisonSetTestObject4EEA90
}

type poisonSetTestWorld4EEA90 struct {
	unit           *poisonSetTestObject4EEA90
	value          int32
	frameValue     uint32
	gameFlagResult int32
	info           *poisonSetTestInfo4EEA90
	subClassValues []uint32
	events         []string
	after          map[string]func()
}

func poisonSetTestObjectName4EEA90(obj *poisonSetTestObject4EEA90) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *poisonSetTestWorld4EEA90) record(event string) {
	w.events = append(w.events, event)
	if fn := w.after[event]; fn != nil {
		fn()
	}
}

func (w *poisonSetTestWorld4EEA90) hooks() poisonSetHooks4EEA90[
	*poisonSetTestObject4EEA90,
	*poisonSetTestHealth4EEA90,
	*poisonSetTestUpdate4EEA90,
	*poisonSetTestPlayer4EEA90,
	*poisonSetTestInfo4EEA90,
] {
	return poisonSetHooks4EEA90[
		*poisonSetTestObject4EEA90,
		*poisonSetTestHealth4EEA90,
		*poisonSetTestUpdate4EEA90,
		*poisonSetTestPlayer4EEA90,
		*poisonSetTestInfo4EEA90,
	]{
		loadUnitArg: func() *poisonSetTestObject4EEA90 {
			unit := w.unit
			w.record("arg")
			return unit
		},
		loadCurrent: func(unit *poisonSetTestObject4EEA90) uint8 {
			current := unit.poison
			w.record("current:" + unit.name)
			return current
		},
		loadValueArg: func() int32 {
			value := w.value
			w.record("value")
			return value
		},
		loadHealth: func(unit *poisonSetTestObject4EEA90) *poisonSetTestHealth4EEA90 {
			health := unit.health
			w.record("health:" + unit.name)
			return health
		},
		frame: func() uint32 {
			frame := w.frameValue
			w.record("frame")
			return frame
		},
		storeHealthFrame: func(health *poisonSetTestHealth4EEA90, frame uint32) {
			health.frame = frame
			w.record(fmt.Sprintf("store-frame:%s:%d", health.name, frame))
		},
		loadClass: func(unit *poisonSetTestObject4EEA90) uint32 {
			class := unit.class
			w.record("class:" + unit.name)
			return class
		},
		storePoison: func(unit *poisonSetTestObject4EEA90, value uint8) {
			unit.poison = value
			w.record(fmt.Sprintf("store-poison:%s:%d", unit.name, value))
		},
		loadUpdateData: func(unit *poisonSetTestObject4EEA90) *poisonSetTestUpdate4EEA90 {
			update := unit.update
			w.record("update:" + unit.name)
			return update
		},
		loadPlayer: func(update *poisonSetTestUpdate4EEA90) *poisonSetTestPlayer4EEA90 {
			if update == nil {
				w.record("player:nil")
				panic("nil player update")
			}
			player := update.player
			w.record("player:" + update.name)
			return player
		},
		needPlayerStatus: func(player *poisonSetTestPlayer4EEA90, status uint32) {
			w.record(fmt.Sprintf("need:%s:%d", player.name, status))
		},
		unsetPlayerStatus: func(player *poisonSetTestPlayer4EEA90, status uint32) {
			w.record(fmt.Sprintf("unset:%s:%d", player.name, status))
		},
		gameFlag: func(flag uint32) int32 {
			result := w.gameFlagResult
			w.record(fmt.Sprintf("game-flag:%d", flag))
			return result
		},
		loadSubClass: func(unit *poisonSetTestObject4EEA90) uint32 {
			value := unit.subClass
			if len(w.subClassValues) != 0 {
				value = w.subClassValues[0]
				w.subClassValues = w.subClassValues[1:]
			}
			w.record("subclass:" + unit.name)
			return value
		},
		playerInfoByIndex: func(index int32) *poisonSetTestInfo4EEA90 {
			info := w.info
			w.record(fmt.Sprintf("player-info:%d", index))
			return info
		},
		loadPlayerUnit: func(info *poisonSetTestInfo4EEA90) *poisonSetTestObject4EEA90 {
			unit := info.unit
			w.record("player-unit:" + info.name)
			return unit
		},
		loadOwner: func(unit *poisonSetTestObject4EEA90) *poisonSetTestObject4EEA90 {
			owner := unit.owner
			w.record("owner:" + unit.name)
			return owner
		},
		reportPoison: func(receiver, unit *poisonSetTestObject4EEA90, active int32) {
			w.record(fmt.Sprintf(
				"report:%s:%s:%d",
				poisonSetTestObjectName4EEA90(receiver),
				poisonSetTestObjectName4EEA90(unit),
				active,
			))
		},
		storePoisonTimer: func(unit *poisonSetTestObject4EEA90, timer uint16) {
			unit.timer = timer
			w.record(fmt.Sprintf("timer:%s:%d", unit.name, timer))
		},
	}
}

func TestSetPoison4EEA90PlayerPositiveTransitionOrder(t *testing.T) {
	health := &poisonSetTestHealth4EEA90{name: "health"}
	player := &poisonSetTestPlayer4EEA90{name: "player"}
	unit := &poisonSetTestObject4EEA90{
		name:   "unit",
		health: health,
		class:  uint32(poisonClearPlayerClassLow4EE8F0 | poisonClearMonsterClassLow4EE8F0),
		update: &poisonSetTestUpdate4EEA90{name: "update", player: player},
	}
	world := &poisonSetTestWorld4EEA90{unit: unit, value: 7, frameValue: 123}
	setPoison4EEA90(world.hooks())
	want := []string{
		"arg", "current:unit", "value", "health:unit", "frame", "store-frame:health:123",
		"class:unit", "store-poison:unit:7", "update:unit", "player:update", "need:player:1024",
		"timer:unit:1000",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %q, want %q", world.events, want)
	}
	if unit.poison != 7 || unit.timer != 1000 || health.frame != 123 {
		t.Fatalf("poison/timer/frame = %d/%d/%d, want 7/1000/123", unit.poison, unit.timer, health.frame)
	}
}

func TestSetPoison4EEA90WholeValueControlsStateButLowByteIsStored(t *testing.T) {
	tests := []struct {
		name      string
		current   uint8
		value     int32
		wantByte  uint8
		wantTimer uint16
		wantFrame bool
		wantEvent string
	}{
		{name: "256 is active zero byte", value: 256, wantByte: 0, wantTimer: 1000, wantFrame: true, wantEvent: "need:player:1024"},
		{name: "negative is active without transition", value: -1, wantByte: 255, wantTimer: 1000, wantEvent: "need:player:1024"},
		{name: "zero is inactive", current: 9, value: 0, wantByte: 0, wantTimer: 0, wantEvent: "unset:player:1024"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			health := &poisonSetTestHealth4EEA90{name: "health"}
			unit := &poisonSetTestObject4EEA90{
				name:   "unit",
				poison: test.current,
				health: health,
				class:  uint32(poisonClearPlayerClassLow4EE8F0),
				update: &poisonSetTestUpdate4EEA90{name: "update", player: &poisonSetTestPlayer4EEA90{name: "player"}},
			}
			world := &poisonSetTestWorld4EEA90{unit: unit, value: test.value, frameValue: 99}
			setPoison4EEA90(world.hooks())
			if unit.poison != test.wantByte || unit.timer != test.wantTimer {
				t.Fatalf("poison/timer = %d/%d, want %d/%d", unit.poison, unit.timer, test.wantByte, test.wantTimer)
			}
			if (health.frame == 99) != test.wantFrame {
				t.Fatalf("frame stamped = %v, want %v; events %q", health.frame == 99, test.wantFrame, world.events)
			}
			if !containsPoisonSetEvent4EEA90(world.events, test.wantEvent) {
				t.Fatalf("missing %q in %q", test.wantEvent, world.events)
			}
		})
	}
}

func TestSetPoison4EEA90CachesClassBeforePoisonStore(t *testing.T) {
	player := &poisonSetTestPlayer4EEA90{name: "player"}
	unit := &poisonSetTestObject4EEA90{
		name:   "unit",
		poison: 1,
		class:  uint32(poisonClearPlayerClassLow4EE8F0),
		update: &poisonSetTestUpdate4EEA90{name: "update", player: player},
	}
	world := &poisonSetTestWorld4EEA90{
		unit:  unit,
		value: 2,
		after: map[string]func(){
			"store-poison:unit:2": func() {
				unit.class = uint32(poisonClearMonsterClassLow4EE8F0)
			},
		},
	}
	setPoison4EEA90(world.hooks())
	if !containsPoisonSetEvent4EEA90(world.events, "need:player:1024") ||
		containsPoisonSetEvent4EEA90(world.events, "game-flag:2048") {
		t.Fatalf("cached Player class was not preserved: %q", world.events)
	}
}

func TestSetPoison4EEA90MonsterQuestAndOwnerReports(t *testing.T) {
	receiver := &poisonSetTestObject4EEA90{name: "receiver"}
	owner := &poisonSetTestObject4EEA90{name: "owner"}
	unit := &poisonSetTestObject4EEA90{
		name:     "unit",
		class:    uint32(poisonClearMonsterClassLow4EE8F0),
		subClass: uint32(poisonClearQuestMonsterLow4EE8F0 | poisonClearOwnedMonsterLow4EE8F0),
		owner:    owner,
	}
	world := &poisonSetTestWorld4EEA90{
		unit:           unit,
		value:          3,
		gameFlagResult: 1,
		info:           &poisonSetTestInfo4EEA90{name: "info", unit: receiver},
	}
	setPoison4EEA90(world.hooks())
	wantTail := []string{"player-unit:info", "report:receiver:unit:1", "timer:unit:1000"}
	if !reflect.DeepEqual(world.events[len(world.events)-len(wantTail):], wantTail) {
		t.Fatalf("quest events = %q, want tail %q", world.events, wantTail)
	}

	unit.poison = 0
	world.events = nil
	world.value = 0
	world.gameFlagResult = 2
	setPoison4EEA90(world.hooks())
	wantTail = []string{"subclass:unit", "owner:unit", "report:owner:unit:0", "timer:unit:0"}
	if !reflect.DeepEqual(world.events[len(world.events)-len(wantTail):], wantTail) {
		t.Fatalf("owner events = %q, want tail %q", world.events, wantTail)
	}
}

func TestSetPoison4EEA90MissingQuestReceiverStillStoresTimer(t *testing.T) {
	unit := &poisonSetTestObject4EEA90{
		name:     "unit",
		class:    uint32(poisonClearMonsterClassLow4EE8F0),
		subClass: uint32(poisonClearQuestMonsterLow4EE8F0),
	}
	world := &poisonSetTestWorld4EEA90{unit: unit, value: 1, gameFlagResult: 1}
	setPoison4EEA90(world.hooks())
	if unit.timer != 1000 || !containsPoisonSetEvent4EEA90(world.events, "player-info:31") ||
		containsPoisonSetPrefix4EEA90(world.events, "report:") {
		t.Fatalf("missing receiver result/events = timer %d, %q", unit.timer, world.events)
	}
}

func TestSetPoison4EEA90NilUnitDefersAllOtherLoads(t *testing.T) {
	world := &poisonSetTestWorld4EEA90{value: 99}
	setPoison4EEA90(world.hooks())
	if !reflect.DeepEqual(world.events, []string{"arg"}) {
		t.Fatalf("nil events = %q, want [arg]", world.events)
	}
}

func containsPoisonSetEvent4EEA90(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}

func containsPoisonSetPrefix4EEA90(events []string, prefix string) bool {
	for _, event := range events {
		if len(event) >= len(prefix) && event[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
