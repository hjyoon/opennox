package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerAdjustStaminaTestWorld4F7DB0 struct {
	class       uint32
	amount      uint8
	update      string
	stamina     map[string]uint8
	player      map[string]string
	playerIndex map[string]uint8
	events      []string
	reported    uint8
}

func (w *playerAdjustStaminaTestWorld4F7DB0) hooks() playerAdjustStaminaHooks4F7DB0[int, string, string] {
	return playerAdjustStaminaHooks4F7DB0[int, string, string]{
		loadClass: func(int) uint32 {
			w.events = append(w.events, "class")
			return w.class
		},
		loadUpdate: func(int) string {
			w.events = append(w.events, "update="+w.update)
			return w.update
		},
		loadAmount: func() uint8 {
			w.events = append(w.events, fmt.Sprintf("amount=%d", w.amount))
			return w.amount
		},
		loadStamina: func(update string) uint8 {
			w.events = append(w.events, "stamina="+update)
			return w.stamina[update]
		},
		storeStamina: func(update string, stamina uint8) {
			w.events = append(w.events, fmt.Sprintf("store=%s:%d", update, stamina))
			w.stamina[update] = stamina
		},
		loadPlayer: func(update string) string {
			w.events = append(w.events, "player="+update)
			return w.player[update]
		},
		loadPlayerIndex: func(player string) uint8 {
			w.events = append(w.events, "index="+player)
			return w.playerIndex[player]
		},
		reportStamina: func(index uint8, _ int) {
			w.events = append(w.events, fmt.Sprintf("report=%d", index))
			w.reported = w.stamina["update"]
		},
	}
}

func TestPlayerAdjustStamina4F7DB0LowByteGateAndWrappingArithmetic(t *testing.T) {
	tests := []struct {
		name         string
		class        uint32
		amount       uint8
		stamina      uint8
		wantStamina  uint8
		wantReported uint8
		wantEvents   []string
	}{
		{
			name:         "player subtracts without sufficiency check",
			class:        4,
			amount:       45,
			stamina:      20,
			wantStamina:  231,
			wantReported: 231,
			wantEvents: []string{
				"class", "update=update", "amount=45", "stamina=update",
				"store=update:231", "player=update", "index=player", "report=7",
			},
		},
		{
			name:         "low byte ff adds one",
			class:        0x80000004,
			amount:       0xff,
			stamina:      5,
			wantStamina:  6,
			wantReported: 6,
			wantEvents: []string{
				"class", "update=update", "amount=255", "stamina=update",
				"store=update:6", "player=update", "index=player", "report=7",
			},
		},
		{
			name:         "player wins with monster bit also set",
			class:        6,
			amount:       0,
			stamina:      9,
			wantStamina:  9,
			wantReported: 9,
			wantEvents: []string{
				"class", "update=update", "amount=0", "stamina=update",
				"store=update:9", "player=update", "index=player", "report=7",
			},
		},
		{
			name:        "non-player returns before update and amount",
			class:       0x80000402,
			amount:      1,
			stamina:     9,
			wantStamina: 9,
			wantEvents:  []string{"class"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := &playerAdjustStaminaTestWorld4F7DB0{
				class:       test.class,
				amount:      test.amount,
				update:      "update",
				stamina:     map[string]uint8{"update": test.stamina},
				player:      map[string]string{"update": "player"},
				playerIndex: map[string]uint8{"player": 7},
			}
			playerAdjustStamina4F7DB0(1, w.hooks())
			if got := w.stamina["update"]; got != test.wantStamina {
				t.Fatalf("stamina = %d, want %d", got, test.wantStamina)
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

func TestPlayerAdjustStamina4F7DB0CachesUpdateAndStoresBeforePlayerRead(t *testing.T) {
	w := &playerAdjustStaminaTestWorld4F7DB0{
		class:   4,
		amount:  1,
		update:  "old-update",
		stamina: map[string]uint8{"old-update": 9, "new-update": 100},
		player: map[string]string{
			"old-update": "old-player",
			"new-update": "wrong-player",
		},
		playerIndex: map[string]uint8{
			"new-player":   11,
			"wrong-player": 99,
		},
	}
	hooks := w.hooks()
	hooks.storeStamina = func(update string, stamina uint8) {
		w.events = append(w.events, fmt.Sprintf("store=%s:%d", update, stamina))
		w.stamina[update] = stamina
		w.update = "new-update"
		w.player["old-update"] = "new-player"
	}
	hooks.reportStamina = func(index uint8, _ int) {
		w.events = append(w.events, fmt.Sprintf("report=%d", index))
		w.reported = w.stamina["old-update"]
	}

	playerAdjustStamina4F7DB0(1, hooks)
	want := []string{
		"class", "update=old-update", "amount=1", "stamina=old-update",
		"store=old-update:8", "player=old-update", "index=new-player", "report=11",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if w.stamina["old-update"] != 8 || w.stamina["new-update"] != 100 || w.reported != 8 {
		t.Fatalf("state = old:%d new:%d reported:%d, want 8/100/8", w.stamina["old-update"], w.stamina["new-update"], w.reported)
	}
}
