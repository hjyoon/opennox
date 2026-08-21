package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestPlayerUnitInit4EFE80NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantUpdateData := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantExtraLives := uintptr(320)
	wantPlayerSize := uintptr(4828)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantUpdateData = 872
		wantUpdateSize = 640
		wantPlayer = 320
		wantExtraLives = 400
		wantPlayerSize = 6160
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.ExtraLives", unsafe.Offsetof(PlayerUpdateData{}.ExtraLives), wantExtraLives},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"ExtraLives width", unsafe.Sizeof(PlayerUpdateData{}.ExtraLives), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerUnitInitNative4EFE80BindsCachedUpdateAndLivePlayers(t *testing.T) {
	players := []*Player{{}, {}, {}, {}}
	cached := &PlayerUpdateData{Player: players[0], ExtraLives: 9}
	replacement := &PlayerUpdateData{ExtraLives: 77}
	unit := &Object{UpdateData: unsafe.Pointer(cached)}
	events := make([]string, 0, 12)

	got := playerUnitInitNative4EFE80(unit, PlayerUnitInitRuntime4EFE80{
		GetGold: func(gotUnit *Object) uint32 {
			events = append(events, "gold")
			if gotUnit != unit {
				t.Fatalf("GetGold unit = %p, want %p", gotUnit, unit)
			}
			unit.UpdateData = unsafe.Pointer(replacement)
			return 0xf1234567
		},
		SubGold: func(gotUnit *Object, value uint32) {
			events = append(events, fmt.Sprintf("sub:%08x", value))
			if gotUnit != unit || value != 0xf1234567 {
				t.Fatalf("SubGold args = %p/%#x", gotUnit, value)
			}
		},
		SyncLevel: func(gotUnit *Object) {
			events = append(events, "sync")
			if gotUnit != unit {
				t.Fatalf("SyncLevel unit = %p, want %p", gotUnit, unit)
			}
		},
		AwardBeastScrolls: func(player *Player) {
			events = append(events, "scroll")
			if player != players[0] {
				t.Fatalf("scroll Player = %p, want %p", player, players[0])
			}
			cached.Player = players[1]
		},
		AwardSpells: func(player *Player) {
			events = append(events, "spell")
			if player != players[1] {
				t.Fatalf("spell Player = %p, want %p", player, players[1])
			}
			cached.Player = players[2]
		},
		ReadValues: func(gotUnit *Object, reward int32) {
			events = append(events, fmt.Sprintf("values:%d", reward))
			if gotUnit != unit || reward != 0 {
				t.Fatalf("ReadValues args = %p/%d", gotUnit, reward)
			}
			cached.Player = players[3]
		},
		AwardWarriorAbilities: func(player *Player) {
			events = append(events, "ability")
			if player != players[3] {
				t.Fatalf("ability Player = %p, want %p", player, players[3])
			}
		},
		GameFlag: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("flag:%x", flag))
			return 1
		},
		BalanceFloat: func(key string) float32 {
			events = append(events, "balance:"+key)
			return math.Float32frombits(0xbfc00000)
		},
		FloatToInt: func(value float32) int32 {
			events = append(events, fmt.Sprintf("convert:%08x", math.Float32bits(value)))
			return -2
		},
		MakeDefaultItems: func(gotUnit *Object, restoreStats, keepItems int32) uint8 {
			events = append(events, fmt.Sprintf("default:%d:%d", restoreStats, keepItems))
			if gotUnit != unit {
				t.Fatalf("MakeDefaultItems unit = %p, want %p", gotUnit, unit)
			}
			return 0xfe
		},
	})

	if got != 0xfe {
		t.Fatalf("result = %#x, want 0xfe", got)
	}
	wantEvents := []string{
		"gold", "sub:f1234567", "sync", "scroll", "spell", "values:0",
		"ability", "flag:1000", "balance:QuestGameStartingExtraLives",
		"convert:bfc00000", "default:1:0",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if cached.ExtraLives != 0xfffffffe || replacement.ExtraLives != 77 {
		t.Fatalf("cached/replacement ExtraLives = %#x/%#x", cached.ExtraLives, replacement.ExtraLives)
	}
}

func TestPlayerUnitInitNative4EFE80NilFaultBoundaries(t *testing.T) {
	newRuntime := func(calls *int) PlayerUnitInitRuntime4EFE80 {
		return PlayerUnitInitRuntime4EFE80{
			GetGold:               func(*Object) uint32 { *calls++; return 0 },
			SubGold:               func(*Object, uint32) { *calls++ },
			SyncLevel:             func(*Object) { *calls++ },
			AwardBeastScrolls:     func(*Player) { *calls++ },
			AwardSpells:           func(*Player) { *calls++ },
			ReadValues:            func(*Object, int32) { *calls++ },
			AwardWarriorAbilities: func(*Player) { *calls++ },
			GameFlag:              func(uint32) int32 { *calls++; return 0 },
			BalanceFloat:          func(string) float32 { *calls++; return 0 },
			FloatToInt:            func(float32) int32 { *calls++; return 0 },
			MakeDefaultItems:      func(*Object, int32, int32) uint8 { *calls++; return 0 },
		}
	}

	t.Run("unit", func(t *testing.T) {
		calls := 0
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("nil unit did not fault")
				}
			}()
			playerUnitInitNative4EFE80(nil, newRuntime(&calls))
		}()
		if calls != 0 {
			t.Fatalf("services before nil-unit fault = %d, want 0", calls)
		}
	})

	t.Run("update", func(t *testing.T) {
		calls := 0
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("nil UpdateData did not fault")
				}
			}()
			playerUnitInitNative4EFE80(&Object{}, newRuntime(&calls))
		}()
		if calls != 3 {
			t.Fatalf("services before nil-UpdateData fault = %d, want 3", calls)
		}
	})
}
