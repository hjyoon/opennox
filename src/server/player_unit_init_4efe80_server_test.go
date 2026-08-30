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
	wantObjClass := uintptr(8)
	wantUpdateData := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayer := uintptr(276)
	wantExtraLives := uintptr(320)
	wantPlayerSize := uintptr(4828)
	wantGold := uintptr(2164)
	wantProtectedGold := uintptr(4588)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjClass = 12
		wantUpdateData = 872
		wantUpdateSize = 656
		wantPlayer = 336
		wantExtraLives = 416
		wantPlayerSize = 6160
		wantGold = 2168
		wantProtectedGold = 5892
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.ExtraLives", unsafe.Offsetof(PlayerUpdateData{}.ExtraLives), wantExtraLives},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.GoldVal", unsafe.Offsetof(Player{}.GoldVal), wantGold},
		{"Player.ProtPlayerGold", unsafe.Offsetof(Player{}.ProtPlayerGold), wantProtectedGold},
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
	players[0].GoldVal = 0xf1234567
	players[0].ProtPlayerGold = 0xcafebabe
	cached := &PlayerUpdateData{Player: players[0], ExtraLives: 9}
	replacement := &PlayerUpdateData{ExtraLives: 77}
	unit := &Object{ObjClass: 0x04, UpdateData: unsafe.Pointer(cached)}
	events := make([]string, 0, 12)

	got := playerUnitInitNative4EFE80(unit, PlayerUnitInitRuntime4EFE80{
		ProtectGold: func(token uint32, delta int32) {
			events = append(events, fmt.Sprintf("protect:%08x:%08x", token, uint32(delta)))
			if token != 0xcafebabe || uint32(delta) != 0x0edcba99 {
				t.Fatalf("ProtectGold args = %#x/%#x", token, delta)
			}
			if players[0].GoldVal != 0 {
				t.Fatalf("gold at ProtectGold = %#x, want 0", players[0].GoldVal)
			}
			unit.UpdateData = unsafe.Pointer(replacement)
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
		"protect:cafebabe:0edcba99", "sync", "scroll", "spell", "values:0",
		"ability", "flag:1000", "balance:QuestGameStartingExtraLives",
		"default:1:0",
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
			ProtectGold:           func(uint32, int32) { *calls++ },
			SyncLevel:             func(*Object) { *calls++ },
			AwardBeastScrolls:     func(*Player) { *calls++ },
			AwardSpells:           func(*Player) { *calls++ },
			ReadValues:            func(*Object, int32) { *calls++ },
			AwardWarriorAbilities: func(*Player) { *calls++ },
			GameFlag:              func(uint32) int32 { *calls++; return 0 },
			BalanceFloat:          func(string) float32 { *calls++; return 0 },
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
		if calls != 0 {
			t.Fatalf("services before nil-UpdateData fault = %d, want 0", calls)
		}
	})
}

func TestPlayerUnitInitGetGoldNative4EFE80PreservesClassGate(t *testing.T) {
	if got := playerUnitInitGetGoldNative4EFE80(nil); got != 0 {
		t.Fatalf("nil unit gold = %#x, want 0", got)
	}
	nonPlayer := &Object{}
	if got := playerUnitInitGetGoldNative4EFE80(nonPlayer); got != 0 {
		t.Fatalf("non-player gold = %#x, want 0", got)
	}

	player := &Player{GoldVal: 0xf1234567}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 0x104, UpdateData: unsafe.Pointer(update)}
	if got := playerUnitInitGetGoldNative4EFE80(unit); got != player.GoldVal {
		t.Fatalf("player gold = %#x, want %#x", got, player.GoldVal)
	}
}

func TestPlayerUnitInitSubGoldNative4EFE80UsesUnsignedClampAndDelta(t *testing.T) {
	tests := []struct {
		name      string
		gold      uint32
		amount    uint32
		wantGold  uint32
		wantDelta int32
	}{
		{name: "zero", gold: 7, amount: 0, wantGold: 7, wantDelta: 0},
		{name: "subtract", gold: 9, amount: 4, wantGold: 5, wantDelta: -4},
		{name: "equal", gold: 9, amount: 9, wantGold: 0, wantDelta: -9},
		{name: "clamp", gold: 9, amount: 10, wantGold: 0, wantDelta: -10},
		{name: "modulo delta", gold: math.MaxUint32, amount: 0x80000001, wantGold: 0x7ffffffe, wantDelta: 0x7fffffff},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const token = uint32(0xfedcba98)
			player := &Player{GoldVal: test.gold, ProtPlayerGold: token}
			update := &PlayerUpdateData{Player: player}
			unit := &Object{UpdateData: unsafe.Pointer(update)}
			calls := 0
			playerUnitInitSubGoldNative4EFE80(unit, test.amount, func(gotToken uint32, gotDelta int32) {
				calls++
				if gotToken != token || gotDelta != test.wantDelta {
					t.Fatalf("protection args = %#x/%#x, want %#x/%#x", gotToken, gotDelta, token, test.wantDelta)
				}
				if player.GoldVal != test.wantGold {
					t.Fatalf("gold at protection call = %#x, want %#x", player.GoldVal, test.wantGold)
				}
			})
			if calls != 1 || player.GoldVal != test.wantGold {
				t.Fatalf("calls/gold = %d/%#x, want 1/%#x", calls, player.GoldVal, test.wantGold)
			}
		})
	}
}

func TestPlayerUnitInitFloatToInt4EFE80MatchesX87FISTP(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		want  int32
	}{
		{name: "positive half even", value: 2.5, want: 2},
		{name: "positive half odd", value: 3.5, want: 4},
		{name: "negative half even", value: -2.5, want: -2},
		{name: "negative half odd", value: -3.5, want: -4},
		{name: "positive limit", value: math.Float32frombits(0x4effffff), want: 2147483520},
		{name: "negative limit", value: -2147483648, want: math.MinInt32},
		{name: "positive overflow", value: 2147483648, want: math.MinInt32},
		{name: "negative overflow", value: math.Float32frombits(0xcf000001), want: math.MinInt32},
		{name: "positive infinity", value: float32(math.Inf(1)), want: math.MinInt32},
		{name: "negative infinity", value: float32(math.Inf(-1)), want: math.MinInt32},
		{name: "nan", value: math.Float32frombits(0x7fc12345), want: math.MinInt32},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := playerUnitInitFloatToInt4EFE80(test.value); got != test.want {
				t.Fatalf("conversion(%08x) = %#x, want %#x", math.Float32bits(test.value), got, test.want)
			}
		})
	}
}
