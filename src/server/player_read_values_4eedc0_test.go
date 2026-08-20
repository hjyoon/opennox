package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerReadValuesTestStats4EEDC0 struct {
	name   string
	values [4]float32
}

type playerReadValuesTestName4EEDC0 struct {
	name string
}

type playerReadValuesTestPlayer4EEDC0 struct {
	name          string
	class         uint8
	level         uint8
	strength      uint32
	speed         uint32
	capacityWord  uint16
	overweight    uint32
	initialized   uint8
	strengthToken uint32
	speedToken    uint32
	manaToken     uint32
	healthToken   uint32
	nameToken     uint32
	originalName  *playerReadValuesTestName4EEDC0
}

type playerReadValuesTestUpdate4EEDC0 struct {
	name        string
	player      *playerReadValuesTestPlayer4EEDC0
	manaCurrent uint16
	manaMaximum uint16
}

type playerReadValuesTestHealth4EEDC0 struct {
	name    string
	current uint16
	maximum uint16
}

type playerReadValuesTestItem4EEDC0 struct {
	name   string
	weight uint8
	next   *playerReadValuesTestItem4EEDC0
}

type playerReadValuesTestObject4EEDC0 struct {
	name           string
	update         *playerReadValuesTestUpdate4EEDC0
	health         *playerReadValuesTestHealth4EEDC0
	speedBase      float32
	mass           float32
	carry          uint16
	first          *playerReadValuesTestItem4EEDC0
	lastSetHP      uint16
	setHPCallCount int
}

type playerReadValuesTestAbility4EEDC0 struct {
	unit      *playerReadValuesTestObject4EEDC0
	count     int8
	rewardArg int32
}

type playerReadValuesTestWorld4EEDC0 struct {
	unit          *playerReadValuesTestObject4EEDC0
	baseStats     *playerReadValuesTestStats4EEDC0
	classStats    map[uint8]*playerReadValuesTestStats4EEDC0
	game          map[uint32]int32
	solo          int32
	rewardArg     int32
	protectResult int32
	events        []string
	abilities     []playerReadValuesTestAbility4EEDC0
	faultAt       int
	after         map[string]func()
}

