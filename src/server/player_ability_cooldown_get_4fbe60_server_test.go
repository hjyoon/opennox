package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/common"
)

func TestPlayerAbilityCooldownGetNative4FBE60Layout(t *testing.T) {
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

func TestPlayerAbilityCooldownGetNative4FBE60UsesFullNetCodeAndPlayerIndex(t *testing.T) {
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

	wide := int64(0x180000009)
	s.Abils = serverAbilities{s: s}
	s.Abils.cooldowns[7][AbilityTreadLightly] = int32(wide)
	s.Abils.cooldowns[5][AbilityTreadLightly] = 17

	if got, want := s.Abils.PlayerAbilityCooldownGet4FBE60(lookupUnit, AbilityTreadLightly), int32(math.MinInt32+9); got != want {
		t.Fatalf("cooldown = %#08x, want %#08x", uint32(got), uint32(want))
	}
}

func TestPlayerAbilityCooldownGetNative4FBE60MissingStateReturnsZero(t *testing.T) {
	s := new(Server)
	s.Players.list = make([]Player, common.MaxPlayers)
	s.Abils = serverAbilities{s: s}
	unit := &Object{NetCode: 0x12345678}

	if got := s.Abils.PlayerAbilityCooldownGet4FBE60(unit, AbilityBerserk); got != 0 {
		t.Fatalf("missing Player cooldown = %d, want 0", got)
	}
	player := &s.Players.list[2]
	player.Active = 1
	player.NetCodeVal = unit.NetCode
	player.PlayerInd = 2
	player.PlayerUnit = new(Object)
	if got := s.Abils.PlayerAbilityCooldownGet4FBE60(unit, AbilityBerserk); got != 0 {
		t.Fatalf("missing runtime cooldown = %d, want 0", got)
	}
}

func TestPlayerAbilityCooldownGetNative4FBE60NilUnitFaults(t *testing.T) {
	s := new(Server)
	s.Players.list = make([]Player, common.MaxPlayers)
	s.Abils = serverAbilities{s: s}
	defer func() {
		if recover() == nil {
			t.Fatal("nil Object.NetCode load did not fault")
		}
	}()
	s.Abils.PlayerAbilityCooldownGet4FBE60(nil, AbilityBerserk)
}
