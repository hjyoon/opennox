package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerCollideTestHealth4E8460 struct {
	name    string
	current uint16
	maximum uint16
}

type playerCollideTestWall4E8460 struct {
	name string
	tile uint8
}

type playerCollideTestObject4E8460 struct {
	name       string
	class      uint32
	flags      uint32
	health     *playerCollideTestHealth4E8460
	mass       float32
	doorState  uint8
	parent     *playerCollideTestObject4E8460
	wall       *playerCollideTestWall4E8460
	newX       float32
	newY       float32
	hasDeath   int32
	deathTimer uint32
	deathPower uint32
}

type playerCollideTestState4E8460 struct {
	events       []string
	ability      int32
	balances     map[string]float64
	wallFlags    uint32
	frameRates   []uint32
	damageMutate func(*playerCollideTestObject4E8460)
}

func (s *playerCollideTestState4E8460) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *playerCollideTestState4E8460) hooks() playerCollideHooks4E8460[
	*playerCollideTestObject4E8460,
	*playerCollideTestHealth4E8460,
	*playerCollideTestWall4E8460,
	*uint32,
] {
	return playerCollideHooks4E8460[
		*playerCollideTestObject4E8460,
		*playerCollideTestHealth4E8460,
		*playerCollideTestWall4E8460,
		*uint32,
	]{
		abilityActive: func(obj *playerCollideTestObject4E8460, ability uint32) int32 {
			s.event("ability:%s:%d", obj.name, ability)
			return s.ability
		},
		class: func(obj *playerCollideTestObject4E8460) uint32 {
			s.event("class:%s", obj.name)
			return obj.class
		},
		flags: func(obj *playerCollideTestObject4E8460) uint32 {
			s.event("flags:%s", obj.name)
			return obj.flags
		},
		flagsLow: func(obj *playerCollideTestObject4E8460) uint8 {
			s.event("flags-low:%s", obj.name)
			return uint8(obj.flags)
		},
		health: func(obj *playerCollideTestObject4E8460) *playerCollideTestHealth4E8460 {
			s.event("health:%s", obj.name)
			return obj.health
		},
		healthCurrent: func(health *playerCollideTestHealth4E8460) uint16 {
			s.event("health-current:%s", health.name)
			return health.current
		},
		healthMax: func(health *playerCollideTestHealth4E8460) uint16 {
			s.event("health-max:%s", health.name)
			return health.maximum
		},
		mass: func(obj *playerCollideTestObject4E8460) float32 {
			s.event("mass:%s", obj.name)
			return obj.mass
		},
		doorState: func(obj *playerCollideTestObject4E8460) uint8 {
			s.event("door:%s", obj.name)
			return obj.doorState
		},
		setState: func(obj *playerCollideTestObject4E8460, state uint32) {
			s.event("set-state:%s:%d", obj.name, state)
		},
		earthquake: func(obj *playerCollideTestObject4E8460, amount int32) {
			s.event("earthquake:%s:%d", obj.name, amount)
		},
		disableAbility: func(obj *playerCollideTestObject4E8460, ability uint32) {
			s.event("disable-ability:%s:%d", obj.name, ability)
		},
		balanceFloat: func(key string) float64 {
			s.event("balance:%s", key)
			return s.balances[key]
		},
		floatToInt: func(value float32) int32 {
			s.event("round:%08x", math.Float32bits(value))
			return playerCollideRound4E8460(value)
		},
		bounce: func(player, other *playerCollideTestObject4E8460) {
			s.event("bounce:%s:%s", player.name, other.name)
		},
		findParent: func(obj *playerCollideTestObject4E8460) *playerCollideTestObject4E8460 {
			s.event("parent:%s", obj.name)
			return obj.parent
		},
		damage: func(target, source, attacker *playerCollideTestObject4E8460, damage int32, damageType uint32) {
			s.event("damage:%s:%s:%s:%d:%d", target.name, source.name, attacker.name, damage, damageType)
			if s.damageMutate != nil {
				s.damageMutate(target)
			}
		},
		applyEnchant: func(obj *playerCollideTestObject4E8460, enchant, duration, power uint32) {
			s.event("apply:%s:%d:%d:%d", obj.name, enchant, duration, power)
		},
		collisionWall: func(obj *playerCollideTestObject4E8460) *playerCollideTestWall4E8460 {
			s.event("collision-wall:%s", obj.name)
			return obj.wall
		},
		wallTile: func(wall *playerCollideTestWall4E8460) uint8 {
			s.event("wall-tile:%s", wall.name)
			return wall.tile
		},
		wallFlags: func(tile uint8) uint32 {
			s.event("wall-flags:%d", tile)
			return s.wallFlags
		},
		audio: func(id uint32, obj *playerCollideTestObject4E8460, kind int32, code uint32) {
			s.event("audio:%d:%s:%d:%d", id, obj.name, kind, code)
		},
		newPosY: func(obj *playerCollideTestObject4E8460) float32 {
			s.event("new-y:%s", obj.name)
			return obj.newY
		},
		newPosX: func(obj *playerCollideTestObject4E8460) float32 {
			s.event("new-x:%s", obj.name)
			return obj.newX
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *playerCollideTestObject4E8460) {
			s.event("damage-map:%d:%d:%d:%d:%s", x, y, damage, damageType, source.name)
		},
		damageClear: func(obj *playerCollideTestObject4E8460, damage int32) {
			s.event("damage-clear:%s:%d", obj.name, damage)
		},
		move: func(obj *playerCollideTestObject4E8460) {
			s.event("move:%s", obj.name)
		},
		hasEnchant: func(obj *playerCollideTestObject4E8460, enchant uint32) int32 {
			s.event("has-enchant:%s:%d", obj.name, enchant)
			return obj.hasDeath
		},
		enchantTimer: func(obj *playerCollideTestObject4E8460, enchant uint32) uint32 {
			s.event("enchant-timer:%s:%d", obj.name, enchant)
			return obj.deathTimer
		},
		frameRate: func() uint32 {
			s.event("frame-rate")
			value := s.frameRates[0]
			s.frameRates = s.frameRates[1:]
			return value
		},
		enchantPower: func(obj *playerCollideTestObject4E8460, enchant uint32) uint32 {
			s.event("enchant-power:%s:%d", obj.name, enchant)
			return obj.deathPower
		},
		disableEnchant: func(obj *playerCollideTestObject4E8460, enchant uint32) {
			s.event("disable-enchant:%s:%d", obj.name, enchant)
		},
	}
}

