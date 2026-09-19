package server

import (
	"fmt"
	"reflect"
	"testing"
)

func matchLimitBaseHooks5096F0(events *[]string) matchLimitStateHooks5096F0[int, int, int] {
	return matchLimitStateHooks5096F0[int, int, int]{
		limitExpired: func() int32 { return 1 },
		hasGameFlags: func(mask uint32) bool {
			*events = append(*events, fmt.Sprintf("flags:%x", mask))
			return false
		},
		switchToNextMap: func() { *events = append(*events, "switch") },
		printAutoExit:   func() { *events = append(*events, "print") },
		firstPlayer:     func() int { return 0 },
		nextPlayer:      func(int) int { return 0 },
		loadUpdate:      func(unit int) int { return unit + 10 },
		loadQuestExit:   func(int) int { return 0 },
		loadPlayer:      func(update int) int { return update + 100 },
		loadQuestState:  func(int) uint32 { return 0 },
		recordQuestProgress: func(unit int) {
			*events = append(*events, fmt.Sprintf("progress:%d", unit))
		},
		matchStateActive: func() int32 { return 0 },
		resolveTeamMode: func() int32 {
			*events = append(*events, "resolve-team")
			return 11
		},
		resolveHighScore: func() int32 {
			*events = append(*events, "resolve-high")
			return 12
		},
		resolveLowScore: func() int32 {
			*events = append(*events, "resolve-low")
			return 13
		},
		setGameFlags: func(flags uint32) {
			*events = append(*events, fmt.Sprintf("set:%x", flags))
		},
		loadPositionX:   func(int) float32 { return 0 },
		loadPositionY:   func(int) float32 { return 0 },
		loadPlayerIndex: func(int) uint8 { return 0 },
		sendPosition: func(index uint8, packet [5]byte) {
			*events = append(*events, fmt.Sprintf("send:%d:%x", index, packet))
		},
		loadNetCode: func(unit int) uint32 { return uint32(unit) },
		audioEvent: func(id uint32, unit int, kind int32, code uint32) {
			*events = append(*events, fmt.Sprintf("audio:%d:%d:%d:%d", id, unit, kind, code))
		},
		clearTimer: func() int32 {
			*events = append(*events, "clear")
			return -37
		},
	}
}

func TestMatchLimitState5096F0InactiveReturnsExactProbeResult(t *testing.T) {
	var events []string
	hooks := matchLimitBaseHooks5096F0(&events)
	hooks.limitExpired = func() int32 { return 0 }
	if got := matchLimitState5096F0(hooks); got != 0 || len(events) != 0 {
		t.Fatalf("result/events = %d/%q, want 0/empty", got, events)
	}
}

func TestMatchLimitState5096F0QuestOrderAndProgress(t *testing.T) {
	var events []string
	hooks := matchLimitBaseHooks5096F0(&events)
	hooks.hasGameFlags = func(mask uint32) bool {
		events = append(events, fmt.Sprintf("flags:%x", mask))
		return mask == matchLimitQuestMask5096F0
	}
	hooks.firstPlayer = func() int {
		events = append(events, "first")
		return 1
	}
	hooks.nextPlayer = func(unit int) int {
		events = append(events, fmt.Sprintf("next:%d", unit))
		if unit == 1 {
			return 2
		}
		return 0
	}
	hooks.loadUpdate = func(unit int) int {
		events = append(events, fmt.Sprintf("update:%d", unit))
		return unit + 10
	}
	hooks.loadQuestExit = func(update int) int {
		events = append(events, fmt.Sprintf("exit:%d", update))
		if update == 11 {
			return 99
		}
		return 0
	}
	hooks.loadPlayer = func(update int) int {
		events = append(events, fmt.Sprintf("player:%d", update))
		return update + 100
	}
	hooks.loadQuestState = func(player int) uint32 {
		events = append(events, fmt.Sprintf("quest:%d", player))
		return 1
	}

	if got := matchLimitState5096F0(hooks); got != -37 {
		t.Fatalf("result = %d, want -37", got)
	}
	want := []string{
		"flags:1000", "switch", "print", "first",
		"update:1", "exit:11", "next:1",
		"update:2", "exit:12", "player:12", "quest:112", "progress:2", "next:2", "clear",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", events, want)
	}
}

