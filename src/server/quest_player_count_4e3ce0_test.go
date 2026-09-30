package server

import (
	"math"
	"reflect"
	"testing"
)

func TestQuestPlayerCount4E3CE0ExactStateAndFullTraversal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		states []uint32
		want   int32
	}{
		{"empty", nil, 0},
		{"inactive", []uint32{0, 0, 0}, 0},
		{"only exact one", []uint32{0, 1, 2, math.MaxUint32, 0x80000000, 1}, 2},
		{"no admission limit", []uint32{1, 1, 1, 1, 1, 1, 1, 1}, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first, next, flags, playerReads := 0, 0, 0, 0
			got := questPlayerCount4E3CE0(questPlayerCountHooks4E3CE0[int, int, int]{
				firstUnit: func() int {
					first++
					if len(tc.states) == 0 {
						return 0
					}
					return 1
				},
				nextUnit: func(unit int) int {
					next++
					if unit != next {
						t.Fatalf("unit identity = %d, want %d", unit, next)
					}
					if unit == len(tc.states) {
						return 0
					}
					return unit + 1
				},
				loadUpdate:  func(unit int) int { return unit },
				gameHost:    func() bool { flags++; return false },
				noRendering: func() bool { t.Fatal("non-host read rendering flag"); return false },
				loadPlayer:  func(update int) int { playerReads++; return update },
				loadPlayerIndex: func(int) uint8 {
					t.Fatal("non-host read player index")
					return 0
				},
				loadQuestState: func(player int) uint32 { return tc.states[player-1] },
			})
			if got != tc.want || first != 1 || next != len(tc.states) || flags != len(tc.states) || playerReads != len(tc.states) {
				t.Fatalf("count/first/next/flags/player reads = %d/%d/%d/%d/%d", got, first, next, flags, playerReads)
			}
		})
	}
}

func TestQuestPlayerCount4E3CE0DedicatedHostShortCircuit(t *testing.T) {
	for _, host := range []bool{false, true} {
		for _, noRendering := range []bool{false, true} {
			indices := []uint8{0, 30, 31, 32, 255}
			next, renderingReads, indexReads, stateReads := 0, 0, 0, 0
			got := questPlayerCount4E3CE0(questPlayerCountHooks4E3CE0[int, int, int]{
				firstUnit: func() int { return 1 },
				nextUnit: func(unit int) int {
					next++
					if unit == len(indices) {
						return 0
					}
					return unit + 1
				},
				loadUpdate:  func(unit int) int { return unit },
				gameHost:    func() bool { return host },
				noRendering: func() bool { renderingReads++; return noRendering },
				loadPlayer:  func(update int) int { return update },
				loadPlayerIndex: func(player int) uint8 {
					indexReads++
					return indices[player-1]
				},
				loadQuestState: func(int) uint32 { stateReads++; return 1 },
			})
			want, wantRendering, wantIndex := int32(5), 0, 0
			if host {
				wantRendering = 5
				if noRendering {
					want, wantIndex = 4, 5
				}
			}
			if got != want || next != 5 || renderingReads != wantRendering || indexReads != wantIndex || stateReads != int(want) {
				t.Fatalf("host=%t rendering-disabled=%t count/next/rendering/index/state = %d/%d/%d/%d/%d",
					host, noRendering, got, next, renderingReads, indexReads, stateReads)
			}
		}
	}
}

func TestQuestPlayerCount4E3CE0CachedUpdateAndReloadedPlayer(t *testing.T) {
	type player struct {
		index uint8
		state uint32
	}
	type update struct{ player *player }
	type unit struct{ update *update }
	original := &player{index: 4, state: 0}
	final := &player{index: 31, state: 1}
	cached := &update{player: original}
	replacement := &update{player: &player{index: 9, state: 0}}
	obj := &unit{update: cached}
	playerReads := 0
	got := questPlayerCount4E3CE0(questPlayerCountHooks4E3CE0[*unit, *update, *player]{
		firstUnit:  func() *unit { return obj },
		nextUnit:   func(*unit) *unit { return nil },
		loadUpdate: func(obj *unit) *update { return obj.update },
		gameHost: func() bool {
			obj.update = replacement
			return true
		},
		noRendering: func() bool { return true },
		loadPlayer: func(got *update) *player {
			if got != cached {
				t.Fatal("update was not cached before game flag callback")
			}
			playerReads++
			return got.player
		},
		loadPlayerIndex: func(got *player) uint8 {
			if got != original {
				t.Fatal("unexpected player at index read")
			}
			cached.player = final
			return got.index
		},
		loadQuestState: func(got *player) uint32 {
			if got != final {
				t.Fatal("player link was not reloaded after index read")
			}
			return got.state
		},
	})
	if got != 1 || playerReads != 2 {
		t.Fatalf("count/player loads = %d/%d, want 1/2", got, playerReads)
	}
}

func TestQuestPlayerCount4E3CE0ObservableFaultPrefixes(t *testing.T) {
	want := []string{"first", "update", "host", "rendering", "player", "index", "player", "state", "next"}
	for failAt := -1; failAt < len(want); failAt++ {
		var trace []string
		mark := func(event string) {
			trace = append(trace, event)
			if len(trace)-1 == failAt {
				panic("injected count fault")
			}
		}
		panicked := false
		var got int32
		func() {
			defer func() {
				if fault := recover(); fault != nil {
					if fault != "injected count fault" {
						t.Fatalf("unexpected fault: %v", fault)
					}
					panicked = true
				}
			}()
			got = questPlayerCount4E3CE0(questPlayerCountHooks4E3CE0[int, int, int]{
				firstUnit:       func() int { mark("first"); return 1 },
				nextUnit:        func(int) int { mark("next"); return 0 },
				loadUpdate:      func(int) int { mark("update"); return 1 },
				gameHost:        func() bool { mark("host"); return true },
				noRendering:     func() bool { mark("rendering"); return true },
				loadPlayer:      func(int) int { mark("player"); return 1 },
				loadPlayerIndex: func(int) uint8 { mark("index"); return 4 },
				loadQuestState:  func(int) uint32 { mark("state"); return 1 },
			})
		}()
		prefix := want
		if failAt >= 0 {
			prefix = want[:failAt+1]
		} else if got != 1 {
			t.Fatalf("successful count = %d, want 1", got)
		}
		if panicked != (failAt >= 0) || !reflect.DeepEqual(trace, prefix) {
			t.Fatalf("fault=%d panic=%t trace=%v, want %v", failAt, panicked, trace, prefix)
		}
	}
}
