package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func playerDieFixture54D2B0() (*Object, *PlayerUpdateData, *Player) {
	player := &Player{Field3600: 44, ProtUnitManaCur: 0x12345678}
	update := &PlayerUpdateData{
		ManaCur:        17,
		Field47_0:      9,
		TrapSpells:     [5]uint32{1, 2, 3, 4, 5},
		TrapSpellsCnt:  0xaabbcc05,
		SpellCastStart: 99,
		Player:         player,
	}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagActive | object.FlagEnabled | object.FlagDead | object.FlagShadow,
		Field131:   8,
		HealthData: &HealthData{Cur: 0, Max: 20},
		UpdateData: unsafe.Pointer(update),
	}
	player.PlayerUnit = unit
	unit.Buffs = 0x1234
	unit.BuffsDur[2] = 77
	unit.BuffsPower[2] = 8
	return unit, update, player
}

func playerDieRuntime54D2B0(t *testing.T, events *[]string) PlayerDieRuntime54D2B0 {
	t.Helper()
	return PlayerDieRuntime54D2B0{
		GameFlag: func(flag uint32) bool {
			return flag == playerDieCoopMode54D2B0
		},
		PrepareAnkhType:   func() { *events = append(*events, "ankh") },
		CancelPendingSave: func() { *events = append(*events, "save") },
		Audio: func(id int, _ *Object) {
			*events = append(*events, "audio:"+string(rune(id)))
		},
		SetPlayerState: func(unit *Object, state PlayerState) bool {
			*events = append(*events, "state")
			unit.UpdateDataPlayer().State = state
			return true
		},
		RemoveActionShadow: func(unit *Object) {
			*events = append(*events, "shadow")
			unit.ObjFlags &^= object.FlagShadow
		},
		DropAllItems: func(*Object) int32 {
			*events = append(*events, "drop")
			return 0
		},
		NotifyPlayerDied: func(*Object) { *events = append(*events, "notify") },
		ProtectMana: func(token uint32, delta int16) {
			if token != 0x12345678 || delta != 0 {
				t.Fatalf("mana protection = (%#x, %d)", token, delta)
			}
			*events = append(*events, "mana")
		},
		SetBuffFlags: func(unit *Object, flags uint32) {
			*events = append(*events, "buff")
			unit.Buffs = flags
		},
		CancelAbilities: func(*Object) { *events = append(*events, "abilities") },
		CancelSpells:    func(*Object) { *events = append(*events, "spells") },
		Unsupported: func(reason string, _ *Object) {
			t.Fatalf("unexpected unsupported branch: %s", reason)
		},
	}
}

func TestPlayerDieNative54D2B0SoloCoopOrderAndState(t *testing.T) {
	unit, update, player := playerDieFixture54D2B0()
	var events []string
	runtime := playerDieRuntime54D2B0(t, &events)
	if !PlayerDieNative54D2B0(unit, runtime) {
		t.Fatal("solo cooperative death was not handled")
	}
	wantEvents := []string{
		"ankh", "save", "audio:" + string(rune(playerDieMaleSound54D2B0)), "state",
		"shadow", "drop", "notify", "mana", "buff", "abilities", "spells", "buff",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	if update.State != PlayerState3 || update.ManaCur != 0 || update.Field47_0 != 0 ||
		update.SpellCastStart != 0 || update.TrapSpells != [5]uint32{} ||
		update.TrapSpellsCnt != 0xaabbcc00 || player.Field3600 != 0 {
		t.Fatalf("cleared update/player = %#v, field3600=%d", *update, player.Field3600)
	}
	wantFlags := object.FlagDead | object.FlagShort
	if unit.ObjFlags&wantFlags != wantFlags || unit.ObjFlags.Has(object.FlagShadow) ||
		unit.Buffs != 0 || unit.BuffsDur[2] != 0 || unit.BuffsPower[2] != 0 {
		t.Fatalf("unit death state = flags:%#x buffs:%#x dur:%d power:%d",
			unit.ObjFlags, unit.Buffs, unit.BuffsDur[2], unit.BuffsPower[2])
	}
}

func TestPlayerDieNative54D2B0ElectricAndFemaleAudio(t *testing.T) {
	for _, tc := range []struct {
		name   string
		typ    uint32
		female bool
		want   int
	}{
		{name: "electric overrides female", typ: playerDieElectricDamage54D2B0, female: true, want: playerDieElectricSound54D2B0},
		{name: "female", typ: 8, female: true, want: playerDieFemaleSound54D2B0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, _, player := playerDieFixture54D2B0()
			unit.Field131 = tc.typ
			player.Info().SetIsFemale(byte(0))
			if tc.female {
				player.Info().SetIsFemale(1)
			}
			var got int
			var events []string
			runtime := playerDieRuntime54D2B0(t, &events)
			runtime.Audio = func(id int, _ *Object) { got = id }
			if !PlayerDieNative54D2B0(unit, runtime) || got != tc.want {
				t.Fatalf("handled/audio = %t/%d, want true/%d", got != 0, got, tc.want)
			}
		})
	}
}

func TestPlayerDieNative54D2B0RejectsBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Object, *PlayerUpdateData, *Player, *PlayerDieRuntime54D2B0)
		want   string
	}{
		{name: "online", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.GameFlag = func(flag uint32) bool { return flag == playerDieCoopMode54D2B0 || flag == playerDieOnlineMode54D2B0 }
		}, want: "non-solo-cooperative mode"},
		{name: "quest", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.GameFlag = func(flag uint32) bool { return flag == playerDieCoopMode54D2B0 || flag == playerDieQuestMode54D2B0 }
		}, want: "non-solo-cooperative mode"},
		{name: "shop", mutate: func(_ *Object, u *PlayerUpdateData, _ *Player, _ *PlayerDieRuntime54D2B0) {
			u.Trade70 = &TradeSession{}
		}, want: "active shop session"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, update, player := playerDieFixture54D2B0()
			var events []string
			runtime := playerDieRuntime54D2B0(t, &events)
			var reason string
			runtime.Unsupported = func(got string, _ *Object) { reason = got }
			tc.mutate(unit, update, player, &runtime)
			beforeUnit := *unit
			beforeUpdate := *update
			beforePlayer := *player
			if PlayerDieNative54D2B0(unit, runtime) || reason != tc.want {
				t.Fatalf("handled/reason = %t/%q, want false/%q", reason == "", reason, tc.want)
			}
			if *unit != beforeUnit || *update != beforeUpdate || *player != beforePlayer || len(events) != 0 {
				t.Fatal("unsupported branch changed state")
			}
		})
	}
}
