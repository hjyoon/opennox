package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerAbilityCooldownSetTestUnit4FBEA0 struct {
	netCode uint32
}

type playerAbilityCooldownSetTestPlayer4FBEA0 struct {
	index uint8
}

type playerAbilityCooldownSetTestKey4FBEA0 struct {
	index   uint8
	ability Ability
}

type playerAbilityCooldownSetTestWorld4FBEA0 struct {
	events     []string
	faultAt    int
	afterEvent map[string]func()
	players    map[uint32]*playerAbilityCooldownSetTestPlayer4FBEA0
	cooldowns  map[playerAbilityCooldownSetTestKey4FBEA0]int32
}

func (w *playerAbilityCooldownSetTestWorld4FBEA0) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerAbilityCooldownSetTestWorld4FBEA0) finish(event string) {
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *playerAbilityCooldownSetTestWorld4FBEA0) hooks() playerAbilityCooldownSetHooks4FBEA0[
	*playerAbilityCooldownSetTestUnit4FBEA0,
	*playerAbilityCooldownSetTestPlayer4FBEA0,
] {
	return playerAbilityCooldownSetHooks4FBEA0[
		*playerAbilityCooldownSetTestUnit4FBEA0,
		*playerAbilityCooldownSetTestPlayer4FBEA0,
	]{
		loadNetCode: func(unit *playerAbilityCooldownSetTestUnit4FBEA0) uint32 {
			if unit == nil {
				w.record("netcode:nil")
				panic("netcode:nil")
			}
			value := unit.netCode
			event := fmt.Sprintf("netcode:%08x", value)
			w.record(event)
			w.finish(event)
			return value
		},
		playerByNetCode: func(netCode uint32) *playerAbilityCooldownSetTestPlayer4FBEA0 {
			player := w.players[netCode]
			event := fmt.Sprintf("player:%08x=%p", netCode, player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadPlayerIndex: func(player *playerAbilityCooldownSetTestPlayer4FBEA0) uint8 {
			if player == nil {
				w.record("index:nil")
				panic("index:nil")
			}
			value := player.index
			event := fmt.Sprintf("index:%02x", value)
			w.record(event)
			w.finish(event)
			return value
		},
		storeCooldown: func(index uint8, ability Ability, cooldown int32) {
			event := fmt.Sprintf("cooldown:%02x:%d=%d", index, ability, cooldown)
			w.record(event)
			w.finish(event)
			w.cooldowns[playerAbilityCooldownSetTestKey4FBEA0{index: index, ability: ability}] = cooldown
		},
	}
}

func TestPlayerAbilityCooldownSet4FBEA0OrderWidthsAndReturn(t *testing.T) {
	unit := &playerAbilityCooldownSetTestUnit4FBEA0{netCode: 0xfedcba98}
	player := &playerAbilityCooldownSetTestPlayer4FBEA0{index: 0xfe}
	key := playerAbilityCooldownSetTestKey4FBEA0{index: 0xfe, ability: Ability(-2)}
	w := playerAbilityCooldownSetTestWorld4FBEA0{
		afterEvent: make(map[string]func()),
		players:    map[uint32]*playerAbilityCooldownSetTestPlayer4FBEA0{unit.netCode: player},
		cooldowns:  make(map[playerAbilityCooldownSetTestKey4FBEA0]int32),
	}
	w.afterEvent["netcode:fedcba98"] = func() {
		unit.netCode = 1
	}
	w.afterEvent["index:fe"] = func() {
		player.index = 3
	}

	if got := playerAbilityCooldownSet4FBEA0(unit, Ability(-2), math.MinInt32, w.hooks()); got != math.MinInt32 {
		t.Fatalf("return = %d, want %d", got, int32(math.MinInt32))
	}
	if got := w.cooldowns[key]; got != math.MinInt32 {
		t.Fatalf("stored cooldown = %d, want %d", got, int32(math.MinInt32))
	}
	want := []string{
		"netcode:fedcba98",
		fmt.Sprintf("player:fedcba98=%p", player),
		"index:fe",
		fmt.Sprintf("cooldown:fe:-2=%d", int32(math.MinInt32)),
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestPlayerAbilityCooldownSet4FBEA0MissingPlayerReturnsZeroWithoutStore(t *testing.T) {
	unit := &playerAbilityCooldownSetTestUnit4FBEA0{netCode: 0x12345678}
	w := playerAbilityCooldownSetTestWorld4FBEA0{
		players:   make(map[uint32]*playerAbilityCooldownSetTestPlayer4FBEA0),
		cooldowns: make(map[playerAbilityCooldownSetTestKey4FBEA0]int32),
	}
	if got := playerAbilityCooldownSet4FBEA0(unit, AbilityInfravis, 77, w.hooks()); got != 0 {
		t.Fatalf("return = %d, want 0", got)
	}
	if len(w.cooldowns) != 0 {
		t.Fatalf("cooldowns = %#v, want no store", w.cooldowns)
	}
	want := []string{"netcode:12345678", "player:12345678=0x0"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestPlayerAbilityCooldownSet4FBEA0AllPlayerAbilitySlots(t *testing.T) {
	unit := new(playerAbilityCooldownSetTestUnit4FBEA0)
	player := new(playerAbilityCooldownSetTestPlayer4FBEA0)
	w := playerAbilityCooldownSetTestWorld4FBEA0{
		players:   map[uint32]*playerAbilityCooldownSetTestPlayer4FBEA0{0: player},
		cooldowns: make(map[playerAbilityCooldownSetTestKey4FBEA0]int32),
	}
	hooks := w.hooks()
	for index := 0; index < abilityRuntimePlayerSlots4FB990; index++ {
		player.index = uint8(index)
		for ability := AbilityInvalid; ability < AbilityMax; ability++ {
			want := int32(index*int(AbilityMax) + int(ability))
			if got := playerAbilityCooldownSet4FBEA0(unit, ability, want, hooks); got != want {
				t.Fatalf("slot %d ability %d return = %d, want %d", index, ability, got, want)
			}
			key := playerAbilityCooldownSetTestKey4FBEA0{uint8(index), ability}
			if got := w.cooldowns[key]; got != want {
				t.Fatalf("slot %d ability %d stored = %d, want %d", index, ability, got, want)
			}
		}
	}
}

func TestPlayerAbilityCooldownSet4FBEA0FaultPrefixes(t *testing.T) {
	unit := &playerAbilityCooldownSetTestUnit4FBEA0{netCode: 7}
	player := &playerAbilityCooldownSetTestPlayer4FBEA0{index: 3}
	for faultAt, want := range [][]string{
		{"netcode:00000007"},
		{"netcode:00000007", fmt.Sprintf("player:00000007=%p", player)},
		{"netcode:00000007", fmt.Sprintf("player:00000007=%p", player), "index:03"},
		{"netcode:00000007", fmt.Sprintf("player:00000007=%p", player), "index:03", "cooldown:03:4=99"},
	} {
		t.Run(fmt.Sprintf("fault-%d", faultAt+1), func(t *testing.T) {
			w := playerAbilityCooldownSetTestWorld4FBEA0{
				faultAt:   faultAt + 1,
				players:   map[uint32]*playerAbilityCooldownSetTestPlayer4FBEA0{7: player},
				cooldowns: make(map[playerAbilityCooldownSetTestKey4FBEA0]int32),
			}
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %#v, want %#v", w.events, want)
				}
				if len(w.cooldowns) != 0 {
					t.Fatalf("cooldowns = %#v, want no completed store", w.cooldowns)
				}
			}()
			playerAbilityCooldownSet4FBEA0(unit, AbilityTreadLightly, 99, w.hooks())
		})
	}
}

func TestPlayerAbilityCooldownSet4FBEA0NilUnitFaultsOnNetCode(t *testing.T) {
	w := playerAbilityCooldownSetTestWorld4FBEA0{
		players:   make(map[uint32]*playerAbilityCooldownSetTestPlayer4FBEA0),
		cooldowns: make(map[playerAbilityCooldownSetTestKey4FBEA0]int32),
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil NetCode load did not fault")
		}
		if want := []string{"netcode:nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %#v, want %#v", w.events, want)
		}
	}()
	playerAbilityCooldownSet4FBEA0[*playerAbilityCooldownSetTestUnit4FBEA0](nil, AbilityBerserk, 1, w.hooks())
}
