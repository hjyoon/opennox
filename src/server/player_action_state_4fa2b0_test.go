package server

import (
	"reflect"
	"testing"
)

type playerActionStateTestObject4FA2B0 struct {
	update *playerActionStateTestUpdate4FA2B0
}

type playerActionStateTestUpdate4FA2B0 struct {
	state  uint8
	player *playerActionStateTestPlayer4FA2B0
	weapon *playerActionStateTestWeapon4FA2B0
}

type playerActionStateTestPlayer4FA2B0 struct {
	equip   uint32
	variant uint8
}

type playerActionStateTestWeapon4FA2B0 struct {
	use *playerActionStateTestUse4FA2B0
}

type playerActionStateTestUse4FA2B0 struct {
	flags uint8
}

type playerActionStateTestRuntime4FA2B0 struct {
	calls         []string
	warcryActive  bool
	warcryCurrent bool
	berserkActive bool
	weaponState   int32
}

func (r *playerActionStateTestRuntime4FA2B0) hooks() playerActionStateHooks4FA2B0[
	*playerActionStateTestObject4FA2B0,
	*playerActionStateTestUpdate4FA2B0,
	*playerActionStateTestPlayer4FA2B0,
	*playerActionStateTestWeapon4FA2B0,
	*playerActionStateTestUse4FA2B0,
] {
	return playerActionStateHooks4FA2B0[
		*playerActionStateTestObject4FA2B0,
		*playerActionStateTestUpdate4FA2B0,
		*playerActionStateTestPlayer4FA2B0,
		*playerActionStateTestWeapon4FA2B0,
		*playerActionStateTestUse4FA2B0,
	]{
		loadUpdateData: func(obj *playerActionStateTestObject4FA2B0) *playerActionStateTestUpdate4FA2B0 {
			r.calls = append(r.calls, "update")
			return obj.update
		},
		loadState: func(update *playerActionStateTestUpdate4FA2B0) uint8 {
			r.calls = append(r.calls, "state")
			return update.state
		},
		isAbilityActive: func(_ *playerActionStateTestObject4FA2B0, ability Ability) bool {
			r.calls = append(r.calls, "active:"+ability.String())
			switch ability {
			case AbilityWarcry:
				return r.warcryActive
			case AbilityBerserk:
				return r.berserkActive
			default:
				panic("unexpected ability")
			}
		},
		isWarcryActive: func(_ *playerActionStateTestObject4FA2B0, ability Ability) bool {
			r.calls = append(r.calls, "warcry:"+ability.String())
			return r.warcryCurrent
		},
		loadPlayer: func(update *playerActionStateTestUpdate4FA2B0) *playerActionStateTestPlayer4FA2B0 {
			r.calls = append(r.calls, "player")
			return update.player
		},
		loadWeaponEquip: func(player *playerActionStateTestPlayer4FA2B0) uint32 {
			r.calls = append(r.calls, "equip")
			return player.equip
		},
		loadEquippedWeapon: func(update *playerActionStateTestUpdate4FA2B0) *playerActionStateTestWeapon4FA2B0 {
			r.calls = append(r.calls, "weapon")
			return update.weapon
		},
		loadWeaponUseData: func(weapon *playerActionStateTestWeapon4FA2B0) *playerActionStateTestUse4FA2B0 {
			r.calls = append(r.calls, "use-data")
			return weapon.use
		},
		loadWeaponFlags: func(use *playerActionStateTestUse4FA2B0) uint8 {
			r.calls = append(r.calls, "weapon-flags")
			return use.flags
		},
		loadAnimationVariant: func(player *playerActionStateTestPlayer4FA2B0) uint8 {
			r.calls = append(r.calls, "variant")
			return player.variant
		},
		weaponAnimation: func(equip uint32) int32 {
			r.calls = append(r.calls, "weapon-animation")
			return r.weaponState + int32(equip&1)
		},
	}
}

func TestPlayerActionState4FA2B0ExhaustiveStateTable(t *testing.T) {
	want := map[uint8]int32{
		0: 4, 1: 46, 2: 21, 3: 1, 4: 2, 5: 6, 10: 21, 12: 3,
		14: 46, 15: 40, 16: 40, 17: 40, 18: 48, 19: 49, 20: 47,
		21: 30, 22: 46, 23: 50, 24: 19, 25: 20, 26: 15, 27: 16,
		28: 16, 29: 16, 30: 52, 32: 54,
	}
	for state := 0; state <= 0xff; state++ {
		runtime := &playerActionStateTestRuntime4FA2B0{warcryActive: true, warcryCurrent: true}
		update := &playerActionStateTestUpdate4FA2B0{
			state:  uint8(state),
			player: &playerActionStateTestPlayer4FA2B0{},
		}
		got := playerActionState4FA2B0(&playerActionStateTestObject4FA2B0{update: update}, runtime.hooks())
		if got != want[uint8(state)] {
			t.Errorf("state %d result = %d, want %d", state, got, want[uint8(state)])
		}
	}
}

