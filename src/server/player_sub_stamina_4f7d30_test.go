package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerSubStaminaTestWorld4F7D30 struct {
	class          uint32
	amount         int32
	playerStamina  uint8
	monsterStamina uint8
	player         string
	playerIndex    map[string]uint8
	events         []string
	reported       uint8
}

func (w *playerSubStaminaTestWorld4F7D30) hooks() playerSubStaminaHooks4F7D30[int, string, string, string] {
	return playerSubStaminaHooks4F7D30[int, string, string, string]{
		loadClass: func(int) uint32 {
			w.events = append(w.events, "class")
			return w.class
		},
		loadPlayerUpdate: func(int) string {
			w.events = append(w.events, "player-update")
			return "player-update"
		},
		loadMonsterUpdate: func(int) string {
			w.events = append(w.events, "monster-update")
			return "monster-update"
		},
		loadAmount: func() int32 {
			w.events = append(w.events, "amount")
			return w.amount
		},
		loadPlayerStamina: func(string) uint8 {
			w.events = append(w.events, "player-stamina")
			return w.playerStamina
		},
		storePlayerStamina: func(_ string, stamina uint8) {
			w.events = append(w.events, fmt.Sprintf("store-player=%d", stamina))
			w.playerStamina = stamina
		},
		loadPlayer: func(string) string {
			w.events = append(w.events, "player")
			return w.player
		},
		loadPlayerIndex: func(player string) uint8 {
			w.events = append(w.events, "player-index="+player)
			return w.playerIndex[player]
		},
		reportStamina: func(index uint8, _ int) {
			w.events = append(w.events, fmt.Sprintf("report=%d", index))
			w.reported = w.playerStamina
		},
		loadMonsterStamina: func(string) uint8 {
			w.events = append(w.events, "monster-stamina")
			return w.monsterStamina
		},
		storeMonsterStamina: func(_ string, stamina uint8) {
			w.events = append(w.events, fmt.Sprintf("store-monster=%d", stamina))
			w.monsterStamina = stamina
		},
	}
}

func TestPlayerSubStamina4F7D30BranchesSignedCompareAndLowByteArithmetic(t *testing.T) {
	tests := []struct {
		name           string
		class          uint32
		amount         int32
		playerStamina  uint8
		monsterStamina uint8
		wantResult     int32
		wantPlayer     uint8
		wantMonster    uint8
		wantReported   uint8
		wantEvents     []string
	}{
		{
			name:           "player success",
			class:          0x80000004,
			amount:         45,
			playerStamina:  100,
			monsterStamina: 77,
			wantResult:     1,
			wantPlayer:     55,
			wantMonster:    77,
			wantReported:   55,
			wantEvents: []string{
				"class", "player-update", "amount", "player-stamina",
				"store-player=55", "player", "player-index=player", "report=7",
			},
		},
		{
			name:           "player takes precedence over monster",
			class:          6,
			amount:         100,
			playerStamina:  100,
			monsterStamina: 88,
			wantResult:     1,
			wantPlayer:     0,
			wantMonster:    88,
			wantEvents: []string{
				"class", "player-update", "amount", "player-stamina",
				"store-player=0", "player", "player-index=player", "report=7",
			},
		},
		{
			name:           "player insufficient",
			class:          4,
			amount:         46,
			playerStamina:  45,
			monsterStamina: 77,
			wantPlayer:     45,
			wantMonster:    77,
			wantEvents:     []string{"class", "player-update", "amount", "player-stamina"},
		},
		{
			name:          "negative player amount wraps by low byte",
			class:         4,
			amount:        -1,
			playerStamina: 5,
			wantResult:    1,
			wantPlayer:    6,
			wantReported:  6,
			wantEvents: []string{
				"class", "player-update", "amount", "player-stamina",
				"store-player=6", "player", "player-index=player", "report=7",
			},
		},
		{
			name:           "monster success",
			class:          2,
			amount:         45,
			playerStamina:  91,
			monsterStamina: 100,
			wantResult:     1,
			wantPlayer:     91,
			wantMonster:    55,
			wantEvents:     []string{"class", "monster-update", "amount", "monster-stamina", "store-monster=55"},
		},
		{
			name:           "monster insufficient",
			class:          2,
			amount:         256,
			playerStamina:  91,
			monsterStamina: 255,
			wantPlayer:     91,
			wantMonster:    255,
			wantEvents:     []string{"class", "monster-update", "amount", "monster-stamina"},
		},
		{
			name:           "negative monster amount with zero low byte",
			class:          2,
			amount:         -256,
			playerStamina:  91,
			monsterStamina: 5,
			wantResult:     1,
			wantPlayer:     91,
			wantMonster:    5,
			wantEvents:     []string{"class", "monster-update", "amount", "monster-stamina", "store-monster=5"},
		},
		{
			name:           "high-byte class bits are ignored",
			class:          0x00000400,
			amount:         1,
			playerStamina:  91,
			monsterStamina: 77,
			wantResult:     1,
			wantPlayer:     91,
			wantMonster:    77,
			wantEvents:     []string{"class"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := &playerSubStaminaTestWorld4F7D30{
				class:          test.class,
				amount:         test.amount,
				playerStamina:  test.playerStamina,
				monsterStamina: test.monsterStamina,
				player:         "player",
				playerIndex:    map[string]uint8{"player": 7},
			}
			if got := playerSubStamina4F7D30(1, w.hooks()); got != test.wantResult {
				t.Fatalf("result = %d, want %d", got, test.wantResult)
			}
			if w.playerStamina != test.wantPlayer || w.monsterStamina != test.wantMonster {
				t.Fatalf("stamina = player:%d monster:%d, want player:%d monster:%d", w.playerStamina, w.monsterStamina, test.wantPlayer, test.wantMonster)
			}
			if w.reported != test.wantReported {
				t.Fatalf("reported stamina = %d, want %d", w.reported, test.wantReported)
			}
			if !reflect.DeepEqual(w.events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", w.events, test.wantEvents)
			}
		})
	}
}

func TestPlayerSubStamina4F7D30StoresBeforeLivePlayerAndReport(t *testing.T) {
	w := &playerSubStaminaTestWorld4F7D30{
		class:         4,
		amount:        1,
		playerStamina: 9,
		player:        "old-player",
		playerIndex: map[string]uint8{
			"old-player": 3,
			"new-player": 11,
		},
	}
	hooks := w.hooks()
	hooks.storePlayerStamina = func(_ string, stamina uint8) {
		w.events = append(w.events, fmt.Sprintf("store-player=%d", stamina))
		w.playerStamina = stamina
		w.player = "new-player"
	}
	if got := playerSubStamina4F7D30(1, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"class", "player-update", "amount", "player-stamina", "store-player=8",
		"player", "player-index=new-player", "report=11",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.reported != 8 {
		t.Fatalf("reported stamina = %d, want stored value 8", w.reported)
	}
}
