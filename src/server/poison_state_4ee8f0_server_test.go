package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func defaultPoisonStateNativeDeps4EE8F0() poisonStateNativeDeps4EE8F0 {
	return poisonStateNativeDeps4EE8F0{
		needPlayerStatus:  func(*Player, uint32) {},
		unsetPlayerStatus: func(*Player, uint32) {},
		priorityMessage:   func(*Object, string, uint8) {},
		gameFlag:          func(uint32) int32 { return 0 },
		playerByIndex:     func(int32) *Player { return nil },
		reportPoison:      func(*Object, *Object, int32) {},
		frame:             func() uint32 { return 0 },
	}
}

func TestPoisonState4EE8F0NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantSubClass := uintptr(12)
	wantPoison := uintptr(540)
	wantTimer := uintptr(542)
	wantHealth := uintptr(556)
	wantOwner := uintptr(508)
	wantUpdate := uintptr(748)
	wantPlayerInUpdate := uintptr(276)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantSubClass = 16
		wantPoison = 600
		wantTimer = 602
		wantHealth = 616
		wantOwner = 552
		wantUpdate = 872
		wantPlayerInUpdate = 336
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.Poison540", unsafe.Offsetof(Object{}.Poison540), wantPoison},
		{"Object.Field542", unsafe.Offsetof(Object{}.Field542), wantTimer},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"HealthData.Field16", unsafe.Offsetof(HealthData{}.Field16), 16},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerInUpdate},
		{"Player.PlayerUnit", unsafe.Offsetof(Player{}.PlayerUnit), 2056},
		{"poison timer width", unsafe.Sizeof(Object{}.Field542), 2},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestUpdatePoisonNative4EE8F0BindsPlayerEffects(t *testing.T) {
	health := &HealthData{Field16: 77}
	player := new(Player)
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		Poison540:  3,
		HealthData: health,
		UpdateData: unsafe.Pointer(update),
	}
	var events []string
	deps := defaultPoisonStateNativeDeps4EE8F0()
	deps.unsetPlayerStatus = func(got *Player, status uint32) {
		events = append(events, "unset")
		if got != player || status != 1024 {
			t.Fatalf("unset args = (%p,%d), want (%p,1024)", got, status, player)
		}
	}
	deps.priorityMessage = func(got *Object, message string, value uint8) {
		events = append(events, "message")
		if got != unit || message != poisonClearFadeMessage4EE8F0 || value != 0 {
			t.Fatalf("message args = (%p,%q,%d)", got, message, value)
		}
	}
	updatePoisonNative4EE8F0(unit, 3, deps)
	if !reflect.DeepEqual(events, []string{"unset", "message"}) {
		t.Fatalf("events = %q, want [unset message]", events)
	}
	if unit.Poison540 != 0 || health.Field16 != 0 {
		t.Fatalf("poison/frame = %d/%d, want 0/0", unit.Poison540, health.Field16)
	}
}

func TestUpdatePoisonNative4EE8F0SignedBytePathDefersServices(t *testing.T) {
	unit := &Object{Poison540: 5}
	deps := defaultPoisonStateNativeDeps4EE8F0()
	deps.gameFlag = func(uint32) int32 {
		t.Fatal("byte subtraction reached clear services")
		return 0
	}
	updatePoisonNative4EE8F0(unit, -1, deps)
	if unit.Poison540 != 6 {
		t.Fatalf("poison = %d, want 6", unit.Poison540)
	}
}

func TestRemovePoisonNative4EE9D0BindsOwnerReport(t *testing.T) {
	owner := new(Object)
	unit := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(poisonClearOwnedMonsterLow4EE8F0),
		Poison540:   1,
		ObjOwner:    owner,
	}
	reports := 0
	deps := defaultPoisonStateNativeDeps4EE8F0()
	deps.reportPoison = func(receiver, got *Object, active int32) {
		reports++
		if receiver != owner || got != unit || active != 0 {
			t.Fatalf("report args = (%p,%p,%d), want (%p,%p,0)", receiver, got, active, owner, unit)
		}
	}
	removePoisonNative4EE9D0(unit, deps)
	if reports != 1 || unit.Poison540 != 0 {
		t.Fatalf("reports/poison = %d/%d, want 1/0", reports, unit.Poison540)
	}
}

