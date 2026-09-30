package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestQuestCanJoin4E4100ThresholdAndNonzeroStates(t *testing.T) {
	for _, test := range []struct {
		name   string
		states []uint32
		want   uint32
	}{
		{"empty", nil, 1},
		{"all inactive", []uint32{0, 0, 0, 0, 0, 0, 0}, 1},
		{"five", []uint32{1, 1, 1, 1, 1}, 1},
		{"six", []uint32{1, 1, 1, 1, 1, 1}, 0},
		{"seven", []uint32{1, 1, 1, 1, 1, 1, 1}, 0},
		{"all nonzero count", []uint32{1, 2, math.MaxUint32, 3, 0x80000000, 7}, 0},
		{"zero excluded", []uint32{1, 2, math.MaxUint32, 0, 0x80000000, 7}, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			players := make([]Player, len(test.states))
			updates := make([]PlayerUpdateData, len(test.states))
			units := make([]Object, len(test.states))
			for i, state := range test.states {
				players[i].Field4792 = state
				updates[i].Player = &players[i]
				units[i].UpdateData = unsafe.Pointer(&updates[i])
				if unsafe.Sizeof(uintptr(0)) > 4 {
					for _, pointer := range []unsafe.Pointer{
						unsafe.Pointer(&players[i]), unsafe.Pointer(&updates[i]), unsafe.Pointer(&units[i]),
					} {
						if uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("pointer = %p, want above 4 GiB", pointer)
						}
					}
				}
			}
			cursor, nextCalls, flagCalls := 0, 0, 0
			got := questCanJoinNative4E4100(questCanJoinNativeDeps4E4100{
				firstUnit: func() *Object {
					if len(units) == 0 {
						return nil
					}
					return &units[0]
				},
				nextUnit: func(unit *Object) *Object {
					if unit != &units[cursor] {
						t.Fatal("native unit identity changed")
					}
					nextCalls++
					cursor++
					if cursor == len(units) {
						return nil
					}
					return &units[cursor]
				},
				gameHost: func() bool { flagCalls++; return false },
				noRendering: func() bool {
					t.Fatal("non-host read rendering flag")
					return false
				},
			})
			if got != test.want || nextCalls != len(units) || flagCalls != len(units) {
				t.Fatalf("result/next/flags = %d/%d/%d, want %d/%d/%d", got, nextCalls, flagCalls,
					test.want, len(units), len(units))
			}
			for i, state := range test.states {
				if players[i].Field4792 != state || updates[i].Player != &players[i] ||
					units[i].UpdateData != unsafe.Pointer(&updates[i]) {
					t.Fatal("read-only admission check mutated native fields")
				}
			}
			runtime.KeepAlive(players)
			runtime.KeepAlive(updates)
			runtime.KeepAlive(units)
		})
	}
}

func TestQuestCanJoin4E4100LiveFlagsAndCachedUpdate(t *testing.T) {
	old := &PlayerUpdateData{Player: &Player{Field4792: 1}}
	replacement := &PlayerUpdateData{Player: &Player{Field4792: 0}}
	unit := &Object{UpdateData: unsafe.Pointer(old)}
	var flags, rendering, next int
	got := questCanJoinNative4E4100(questCanJoinNativeDeps4E4100{
		firstUnit: func() *Object { return unit },
		nextUnit: func(got *Object) *Object {
			if got != unit {
				t.Fatal("unit pointer changed")
			}
			next++
			if next == 7 {
				return nil
			}
			unit.UpdateData = unsafe.Pointer(old)
			return unit
		},
		gameHost: func() bool {
			flags++
			unit.UpdateData = unsafe.Pointer(replacement)
			return flags%2 == 0
		},
		noRendering: func() bool { rendering++; return false },
	})
	if got != 0 || flags != 7 || rendering != 3 || next != 7 {
		t.Fatalf("result/flags/rendering/next = %d/%d/%d/%d, want 0/7/3/7", got, flags, rendering, next)
	}
	runtime.KeepAlive(old)
	runtime.KeepAlive(replacement)
}

func TestQuestCanJoin4E4100DedicatedHostAndNativeLayout(t *testing.T) {
	oldGame, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	})
	s := new(Server)
	s.Players.list = make([]Player, 32)
	units := make([]Object, 6)
	updates := make([]PlayerUpdateData, 6)
	for i, index := range []int{0, 1, 2, 3, 4, HostPlayerIndex} {
		player := &s.Players.list[index]
		*player = Player{Active: 1, PlayerInd: uint8(index), Field4792: 2}
		updates[i].Player = player
		units[i] = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(&updates[i])}
		player.PlayerUnit = &units[i]
	}
	for _, test := range []struct {
		host, noRendering bool
		want              uint32
	}{
		{false, false, 0}, {false, true, 0}, {true, false, 0}, {true, true, 1},
	} {
		noxflags.ResetGame()
		noxflags.ResetEngine()
		if test.host {
			noxflags.SetGame(noxflags.GameHost)
		}
		if test.noRendering {
			noxflags.SetEngine(noxflags.EngineNoRendering)
		}
		if got := s.QuestCanJoin4E4100(); got != test.want {
			t.Fatalf("host=%t noRendering=%t result=%d, want %d", test.host, test.noRendering, got, test.want)
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if unsafe.Offsetof(Object{}.UpdateData) != 872 || unsafe.Offsetof(PlayerUpdateData{}.Player) != 336 ||
			unsafe.Offsetof(Player{}.PlayerInd) != 2068 || unsafe.Offsetof(Player{}.Field4792) != 6096 {
			t.Fatal("native Quest admission field layout changed")
		}
	} else if unsafe.Offsetof(Object{}.UpdateData) != 748 || unsafe.Offsetof(PlayerUpdateData{}.Player) != 276 ||
		unsafe.Offsetof(Player{}.PlayerInd) != 2064 || unsafe.Offsetof(Player{}.Field4792) != 4792 {
		t.Fatal("PE32 Quest admission field layout changed")
	}
	runtime.KeepAlive(updates)
	runtime.KeepAlive(units)
}
