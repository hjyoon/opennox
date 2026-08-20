package server

import (
	"fmt"
	"reflect"
	"testing"
)

type poisonClearTestObject4EE8F0 struct {
	name      string
	poison    uint8
	health    *poisonClearTestHealth4EE8F0
	class     uint32
	subClass  uint32
	update    *poisonClearTestUpdate4EE8F0
	owner     *poisonClearTestObject4EE8F0
	frameZero bool
}

type poisonClearTestHealth4EE8F0 struct {
	name  string
	frame uint32
}

type poisonClearTestUpdate4EE8F0 struct {
	name   string
	player *poisonClearTestPlayer4EE8F0
}

type poisonClearTestPlayer4EE8F0 struct{ name string }

type poisonClearTestInfo4EE8F0 struct {
	name string
	unit *poisonClearTestObject4EE8F0
}

type poisonClearTestWorld4EE8F0 struct {
	unit           *poisonClearTestObject4EE8F0
	amount         int32
	gameFlagResult int32
	info           *poisonClearTestInfo4EE8F0
	subClassValues []uint32
	events         []string
	after          map[string]func()
}

func poisonClearTestObjectName4EE8F0(obj *poisonClearTestObject4EE8F0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *poisonClearTestWorld4EE8F0) record(event string) {
	w.events = append(w.events, event)
	if fn := w.after[event]; fn != nil {
		fn()
	}
}

func (w *poisonClearTestWorld4EE8F0) hooks() poisonClearHooks4EE8F0[
	*poisonClearTestObject4EE8F0,
	*poisonClearTestHealth4EE8F0,
	*poisonClearTestUpdate4EE8F0,
	*poisonClearTestPlayer4EE8F0,
	*poisonClearTestInfo4EE8F0,
] {
	return poisonClearHooks4EE8F0[
		*poisonClearTestObject4EE8F0,
		*poisonClearTestHealth4EE8F0,
		*poisonClearTestUpdate4EE8F0,
		*poisonClearTestPlayer4EE8F0,
		*poisonClearTestInfo4EE8F0,
	]{
		loadUnitArg: func() *poisonClearTestObject4EE8F0 {
			unit := w.unit
			w.record("arg")
			return unit
		},
		loadAmountArg: func() int32 {
			amount := w.amount
			w.record("amount")
			return amount
		},
		loadPoison: func(unit *poisonClearTestObject4EE8F0) uint8 {
			value := unit.poison
			w.record("poison:" + unit.name)
			return value
		},
		storePoison: func(unit *poisonClearTestObject4EE8F0, value uint8) {
			unit.poison = value
			w.record(fmt.Sprintf("store-poison:%s:%d", unit.name, value))
		},
		loadHealth: func(unit *poisonClearTestObject4EE8F0) *poisonClearTestHealth4EE8F0 {
			health := unit.health
			w.record("health:" + unit.name)
			return health
		},
		clearHealthFrame: func(health *poisonClearTestHealth4EE8F0) {
			health.frame = 0
			w.record("clear-frame:" + health.name)
		},
		loadClass: func(unit *poisonClearTestObject4EE8F0) uint32 {
			class := unit.class
			w.record("class:" + unit.name)
			return class
		},
		loadUpdateData: func(unit *poisonClearTestObject4EE8F0) *poisonClearTestUpdate4EE8F0 {
			update := unit.update
			w.record("update:" + unit.name)
			return update
		},
		loadPlayer: func(update *poisonClearTestUpdate4EE8F0) *poisonClearTestPlayer4EE8F0 {
			if update == nil {
				w.record("player:nil")
				panic("nil player update")
			}
			player := update.player
			w.record("player:" + update.name)
			return player
		},
		unsetPlayerStatus: func(player *poisonClearTestPlayer4EE8F0, status uint32) {
			name := "nil"
			if player != nil {
				name = player.name
			}
			w.record(fmt.Sprintf("unset:%s:%d", name, status))
		},
		priorityMessage: func(unit *poisonClearTestObject4EE8F0, message string, value uint8) {
			w.record(fmt.Sprintf("message:%s:%s:%d", unit.name, message, value))
		},
		gameFlag: func(flag uint32) int32 {
			result := w.gameFlagResult
			w.record(fmt.Sprintf("game-flag:%d", flag))
			return result
		},
		loadSubClass: func(unit *poisonClearTestObject4EE8F0) uint32 {
			value := unit.subClass
			if len(w.subClassValues) != 0 {
				value = w.subClassValues[0]
				w.subClassValues = w.subClassValues[1:]
			}
			w.record("subclass:" + unit.name)
			return value
		},
		playerInfoByIndex: func(index int32) *poisonClearTestInfo4EE8F0 {
			info := w.info
			w.record(fmt.Sprintf("player-info:%d", index))
			return info
		},
		loadPlayerUnit: func(info *poisonClearTestInfo4EE8F0) *poisonClearTestObject4EE8F0 {
			unit := info.unit
			w.record("player-unit:" + info.name)
			return unit
		},
		loadOwner: func(unit *poisonClearTestObject4EE8F0) *poisonClearTestObject4EE8F0 {
			owner := unit.owner
			w.record("owner:" + unit.name)
			return owner
		},
		reportPoison: func(receiver, unit *poisonClearTestObject4EE8F0, active int32) {
			w.record(fmt.Sprintf(
				"report:%s:%s:%d",
				poisonClearTestObjectName4EE8F0(receiver),
				poisonClearTestObjectName4EE8F0(unit),
				active,
			))
		},
	}
}

