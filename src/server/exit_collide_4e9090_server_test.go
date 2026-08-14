package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestExitCollideNativeHelpers4E9090PreserveFixedWidthState(t *testing.T) {
	update := &PlayerUpdateData{CurTraps: 0xa1b2c300}
	if got := exitCollideCurTrapsByte4E9090(update); got != 0 {
		t.Fatalf("CurTraps low byte = %02x, want 00", got)
	}
	exitCollideStoreCurTrapsByte4E9090(update, 0xfe)
	if update.CurTraps != 0xa1b2c3fe {
		t.Fatalf("CurTraps = %08x, want a1b2c3fe", update.CurTraps)
	}

	player := &Player{Field4792: 1, field4652: ^uint32(0), field4692: 0x10}
	unit := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player})}
	server := &Server{}
	server.exitCollideRecordProgress4D60E0(unit)
	if player.field4652 != 0 || player.field4692 != 0x11 {
		t.Fatalf("progress = (%08x,%08x), want (00000000,00000011)", player.field4652, player.field4692)
	}

	player.Field4792 = 2
	player.field4652 = 9
	player.field4692 = 0x20
	server.exitCollideRecordProgress4D60E0(unit)
	if player.field4652 != 9 || player.field4692 != 0x20 {
		t.Fatalf("non-exact Quest state changed progress to (%d,%08x)", player.field4652, player.field4692)
	}

	unit.ObjFlags = 0x20
	player.Field4792 = 1
	server.exitCollideRecordProgress4D60E0(unit)
	if player.field4652 != 9 || player.field4692 != 0x20 {
		t.Fatalf("flagged unit changed progress to (%d,%08x)", player.field4652, player.field4692)
	}
}

func TestExitCollideUnitPacket4E9090MatchesInform18(t *testing.T) {
	got := exitCollideUnitPacket4E9090(exitCollideMessageExit4E9090, 0x89abcdef)
	want := [6]byte{0xa9, 18, 0xef, 0xcd, 0xab, 0x89}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("packet = % x, want % x", got, want)
	}
}

func TestExitCollideResetQuestPlayer4D6000UsesNativePlayer(t *testing.T) {
	first := &Player{
		field4652: 1, field4656: 2, field4660: 3, field4664: 4, field4668: 5,
		field4672: 6, field4676: 7, field4680: 8, field4684: 9, field4688: 10,
		field4692: 11,
	}
	second := &Player{field4688: 12, field4692: 13}
	update := &PlayerUpdateData{Player: first}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	calls := 0
	(&Server{}).exitCollideResetQuestPlayer4D6000(unit, func() uint32 {
		calls++
		update.Player = second
		return 0x80000005
	})
	if calls != 1 {
		t.Fatalf("current stage calls = %d, want 1", calls)
	}
	if first.field4652 != 0 || first.field4656 != 0 || first.field4660 != 0 ||
		first.field4664 != 0 || first.field4668 != 0 || first.field4672 != 0 ||
		first.field4676 != 0 || first.field4680 != 0 || first.field4684 != 0 ||
		first.field4688 != 10 || first.field4692 != 11 {
		t.Fatalf("pre-callback player fields = %#v", first)
	}
	if second.field4688 != 0x80000005 || second.field4692 != 63 {
		t.Fatalf("reloaded player tail = (%08x,%08x), want (80000005,0000003f)",
			second.field4688, second.field4692)
	}
}

func TestExitCollideMapString4E9090RetainsCStringContract(t *testing.T) {
	data := &ExitCollideData{}
	copy(data.MapName[:], "quest42.map")
	if got := exitCollideMapString4E9090(unsafe.Pointer(data)); got != "quest42.map" {
		t.Fatalf("map = %q, want quest42.map", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil C string must fault")
		}
	}()
	_ = exitCollideMapString4E9090(nil)
}

func TestExitCollideNativeLayout4E9090(t *testing.T) {
	wantObjectSize, wantCollideData, wantUpdateData := uintptr(780), uintptr(700), uintptr(748)
	wantOwnedNext, wantOwnedFirst := uintptr(512), uintptr(516)
	wantUpdateSize, wantCurTraps, wantPlayer := uintptr(556), uintptr(244), uintptr(276)
	wantQuestExit, wantQuestWarp := uintptr(312), uintptr(316)
	wantPlayerSize, wantPlayerIndex, wantQuestStage := uintptr(4828), uintptr(2064), uintptr(4696)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize, wantCollideData, wantUpdateData = 928, 776, 872
		wantOwnedNext, wantOwnedFirst = 560, 568
		wantUpdateSize, wantCurTraps, wantPlayer = 640, 288, 320
		wantQuestExit, wantQuestWarp = 384, 392
		wantPlayerSize, wantPlayerIndex, wantQuestStage = 6136, 2068, 6000
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"ExitCollideData size", unsafe.Sizeof(ExitCollideData{}), 88},
		{"ExitCollideData.MapName", unsafe.Offsetof(ExitCollideData{}.MapName), 0},
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantOwnedNext},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantOwnedFirst},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.CurTraps", unsafe.Offsetof(PlayerUpdateData{}.CurTraps), wantCurTraps},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.QuestExit", unsafe.Offsetof(PlayerUpdateData{}.QuestExit), wantQuestExit},
		{"PlayerUpdateData.QuestWarpGate", unsafe.Offsetof(PlayerUpdateData{}.QuestWarpGate), wantQuestWarp},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.QuestStage", unsafe.Offsetof(Player{}.QuestStage), wantQuestStage},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}
