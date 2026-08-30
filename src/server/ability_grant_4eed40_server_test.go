package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func defaultAbilityGivePlayerAllNativeDeps4EED40() abilityGivePlayerAllNativeDeps4EED40 {
	return abilityGivePlayerAllNativeDeps4EED40{
		loadAbilityID:  func(int32) uint32 { return 0 },
		gameFlagsCheck: func(uint32) int32 { return 0 },
		isQuest:        func() int32 { return 0 },
		questMode:      func() int32 { return 0 },
		rewardAbility:  func(*Object, int32, int32) {},
	}
}

func TestAbilityGivePlayerAll4EED40NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantSpellLevels := uintptr(3696)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectUpdate = 872
		wantUpdateSize = 656
		wantUpdatePlayer = 336
		wantPlayerSize = 6160
		wantSpellLevels = 4992
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantSpellLevels},
		{"Player.SpellLvl element", unsafe.Sizeof(Player{}.SpellLvl[0]), 4},
		{"Player.SpellLvl count", uintptr(len(Player{}.SpellLvl)), 137},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestAbilityGivePlayerAllNative4EED40BindsRewardPath(t *testing.T) {
	table := [...]uint32{0, 1, 2, 4, 5, 3, 0, 0, 0, 0}
	player := &Player{}
	for i := range player.SpellLvl {
		player.SpellLvl[i] = uint32(i + 100)
	}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	type reward struct {
		unit    *Object
		ability int32
		arg     int32
	}
	var (
		events  []string
		rewards []reward
	)
	deps := defaultAbilityGivePlayerAllNativeDeps4EED40()
	deps.loadAbilityID = func(index int32) uint32 {
		events = append(events, "ability")
		return table[index]
	}
	deps.gameFlagsCheck = func(mask uint32) int32 {
		events = append(events, "game")
		if mask != 0x1000 {
			t.Fatalf("game mask = %#x, want 0x1000", mask)
		}
		return 0
	}
	deps.isQuest = func() int32 {
		events = append(events, "quest")
		return 0
	}
	deps.questMode = func() int32 {
		events = append(events, "quest-mode")
		return 0
	}
	deps.rewardAbility = func(gotUnit *Object, ability, arg int32) {
		events = append(events, "reward")
		rewards = append(rewards, reward{unit: gotUnit, ability: ability, arg: arg})
	}

	abilityGivePlayerAllNative4EED40(unit, int8(len(table)), -0x7654321, deps)
	wantRewards := []reward{
		{unit, 1, -0x7654321},
		{unit, 2, -0x7654321},
		{unit, 4, -0x7654321},
		{unit, 5, -0x7654321},
		{unit, 3, -0x7654321},
	}
	if !reflect.DeepEqual(rewards, wantRewards) {
		t.Fatalf("rewards = %#v, want %#v", rewards, wantRewards)
	}
	if got := player.SpellLvl[:10]; !reflect.DeepEqual(got, []uint32{100, 101, 102, 103, 104, 105, 106, 107, 108, 109}) {
		t.Fatalf("spell levels = %#v", got)
	}
	if got, want := len(events), 10+5*5; got != want {
		t.Fatalf("event count = %d, want %d", got, want)
	}
}

func TestAbilityGivePlayerAllNative4EED40RestrictedClearsTableIndexes(t *testing.T) {
	table := [...]uint32{0, 1, 2, 4, 5, 3, 0, 0, 0, 0}
	player := &Player{}
	for i := range table {
		player.SpellLvl[i] = uint32(i + 1)
	}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	deps := defaultAbilityGivePlayerAllNativeDeps4EED40()
	deps.loadAbilityID = func(index int32) uint32 { return table[index] }
	deps.gameFlagsCheck = func(uint32) int32 { return -1 }
	deps.isQuest = func() int32 {
		t.Fatal("quest check must be short-circuited")
		return 0
	}
	deps.rewardAbility = func(*Object, int32, int32) {
		t.Fatal("restricted mode must not reward abilities")
	}

	abilityGivePlayerAllNative4EED40(unit, int8(len(table)), 99, deps)
	if got, want := player.SpellLvl[:10], []uint32{1, 0, 0, 0, 0, 0, 7, 8, 9, 10}; !reflect.DeepEqual(got, want) {
		t.Fatalf("spell levels = %#v, want %#v", got, want)
	}
}

func TestAbilityGivePlayerAllNative4EED40CachesPlayer(t *testing.T) {
	firstPlayer := &Player{}
	firstPlayer.SpellLvl[0] = 17
	secondPlayer := &Player{}
	secondPlayer.SpellLvl[0] = 29
	update := &PlayerUpdateData{Player: firstPlayer}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	deps := defaultAbilityGivePlayerAllNativeDeps4EED40()
	deps.loadAbilityID = func(int32) uint32 {
		unit.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: secondPlayer})
		return 3
	}
	deps.gameFlagsCheck = func(uint32) int32 { return 1 }

	abilityGivePlayerAllNative4EED40(unit, 1, 0, deps)
	if firstPlayer.SpellLvl[0] != 0 || secondPlayer.SpellLvl[0] != 29 {
		t.Fatalf("cached/replacement levels = %d/%d, want 0/29", firstPlayer.SpellLvl[0], secondPlayer.SpellLvl[0])
	}
}

func TestAbilityGivePlayerAllNative4EED40EntryBehavior(t *testing.T) {
	deps := defaultAbilityGivePlayerAllNativeDeps4EED40()
	deps.loadAbilityID = func(int32) uint32 {
		t.Fatal("nil unit or non-positive count must not load table")
		return 0
	}
	abilityGivePlayerAllNative4EED40(nil, 1, 0, deps)

	update := &PlayerUpdateData{}
	abilityGivePlayerAllNative4EED40(&Object{UpdateData: unsafe.Pointer(update)}, 0, 0, deps)

	defer func() {
		if recover() == nil {
			t.Fatal("nil UpdateData did not fault before non-positive count return")
		}
	}()
	abilityGivePlayerAllNative4EED40(&Object{}, -1, 0, deps)
}
