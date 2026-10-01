package server

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func playerDieQuestRuntimeTest54D2B0(t *testing.T, events *[]string) PlayerDieRuntime54D2B0 {
	t.Helper()
	r := playerDieRuntime54D2B0(t, events)
	r.GameFlag = func(flag uint32) bool { return flag == playerDieQuestMode54D2B0 }
	r.Quest = &PlayerDieQuestRuntime54D2B0{
		SendStats: func(uint8, [14]byte) { *events = append(*events, "stats") },
		RecordDeath: func(unit *Object) unsafe.Pointer {
			*events = append(*events, "record")
			return QuestRecordDeath4D6130(unit)
		},
		ResetPlayer: func(unit *Object) {
			*events = append(*events, "reset")
			new(Server).ResetQuestPlayer4D6000(unit, func() uint32 { return 19 })
		},
		Penalty: func(*Object) { *events = append(*events, "penalty") },
		BalanceFloat: func(key string) float32 {
			if key != "QuestGameStartingExtraLives" {
				t.Fatalf("balance key = %q", key)
			}
			*events = append(*events, "balance")
			return 3
		},
	}
	return r
}

func TestPlayerDieNative54D2B0QuestLivesAndZeroLifeSequence(t *testing.T) {
	for _, lives := range []uint32{0, 1, 2, 3, 0x80000000, math.MaxUint32} {
		unit, update, p := playerDieFixture54D2B0()
		unit.NetCode, update.ExtraLives, p.PlayerInd = 0x1234, lives, 31
		p.field4660, p.field4692 = math.MaxUint32, 0x80000021
		p.field4668, p.field4672, p.field4664, p.field4688 = 0xaabb1234, 0xccdd5678, 0xeeff9abc, 0x1122def0
		update.RespawnMarkers[31] = 0x77
		var events []string
		r := playerDieQuestRuntimeTest54D2B0(t, &events)
		r.Quest.SendStats = func(index uint8, packet [14]byte) {
			events = append(events, "stats")
			want := [14]byte{0xf0, 2, 0x34, 0x12, 0x78, 0x56, 0xbc, 0x9a, 0xf0, 0xde}
			if index != 31 || packet != want || p.field4660 != math.MaxUint32 || update.Field137 != 1000 {
				t.Fatalf("pre-reset stats = %d/%x", index, packet)
			}
		}
		r.Quest.Penalty = func(got *Object) {
			events = append(events, "penalty")
			if got != unit || p.field4660 != 0 || p.field4664 != 0 || p.field4668 != 0 || p.field4672 != 0 || p.field4688 != 19 || p.field4692 != 63 || update.ExtraLives != 0 {
				t.Fatal("penalty did not follow native statistics reset")
			}
		}
		if !PlayerDieNative54D2B0(unit, r) {
			t.Fatal("Quest death rejected")
		}
		want := []string{"ankh", "audio:" + string(rune(playerDieMaleSound54D2B0)), "state", "shadow", "notify", "mana", "buff", "abilities", "spells", "buff"}
		if lives != 0 {
			want = append(want, "record")
			if update.ExtraLives != lives-1 || p.field4660 != 0 || p.field4692 != 0x80000023 || update.Field137 != 0 || update.RespawnMarkers[31] != 0x77 {
				t.Fatalf("positive-life state = %d/%d/%#x", update.ExtraLives, p.field4660, p.field4692)
			}
		} else {
			want = append(want, "stats", "reset", "penalty", "balance")
			if update.ExtraLives != 3 || update.RespawnMarkers[31] != 3 {
				t.Fatalf("new life state = %d/%d", update.ExtraLives, update.RespawnMarkers[31])
			}
		}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("lives=%#x events=%v, want %v", lives, events, want)
		}
	}
}

func TestPlayerDieNative54D2B0QuestRetainsEntryUpdateAndReloadsCommonPlayers(t *testing.T) {
	unit, entry, first := playerDieFixture54D2B0()
	entry.ExtraLives, entry.Trade70 = 3, &TradeSession{}
	female := &Player{PlayerUnit: unit, ProtUnitManaCur: 11, Field3600: 0}
	female.Info().SetIsFemale(1)
	mana := &Player{PlayerUnit: unit, ProtUnitManaCur: 0xabcdef12, Field3600: 8}
	abs := &Player{PlayerUnit: unit, Field3600: 12}
	final := &Player{PlayerUnit: unit, field4660: 21, field4692: 0x80}
	live := &PlayerUpdateData{Player: final, ManaCur: 59, ExtraLives: 99, Field47_0: 3, SpellCastStart: 17}
	var events []string
	r := playerDieQuestRuntimeTest54D2B0(t, &events)
	r.PrepareAnkhType = func() { entry.Player = female; unit.UpdateData = unsafe.Pointer(live) }
	r.SetPlayerState = func(got *Object, state PlayerState) bool {
		if got != unit || state != PlayerState3 {
			t.Fatal("state arguments")
		}
		entry.Player = mana
		return false // The original ignores this result.
	}
	r.ProtectMana = func(token uint32, delta int16) {
		if token != 0xabcdef12 || delta != 0 || entry.ManaCur != 0 || live.ManaCur != 59 {
			t.Fatal("mana must use live Player of the entry-cached update")
		}
	}
	r.CancelAbilities = func(*Object) { entry.Player = abs }
	r.CancelTrade = func(session *TradeSession) {
		if session != entry.Trade70 || abs.Field3600 != 0 {
			t.Fatal("cached trade/common clear")
		}
	}
	if !PlayerDieNative54D2B0(unit, r) {
		t.Fatal("Quest death rejected")
	}
	if len(events) == 0 || events[0] != "audio:"+string(rune(playerDieFemaleSound54D2B0)) || first.Field3600 != 44 || mana.Field3600 != 8 || abs.Field3600 != 0 {
		t.Fatalf("fresh common players = %v/%d/%d/%d", events, first.Field3600, mana.Field3600, abs.Field3600)
	}
	if entry.ExtraLives != 2 || entry.Trade70 != nil || live.ExtraLives != 99 || final.field4660 != 22 || final.field4692 != 0x82 || live.Field47_0 != 3 || live.SpellCastStart != 17 {
		t.Fatalf("entry/fresh-helper updates = %+v/%+v player=%+v", entry, live, final)
	}
}