func TestSetPoisonNative4EEA90BindsWholeValueAndPlayerStatus(t *testing.T) {
	health := new(HealthData)
	player := new(Player)
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		HealthData: health,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
	}
	var events []string
	deps := defaultPoisonStateNativeDeps4EE8F0()
	deps.frame = func() uint32 {
		events = append(events, "frame")
		return 55
	}
	deps.needPlayerStatus = func(got *Player, status uint32) {
		events = append(events, "need")
		if got != player || status != 1024 {
			t.Fatalf("need args = (%p,%d), want (%p,1024)", got, status, player)
		}
	}
	setPoisonNative4EEA90(unit, 256, deps)
	if !reflect.DeepEqual(events, []string{"frame", "need"}) {
		t.Fatalf("events = %q, want [frame need]", events)
	}
	if unit.Poison540 != 0 || unit.Field542 != 1000 || health.Field16 != 55 {
		t.Fatalf("poison/timer/frame = %d/%d/%d, want 0/1000/55", unit.Poison540, unit.Field542, health.Field16)
	}
}

func TestSetPoisonNative4EEA90BindsQuestPlayerUnit(t *testing.T) {
	receiver := new(Object)
	questPlayer := &Player{PlayerUnit: receiver}
	unit := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(poisonClearQuestMonsterLow4EE8F0),
	}
	deps := defaultPoisonStateNativeDeps4EE8F0()
	deps.gameFlag = func(flag uint32) int32 {
		if flag != 2048 {
			t.Fatalf("game flag = %d, want 2048", flag)
		}
		return 1
	}
	deps.playerByIndex = func(index int32) *Player {
		if index != 31 {
			t.Fatalf("player index = %d, want 31", index)
		}
		return questPlayer
	}
	deps.reportPoison = func(gotReceiver, gotUnit *Object, active int32) {
		if gotReceiver != receiver || gotUnit != unit || active != 1 {
			t.Fatalf("report args = (%p,%p,%d), want (%p,%p,1)", gotReceiver, gotUnit, active, receiver, unit)
		}
	}
	setPoisonNative4EEA90(unit, 1, deps)
	if unit.Poison540 != 1 || unit.Field542 != 1000 {
		t.Fatalf("poison/timer = %d/%d, want 1/1000", unit.Poison540, unit.Field542)
	}
}

func TestPoisonState4EE8F0ServerMethodsUseNativeFields(t *testing.T) {
	s := new(Server)
	s.SetFrame(88)
	unit := &Object{Poison540: 4, HealthData: new(HealthData)}
	s.UpdatePoison4EE8F0(unit, 1)
	if unit.Poison540 != 3 {
		t.Fatalf("updated poison = %d, want 3", unit.Poison540)
	}
	s.RemovePoison4EE9D0(unit)
	if unit.Poison540 != 0 {
		t.Fatalf("removed poison = %d, want 0", unit.Poison540)
	}
	s.SetPoison4EEA90(unit, 2)
	if unit.Poison540 != 2 || unit.Field542 != 1000 || unit.HealthData.Field16 != 88 {
		t.Fatalf("set poison/timer/frame = %d/%d/%d, want 2/1000/88", unit.Poison540, unit.Field542, unit.HealthData.Field16)
	}
}

func TestPoisonState4EE8F0ServerUsesNativePlayerStatus(t *testing.T) {
	previous := noxflags.GetGame()
	noxflags.ResetGame()
	defer func() {
		noxflags.ResetGame()
		noxflags.SetGame(previous)
	}()

	s := new(Server)
	player := new(Player)
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		HealthData: new(HealthData),
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
	}
	s.SetPoison4EEA90(unit, 3)
	if player.Field3680&poisonClearPlayerStatus4EE8F0 == 0 {
		t.Fatalf("player status = %#x, want poison bit", player.Field3680)
	}
	s.RemovePoison4EE9D0(unit)
	if player.Field3680&poisonClearPlayerStatus4EE8F0 != 0 {
		t.Fatalf("player status = %#x, want poison bit cleared", player.Field3680)
	}
}