func (w *playerReadValuesTestWorld4EEDC0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func playerReadValuesTestPlayerName4EEDC0(player *playerReadValuesTestPlayer4EEDC0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func playerReadValuesTestRound4EEDC0(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

func playerReadValuesTestTrunc4EEDC0(value float64) int64 {
	if math.IsNaN(value) || value >= 9223372036854775808 || value < -9223372036854775808 {
		return math.MinInt64
	}
	return int64(math.Trunc(value))
}

func (w *playerReadValuesTestWorld4EEDC0) hooks() playerReadValuesHooks4EEDC0[
	*playerReadValuesTestObject4EEDC0,
	*playerReadValuesTestUpdate4EEDC0,
	*playerReadValuesTestPlayer4EEDC0,
	*playerReadValuesTestStats4EEDC0,
	*playerReadValuesTestHealth4EEDC0,
	*playerReadValuesTestName4EEDC0,
	*playerReadValuesTestItem4EEDC0,
] {
	return playerReadValuesHooks4EEDC0[
		*playerReadValuesTestObject4EEDC0,
		*playerReadValuesTestUpdate4EEDC0,
		*playerReadValuesTestPlayer4EEDC0,
		*playerReadValuesTestStats4EEDC0,
		*playerReadValuesTestHealth4EEDC0,
		*playerReadValuesTestName4EEDC0,
		*playerReadValuesTestItem4EEDC0,
	]{
		loadUnitArg: func() *playerReadValuesTestObject4EEDC0 {
			unit := w.unit
			name := "nil"
			if unit != nil {
				name = unit.name
			}
			w.record("arg:" + name)
			return unit
		},
		loadUpdateData: func(unit *playerReadValuesTestObject4EEDC0) *playerReadValuesTestUpdate4EEDC0 {
			update := unit.update
			name := "nil"
			if update != nil {
				name = update.name
			}
			w.record("update:" + name)
			return update
		},
		loadPlayer: func(update *playerReadValuesTestUpdate4EEDC0) *playerReadValuesTestPlayer4EEDC0 {
			player := update.player
			w.record("player:" + playerReadValuesTestPlayerName4EEDC0(player))
			return player
		},
		loadBaseStats: func() *playerReadValuesTestStats4EEDC0 {
			stats := w.baseStats
			w.record("base-stats:" + stats.name)
			return stats
		},
		loadPlayerClass: func(player *playerReadValuesTestPlayer4EEDC0) uint8 {
			class := player.class
			w.record(fmt.Sprintf("class:%s=%d", playerReadValuesTestPlayerName4EEDC0(player), class))
			return class
		},
		loadClassStats: func(class uint8) *playerReadValuesTestStats4EEDC0 {
			stats := w.classStats[class]
			w.record(fmt.Sprintf("class-stats:%d=%s", class, stats.name))
			return stats
		},
		loadStat: func(stats *playerReadValuesTestStats4EEDC0, stat playerReadValuesStat4EEDC0) float32 {
			value := stats.values[stat]
			w.record(fmt.Sprintf("stat:%s:%d=%08x", stats.name, stat, math.Float32bits(value)))
			return value
		},
		gameFlagsCheck: func(mask uint32) int32 {
			value := w.game[mask]
			w.record(fmt.Sprintf("game:%#x=%d", mask, value))
			return value
		},
		floatToInt: func(value float32) int32 {
			result := playerReadValuesTestRound4EEDC0(value)
			w.record(fmt.Sprintf("round:%08x=%d", math.Float32bits(value), result))
			return result
		},
		floatToInt16Abs: func(value float32) int16 {
			result := int16(playerReadValuesTestRound4EEDC0(float32(math.Abs(float64(value)))))
			w.record(fmt.Sprintf("round-abs16:%08x=%d", math.Float32bits(value), result))
			return result
		},
		loadHealthData: func(unit *playerReadValuesTestObject4EEDC0) *playerReadValuesTestHealth4EEDC0 {
			health := unit.health
			name := "nil"
			if health != nil {
				name = health.name
			}
			w.record("health:" + name)
			return health
		},
		storeHealthMax: func(health *playerReadValuesTestHealth4EEDC0, value uint16) {
			w.record(fmt.Sprintf("store-health-max:%s=%d", health.name, value))
			health.maximum = value
		},
		loadHealthMax: func(health *playerReadValuesTestHealth4EEDC0) uint16 {
			value := health.maximum
			w.record(fmt.Sprintf("health-max:%s=%d", health.name, value))
			return value
		},
		setHP: func(unit *playerReadValuesTestObject4EEDC0, value uint16) {
			w.record(fmt.Sprintf("set-hp:%s=%d", unit.name, value))
			unit.lastSetHP = value
			unit.setHPCallCount++
			unit.health.current = value
		},
		storeManaMax: func(update *playerReadValuesTestUpdate4EEDC0, value uint16) {
			w.record(fmt.Sprintf("store-mana-max:%s=%d", update.name, value))
			update.manaMaximum = value
		},
		loadManaMax: func(update *playerReadValuesTestUpdate4EEDC0) uint16 {
			value := update.manaMaximum
			w.record(fmt.Sprintf("mana-max:%s=%d", update.name, value))
			return value
		},
		storeManaCurrent: func(update *playerReadValuesTestUpdate4EEDC0, value uint16) {
			w.record(fmt.Sprintf("store-mana-current:%s=%d", update.name, value))
			update.manaCurrent = value
		},
		storeStrength: func(player *playerReadValuesTestPlayer4EEDC0, value uint32) {
			w.record(fmt.Sprintf("store-strength:%s=%#x", player.name, value))
			player.strength = value
		},
		loadStrength: func(player *playerReadValuesTestPlayer4EEDC0) uint32 {
			value := player.strength
			w.record(fmt.Sprintf("strength:%s=%#x", player.name, value))
			return value
		},
		storeSpeedStat: func(player *playerReadValuesTestPlayer4EEDC0, value uint32) {
			w.record(fmt.Sprintf("store-speed:%s=%#x", player.name, value))
			player.speed = value
		},
		loadSpeedStat: func(player *playerReadValuesTestPlayer4EEDC0) uint32 {
			value := player.speed
			w.record(fmt.Sprintf("speed:%s=%#x", player.name, value))
			return value
		},
		storeSpeedBase: func(unit *playerReadValuesTestObject4EEDC0, value float32) {
			w.record(fmt.Sprintf("store-speed-base:%s=%08x", unit.name, math.Float32bits(value)))
			unit.speedBase = value
		},
		loadLevel: func(player *playerReadValuesTestPlayer4EEDC0) uint8 {
			value := player.level
			w.record(fmt.Sprintf("level:%s=%#x", player.name, value))
			return value
		},
		soloMode: func() int32 {
			value := w.solo
			w.record(fmt.Sprintf("solo=%d", value))
			return value
		},
		loadRewardArg: func() int32 {
			value := w.rewardArg
			w.record(fmt.Sprintf("reward-arg=%d", value))
			return value
		},
		abilityGiveAll: func(unit *playerReadValuesTestObject4EEDC0, count int8, rewardArg int32) {
			w.record(fmt.Sprintf("abilities:%s:%d:%d", unit.name, count, rewardArg))
			w.abilities = append(w.abilities, playerReadValuesTestAbility4EEDC0{unit, count, rewardArg})
		},
		storeMass: func(unit *playerReadValuesTestObject4EEDC0, value float32) {
			w.record(fmt.Sprintf("store-mass:%s=%08x", unit.name, math.Float32bits(value)))
			unit.mass = value
		},
		floatToInt64Trunc: func(value float64) int64 {
			result := playerReadValuesTestTrunc4EEDC0(value)
			w.record(fmt.Sprintf("trunc:%016x=%d", math.Float64bits(value), result))
			return result
		},
		storeCapacityWord: func(player *playerReadValuesTestPlayer4EEDC0, value uint16) {
			w.record(fmt.Sprintf("store-capacity-word:%s=%d", player.name, value))
			player.capacityWord = value
		},
		loadCapacityWord: func(player *playerReadValuesTestPlayer4EEDC0) uint16 {
			value := player.capacityWord
			w.record(fmt.Sprintf("capacity-word:%s=%d", player.name, value))
			return value
		},
		storeCarry: func(unit *playerReadValuesTestObject4EEDC0, value uint16) {
			w.record(fmt.Sprintf("store-carry:%s=%d", unit.name, value))
			unit.carry = value
		},
		loadStrengthToken: func(player *playerReadValuesTestPlayer4EEDC0) uint32 {
			value := player.strengthToken
			w.record(fmt.Sprintf("strength-token:%s=%#x", player.name, value))
			return value
		},
		loadSpeedToken: func(player *playerReadValuesTestPlayer4EEDC0) uint32 {
			value := player.speedToken
			w.record(fmt.Sprintf("speed-token:%s=%#x", player.name, value))
			return value
		},
		loadManaMaxToken: func(player *playerReadValuesTestPlayer4EEDC0) uint32 {
			value := player.manaToken
			w.record(fmt.Sprintf("mana-token:%s=%#x", player.name, value))
			return value
		},
		loadHealthToken: func(player *playerReadValuesTestPlayer4EEDC0) uint32 {
			value := player.healthToken
			w.record(fmt.Sprintf("health-token:%s=%#x", player.name, value))
			return value
		},
		protectInt: func(token, value uint32) {
			w.record(fmt.Sprintf("protect-int:%#x=%#x", token, value))
		},
		protectUint16: func(token uint32, value uint16) {
			w.record(fmt.Sprintf("protect-u16:%#x=%d", token, value))
		},
		loadFirstItem: func(unit *playerReadValuesTestObject4EEDC0) *playerReadValuesTestItem4EEDC0 {
			item := unit.first
			name := "nil"
			if item != nil {
				name = item.name
			}
			w.record("first-item:" + name)
			return item
		},
		loadItemWeight: func(item *playerReadValuesTestItem4EEDC0) uint8 {
			value := item.weight
			w.record(fmt.Sprintf("weight:%s=%d", item.name, value))
			return value
		},
		loadNextItem: func(item *playerReadValuesTestItem4EEDC0) *playerReadValuesTestItem4EEDC0 {
			next := item.next
			name := "nil"
			if next != nil {
				name = next.name
			}
			w.record(fmt.Sprintf("next:%s=%s", item.name, name))
			return next
		},
		loadCarry: func(unit *playerReadValuesTestObject4EEDC0) uint16 {
			value := unit.carry
			w.record(fmt.Sprintf("carry:%s=%d", unit.name, value))
			return value
		},
		storeOverweight: func(player *playerReadValuesTestPlayer4EEDC0, value uint32) {
			w.record(fmt.Sprintf("store-overweight:%s=%d", player.name, value))
			player.overweight = value
		},
		loadNameToken: func(player *playerReadValuesTestPlayer4EEDC0) uint32 {
			value := player.nameToken
			w.record(fmt.Sprintf("name-token:%s=%#x", player.name, value))
			return value
		},
		loadName: func(player *playerReadValuesTestPlayer4EEDC0) *playerReadValuesTestName4EEDC0 {
			name := player.originalName
			w.record(fmt.Sprintf("name:%s=%s", player.name, name.name))
			return name
		},
		wideLen: func(name *playerReadValuesTestName4EEDC0) uint32 {
			length := uint32(len(name.name))
			w.record(fmt.Sprintf("wide-len:%s=%d", name.name, length))
			return length
		},
		protectName: func(name *playerReadValuesTestName4EEDC0, size, token uint32) int32 {
			w.record(fmt.Sprintf("protect-name:%s:%d:%#x=%d", name.name, size, token, w.protectResult))
			return w.protectResult
		},
		storeInitialized: func(player *playerReadValuesTestPlayer4EEDC0, value uint8) {
			w.record(fmt.Sprintf("store-initialized:%s=%d", player.name, value))
			player.initialized = value
		},
	}
}

func newPlayerReadValuesTestWorld4EEDC0() *playerReadValuesTestWorld4EEDC0 {
	name := &playerReadValuesTestName4EEDC0{name: "Alpha"}
	player := &playerReadValuesTestPlayer4EEDC0{
		name:          "player",
		class:         1,
		level:         5,
		strengthToken: 0x11111111,
		speedToken:    0x22222222,
		manaToken:     0x33333333,
		healthToken:   0x44444444,
		nameToken:     0x55555555,
		originalName:  name,
	}
	update := &playerReadValuesTestUpdate4EEDC0{name: "update", player: player}
	item2 := &playerReadValuesTestItem4EEDC0{name: "item-2", weight: 250}
	item1 := &playerReadValuesTestItem4EEDC0{name: "item-1", weight: 200, next: item2}
	unit := &playerReadValuesTestObject4EEDC0{
		name:   "unit",
		update: update,
		health: &playerReadValuesTestHealth4EEDC0{name: "health"},
		first:  item1,
	}
	base := &playerReadValuesTestStats4EEDC0{name: "base", values: [4]float32{100, 80, 400, 50}}
	class0 := &playerReadValuesTestStats4EEDC0{name: "class-0", values: [4]float32{200, 100, 500, 100}}
	class1 := &playerReadValuesTestStats4EEDC0{name: "class-1", values: [4]float32{190, 170, 490, 95}}
	return &playerReadValuesTestWorld4EEDC0{
		unit:          unit,
		baseStats:     base,
		classStats:    map[uint8]*playerReadValuesTestStats4EEDC0{0: class0, 1: class1},
		game:          make(map[uint32]int32),
		rewardArg:     -7,
		protectResult: -0x1234567,
		after:         make(map[string]func()),
	}
}

func TestPlayerReadValues4EEDC0InterpolatesAndFinalizesInOracleOrder(t *testing.T) {
	world := newPlayerReadValuesTestWorld4EEDC0()
	result := playerReadValues4EEDC0(world.hooks())
	player := world.unit.update.player
	if result != world.protectResult {
		t.Fatalf("result = %d, want %d", result, world.protectResult)
	}
	if world.unit.health.maximum != 140 || world.unit.lastSetHP != 140 || world.unit.setHPCallCount != 1 {
		t.Fatalf("health max/current/calls = %d/%d/%d, want 140/140/1", world.unit.health.maximum, world.unit.lastSetHP, world.unit.setHPCallCount)
	}
	if world.unit.update.manaMaximum != 120 || world.unit.update.manaCurrent != 120 {
		t.Fatalf("mana max/current = %d/%d, want 120/120", world.unit.update.manaMaximum, world.unit.update.manaCurrent)
	}
	if player.strength != 70 || player.speed != 440 {
		t.Fatalf("strength/speed = %d/%d, want 70/440", player.strength, player.speed)
	}
	if got, want := math.Float32bits(world.unit.speedBase), math.Float32bits(playerReadValuesScaleSpeedExtended4EEDC0(440)); got != want {
		t.Fatalf("speed base bits = %#x, want %#x", got, want)
	}
	if got, want := math.Float32bits(world.unit.mass), math.Float32bits(24); got != want {
		t.Fatalf("mass bits = %#x, want %#x", got, want)
	}
	if player.capacityWord != 2437 || world.unit.carry != 2437 || player.overweight != 0 {
		t.Fatalf("capacity/carry/overweight = %d/%d/%d, want 2437/2437/0", player.capacityWord, world.unit.carry, player.overweight)
	}
	if player.initialized != 1 {
		t.Fatalf("initialized = %d, want 1", player.initialized)
	}
	if len(world.abilities) != 0 {
		t.Fatalf("non-warrior abilities = %#v, want none", world.abilities)
	}
	wantTail := []string{
		"player:player", "strength:player=0x46", "strength-token:player=0x11111111", "protect-int:0x11111111=0x46",
		"player:player", "speed:player=0x1b8", "speed-token:player=0x22222222", "protect-int:0x22222222=0x1b8",
		"player:player", "mana-max:update=120", "mana-token:player=0x33333333", "protect-u16:0x33333333=120",
		"health:health", "player:player", "health-max:health=140", "health-token:player=0x44444444", "protect-u16:0x44444444=140",
		"first-item:item-1", "weight:item-1=200", "next:item-1=item-2", "weight:item-2=250", "next:item-2=nil",
		"carry:unit=2437", "store-overweight:player=0",
		"player:player", "name-token:player=0x55555555", "name:player=Alpha", "wide-len:Alpha=5",
		"player:player", "name:player=Alpha", "protect-name:Alpha:10:0x55555555=-19088743", "store-initialized:player=1",
	}
	if got := world.events[len(world.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("tail events =\n%#v\nwant\n%#v", got, wantTail)
	}
	if first := world.events[:8]; !reflect.DeepEqual(first, []string{
		"arg:unit", "update:update", "player:player", "base-stats:base",
		"class:player=1", "class-stats:1=class-1", "class-stats:0=class-0", "game:0x2000=0",
	}) {
		t.Fatalf("entry events = %#v", first)
	}
}

func TestPlayerReadValues4EEDC0FlagPathAndAbilityShortCircuit(t *testing.T) {
	tests := []struct {
		name        string
		class       uint8
		game1000    int32
		solo        int32
		wantAbility bool
		wantGame    bool
		wantSolo    bool
	}{
		{name: "warrior grants", class: 0, wantAbility: true, wantGame: true, wantSolo: true},
		{name: "class skips", class: 1},
		{name: "game skips solo", class: 0, game1000: -1, wantGame: true},
		{name: "solo skips grant", class: 0, solo: 2, wantGame: true, wantSolo: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			world := newPlayerReadValuesTestWorld4EEDC0()
			world.unit.update.player.class = tc.class
			world.classStats[tc.class] = &playerReadValuesTestStats4EEDC0{name: "selected", values: [4]float32{101.5, 202.5, 303.5, 404.5}}
			world.game[0x2000] = 1
			world.game[0x1000] = tc.game1000
			world.solo = tc.solo
			playerReadValues4EEDC0(world.hooks())

			player := world.unit.update.player
			if world.unit.health.maximum != 102 || world.unit.lastSetHP != 102 || world.unit.update.manaMaximum != 202 || world.unit.update.manaCurrent != 202 || player.strength != 404 || player.speed != 304 {
				t.Fatalf("flag values hp/mana/strength/speed = %d/%d/%d/%d/%d/%d", world.unit.health.maximum, world.unit.lastSetHP, world.unit.update.manaMaximum, world.unit.update.manaCurrent, player.strength, player.speed)
			}
			if got := len(world.abilities) != 0; got != tc.wantAbility {
				t.Fatalf("ability present = %t, want %t", got, tc.wantAbility)
			}
			if tc.wantAbility && world.abilities[0] != (playerReadValuesTestAbility4EEDC0{world.unit, 10, 0}) {
				t.Fatalf("ability = %#v", world.abilities[0])
			}
			gotGame := false
			gotSolo := false
			for _, event := range world.events {
				gotGame = gotGame || event == fmt.Sprintf("game:0x1000=%d", tc.game1000)
				gotSolo = gotSolo || event == fmt.Sprintf("solo=%d", tc.solo)
			}
			if gotGame != tc.wantGame || gotSolo != tc.wantSolo {
				t.Fatalf("game/solo calls = %t/%t, want %t/%t", gotGame, gotSolo, tc.wantGame, tc.wantSolo)
			}
		})
	}
}

func TestPlayerReadValues4EEDC0SignedLevelRule(t *testing.T) {
	tests := []struct {
		input uint8
		want  int8
	}{
		{0, 0}, {10, 10}, {11, 10}, {127, 10}, {128, -128}, {200, -56}, {255, -1},
	}
	for _, tc := range tests {
		if got := playerReadValuesClampLevel4EEDC0(tc.input); got != tc.want {
			t.Errorf("clamp(%#x) = %d, want %d", tc.input, got, tc.want)
		}
	}

	world := newPlayerReadValuesTestWorld4EEDC0()
	world.unit.update.player.class = 0
	world.unit.update.player.level = 200
	playerReadValues4EEDC0(world.hooks())
	if len(world.abilities) != 1 || world.abilities[0].count != -56 {
		t.Fatalf("signed ability count = %#v, want -56", world.abilities)
	}
}

func TestPlayerReadValues4EEDC0ReloadsPlayersButKeepsEntryPlayer(t *testing.T) {
	world := newPlayerReadValuesTestWorld4EEDC0()
	entryPlayer := world.unit.update.player
	secondName := &playerReadValuesTestName4EEDC0{name: "Beta"}
	replacement := &playerReadValuesTestPlayer4EEDC0{
		name:         "replacement",
		capacityWord: 600,
		nameToken:    0x99999999,
		originalName: secondName,
	}
	world.after["store-capacity-word:player=2437"] = func() {
		world.unit.update.player = replacement
	}
	world.after["wide-len:Beta=4"] = func() {
		replacement2 := *replacement
		replacement2.name = "replacement-2"
		replacement2.originalName = &playerReadValuesTestName4EEDC0{name: "Gamma"}
		world.unit.update.player = &replacement2
	}
	playerReadValues4EEDC0(world.hooks())

	if entryPlayer.initialized != 1 || entryPlayer.overweight != 0 {
		t.Fatalf("cached entry player initialized/overweight = %d/%d", entryPlayer.initialized, entryPlayer.overweight)
	}
	if world.unit.carry != 600 {
		t.Fatalf("carry from reloaded player = %d, want 600", world.unit.carry)
	}
	wantNameTail := []string{
		"player:replacement", "name-token:replacement=0x99999999", "name:replacement=Beta", "wide-len:Beta=4",
		"player:replacement-2", "name:replacement-2=Gamma", "protect-name:Gamma:8:0x99999999=-19088743", "store-initialized:player=1",
	}
	if got := world.events[len(world.events)-len(wantNameTail):]; !reflect.DeepEqual(got, wantNameTail) {
		t.Fatalf("name reload events = %#v, want %#v", got, wantNameTail)
	}
}

func TestPlayerReadValues4EEDC0FaultsStopAtEveryObservedBoundary(t *testing.T) {
	baseline := newPlayerReadValuesTestWorld4EEDC0()
	playerReadValues4EEDC0(baseline.hooks())
	want := baseline.events
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		world := newPlayerReadValuesTestWorld4EEDC0()
		world.faultAt = faultAt
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("fault %d did not stop execution", faultAt)
				}
			}()
			playerReadValues4EEDC0(world.hooks())
		}()
		if !reflect.DeepEqual(world.events, want[:faultAt]) {
			t.Fatalf("fault %d events = %#v, want %#v", faultAt, world.events, want[:faultAt])
		}
	}
}

