package server

import (
	"encoding/binary"
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
		Frame: func() uint32 { return 1000 },
		TickRate: func() uint32 {
			return 30
		},
		PlayerByIndex:   func(uint32) *Player { return nil },
		ObjectByNetCode: func(uint32) *Object { return nil },
		InformText:      func(int, [14]byte) { *events = append(*events, "text") },
		ResetAbility:    func(*Object, int32) { *events = append(*events, "ability") },
		GameplayHasRivals: func() bool {
			return false
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
		CancelTrade:     func(*TradeSession) { *events = append(*events, "trade") },
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

func TestPlayerDieNative54D2B0HostedOnlineMonsterDeath(t *testing.T) {
	unit, update, _ := playerDieFixture54D2B0()
	unit.NetCode = 23
	unit.Obj130 = &Object{ObjClass: object.ClassMonster, TypeInd: 1353}
	update.Field75 = 0xaabb
	update.Field76 = 9

	var events []string
	runtime := playerDieRuntime54D2B0(t, &events)
	runtime.GameFlag = func(flag uint32) bool {
		return flag == playerDieOnlineMode54D2B0 || flag == playerDieArenaMode54D2B0
	}
	var code int
	var packet [14]byte
	runtime.InformText = func(gotCode int, gotPacket [14]byte) {
		code, packet = gotCode, gotPacket
		events = append(events, "text")
	}
	if !PlayerDieNative54D2B0(unit, runtime) {
		t.Fatal("hosted online death was not handled")
	}

	wantEvents := []string{
		"ankh", "text", "audio:" + string(rune(playerDieMaleSound54D2B0)), "state",
		"shadow", "drop", "notify", "mana", "buff", "abilities", "spells", "buff",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	var wantPacket [14]byte
	binary.LittleEndian.PutUint16(wantPacket[6:], 23)
	binary.LittleEndian.PutUint16(wantPacket[8:], 1353)
	wantPacket[10] = 1
	if code != 14 || packet != wantPacket {
		t.Fatalf("death text = %d/%v, want 14/%v", code, packet, wantPacket)
	}
	if update.Field76 != 0 {
		t.Fatalf("online damage marker was not cleared: %d", update.Field76)
	}
}

func TestPlayerDieNative54D2B0OnlinePlayerAttribution(t *testing.T) {
	unit, update, player := playerDieFixture54D2B0()
	unit.NetCode = 23
	killer := &Object{ObjClass: object.ClassPlayer, NetCode: 101}
	unit.Obj130 = &Object{ObjOwner: killer}
	update.Field75 = 2
	update.Field76 = 2
	player.Field3600 = 1
	player.Field3604 = 7
	player.SetLastAggressorFrame(950)
	assistUnit := &Object{ObjClass: object.ClassPlayer, NetCode: 202}
	aggressor := &Player{Active: 1, PlayerUnit: assistUnit, NetCodeVal: 202}

	var events []string
	runtime := playerDieRuntime54D2B0(t, &events)
	runtime.GameFlag = func(flag uint32) bool { return flag == playerDieOnlineMode54D2B0 }
	runtime.PlayerByIndex = func(index uint32) *Player {
		if index != 7 {
			t.Fatalf("aggressor index = %d, want 7", index)
		}
		return aggressor
	}
	runtime.ObjectByNetCode = func(code uint32) *Object {
		if code != 202 {
			t.Fatalf("aggressor net code = %d, want 202", code)
		}
		return assistUnit
	}
	var packet [14]byte
	runtime.InformText = func(code int, got [14]byte) {
		if code != 14 {
			t.Fatalf("death text code = %d, want 14", code)
		}
		packet = got
	}
	var resetObject *Object
	var resetAbility int32
	runtime.ResetAbility = func(obj *Object, ability int32) {
		resetObject, resetAbility = obj, ability
	}
	if !PlayerDieNative54D2B0(unit, runtime) {
		t.Fatal("online attributed death was not handled")
	}
	if got := binary.LittleEndian.Uint16(packet[2:]); got != 101 {
		t.Fatalf("primary killer = %d, want 101", got)
	}
	if got := binary.LittleEndian.Uint16(packet[4:]); got != 202 {
		t.Fatalf("assisting killer = %d, want 202", got)
	}
	if got := binary.LittleEndian.Uint16(packet[6:]); got != 23 {
		t.Fatalf("victim = %d, want 23", got)
	}
	if got := binary.LittleEndian.Uint16(packet[8:]); got != 2 || packet[10] != 2 {
		t.Fatalf("damage attribution = %d/%d, want 2/2", got, packet[10])
	}
	if resetObject != killer || resetAbility != 1 {
		t.Fatalf("ability reset = %p/%d, want %p/1", resetObject, resetAbility, killer)
	}
}

func TestPlayerDieNative54D2B0CancelsActiveShop(t *testing.T) {
	unit, update, _ := playerDieFixture54D2B0()
	session := &TradeSession{}
	update.Trade70 = session
	var events []string
	runtime := playerDieRuntime54D2B0(t, &events)
	var canceled *TradeSession
	runtime.CancelTrade = func(got *TradeSession) { canceled = got }
	if !PlayerDieNative54D2B0(unit, runtime) {
		t.Fatal("death with an active shop was not handled")
	}
	if canceled != session || update.Trade70 != nil {
		t.Fatalf("canceled/current trade = %p/%p, want %p/nil", canceled, update.Trade70, session)
	}
}

func TestPlayerDieNative54D2B0RejectsBeforeMutation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Object, *PlayerUpdateData, *Player, *PlayerDieRuntime54D2B0)
		want   string
	}{
		{name: "unsupported mode", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.GameFlag = func(uint32) bool { return false }
		}, want: "unsupported game mode"},
		{name: "quest", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.GameFlag = func(flag uint32) bool { return flag == playerDieCoopMode54D2B0 || flag == playerDieQuestMode54D2B0 }
		}, want: "quest mode"},
		{name: "elimination", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.GameFlag = func(flag uint32) bool {
				return flag == playerDieOnlineMode54D2B0 || flag == playerDieElimMode54D2B0
			}
		}, want: "elimination mode"},
		{name: "missing cooperative service", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.CancelPendingSave = nil
		}, want: "missing cooperative death service"},
		{name: "missing online service", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.GameFlag = func(flag uint32) bool { return flag == playerDieOnlineMode54D2B0 }
			r.InformText = nil
		}, want: "missing online death service"},
		{name: "competitive scoring", mutate: func(_ *Object, _ *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			r.GameFlag = func(flag uint32) bool {
				return flag == playerDieOnlineMode54D2B0 || flag == playerDieArenaMode54D2B0
			}
			r.GameplayHasRivals = func() bool { return true }
		}, want: "competitive scoring"},
		{name: "missing shop service", mutate: func(_ *Object, u *PlayerUpdateData, _ *Player, r *PlayerDieRuntime54D2B0) {
			u.Trade70 = &TradeSession{}
			r.CancelTrade = nil
		}, want: "missing shop death service"},
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
