package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
)

func defaultAbilityRewardNativeDeps4FB9C0() abilityRewardNativeDeps4FB9C0 {
	return abilityRewardNativeDeps4FB9C0{
		loadString:       func(string, string, int) string { return "" },
		sendLineMessage:  func(*Object, string) {},
		primaryMessage:   func(*Object, string, uint8) {},
		awardProtection:  func(uint32, int32, int32) {},
		reportAbility:    func(*Object, int32, int32) {},
		gameFlagsCheck:   func(uint32) int32 { return 0 },
		rewardNotify:     func(*Object, int32, *Object, int32) {},
		checkPlayerState: func(*Object) int32 { return 0 },
		firstPlayerUnit:  func() *Object { return nil },
		nextPlayerUnit:   func(*Object) *Object { return nil },
	}
}

func TestAbilityReward4FB9C0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantObjectNetCode := uintptr(36)
	wantObjectUseData := uintptr(736)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerIndex := uintptr(2064)
	wantPlayerInfo := uintptr(2185)
	wantSpellLevels := uintptr(3696)
	wantProtection := uintptr(4636)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantObjectNetCode = 40
		wantObjectUseData = 848
		wantObjectUpdate = 872
		wantUpdateSize = 656
		wantUpdatePlayer = 336
		wantPlayerSize = 6160
		wantPlayerIndex = 2068
		wantPlayerInfo = 2189
		wantSpellLevels = 4992
		wantProtection = 5940
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantObjectNetCode},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantObjectUseData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.info", unsafe.Offsetof(Player{}.info), wantPlayerInfo},
		{"Player class byte", unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass), wantPlayerInfo + 66},
		{"Player.SpellLvl", unsafe.Offsetof(Player{}.SpellLvl), wantSpellLevels},
		{"Player.Prot4636", unsafe.Offsetof(Player{}.Prot4636), wantProtection},
		{"AbilityRewardUseData size", unsafe.Sizeof(AbilityRewardUseData{}), 1},
		{"AbilityRewardUseData.Ability", unsafe.Offsetof(AbilityRewardUseData{}.Ability), 0},
		{"ability level width", unsafe.Sizeof(Player{}.SpellLvl[0]), 4},
		{"ability level count", uintptr(len(Player{}.SpellLvl)), 137},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestAbilityRewardServNative4FB9C0PreservesPointersAndScalars(t *testing.T) {
	player := &Player{Prot4636: 0x89abcdef}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 0xf4, NetCode: 0x10203040, UpdateData: unsafe.Pointer(update)}
	other := &Object{ObjClass: 4}

	type notification struct {
		recipient *Object
		kind      int32
		source    *Object
		ability   int32
	}
	var (
		protectedToken uint32
		protectedArgs  [2]int32
		reportUnit     *Object
		reportArgs     [2]int32
		notifications  []notification
	)
	deps := defaultAbilityRewardNativeDeps4FB9C0()
	deps.awardProtection = func(token uint32, ability, level int32) {
		protectedToken = token
		protectedArgs = [2]int32{ability, level}
	}
	deps.reportAbility = func(gotUnit *Object, ability, rewardArg int32) {
		reportUnit = gotUnit
		reportArgs = [2]int32{ability, rewardArg}
	}
	deps.gameFlagsCheck = func(mask uint32) int32 {
		if mask != abilityRewardQuestFlag4FB9C0 {
			t.Fatalf("flag mask = %#x, want %#x", mask, abilityRewardQuestFlag4FB9C0)
		}
		return math.MinInt32
	}
	deps.rewardNotify = func(recipient *Object, kind int32, source *Object, ability int32) {
		notifications = append(notifications, notification{recipient, kind, source, ability})
	}
	deps.firstPlayerUnit = func() *Object { return unit }
	deps.nextPlayerUnit = func(got *Object) *Object {
		switch got {
		case unit:
			return other
		case other:
			return nil
		default:
			t.Fatalf("unexpected iterator pointer %p", got)
			return nil
		}
	}

	const rewardArg = int32(-0x1234567)
	if got := abilityRewardServNative4FB9C0(unit, 5, rewardArg, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if player.SpellLvl[5] != 5 {
		t.Fatalf("ability level = %d, want 5", player.SpellLvl[5])
	}
	if protectedToken != 0x89abcdef || protectedArgs != [2]int32{5, 5} {
		t.Fatalf("protection = %#x/%v, want 0x89abcdef/[5 5]", protectedToken, protectedArgs)
	}
	if reportUnit != unit || reportArgs != [2]int32{5, rewardArg} {
		t.Fatalf("report = %p/%v, want %p/[5 %d]", reportUnit, reportArgs, unit, rewardArg)
	}
	wantNotifications := []notification{
		{unit, 2, unit, 5},
		{other, 2, unit, 5},
	}
	if !reflect.DeepEqual(notifications, wantNotifications) {
		t.Fatalf("notifications = %#v, want %#v", notifications, wantNotifications)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"player": uintptr(unsafe.Pointer(player)),
			"update": uintptr(unsafe.Pointer(update)),
			"unit":   uintptr(unsafe.Pointer(unit)),
			"other":  uintptr(unsafe.Pointer(other)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(other)
}

func TestAbilityRewardReportNative4D8060BuildsExactPacket(t *testing.T) {
	s := new(Server)
	player := &Player{PlayerInd: 0xfe}
	player.SpellLvl[3] = 0x102
	update := &PlayerUpdateData{Player: player}
	unit := &Object{ObjClass: 4, UpdateData: unsafe.Pointer(update)}

	var (
		gotRecipient int
		gotPacket    []byte
		gotRelated   *Object
		gotRemove    int
		gotSequence  int
	)
	s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
		gotRecipient = index
		gotPacket = append([]byte(nil), packet...)
		gotRelated = related
		gotRemove = remove
		gotSequence = sequence
		return math.MinInt32
	}

	if got := abilityRewardReportNative4D8060(s, unit, 3, -1); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, math.MinInt32)
	}
	want := []byte{byte(netmsg.MSG_REPORT_ABILITY_AWARD), 3, 0x82}
	if gotRecipient != 0xfe || !reflect.DeepEqual(gotPacket, want) || gotRelated != nil || gotRemove != 1 || gotSequence != 1 {
		t.Fatalf("packet = recipient %d bytes %v related %p remove %d sequence %d", gotRecipient, gotPacket, gotRelated, gotRemove, gotSequence)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}