func TestPlayerCollide4E8460PlayerHitMovesAndTransfersDeathEnchant(t *testing.T) {
	parent := &playerCollideTestObject4E8460{name: "parent"}
	player := &playerCollideTestObject4E8460{
		name: "player", health: &playerCollideTestHealth4E8460{name: "player", current: 5, maximum: 20},
		mass: 2, parent: parent, hasDeath: 1, deathTimer: 419, deathPower: 0x1ab,
	}
	other := &playerCollideTestObject4E8460{
		name: "other", class: playerCollidePlayerClass4E8460,
		health: &playerCollideTestHealth4E8460{name: "other", current: 10, maximum: 10}, mass: 3,
	}
	collision := uint32(0xa5a55a5a)
	state := &playerCollideTestState4E8460{
		ability: 1,
		balances: map[string]float64{
			"BerserkerDamage":       12.5,
			"BerserkerStunDuration": 7.5,
			"BerserkerPainRatio":    0.5,
		},
		frameRates: []uint32{30, 31},
	}
	playerCollide4E8460(player, other, &collision, state.hooks())

	want := []string{
		"ability:player:1", "class:other", "health:other", "health-current:other", "flags-low:other",
		"set-state:player:13", "earthquake:player:10", "disable-ability:player:1",
		"balance:BerserkerDamage", "round:41480000", "class:other", "bounce:player:other",
		"parent:player", "damage:other:parent:player:12:2", "class:other", "move:player",
		"class:other", "flags:other",
		"has-enchant:player:16", "enchant-timer:player:16", "frame-rate",
		"enchant-power:player:16", "frame-rate", "apply:other:16:465:171", "disable-enchant:player:16",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", state.events, want)
	}
	if collision != 0xa5a55a5a {
		t.Fatalf("collision data changed: %#x", collision)
	}
	if len(state.frameRates) != 0 {
		t.Fatalf("unused frame rates = %v", state.frameRates)
	}
}

