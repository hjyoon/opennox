package server

import (
	"testing"
	"unsafe"
)

func TestPlayerRespawnStateReset4EF660NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantUpdateSize := uintptr(556)
	wantUpdateData := uintptr(748)
	wantAttribution := uintptr(520)
	wantPending := uintptr(116)
	wantCurTraps := uintptr(244)
	wantField66 := uintptr(264)
	wantPlayer := uintptr(276)
	wantSoulGate := uintptr(308)
	wantPlayerSize := uintptr(4828)
	wantMarker0 := uintptr(3660)
	wantMarker1 := uintptr(3664)
	wantQuestAnkhs := uintptr(4796)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantUpdateSize = 640
		wantUpdateData = 872
		wantAttribution = 576
		wantPending = 120
		wantCurTraps = 288
		wantField66 = 308
		wantPlayer = 320
		wantSoulGate = 376
		wantPlayerSize = 6160
		wantMarker0 = 4956
		wantMarker1 = 4960
		wantQuestAnkhs = 6104
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"Object.Obj130", unsafe.Offsetof(Object{}.Obj130), wantAttribution},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Field29", unsafe.Offsetof(PlayerUpdateData{}.Field29), wantPending},
		{"PlayerUpdateData.Field29 size", unsafe.Sizeof(PlayerUpdateData{}.Field29), 4 * unsafe.Sizeof(uintptr(0))},
		{"PlayerUpdateData.CurTraps", unsafe.Offsetof(PlayerUpdateData{}.CurTraps), wantCurTraps},
		{"PlayerUpdateData.Field66", unsafe.Offsetof(PlayerUpdateData{}.Field66), wantField66},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.SoulGate", unsafe.Offsetof(PlayerUpdateData{}.SoulGate), wantSoulGate},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.field3660", unsafe.Offsetof(Player{}.field3660), wantMarker0},
		{"Player.field3664", unsafe.Offsetof(Player{}.field3664), wantMarker1},
		{"Player.QuestAnkhs", unsafe.Offsetof(Player{}.QuestAnkhs), wantQuestAnkhs},
		{"Player.QuestAnkhs size", unsafe.Sizeof(Player{}.QuestAnkhs), 5 * unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerRespawnStateReset4EF660NativeFieldsAndLowByte(t *testing.T) {
	value := &Object{}
	player := &Player{field3660: 1, field3664: 2}
	for index := range player.QuestAnkhs {
		player.QuestAnkhs[index] = value
	}
	update := &PlayerUpdateData{
		SoulGate: value,
		Player:   player,
		CurTraps: 0xa1b2c3dd,
		Field66:  0xffffffff,
	}
	for index := range update.Field29 {
		update.Field29[index] = value
	}
	unit := &Object{UpdateData: unsafe.Pointer(update), Obj130: value}

	got := playerRespawnStateResetNative4EF660(unit, playerRespawnStateResetNativeDeps4EF660{
		gameFlag:    func(uint32) int32 { return 0 },
		countGlyphs: func(*Object) int32 { return 0x123 },
	})
	if got != player {
		t.Fatalf("result = %p, want %p", got, player)
	}
	if update.Field29 != [4]*Object{} || update.SoulGate != nil || player.QuestAnkhs != [5]*Object{} {
		t.Fatalf("pointer reset = pending %v soul %p ankhs %v", update.Field29, update.SoulGate, player.QuestAnkhs)
	}
	if update.CurTraps != 0xa1b2c323 || update.Field66 != 0 || unit.Obj130 != nil {
		t.Fatalf("fixed fields = traps %#x field66 %#x attribution %p", update.CurTraps, update.Field66, unit.Obj130)
	}
	if player.field3660 != 0xdeadface || player.field3664 != 0xdeadface {
		t.Fatalf("markers = %#x/%#x", player.field3660, player.field3664)
	}
}

func TestPlayerRespawnStateReset4EF660NativeCallbacksKeepCachedUpdate(t *testing.T) {
	entryPlayer := &Player{}
	livePlayer := &Player{}
	cached := &PlayerUpdateData{Player: entryPlayer, CurTraps: 0x11223344, Field66: 9}
	replacement := &PlayerUpdateData{CurTraps: 0xffffffff, Field66: 7}
	unit := &Object{UpdateData: unsafe.Pointer(cached), Obj130: &Object{}}

	got := playerRespawnStateResetNative4EF660(unit, playerRespawnStateResetNativeDeps4EF660{
		gameFlag: func(flag uint32) int32 {
			if flag != playerRespawnStateResetCoopFlag4EF660 {
				t.Fatalf("flag = %#x", flag)
			}
			unit.UpdateData = unsafe.Pointer(replacement)
			cached.Player = livePlayer
			return 0
		},
		countGlyphs: func(got *Object) int32 {
			if got != unit {
				t.Fatalf("glyph unit = %p, want %p", got, unit)
			}
			return -1
		},
	})
	if got != livePlayer || livePlayer.field3660 != 0xdeadface || livePlayer.field3664 != 0xdeadface {
		t.Fatalf("live player/result = %p/%#x/%#x", got, livePlayer.field3660, livePlayer.field3664)
	}
	if entryPlayer.field3660 != 0 || entryPlayer.field3664 != 0 {
		t.Fatalf("entry player markers changed = %#x/%#x", entryPlayer.field3660, entryPlayer.field3664)
	}
	if cached.CurTraps != 0x112233ff || cached.Field66 != 0 {
		t.Fatalf("cached update = traps %#x field66 %#x", cached.CurTraps, cached.Field66)
	}
	if replacement.CurTraps != 0xffffffff || replacement.Field66 != 7 {
		t.Fatalf("replacement update changed = traps %#x field66 %#x", replacement.CurTraps, replacement.Field66)
	}
}
