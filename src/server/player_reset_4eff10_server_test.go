package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestPlayerReset4EFF10NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectFlags := uintptr(16)
	wantObject130 := uintptr(520)
	wantObjectField541 := uintptr(541)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantTrapSpells := uintptr(192)
	wantTrapCount := uintptr(212)
	wantPlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerIndex := uintptr(2064)
	wantPlayerLevel := uintptr(3684)
	wantMarker3660 := uintptr(3660)
	wantMarker3664 := uintptr(3664)
	wantManaToken := uintptr(4596)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectFlags = 20
		wantObject130 = 576
		wantObjectField541 = 601
		wantObjectUpdate = 872
		wantUpdateSize = 640
		wantTrapSpells = 228
		wantTrapCount = 248
		wantPlayer = 320
		wantPlayerSize = 6160
		wantPlayerIndex = 2068
		wantPlayerLevel = 4980
		wantMarker3660 = 4956
		wantMarker3664 = 4960
		wantManaToken = 5900
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantObjectFlags},
		{"Object.Obj130", unsafe.Offsetof(Object{}.Obj130), wantObject130},
		{"Object.Field541", unsafe.Offsetof(Object{}.Field541), wantObjectField541},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.ManaCur", unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 4},
		{"PlayerUpdateData.ManaPrev", unsafe.Offsetof(PlayerUpdateData{}.ManaPrev), 6},
		{"PlayerUpdateData.ManaMax", unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 8},
		{"PlayerUpdateData.TrapSpells", unsafe.Offsetof(PlayerUpdateData{}.TrapSpells), wantTrapSpells},
		{"PlayerUpdateData.TrapSpellsCnt", unsafe.Offsetof(PlayerUpdateData{}.TrapSpellsCnt), wantTrapCount},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.Level", unsafe.Offsetof(Player{}.Level), wantPlayerLevel},
		{"Player.field3660", unsafe.Offsetof(Player{}.field3660), wantMarker3660},
		{"Player.field3664", unsafe.Offsetof(Player{}.field3664), wantMarker3664},
		{"Player.ProtUnitManaCur", unsafe.Offsetof(Player{}.ProtUnitManaCur), wantManaToken},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerReset4EFF10NativeStateAndLivePlayers(t *testing.T) {
	players := make([]*Player, 9)
	for index := range players {
		players[index] = &Player{
			PlayerInd:       uint8(20 + index),
			ProtUnitManaCur: uint32(0x1000 + index),
		}
	}
	update := &PlayerUpdateData{
		ManaCur:       1,
		ManaPrev:      2,
		ManaMax:       0xabcd,
		TrapSpells:    [5]uint32{1, 2, 3, 4, 5},
		TrapSpellsCnt: 0xa1b2c3dd,
		Player:        players[0],
	}
	unit := &Object{
		ObjFlags:   0xffffffff,
		Obj130:     &Object{},
		Field541:   0xff,
		UpdateData: unsafe.Pointer(update),
	}

	var events []string
	record := func(event string) { events = append(events, event) }
	got := playerResetNative4EFF10(unit, PlayerResetRuntime4EFF10{
		AwardBeastScrolls: func(player *Player) {
			if player != players[0] {
				t.Fatalf("beast Player = %p", player)
			}
			record("beast")
			update.Player = players[1]
		},
		AwardSpells: func(player *Player) {
			if player != players[1] {
				t.Fatalf("spell Player = %p", player)
			}
			record("spells")
			update.Player = players[2]
		},
		CancelAbilities: func(got *Object) {
			if got != unit || players[2].Level != 1 {
				t.Fatalf("cancel state = %p/%d", got, players[2].Level)
			}
			record("cancel-abilities")
			update.Player = players[3]
		},
		ReadValues: func(got *Object, reward int32) {
			if got != unit || reward != 0 {
				t.Fatalf("read values = %p/%d", got, reward)
			}
			record("read-values")
			update.Player = players[4]
		},
		AwardWarriorAbilities: func(player *Player) {
			if player != players[4] {
				t.Fatalf("warrior Player = %p", player)
			}
			record("warrior")
			update.Player = players[5]
		},
		ProtectMana: func(token uint32, value uint16) {
			if token != players[5].ProtUnitManaCur || value != update.ManaMax || update.ManaCur != value || update.ManaPrev != value {
				t.Fatalf("protect state = %#x/%#x/%#x/%#x", token, value, update.ManaCur, update.ManaPrev)
			}
			record("protect-mana")
			update.Player = players[6]
		},
		SetHealthMaximum: func(got *Object) {
			if got != unit || update.TrapSpells != [5]uint32{} || update.TrapSpellsCnt != 0xa1b2c300 {
				t.Fatalf("health max state = %p/%v/%#x", got, update.TrapSpells, update.TrapSpellsCnt)
			}
			record("health-max")
		},
		SetPlayerState: func(got *Object, state PlayerState) {
			if got != unit || state != PlayerState13 || unit.Field541 != 0 || uint32(unit.ObjFlags) != playerResetObjectFlagMask4EFF10 {
				t.Fatalf("state callback = %p/%d/%d/%#x", got, state, unit.Field541, unit.ObjFlags)
			}
			record("state")
		},
		ClearBuffs:         func(*Object) { record("clear-buffs") },
		CancelSpells:       func(*Object) { record("cancel-spells") },
		RemovePoison:       func(*Object) { record("remove-poison") },
		ResetPlayerRuntime: func(*Object) { record("reset-runtime") },
		ReportTotalHealth: func(index uint8, got *Object) {
			if got != unit || index != players[6].PlayerInd {
				t.Fatalf("health report = %d/%p", index, got)
			}
			record("report-health")
			update.Player = players[7]
		},
		ReportTotalMana: func(index uint8, got *Object) {
			if got != unit || index != players[7].PlayerInd {
				t.Fatalf("mana report = %d/%p", index, got)
			}
			record("report-mana")
			update.Player = players[8]
		},
	})

	wantEvents := []string{
		"beast", "spells", "cancel-abilities", "read-values", "warrior",
		"protect-mana", "health-max", "state", "clear-buffs", "cancel-spells",
		"remove-poison", "reset-runtime", "report-health", "report-mana",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if got != playerResetResult4EFF10 || unit.Obj130 != nil {
		t.Fatalf("result/Obj130 = %#x/%p", uint32(got), unit.Obj130)
	}
	if players[8].field3664 != playerResetMarker4EFF10 || players[8].field3660 != playerResetMarker4EFF10 {
		t.Fatalf("markers = %#x/%#x", players[8].field3664, players[8].field3660)
	}
}

func TestPlayerReset4EFF10NativeHasNoEntryPointerGuards(t *testing.T) {
	for _, unit := range []*Object{nil, &Object{}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("unit %p did not fault", unit)
				}
			}()
			playerResetNative4EFF10(unit, PlayerResetRuntime4EFF10{})
		}()
	}
}