func TestPlayerCollide4E8460NonUnitHitStunsAndClearsPain(t *testing.T) {
	parent := &playerCollideTestObject4E8460{name: "parent"}
	player := &playerCollideTestObject4E8460{
		name: "player", mass: 2, parent: parent,
		health: &playerCollideTestHealth4E8460{name: "player", current: 5, maximum: 20},
	}
	other := &playerCollideTestObject4E8460{name: "other", class: 8, mass: 3}
	state := &playerCollideTestState4E8460{
		ability: 1,
		balances: map[string]float64{
			"BerserkerDamage":       12.5,
			"BerserkerStunDuration": 65537.5,
			"BerserkerPainRatio":    0.5,
		},
	}
	playerCollide4E8460(player, other, (*uint32)(nil), state.hooks())

	want := []string{
		"ability:player:1", "class:other", "flags-low:other", "mass:other", "mass:player",
		"flags-low:other", "set-state:player:13", "earthquake:player:10", "disable-ability:player:1",
		"balance:BerserkerDamage", "round:41480000", "class:other", "bounce:player:other",
		"parent:player", "damage:other:parent:player:12:2", "class:other",
		"balance:BerserkerStunDuration", "round:478000c0", "apply:player:5:2:5",
		"balance:BerserkerPainRatio", "health:player", "health-current:player", "round:40200000",
		"damage-clear:player:2", "move:player", "class:other",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", state.events, want)
	}
}

func TestPlayerCollide4E8460WallImpactOrderAndGridConversion(t *testing.T) {
	player := &playerCollideTestObject4E8460{
		name: "player", newX: 23, newY: 46,
		health: &playerCollideTestHealth4E8460{name: "player", current: 1, maximum: 1},
		wall:   &playerCollideTestWall4E8460{name: "wall", tile: 9},
	}
	state := &playerCollideTestState4E8460{
		ability:   1,
		wallFlags: 5,
		balances: map[string]float64{
			"BerserkerStunDuration": 1.5,
			"BerserkerPainRatio":    0,
		},
	}
	playerCollide4E8460(player, (*playerCollideTestObject4E8460)(nil), (*uint32)(nil), state.hooks())
	want := []string{
		"ability:player:1", "set-state:player:13", "earthquake:player:10", "disable-ability:player:1",
		"collision-wall:player", "wall-tile:wall", "wall-flags:9", "audio:171:player:0:0",
		"balance:BerserkerStunDuration", "round:3fc00000", "apply:player:5:2:5",
		"new-y:player", "round:40000000", "new-x:player", "round:3f800000",
		"damage-map:1:2:100:2:player", "balance:BerserkerPainRatio", "health:player",
		"health-current:player", "round:00000000", "damage-clear:player:1", "move:player",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", state.events, want)
	}
}

func TestPlayerCollide4E8460MassDoorAndFlagGates(t *testing.T) {
	tests := []struct {
		name       string
		class      uint32
		flags      uint32
		current    uint16
		maximum    uint16
		playerMass float32
		otherMass  float32
		doorState  uint8
		wantHit    bool
		wantMass   bool
	}{
		{name: "non-unit equal mass", class: 8, playerMass: 4, otherMass: 4, wantMass: true},
		{name: "non-unit greater mass", class: 8, playerMass: 4, otherMass: 5, wantHit: true, wantMass: true},
		{name: "non-unit unordered mass", class: 8, playerMass: 4, otherMass: float32(math.NaN()), wantMass: true},
		{name: "class bypasses mass", class: 0x400008, playerMass: 9, otherMass: 1, wantHit: true},
		{name: "flag bypasses mass", class: 8, flags: 0x80, playerMass: 9, otherMass: 1, wantHit: true},
		{name: "live unit skips mass", class: 2, current: 1, maximum: 1, playerMass: 9, otherMass: 1, wantHit: true},
		{name: "zero health with maximum checks mass", class: 2, maximum: 1, playerMass: 9, otherMass: 1, wantMass: true},
		{name: "door closed", class: 0x80, flags: 0x80, doorState: 0},
		{name: "door open", class: 0x80, flags: 0x80, doorState: 1, wantHit: true},
		{name: "reject object flag", class: 8, flags: 0x81, doorState: 1},
		{name: "reject class bit", class: 0x400001, doorState: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			player := &playerCollideTestObject4E8460{
				name: "player", mass: tc.playerMass, parent: &playerCollideTestObject4E8460{name: "parent"},
				health: &playerCollideTestHealth4E8460{name: "player", current: 1, maximum: 1},
			}
			other := &playerCollideTestObject4E8460{
				name: "other", class: tc.class, flags: tc.flags, mass: tc.otherMass, doorState: tc.doorState,
				health: &playerCollideTestHealth4E8460{name: "other", current: tc.current, maximum: tc.maximum},
			}
			state := &playerCollideTestState4E8460{ability: 1, balances: map[string]float64{}}
			playerCollide4E8460(player, other, (*uint32)(nil), state.hooks())
			gotHit, gotMass := false, false
			for _, event := range state.events {
				gotHit = gotHit || event == "set-state:player:13"
				gotMass = gotMass || event == "mass:other"
			}
			if gotHit != tc.wantHit || gotMass != tc.wantMass {
				t.Fatalf("hit/mass = %v/%v, want %v/%v; events=%v", gotHit, gotMass, tc.wantHit, tc.wantMass, state.events)
			}
		})
	}
}