func TestPlayerReadValues4EEDC0ArithmeticEdges(t *testing.T) {
	if got := math.Float64bits(playerReadValuesInterpolate4EEDC0(190, 100, 100, 4)); got != 0x4061800000a00000 {
		t.Fatalf("interpolation bits = %#x, want 0x4061800000a00000", got)
	}
	if got := math.Float32bits(playerReadValuesScaleSpeed4EEDC0(440)); got != 0x3d343958 {
		t.Fatalf("scaled speed bits = %#x, want 0x3d343958", got)
	}
	if got := math.Float32bits(playerReadValuesMass4EEDC0(0xffffffce, 100)); got != math.Float32bits(0) {
		t.Fatalf("signed strength mass bits = %#x, want zero", got)
	}
	if got := playerReadValuesCapacity4EEDC0(70, 100); got != 2437.5 {
		t.Fatalf("capacity = %g, want 2437.5", got)
	}
	if got := playerReadValuesAddWeight4EEDC0(math.MaxInt32, 255); got != math.MinInt32+254 {
		t.Fatalf("wrapped inventory sum = %d, want %d", got, int32(math.MinInt32+254))
	}
	if playerReadValuesOverweight4EEDC0(-1, 0) || !playerReadValuesOverweight4EEDC0(65536, math.MaxUint16) {
		t.Fatal("signed inventory/capacity comparison mismatch")
	}
}