func TestUpdatePoison4EE8F0SignedCompareAndByteSubtraction(t *testing.T) {
	tests := []struct {
		name   string
		poison uint8
		amount int32
		want   uint8
	}{
		{name: "ordinary", poison: 9, amount: 4, want: 5},
		{name: "negative adds through low byte", poison: 5, amount: -1, want: 6},
		{name: "minimum wraps low byte", poison: 1, amount: -128, want: 129},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			unit := &poisonClearTestObject4EE8F0{name: "unit", poison: test.poison}
			world := &poisonClearTestWorld4EE8F0{unit: unit, amount: test.amount}
			updatePoison4EE8F0(world.hooks())
			if unit.poison != test.want {
				t.Fatalf("poison = %d, want %d", unit.poison, test.want)
			}
			wantEvents := []string{"arg", "poison:unit", "amount", fmt.Sprintf("store-poison:unit:%d", test.want)}
			if !reflect.DeepEqual(world.events, wantEvents) {
				t.Fatalf("events = %q, want %q", world.events, wantEvents)
			}
		})
	}
}

func TestUpdatePoison4EE8F0PlayerClearOrder(t *testing.T) {
	health := &poisonClearTestHealth4EE8F0{name: "health", frame: 77}
	player := &poisonClearTestPlayer4EE8F0{name: "player"}
	unit := &poisonClearTestObject4EE8F0{
		name:   "unit",
		poison: 5,
		health: health,
		class:  uint32(poisonClearPlayerClassLow4EE8F0 | poisonClearMonsterClassLow4EE8F0),
		update: &poisonClearTestUpdate4EE8F0{name: "update", player: player},
	}
	world := &poisonClearTestWorld4EE8F0{unit: unit, amount: 5}
	updatePoison4EE8F0(world.hooks())
	want := []string{
		"arg", "poison:unit", "amount", "health:unit", "store-poison:unit:0",
		"clear-frame:health", "class:unit", "update:unit", "player:update", "unset:player:1024",
		"message:unit:Health.c:PoisonFade:0",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %q, want %q", world.events, want)
	}
	if unit.poison != 0 || health.frame != 0 {
		t.Fatalf("poison/frame = %d/%d, want 0/0", unit.poison, health.frame)
	}
}

func TestRemovePoison4EE9D0PlayerOmitsFadeMessage(t *testing.T) {
	player := &poisonClearTestPlayer4EE8F0{name: "player"}
	unit := &poisonClearTestObject4EE8F0{
		name:   "unit",
		poison: 1,
		class:  uint32(poisonClearPlayerClassLow4EE8F0),
		update: &poisonClearTestUpdate4EE8F0{name: "update", player: player},
	}
	world := &poisonClearTestWorld4EE8F0{unit: unit}
	removePoison4EE9D0(world.hooks())
	want := []string{
		"arg", "poison:unit", "health:unit", "store-poison:unit:0", "class:unit",
		"update:unit", "player:update", "unset:player:1024",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %q, want %q", world.events, want)
	}
}

func TestPoisonClear4EE8F0MonsterQuestReceiverAndNoFallback(t *testing.T) {
	receiver := &poisonClearTestObject4EE8F0{name: "receiver"}
	unit := &poisonClearTestObject4EE8F0{
		name:     "unit",
		poison:   1,
		class:    uint32(poisonClearMonsterClassLow4EE8F0),
		subClass: uint32(poisonClearQuestMonsterLow4EE8F0 | poisonClearOwnedMonsterLow4EE8F0),
		owner:    &poisonClearTestObject4EE8F0{name: "owner"},
	}
	world := &poisonClearTestWorld4EE8F0{
		unit:           unit,
		gameFlagResult: 1,
		info:           &poisonClearTestInfo4EE8F0{name: "info", unit: receiver},
	}
	removePoison4EE9D0(world.hooks())
	wantTail := []string{
		"class:unit", "game-flag:2048", "subclass:unit", "player-info:31",
		"player-unit:info", "report:receiver:unit:0",
	}
	if !reflect.DeepEqual(world.events[len(world.events)-len(wantTail):], wantTail) {
		t.Fatalf("events = %q, want tail %q", world.events, wantTail)
	}

	unit.poison = 1
	world.events = nil
	world.info = nil
	removePoison4EE9D0(world.hooks())
	if containsPoisonClearEvent4EE8F0(world.events, "owner:unit") {
		t.Fatalf("nil quest player info fell back to owner: %q", world.events)
	}
}