func TestPlayerDieQuestNative54D2B0NativeZeroLifeLinksAndReturn(t *testing.T) {
	unit, entry, original := playerDieFixture54D2B0()
	packetPlayer := &Player{PlayerUnit: unit, PlayerInd: 7, field4668: 0xaabb1234, field4672: 0xccdd5678, field4664: 0xeeff9abc, field4688: 0x1122def0}
	final := &Player{PlayerUnit: unit, PlayerInd: 31}
	live := &PlayerUpdateData{Player: final, ExtraLives: 77, ManaCur: 55}
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(entry), unsafe.Pointer(original), unsafe.Pointer(packetPlayer), unsafe.Pointer(final), unsafe.Pointer(live)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native pointer %p below 4 GiB", pointer)
		}
	}
	var events []string
	r := playerDieQuestRuntimeTest54D2B0(t, &events)
	r.Frame = func() uint32 { entry.Player = packetPlayer; return 0xffffffff }
	r.Quest.SendStats = func(index uint8, packet [14]byte) {
		if index != 7 || binary.LittleEndian.Uint16(packet[8:]) != 0xdef0 || entry.Field137 != math.MaxUint32 {
			t.Fatal("frame-before-Player/stat narrowing")
		}
		unit.UpdateData = unsafe.Pointer(live)
	}
	r.Quest.ResetPlayer = func(got *Object) { new(Server).ResetQuestPlayer4D6000(got, func() uint32 { return 23 }) }
	r.Quest.Penalty = func(got *Object) {
		if got != unit || final.field4688 != 23 || final.field4692 != 63 {
			t.Fatal("reset followed current unit")
		}
		entry.Player = final
	}
	r.Quest.BalanceFloat = func(string) float32 { return 257.5 }
	got := playerDieQuestNative54D2B0(unit, entry, r)
	if got.kind != playerDieQuestPlayerReturn54D2B0 || got.player != final || entry.ExtraLives != 258 || entry.RespawnMarkers[31] != 2 || live.ExtraLives != 77 || live.ManaCur != 55 || packetPlayer.field4688 != 0x1122def0 {
		t.Fatalf("native result/cached stores = %+v/%+v/%+v", got, entry, live)
	}
}

func TestPlayerDieNative54D2B0MissingQuestServicesDoNotMutate(t *testing.T) {
	for _, name := range []string{"Quest", "SendStats", "RecordDeath", "ResetPlayer", "Penalty", "BalanceFloat"} {
		t.Run(name, func(t *testing.T) {
			unit, update, p := playerDieFixture54D2B0()
			var events []string
			r := playerDieQuestRuntimeTest54D2B0(t, &events)
			if name == "Quest" {
				r.Quest = nil
			} else {
				field := reflect.ValueOf(r.Quest).Elem().FieldByName(name)
				field.Set(reflect.Zero(field.Type()))
			}
			var reason string
			r.Unsupported = func(got string, _ *Object) { reason = got }
			beforeUnit, beforeUpdate, beforePlayer := *unit, *update, *p
			if PlayerDieNative54D2B0(unit, r) || reason != "missing quest death service" || len(events) != 0 || *unit != beforeUnit || *update != beforeUpdate || *p != beforePlayer {
				t.Fatalf("missing %s changed state/reason %q", name, reason)
			}
		})
	}
}

func TestPlayerDieQuestNative54D2B0PreservesMissingPlayerFault(t *testing.T) {
	unit, update, _ := playerDieFixture54D2B0()
	var events []string
	r := playerDieQuestRuntimeTest54D2B0(t, &events)
	r.Frame = func() uint32 { events = append(events, "frame"); return 123 }
	update.Player = nil
	defer func() {
		if recover() == nil || !reflect.DeepEqual(events, []string{"frame"}) || update.Field137 != 123 || update.ExtraLives != 0 {
			t.Fatalf("missing Player fault prefix/state = %v/%+v", events, update)
		}
	}()
	playerDieQuestNative54D2B0(unit, update, r)
}

func TestPlayerDieNative54D2B0OfflineAttributionReadsIndexAfterFrame(t *testing.T) {
	unit, _, p := playerDieFixture54D2B0()
	p.Field3600, p.Field3604 = 1, 4
	p.SetLastAggressorFrame(980)
	var events []string
	r := playerDieRuntime54D2B0(t, &events)
	r.Frame = func() uint32 { p.Field3604 = 7; return 1000 }
	var index uint32
	r.PlayerByIndex = func(got uint32) *Player { index = got; return nil }
	if !PlayerDieNative54D2B0(unit, r) || index != 7 {
		t.Fatalf("offline late index = %d", index)
	}
	if unit.ObjFlags&object.FlagDead == 0 {
		t.Fatal("death did not finish")
	}
}