func TestPlayerActionState4FA2B0AbilityPriorityAndTrace(t *testing.T) {
	tests := []struct {
		name          string
		warcryActive  bool
		warcryCurrent bool
		berserkActive bool
		want          int32
		wantCalls     []string
	}{
		{
			name: "current warcry", warcryActive: true, warcryCurrent: true, berserkActive: true, want: 46,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "warcry:ABILITY_WARCRY"},
		},
		{
			name: "stale warcry then berserk", warcryActive: true, berserkActive: true, want: 45,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "warcry:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE"},
		},
		{
			name: "inactive warcry skips current check", berserkActive: true, want: 45,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runtime := &playerActionStateTestRuntime4FA2B0{
				warcryActive: tc.warcryActive, warcryCurrent: tc.warcryCurrent, berserkActive: tc.berserkActive,
			}
			update := &playerActionStateTestUpdate4FA2B0{state: 1}
			got := playerActionState4FA2B0(&playerActionStateTestObject4FA2B0{update: update}, runtime.hooks())
			if got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if !reflect.DeepEqual(runtime.calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", runtime.calls, tc.wantCalls)
			}
		})
	}
}

func TestPlayerActionState4FA2B0WeaponBranches(t *testing.T) {
	tests := []struct {
		name      string
		equip     uint32
		variant   uint8
		flags     uint8
		want      int32
		wantCalls []string
	}{
		{
			name: "ranged ready", equip: 0x10000, flags: 0, want: 31,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE", "player", "equip", "weapon", "use-data", "weapon-flags"},
		},
		{
			name: "ranged firing", equip: 0x10000, flags: 2, want: 29,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE", "player", "equip", "weapon", "use-data", "weapon-flags"},
		},
		{
			name: "unarmed variant", variant: 77, want: 77,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE", "player", "equip", "variant"},
		},
		{
			name: "bit zero variant", equip: 1, variant: 78, want: 78,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE", "player", "equip", "variant"},
		},
		{
			name: "zero variant fallback", want: 90,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE", "player", "equip", "variant", "weapon-animation"},
		},
		{
			name: "melee skips variant", equip: 4, variant: 79, want: 90,
			wantCalls: []string{"update", "state", "active:ABILITY_WARCRY", "active:ABILITY_BERSERKER_CHARGE", "player", "equip", "weapon-animation"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runtime := &playerActionStateTestRuntime4FA2B0{weaponState: 90}
			player := &playerActionStateTestPlayer4FA2B0{equip: tc.equip, variant: tc.variant}
			update := &playerActionStateTestUpdate4FA2B0{
				state: 1, player: player,
				weapon: &playerActionStateTestWeapon4FA2B0{use: &playerActionStateTestUse4FA2B0{flags: tc.flags}},
			}
			got := playerActionState4FA2B0(&playerActionStateTestObject4FA2B0{update: update}, runtime.hooks())
			if got != tc.want {
				t.Fatalf("result = %d, want %d", got, tc.want)
			}
			if !reflect.DeepEqual(runtime.calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", runtime.calls, tc.wantCalls)
			}
		})
	}
}

func TestPlayerActionState4FA2B0BlockState(t *testing.T) {
	for _, tc := range []struct {
		equip uint32
		want  int32
	}{{equip: 0x400, want: 38}, {equip: 0x200, want: 0}} {
		runtime := new(playerActionStateTestRuntime4FA2B0)
		update := &playerActionStateTestUpdate4FA2B0{
			state: 13, player: &playerActionStateTestPlayer4FA2B0{equip: tc.equip},
		}
		got := playerActionState4FA2B0(&playerActionStateTestObject4FA2B0{update: update}, runtime.hooks())
		if got != tc.want {
			t.Errorf("equip %#x result = %d, want %d", tc.equip, got, tc.want)
		}
		wantCalls := []string{"update", "state", "player", "equip"}
		if !reflect.DeepEqual(runtime.calls, wantCalls) {
			t.Errorf("equip %#x calls = %v, want %v", tc.equip, runtime.calls, wantCalls)
		}
	}
}

func TestPlayerWeaponAnimation4FA280FirstBitAndBounds(t *testing.T) {
	var calls []int
	load := func(bit int) uint32 {
		calls = append(calls, bit)
		return 0xf1230000 | uint32(bit)
	}
	want := uint32(0xf1230002)
	if got := playerWeaponAnimation4FA280((1<<26)|(1<<17)|(1<<2), load); uint32(got) != want {
		t.Fatalf("first-bit result = %#x, want %#x", uint32(got), want)
	}
	if !reflect.DeepEqual(calls, []int{2}) {
		t.Fatalf("first-bit table calls = %v, want [2]", calls)
	}
	calls = nil
	if got := playerWeaponAnimation4FA280((1<<27)|(1<<1), load); got != 0 {
		t.Fatalf("out-of-range result = %d, want 0", got)
	}
	if calls != nil {
		t.Fatalf("out-of-range table calls = %v, want nil", calls)
	}
}

func TestPlayerActionState4FA2B0DoesNotGuardPointers(t *testing.T) {
	runtime := new(playerActionStateTestRuntime4FA2B0)
	defer func() {
		if recover() == nil {
			t.Fatal("nil update pointer did not preserve original fault contract")
		}
	}()
	_ = playerActionState4FA2B0((*playerActionStateTestObject4FA2B0)(nil), runtime.hooks())
}
