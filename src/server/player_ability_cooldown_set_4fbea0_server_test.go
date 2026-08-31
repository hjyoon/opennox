package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/common"
)

func TestPlayerAbilityCooldownSetNative4FBEA0Layout(t *testing.T) {
	ptrBytes := unsafe.Sizeof(uintptr(0))
	wantObjectNetCode := uintptr(36)
	wantPlayerIndex := uintptr(2064)
	if ptrBytes == 8 {
		wantObjectNetCode = 40
		wantPlayerIndex = 2068
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantObjectNetCode},
		{"Object.NetCode width", unsafe.Sizeof(Object{}.NetCode), 4},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.PlayerInd width", unsafe.Sizeof(Player{}.PlayerInd), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d for %d-byte pointers", check.name, check.got, check.want, ptrBytes)
		}
	}
}

func TestPlayerAbilityCooldownSetNative4FBEA0UsesFullNetCodeAndPlayerIndex(t *testing.T) {
	s := new(Server)
	s.Players.list = make([]Player, common.MaxPlayers)
	lookupUnit := &Object{NetCode: 0xfedcba98}
	indexedUnit := new(Object)
	wrongUnit := new(Object)
	lookupPlayer := &s.Players.list[5]
	lookupPlayer.Active = 1
	lookupPlayer.NetCodeVal = lookupUnit.NetCode
	lookupPlayer.PlayerInd = 7
	lookupPlayer.PlayerUnit = wrongUnit
	s.Players.list[7].PlayerUnit = indexedUnit

	s.Abils = serverAbilities{s: s}
	s.Abils.Init4FB990()
	want := int32(math.MinInt32 + 0x1234)
	if got := s.Abils.PlayerAbilityCooldownSet4FBEA0(lookupUnit, AbilityTreadLightly, want); got != want {
		t.Fatalf("return = %#08x, want %#08x", uint32(got), uint32(want))
	}
	if runtime := s.Abils.ByUnit[indexedUnit]; runtime == nil || int32(runtime.Cooldowns[AbilityTreadLightly]) != want {
		t.Fatalf("indexed runtime = %#v, want cooldown %#08x", runtime, uint32(want))
	}
	if _, ok := s.Abils.ByUnit[wrongUnit]; ok {
		t.Fatal("stored through lookup PlayerUnit instead of observed PlayerInd")
	}
	if _, ok := s.Abils.ByUnit[lookupUnit]; ok {
		t.Fatal("stored through lookup unit instead of observed PlayerInd")
	}
	if got := s.Abils.PlayerAbilityCooldownGet4FBE60(lookupUnit, AbilityTreadLightly); got != want {
		t.Fatalf("round-trip cooldown = %#08x, want %#08x", uint32(got), uint32(want))
	}
}

func TestPlayerAbilityCooldownSetNative4FBEA0MissingPlayerReturnsZeroWithoutState(t *testing.T) {
	s := new(Server)
	s.Players.list = make([]Player, common.MaxPlayers)
	s.Abils = serverAbilities{s: s}
	s.Abils.Init4FB990()
	unit := &Object{NetCode: 0x12345678}

	if got := s.Abils.PlayerAbilityCooldownSet4FBEA0(unit, AbilityBerserk, 99); got != 0 {
		t.Fatalf("return = %d, want 0", got)
	}
	if len(s.Abils.ByUnit) != 0 {
		t.Fatalf("runtime count = %d, want 0", len(s.Abils.ByUnit))
	}
}

func TestPlayerAbilityCooldownSetNative4FBEA0NilUnitFaults(t *testing.T) {
	s := new(Server)
	s.Players.list = make([]Player, common.MaxPlayers)
	s.Abils = serverAbilities{s: s}
	s.Abils.Init4FB990()
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object.NetCode load did not fault")
		}
	}()
	s.Abils.PlayerAbilityCooldownSet4FBEA0(nil, AbilityBerserk, 1)
}

func TestPlayerAbilityCooldownSetForUnit4FBEA0NarrowsHostInt(t *testing.T) {
	if unsafe.Sizeof(int(0)) != 8 {
		t.Skip("host int has no high bits to narrow")
	}
	s := new(Server)
	s.Players.list = make([]Player, common.MaxPlayers)
	unit := &Object{NetCode: 7}
	player := &s.Players.list[3]
	player.Active = 1
	player.NetCodeVal = unit.NetCode
	player.PlayerInd = 3
	player.PlayerUnit = unit
	s.Abils = serverAbilities{s: s}
	s.Abils.Init4FB990()

	wide64 := int64(0x180000009)
	wide := int(wide64)
	s.Abils.SetCooldownForUnit(unit, AbilityHarpoon, wide)
	if got, want := s.Abils.PlayerAbilityCooldownGet4FBE60(unit, AbilityHarpoon), int32(math.MinInt32+9); got != want {
		t.Fatalf("cooldown = %#08x, want narrowed %#08x", uint32(got), uint32(want))
	}
}