func TestPoisonClear4EE8F0MonsterOwnerBranchAndSubclassReload(t *testing.T) {
	owner := &poisonClearTestObject4EE8F0{name: "owner"}
	unit := &poisonClearTestObject4EE8F0{
		name:   "unit",
		poison: 1,
		class:  uint32(poisonClearMonsterClassLow4EE8F0),
		owner:  owner,
	}
	world := &poisonClearTestWorld4EE8F0{
		unit:           unit,
		gameFlagResult: 1,
		subClassValues: []uint32{
			uint32(poisonClearOwnedMonsterLow4EE8F0),
			uint32(poisonClearOwnedMonsterLow4EE8F0),
		},
	}
	removePoison4EE9D0(world.hooks())
	wantTail := []string{
		"game-flag:2048", "subclass:unit", "subclass:unit", "owner:unit", "report:owner:unit:0",
	}
	if !reflect.DeepEqual(world.events[len(world.events)-len(wantTail):], wantTail) {
		t.Fatalf("events = %q, want tail %q", world.events, wantTail)
	}

	unit.poison = 1
	world.events = nil
	world.gameFlagResult = 2
	world.subClassValues = []uint32{uint32(poisonClearOwnedMonsterLow4EE8F0)}
	removePoison4EE9D0(world.hooks())
	if countPoisonClearEvent4EE8F0(world.events, "subclass:unit") != 1 {
		t.Fatalf("noncanonical game-flag result subclass loads = %q, want one", world.events)
	}
}

func containsPoisonClearEvent4EE8F0(events []string, want string) bool {
	return countPoisonClearEvent4EE8F0(events, want) != 0
}

func countPoisonClearEvent4EE8F0(events []string, want string) int {
	count := 0
	for _, event := range events {
		if event == want {
			count++
		}
	}
	return count
}

func TestPoisonClear4EE8F0CachesHealthAndLoadsClassAfterStores(t *testing.T) {
	entryHealth := &poisonClearTestHealth4EE8F0{name: "entry", frame: 10}
	liveHealth := &poisonClearTestHealth4EE8F0{name: "live", frame: 20}
	player := &poisonClearTestPlayer4EE8F0{name: "player"}
	unit := &poisonClearTestObject4EE8F0{name: "unit", poison: 1, health: entryHealth}
	world := &poisonClearTestWorld4EE8F0{
		unit: unit,
		after: map[string]func(){
			"health:unit": func() {
				unit.health = liveHealth
			},
			"store-poison:unit:0": func() {
				unit.class = uint32(poisonClearPlayerClassLow4EE8F0)
				unit.update = &poisonClearTestUpdate4EE8F0{name: "update", player: player}
			},
		},
	}
	removePoison4EE9D0(world.hooks())
	if entryHealth.frame != 0 || liveHealth.frame != 20 {
		t.Fatalf("entry/live frame = %d/%d, want 0/20", entryHealth.frame, liveHealth.frame)
	}
	if !containsPoisonClearEvent4EE8F0(world.events, "unset:player:1024") {
		t.Fatalf("live class was not used after poison store: %q", world.events)
	}
}

func TestPoisonClear4EE8F0EarlyReturnsAndFaultPrefix(t *testing.T) {
	nilWorld := &poisonClearTestWorld4EE8F0{}
	updatePoison4EE8F0(nilWorld.hooks())
	if !reflect.DeepEqual(nilWorld.events, []string{"arg"}) {
		t.Fatalf("nil update events = %q", nilWorld.events)
	}

	zeroUnit := &poisonClearTestObject4EE8F0{name: "unit"}
	zeroWorld := &poisonClearTestWorld4EE8F0{unit: zeroUnit}
	removePoison4EE9D0(zeroWorld.hooks())
	if !reflect.DeepEqual(zeroWorld.events, []string{"arg", "poison:unit"}) {
		t.Fatalf("zero remove events = %q", zeroWorld.events)
	}

	faultUnit := &poisonClearTestObject4EE8F0{
		name:   "unit",
		poison: 1,
		class:  uint32(poisonClearPlayerClassLow4EE8F0),
	}
	faultWorld := &poisonClearTestWorld4EE8F0{unit: faultUnit}
	defer func() {
		if recover() == nil {
			t.Fatal("nil PlayerUpdateData did not fault")
		}
		want := []string{
			"arg", "poison:unit", "health:unit", "store-poison:unit:0", "class:unit", "update:unit", "player:nil",
		}
		if !reflect.DeepEqual(faultWorld.events, want) {
			t.Fatalf("fault events = %q, want %q", faultWorld.events, want)
		}
	}()
	removePoison4EE9D0(faultWorld.hooks())
}
