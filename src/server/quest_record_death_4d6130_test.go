package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestQuestRecordDeath4D6130LoadStoreOrderAndLivePlayer(t *testing.T) {
	type player struct{ deaths, mask uint32 }
	type update struct{ player *player }
	type unit struct {
		flags  uint32
		update *update
	}
	first := &player{deaths: math.MaxUint32, mask: 0xa0}
	second := &player{deaths: 17, mask: 0x80000005}
	cached := &update{player: first}
	replacement := &update{player: &player{deaths: 77, mask: 0x10}}
	u := &unit{flags: 0x8000ffdf, update: cached}
	var trace []string
	loads := 0
	got := questRecordDeath4D6130(u, questRecordDeathHooks4D6130[*unit, *update, *player]{
		loadFlags: func(got *unit) uint32 {
			if got != u {
				t.Fatal("flags loaded from a different unit")
			}
			trace = append(trace, "flags")
			return got.flags
		},
		loadUpdate: func(got *unit) *update {
			trace = append(trace, "update")
			return got.update
		},
		loadPlayer: func(got *update) *player {
			if got != cached {
				t.Fatal("UpdateData was reloaded after the death-counter store")
			}
			loads++
			trace = append(trace, "player")
			return got.player
		},
		loadDeaths: func(got *player) uint32 {
			if got != first {
				t.Fatal("death counter was read from the wrong player")
			}
			trace = append(trace, "deaths")
			return got.deaths
		},
		storeDeaths: func(got *player, value uint32) {
			trace = append(trace, "increment")
			if got != first || value != 0 {
				t.Fatalf("wrapping increment = %p/%#x", got, value)
			}
			got.deaths = value
			cached.player, u.update = second, replacement
		},
		loadMask: func(got *player) uint32 {
			if got != second {
				t.Fatal("progress mask did not use the second live Player")
			}
			trace = append(trace, "mask")
			return got.mask
		},
		storeMask: func(got *player, value uint32) {
			trace = append(trace, "mark")
			got.mask = value
		},
	})
	want := []string{"flags", "update", "player", "deaths", "increment", "player", "mask", "mark"}
	if !reflect.DeepEqual(trace, want) || loads != 2 || !got.isPlayer || got.player != second || got.unit != nil {
		t.Fatalf("trace/result = %v/%+v, want %v/second Player", trace, got, want)
	}
	if *first != (player{deaths: 0, mask: 0xa0}) || *second != (player{deaths: 17, mask: 0x80000007}) ||
		*replacement.player != (player{deaths: 77, mask: 0x10}) {
		t.Fatalf("first/second/replacement = %+v/%+v/%+v", first, second, replacement.player)
	}
}

func TestQuestRecordDeath4D6130NilAndDestroyedReturnInput(t *testing.T) {
	var nilUnit *Object
	nilResult := questRecordDeath4D6130(nilUnit, questRecordDeathHooks4D6130[*Object, *PlayerUpdateData, *Player]{})
	if nilResult.isPlayer || nilResult.unit != nil || nilResult.player != nil || QuestRecordDeath4D6130(nil) != nil {
		t.Fatalf("nil result = %+v", nilResult)
	}
	for _, flags := range []object.Flags{object.FlagDestroyed, object.Flags(math.MaxUint32)} {
		u := &Object{ObjFlags: flags}
		reads := 0
		got := questRecordDeath4D6130(u, questRecordDeathHooks4D6130[*Object, *PlayerUpdateData, *Player]{
			loadFlags: func(got *Object) uint32 { reads++; return uint32(got.ObjFlags) },
		})
		if reads != 1 || got.isPlayer || got.unit != u || got.player != nil || QuestRecordDeath4D6130(u) != unsafe.Pointer(u) {
			t.Fatalf("destroyed result/reads = %+v/%d", got, reads)
		}
	}
}

func TestQuestRecordDeath4D6130NativePointersAndUntouchedFields(t *testing.T) {
	for _, deaths := range []uint32{0, 1, 0x80000000, math.MaxUint32} {
		p := &Player{field4656: 0x11223344, field4660: deaths, field4664: 0x55667788, field4692: 0x80000021, QuestStage: 0xaabbccdd}
		update := &PlayerUpdateData{Player: p, ExtraLives: 17, Field137: 0x76543210}
		u := &Object{ObjFlags: object.FlagDead | object.FlagActive, UpdateData: unsafe.Pointer(update)}
		beforeUnit, beforeUpdate, wantPlayer := *u, *update, *p
		wantPlayer.field4660++
		wantPlayer.field4692 |= 2
		if got := QuestRecordDeath4D6130(u); got != unsafe.Pointer(p) {
			t.Fatalf("result = %p, want native Player %p", got, p)
		}
		if *u != beforeUnit || *update != beforeUpdate || *p != wantPlayer {
			t.Fatal("death statistic changed fields other than its DWORD and mask")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for name, pointer := range map[string]unsafe.Pointer{"unit": unsafe.Pointer(u), "update": unsafe.Pointer(update), "player": unsafe.Pointer(p)} {
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("%s pointer = %p, want above 4 GiB", name, pointer)
				}
			}
		}
		runtime.KeepAlive(u)
		runtime.KeepAlive(update)
		runtime.KeepAlive(p)
	}
}

func TestQuestRecordDeath4D6130PreservesOriginalMissingBindingFaults(t *testing.T) {
	for _, u := range []*Object{{}, {UpdateData: unsafe.Pointer(&PlayerUpdateData{})}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("a missing live binding was silently ignored")
				}
			}()
			QuestRecordDeath4D6130(u)
		}()
	}
}

func TestQuestRecordDeath4D6130NativeLayout(t *testing.T) {
	wantUpdate, wantPlayer, wantDeaths, wantMask := uintptr(748), uintptr(276), uintptr(4660), uintptr(4692)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantUpdate, wantPlayer, wantDeaths, wantMask = 872, 336, 5964, 5996
	}
	for _, check := range []struct{ got, want uintptr }{
		{unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{unsafe.Offsetof(Player{}.field4660), wantDeaths},
		{unsafe.Offsetof(Player{}.field4692), wantMask},
		{unsafe.Sizeof(Player{}.field4660), 4},
		{unsafe.Sizeof(Player{}.field4692), 4},
	} {
		if check.got != check.want {
			t.Errorf("native layout = %d, want %d", check.got, check.want)
		}
	}
}