func TestMatchLimitState5096F0WinnerModePriority(t *testing.T) {
	tests := []struct {
		name      string
		active    map[uint32]bool
		wantEvent string
		wantMasks []uint32
	}{
		{name: "team before high and low", active: map[uint32]bool{
			matchLimitTeamModeMask5096F0: true, matchLimitHighScoreMask5096F0: true, matchLimitLowScoreMask5096F0: true,
		}, wantEvent: "resolve-team", wantMasks: []uint32{matchLimitQuestMask5096F0, matchLimitTeamModeMask5096F0}},
		{name: "high before low", active: map[uint32]bool{
			matchLimitHighScoreMask5096F0: true, matchLimitLowScoreMask5096F0: true,
		}, wantEvent: "resolve-high", wantMasks: []uint32{matchLimitQuestMask5096F0, matchLimitTeamModeMask5096F0, matchLimitHighScoreMask5096F0}},
		{name: "low", active: map[uint32]bool{matchLimitLowScoreMask5096F0: true}, wantEvent: "resolve-low", wantMasks: []uint32{
			matchLimitQuestMask5096F0, matchLimitTeamModeMask5096F0, matchLimitHighScoreMask5096F0, matchLimitLowScoreMask5096F0,
		}},
		{name: "none", active: map[uint32]bool{}, wantMasks: []uint32{
			matchLimitQuestMask5096F0, matchLimitTeamModeMask5096F0, matchLimitHighScoreMask5096F0, matchLimitLowScoreMask5096F0,
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			var masks []uint32
			hooks := matchLimitBaseHooks5096F0(&events)
			hooks.hasGameFlags = func(mask uint32) bool {
				masks = append(masks, mask)
				return tc.active[mask]
			}
			if got := matchLimitState5096F0(hooks); got != -37 {
				t.Fatalf("result = %d, want -37", got)
			}
			if !reflect.DeepEqual(masks, tc.wantMasks) {
				t.Fatalf("masks = %#x, want %#x", masks, tc.wantMasks)
			}
			if tc.wantEvent == "" {
				if !reflect.DeepEqual(events, []string{"clear"}) {
					t.Fatalf("events = %q", events)
				}
			} else if !reflect.DeepEqual(events, []string{tc.wantEvent, "clear"}) {
				t.Fatalf("events = %q, want resolver then clear", events)
			}
		})
	}
}

func TestMatchLimitState5096F0SuddenDeathReloadsPlayerAndBroadcastsAudio(t *testing.T) {
	var events []string
	hooks := matchLimitBaseHooks5096F0(&events)
	hooks.matchStateActive = func() int32 { return 1 }
	hooks.firstPlayer = func() int { return 1 }
	hooks.nextPlayer = func(unit int) int {
		events = append(events, fmt.Sprintf("next:%d", unit))
		if unit == 1 {
			return 2
		}
		return 0
	}
	hooks.loadUpdate = func(unit int) int {
		events = append(events, fmt.Sprintf("update:%d", unit))
		if unit == 1 {
			return 11
		}
		return 0
	}
	playerLoads := 0
	hooks.loadPlayer = func(update int) int {
		playerLoads++
		events = append(events, fmt.Sprintf("player:%d:%d", update, playerLoads))
		return 100 + playerLoads
	}
	hooks.loadPositionX = func(player int) float32 {
		events = append(events, fmt.Sprintf("x:%d", player))
		return -2.9
	}
	hooks.loadPositionY = func(player int) float32 {
		events = append(events, fmt.Sprintf("y:%d", player))
		return 65537.9
	}
	hooks.loadPlayerIndex = func(player int) uint8 {
		events = append(events, fmt.Sprintf("index:%d", player))
		return 7
	}
	hooks.loadNetCode = func(unit int) uint32 {
		events = append(events, fmt.Sprintf("net:%d", unit))
		return uint32(0x1000 + unit)
	}

	if got := matchLimitState5096F0(hooks); got != -37 {
		t.Fatalf("result = %d, want -37", got)
	}
	want := []string{
		"flags:1000", "set:4000000", "update:1",
		"player:11:1", "x:101", "player:11:2", "y:102", "player:11:3", "index:103",
		"send:7:9afeff0100", "net:1", "audio:582:1:2:4097", "next:1",
		"update:2", "net:2", "audio:582:2:2:4098", "next:2", "clear",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", events, want)
	}
}

func TestMatchLimitPositionPacket5096F0(t *testing.T) {
	if got, want := matchLimitPositionPacket5096F0(32768.9, -32769.9), [5]byte{0x9a, 0x00, 0x80, 0xff, 0x7f}; got != want {
		t.Fatalf("packet = %x, want %x", got, want)
	}
}
