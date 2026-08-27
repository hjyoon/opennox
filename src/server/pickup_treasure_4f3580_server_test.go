package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

func defaultPickupTreasureNativeDeps4F3580() pickupTreasureNativeDeps4F3580 {
	return pickupTreasureNativeDeps4F3580{
		defaultPickup:      func(*Object, *Object, int32, int32) int32 { return 0 },
		gameFlag:           func(uint32) int32 { return 0 },
		audio:              func(uint32, *Object, int32, uint32) {},
		treasureMax:        func() uint32 { return 0 },
		report:             func(*Object) {},
		findTeam:           func(uint8) *Team { return nil },
		firstPlayer:        func() *Object { return nil },
		nextPlayer:         func(*Object) *Object { return nil },
		setGameFlags:       func(uint32) {},
		changeScore:        func(*Object, int32) {},
		reportLesson:       func(*Object) {},
		incrementElimDeath: func(*Object) {},
	}
}

func TestPickupTreasure4F3580NativeLayoutsAndConstants(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantObjectTeam := uintptr(48)
	wantUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantDeaths := uintptr(2140)
	wantCount := uintptr(2152)
	wantMaximum := uintptr(2156)
	wantTeamSize := uintptr(80)
	if ptrSize == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantObjectTeam = 52
		wantUpdate = 872
		wantUpdateSize = 640
		wantUpdatePlayer = 320
		wantPlayerSize = 6160
		wantDeaths = 2144
		wantCount = 2156
		wantMaximum = 2160
		wantTeamSize = 88
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.TeamVal", unsafe.Offsetof(Object{}.TeamVal), wantObjectTeam},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"ObjectTeam size", unsafe.Sizeof(ObjectTeam{}), 8},
		{"ObjectTeam.ID", unsafe.Offsetof(ObjectTeam{}.ID), 4},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.Field2140", unsafe.Offsetof(Player{}.Field2140), wantDeaths},
		{"Player.Field2152", unsafe.Offsetof(Player{}.Field2152), wantCount},
		{"Player.Field2156", unsafe.Offsetof(Player{}.Field2156), wantMaximum},
		{"Team size", unsafe.Sizeof(Team{}), wantTeamSize},
		{"Team.IDVal", unsafe.Offsetof(Team{}.IDVal), 57},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	if got := uint32(sound.SoundTreasurePickup); got != pickupTreasureAudio4F3580 {
		t.Errorf("SoundTreasurePickup = %d, want %d", got, pickupTreasureAudio4F3580)
	}
	if uint32(noxflags.GameModeFlagBall) != pickupTreasureGameMode4F3580 {
		t.Errorf("Scavenger flag = %08x", uint32(noxflags.GameModeFlagBall))
	}
	if uint32(noxflags.GameFlag4) != pickupTreasureCompleteFlag4F3580 {
		t.Errorf("completion flag = %08x", uint32(noxflags.GameFlag4))
	}
}