func TestPlayerCollide4E8460PostDamageMoveSkipsPainButStillTransfers(t *testing.T) {
	parent := &playerCollideTestObject4E8460{name: "parent"}
	player := &playerCollideTestObject4E8460{
		name: "player", parent: parent, hasDeath: 1, deathTimer: 0, deathPower: 3,
		health: &playerCollideTestHealth4E8460{name: "player", current: 10, maximum: 10},
	}
	other := &playerCollideTestObject4E8460{
		name: "other", class: 4, health: &playerCollideTestHealth4E8460{name: "other", current: 1, maximum: 1},
	}
	state := &playerCollideTestState4E8460{
		ability: 1, balances: map[string]float64{"BerserkerDamage": 1}, frameRates: []uint32{30, 30},
		damageMutate: func(target *playerCollideTestObject4E8460) { target.class = 0x20004 },
	}
	playerCollide4E8460(player, other, (*uint32)(nil), state.hooks())
	for _, event := range state.events {
		if event == "balance:BerserkerStunDuration" || event == "balance:BerserkerPainRatio" || event == "damage-clear:player:1" {
			t.Fatalf("post-damage early move crossed skipped work: %v", state.events)
		}
	}
	wantSuffix := []string{
		"move:player", "class:other", "flags:other", "has-enchant:player:16", "enchant-timer:player:16",
		"frame-rate", "enchant-power:player:16", "frame-rate", "apply:other:16:450:3", "disable-enchant:player:16",
	}
	if got := state.events[len(state.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("suffix = %v, want %v", got, wantSuffix)
	}
}

func TestPlayerCollide4E8460DeathTransferThresholdAndGuards(t *testing.T) {
	tests := []struct {
		name      string
		class     uint32
		flags     uint32
		hasDeath  int32
		timer     uint32
		wantApply bool
	}{
		{name: "not player", class: 2, hasDeath: 1},
		{name: "dead", class: 4, flags: playerCollideDeadFlag4E8460, hasDeath: 1},
		{name: "missing enchant", class: 4},
		{name: "equal threshold", class: 4, hasDeath: 1, timer: 420},
		{name: "strictly below", class: 4, hasDeath: 1, timer: 419, wantApply: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			player := &playerCollideTestObject4E8460{name: "player", deathPower: 0x1ff}
			other := &playerCollideTestObject4E8460{name: "other", class: tc.class, flags: tc.flags}
			player.hasDeath, player.deathTimer = tc.hasDeath, tc.timer
			state := &playerCollideTestState4E8460{balances: map[string]float64{}, frameRates: []uint32{30, 31}}
			playerCollide4E8460(player, other, (*uint32)(nil), state.hooks())
			gotApply := false
			for _, event := range state.events {
				gotApply = gotApply || event == "apply:other:16:465:255"
			}
			if gotApply != tc.wantApply {
				t.Fatalf("apply = %v, want %v; events=%v", gotApply, tc.wantApply, state.events)
			}
			if tc.wantApply && len(state.frameRates) != 0 {
				t.Fatalf("two live FPS values were not consumed: %v", state.frameRates)
			}
		})
	}
}

func TestPlayerCollideRound4E8460MatchesX87FISTPDefaultMode(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {3.5, 4},
		{-0.5, 0}, {-1.5, -2}, {-2.5, -2},
		{math.Float32frombits(0x4effffff), 2147483520},
		{math.Float32frombits(0x4f000000), math.MinInt32},
		{math.Float32frombits(0xcf000000), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
		{float32(math.NaN()), math.MinInt32},
	}
	for _, tc := range tests {
		if got := playerCollideRound4E8460(tc.value); got != tc.want {
			t.Errorf("round(%08x) = %d, want %d", math.Float32bits(tc.value), got, tc.want)
		}
	}
}
