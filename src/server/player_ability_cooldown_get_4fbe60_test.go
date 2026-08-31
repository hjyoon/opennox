package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type playerAbilityCooldownGetTestUnit4FBE60 struct {
	netCode uint32
}

type playerAbilityCooldownGetTestPlayer4FBE60 struct {
	index uint8
}

type playerAbilityCooldownGetTestKey4FBE60 struct {
	index   uint8
	ability Ability
}

type playerAbilityCooldownGetTestWorld4FBE60 struct {
	events     []string
	faultAt    int
	afterEvent map[string]func()
	players    map[uint32]*playerAbilityCooldownGetTestPlayer4FBE60
	cooldowns  map[playerAbilityCooldownGetTestKey4FBE60]int32
}

func (w *playerAbilityCooldownGetTestWorld4FBE60) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerAbilityCooldownGetTestWorld4FBE60) finish(event string) {
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *playerAbilityCooldownGetTestWorld4FBE60) hooks() playerAbilityCooldownGetHooks4FBE60[
	*playerAbilityCooldownGetTestUnit4FBE60,
	*playerAbilityCooldownGetTestPlayer4FBE60,
] {
	return playerAbilityCooldownGetHooks4FBE60[
		*playerAbilityCooldownGetTestUnit4FBE60,
		*playerAbilityCooldownGetTestPlayer4FBE60,
	]{
		loadNetCode: func(unit *playerAbilityCooldownGetTestUnit4FBE60) uint32 {
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
		playerByNetCode: func(netCode uint32) *playerAbilityCooldownGetTestPlayer4FBE60 {
			player := w.players[netCode]
			event := fmt.Sprintf("player:%08x=%p", netCode, player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadPlayerIndex: func(player *playerAbilityCooldownGetTestPlayer4FBE60) uint8 {
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
		loadCooldown: func(index uint8, ability Ability) int32 {
			value := w.cooldowns[playerAbilityCooldownGetTestKey4FBE60{index: index, ability: ability}]
			event := fmt.Sprintf("cooldown:%02x:%d=%d", index, ability, value)
			w.record(event)
			w.finish(event)
			return value
		},
	}
}

func TestPlayerAbilityCooldownGet4FBE60OrderWidthsAndLiveValue(t *testing.T) {
	unit := &playerAbilityCooldownGetTestUnit4FBE60{netCode: 0xfedcba98}
	player := &playerAbilityCooldownGetTestPlayer4FBE60{index: 0xfe}
	key := playerAbilityCooldownGetTestKey4FBE60{index: 0xfe, ability: Ability(-2)}
	w := playerAbilityCooldownGetTestWorld4FBE60{
		afterEvent: make(map[string]func()),
		players:    map[uint32]*playerAbilityCooldownGetTestPlayer4FBE60{unit.netCode: player},
		cooldowns:  map[playerAbilityCooldownGetTestKey4FBE60]int32{key: 7},
	}
	w.afterEvent["netcode:fedcba98"] = func() {
		unit.netCode = 1
	}
	w.afterEvent["index:fe"] = func() {
		w.cooldowns[key] = math.MinInt32
	}

	got := playerAbilityCooldownGet4FBE60(unit, Ability(-2), w.hooks())
	if got != math.MinInt32 {
		t.Fatalf("cooldown = %d, want %d", got, int32(math.MinInt32))
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

func TestPlayerAbilityCooldownGet4FBE60MissingPlayerReturnsZero(t *testing.T) {
	unit := &playerAbilityCooldownGetTestUnit4FBE60{netCode: 0x12345678}
	w := playerAbilityCooldownGetTestWorld4FBE60{
		players:   make(map[uint32]*playerAbilityCooldownGetTestPlayer4FBE60),
		cooldowns: make(map[playerAbilityCooldownGetTestKey4FBE60]int32),
	}
	if got := playerAbilityCooldownGet4FBE60(unit, AbilityInfravis, w.hooks()); got != 0 {
		t.Fatalf("cooldown = %d, want 0", got)
	}
	want := []string{"netcode:12345678", "player:12345678=0x0"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestPlayerAbilityCooldownGet4FBE60AllPlayerAbilitySlots(t *testing.T) {
	unit := new(playerAbilityCooldownGetTestUnit4FBE60)
	player := new(playerAbilityCooldownGetTestPlayer4FBE60)
	w := playerAbilityCooldownGetTestWorld4FBE60{
		players:   map[uint32]*playerAbilityCooldownGetTestPlayer4FBE60{0: player},
		cooldowns: make(map[playerAbilityCooldownGetTestKey4FBE60]int32),
	}
	hooks := w.hooks()
	for index := 0; index < abilityRuntimePlayerSlots4FB990; index++ {
		player.index = uint8(index)
		for ability := AbilityInvalid; ability < AbilityMax; ability++ {
			want := int32(index*int(AbilityMax) + int(ability))
			w.cooldowns[playerAbilityCooldownGetTestKey4FBE60{uint8(index), ability}] = want
			if got := playerAbilityCooldownGet4FBE60(unit, ability, hooks); got != want {
				t.Fatalf("slot %d ability %d = %d, want %d", index, ability, got, want)
			}
		}
	}
}

func TestPlayerAbilityCooldownGet4FBE60FaultPrefixes(t *testing.T) {
	unit := &playerAbilityCooldownGetTestUnit4FBE60{netCode: 7}
	player := &playerAbilityCooldownGetTestPlayer4FBE60{index: 3}
	for faultAt, want := range [][]string{
		{"netcode:00000007"},
		{"netcode:00000007", fmt.Sprintf("player:00000007=%p", player)},
		{"netcode:00000007", fmt.Sprintf("player:00000007=%p", player), "index:03"},
		{"netcode:00000007", fmt.Sprintf("player:00000007=%p", player), "index:03", "cooldown:03:4=99"},
	} {
		t.Run(fmt.Sprintf("fault-%d", faultAt+1), func(t *testing.T) {
			w := playerAbilityCooldownGetTestWorld4FBE60{
				faultAt:   faultAt + 1,
				players:   map[uint32]*playerAbilityCooldownGetTestPlayer4FBE60{7: player},
				cooldowns: map[playerAbilityCooldownGetTestKey4FBE60]int32{{3, AbilityTreadLightly}: 99},
			}
			defer func() {
				if recover() == nil {
					t.Fatal("fault was not propagated")
				}
				if !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %#v, want %#v", w.events, want)
				}
			}()
			playerAbilityCooldownGet4FBE60(unit, AbilityTreadLightly, w.hooks())
		})
	}
}

func TestPlayerAbilityCooldownGet4FBE60NilUnitFaultsOnNetCode(t *testing.T) {
	w := playerAbilityCooldownGetTestWorld4FBE60{
		players:   make(map[uint32]*playerAbilityCooldownGetTestPlayer4FBE60),
		cooldowns: make(map[playerAbilityCooldownGetTestKey4FBE60]int32),
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil NetCode load did not fault")
		}
		if want := []string{"netcode:nil"}; !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %#v, want %#v", w.events, want)
		}
	}()
	playerAbilityCooldownGet4FBE60[*playerAbilityCooldownGetTestUnit4FBE60](nil, AbilityBerserk, w.hooks())
}
