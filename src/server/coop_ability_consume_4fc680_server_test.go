package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func TestCoopAbilityConsumeNative4FC680StateWidthMasksAndFirstAvailableUnit(t *testing.T) {
	s := new(Server)
	if got := unsafe.Sizeof(s.coopAbilityState4FC670); got != 4 {
		t.Fatalf("cooperative-ability state size = %d, want 4", got)
	}
	first := new(Object)
	second := new(Object)
	s.Players.list = []Player{
		{Active: 1, PlayerInd: 0},
		{Active: 1, PlayerInd: 1, PlayerUnit: first},
		{Active: 1, PlayerInd: 2, PlayerUnit: second},
	}
	s.SetCoopAbilityState4FC670(-1985229329) // 0x89abcdef

	var masks []uint32
	var gotUnit *Object
	var gotState int32
	s.CoopAbilityConsume4FC680(CoopAbilityConsumeRuntime4FC680{
		GameFlag: func(mask uint32) int32 {
			masks = append(masks, mask)
			if mask == coopAbilityModeFlag4FC680 {
				return 1
			}
			return 0
		},
		ExecuteAbility: func(unit *Object, state int32) {
			gotUnit = unit
			gotState = state
			s.SetCoopAbilityState4FC670(math.MaxInt32)
		},
	})

	if want := []uint32{0x800, 0x80000}; !reflect.DeepEqual(masks, want) {
		t.Fatalf("masks = %#v, want %#v", masks, want)
	}
	if gotUnit != first {
		t.Fatalf("unit = %p, want first available %p", gotUnit, first)
	}
	if uint32(gotState) != 0x89abcdef {
		t.Fatalf("state = %#08x, want 0x89abcdef", uint32(gotState))
	}
	if got := s.CoopAbilityState4FC670(); got != 0 {
		t.Fatalf("state after execution = %#08x, want cleared", uint32(got))
	}
}

func TestCoopAbilityConsumeNative4FC680EmptyAndNilUnitListsPreserveState(t *testing.T) {
	for _, tc := range []struct {
		name    string
		players []Player
	}{
		{name: "empty"},
		{name: "active players without units", players: []Player{{Active: 1, PlayerInd: 0}, {Active: 1, PlayerInd: 1}}},
		{name: "inactive unit is ignored", players: []Player{{PlayerUnit: new(Object)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := new(Server)
			s.Players.list = tc.players
			s.SetCoopAbilityState4FC670(math.MinInt32)
			executed := false
			s.CoopAbilityConsume4FC680(CoopAbilityConsumeRuntime4FC680{
				GameFlag: func(mask uint32) int32 {
					if mask == coopAbilityModeFlag4FC680 {
						return 1
					}
					return 0
				},
				ExecuteAbility: func(*Object, int32) { executed = true },
			})
			if executed {
				t.Fatal("ability executed without an active player unit")
			}
			if got := s.CoopAbilityState4FC670(); got != math.MinInt32 {
				t.Fatalf("state = %#08x, want unchanged", uint32(got))
			}
		})
	}
}

func TestCoopAbilityConsumeNative4FC680ReloadsStateAcrossPlayerLookup(t *testing.T) {
	s := new(Server)
	unit := new(Object)
	s.SetCoopAbilityState4FC670(1)
	var gotUnit *Object
	var gotState int32
	coopAbilityConsumeNative4FC680(s, coopAbilityConsumeNativeDeps4FC680{
		gameFlag: func(mask uint32) int32 {
			if mask == coopAbilityModeFlag4FC680 {
				return 1
			}
			return 0
		},
		firstPlayerUnit: func() *Object {
			s.SetCoopAbilityState4FC670(-1985229329)
			return unit
		},
		executeAbility: func(unit *Object, state int32) {
			gotUnit = unit
			gotState = state
		},
	})
	if gotUnit != unit || uint32(gotState) != 0x89abcdef {
		t.Fatalf("unit/state = %p/%#08x, want %p/0x89abcdef", gotUnit, uint32(gotState), unit)
	}
	if got := s.CoopAbilityState4FC670(); got != 0 {
		t.Fatalf("state = %#08x, want cleared", uint32(got))
	}
}

func TestCoopAbilityConsumeNative4FC680ExecuteFaultPreservesState(t *testing.T) {
	s := new(Server)
	unit := new(Object)
	s.Players.list = []Player{{Active: 1, PlayerUnit: unit}}
	s.SetCoopAbilityState4FC670(-1985229329)

	defer func() {
		if got := recover(); got != "execute" {
			t.Fatalf("panic = %v, want execute", got)
		}
		if got := s.CoopAbilityState4FC670(); uint32(got) != 0x89abcdef {
			t.Fatalf("state after execute fault = %#08x, want unchanged", uint32(got))
		}
	}()
	s.CoopAbilityConsume4FC680(CoopAbilityConsumeRuntime4FC680{
		GameFlag: func(mask uint32) int32 {
			if mask == coopAbilityModeFlag4FC680 {
				return 1
			}
			return 0
		},
		ExecuteAbility: func(got *Object, state int32) {
			if got != unit || uint32(state) != 0x89abcdef {
				panic(fmt.Sprintf("unexpected execute args %p/%#08x", got, uint32(state)))
			}
			panic("execute")
		},
	})
}