func TestPickupTreasureNative4F3580SoloBindsFieldsAndEffects(t *testing.T) {
	player := &Player{Field2152: 4}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	item := &Object{}
	other := &Object{}
	events := make([]string, 0, 12)
	maximumCall := 0
	deps := defaultPickupTreasureNativeDeps4F3580()
	deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
			t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
		}
		return math.MinInt32
	}
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, "game-flag")
		if flag != uint32(noxflags.GameModeFlagBall) {
			t.Fatalf("game flag = %08x", flag)
		}
		return math.MinInt32
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != uint32(sound.SoundTreasurePickup) || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	deps.treasureMax = func() uint32 {
		events = append(events, "maximum")
		maximumCall++
		return 5
	}
	deps.report = func(gotOwner *Object) {
		events = append(events, "report")
		if gotOwner != owner || player.Field2152 != 5 || player.Field2156 != 5 {
			t.Fatalf("report owner/counters = %p/%d/%d", gotOwner, player.Field2152, player.Field2156)
		}
	}
	deps.setGameFlags = func(flags uint32) {
		events = append(events, "set-flags")
		if flags != uint32(noxflags.GameFlag4) {
			t.Fatalf("set flags = %08x", flags)
		}
	}
	deps.changeScore = func(gotOwner *Object, value int32) {
		events = append(events, "change-score")
		if gotOwner != owner || value != 1 {
			t.Fatalf("score args = %p/%d", gotOwner, value)
		}
	}
	deps.reportLesson = func(obj *Object) {
		if obj == owner {
			events = append(events, "lesson-owner")
		} else if obj == other {
			events = append(events, "lesson-other")
		} else {
			t.Fatalf("unexpected lesson object %p", obj)
		}
	}
	first := true
	deps.firstPlayer = func() *Object {
		events = append(events, "first")
		return owner
	}
	deps.nextPlayer = func(obj *Object) *Object {
		events = append(events, "next")
		if first && obj == owner {
			first = false
			return other
		}
		if obj != other {
			t.Fatalf("next object = %p", obj)
		}
		return nil
	}
	deps.incrementElimDeath = func(obj *Object) {
		events = append(events, "death-other")
		if obj != other {
			t.Fatalf("death object = %p", obj)
		}
	}

	if got := pickupTreasureNative4F3580(owner, item, math.MinInt32, math.MaxInt32, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if player.Field2152 != 5 || player.Field2156 != 5 || maximumCall != 2 {
		t.Fatalf("counters/calls = %d/%d/%d", player.Field2152, player.Field2156, maximumCall)
	}
	want := []string{
		"default", "game-flag", "audio", "maximum", "report", "maximum",
		"set-flags", "change-score", "lesson-owner", "first", "next",
		"death-other", "lesson-other", "next",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPickupTreasureNative4F3580TeamUsesNativeLayoutsAndWraps(t *testing.T) {
	ownerPlayer := &Player{}
	owner := &Object{
		ObjClass:   object.ClassPlayer,
		TeamVal:    ObjectTeam{ID: 7},
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: ownerPlayer}),
	}
	item := &Object{}
	first := &Object{
		TeamVal:    ObjectTeam{ID: 7},
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{Field2152: math.MaxUint32}}),
	}
	second := &Object{
		TeamVal:    ObjectTeam{ID: 9},
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{Field2152: 99}}),
	}
	third := &Object{
		TeamVal:    ObjectTeam{ID: 7},
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{Field2152: 2}}),
	}
	team := &Team{IDVal: 7}
	deps := defaultPickupTreasureNativeDeps4F3580()
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 { return 1 }
	deps.gameFlag = func(uint32) int32 { return 1 }
	maxCall := 0
	deps.treasureMax = func() uint32 {
		maxCall++
		if maxCall == 1 {
			return 11
		}
		return 1
	}
	deps.findTeam = func(id uint8) *Team {
		if id != 7 {
			t.Fatalf("team lookup ID = %d", id)
		}
		return team
	}
	deps.firstPlayer = func() *Object { return first }
	deps.nextPlayer = func(obj *Object) *Object {
		switch obj {
		case first:
			return second
		case second:
			return third
		case third:
			return nil
		default:
			t.Fatalf("unexpected next object %p", obj)
			return nil
		}
	}
	set := 0
	deps.setGameFlags = func(flags uint32) {
		set++
		if flags != 8 {
			t.Fatalf("flags = %d", flags)
		}
	}
	if got := pickupTreasureNative4F3580(owner, item, 3, 4, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if ownerPlayer.Field2152 != 1 || ownerPlayer.Field2156 != 11 || maxCall != 2 || set != 1 {
		t.Fatalf("owner counters/max calls/set = %d/%d/%d/%d", ownerPlayer.Field2152, ownerPlayer.Field2156, maxCall, set)
	}
}

func TestPickupTreasureNative4F3580NilOwnerFaultsAfterDefault(t *testing.T) {
	deps := defaultPickupTreasureNativeDeps4F3580()
	called := false
	deps.defaultPickup = func(owner, item *Object, arg3, arg4 int32) int32 {
		called = true
		if owner != nil || item != nil || arg3 != 3 || arg4 != 4 {
			t.Fatalf("default args = %p/%p/%d/%d", owner, item, arg3, arg4)
		}
		return 1
	}
	defer func() {
		if recover() == nil || !called {
			t.Fatalf("nil owner fault/default = false/%v", called)
		}
	}()
	pickupTreasureNative4F3580(nil, nil, 3, 4, deps)
}

func TestPickupTreasure4F3580ServerBindingDefaultPickupBeforeModeGate(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := &Server{}
	owner := &Object{
		ObjClass:      object.ClassPlayer,
		CarryCapacity: 50,
		UpdateData:    unsafe.Pointer(&PlayerUpdateData{}),
	}
	item := &Object{
		TypeInd:  17,
		ObjClass: object.ClassPickup,
		ObjFlags: object.FlagActive,
		Weight:   3,
	}
	events := make([]string, 0, 2)
	runtime := PickupTreasureRuntime4F3580{
		DefaultPickup: PickupDefaultRuntime4F31E0{
			DeleteWorldObject: func(gotItem *Object) {
				events = append(events, "delete")
				if gotItem != item {
					t.Fatalf("deleted = %p, want %p", gotItem, item)
				}
				item.ObjFlags &^= object.FlagActive
			},
			InventoryPut: func(gotOwner, gotItem *Object, report int32) {
				events = append(events, "put")
				if gotOwner != owner || gotItem != item || report != -9 {
					t.Fatalf("put args = %p/%p/%d", gotOwner, gotItem, report)
				}
			},
		},
	}
	if got := s.PickupTreasure4F3580(owner, item, -9, 23, runtime); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(events, []string{"delete", "put"}) {
		t.Fatalf("events = %v", events)
	}
	if len(s.Audio.delayedObj) != 0 {
		t.Fatalf("audio queued outside Scavenger mode: %v", s.Audio.delayedObj)
	}
}
