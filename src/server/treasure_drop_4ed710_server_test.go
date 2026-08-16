package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultTreasureDropNativeDeps4ED710() treasureDropNativeDeps4ED710 {
	return treasureDropNativeDeps4ED710{
		defaultDrop: func(*Object, *Object, *types.Pointf) int32 { return 0 },
		gameFlag:    func(uint32) int32 { return 0 },
		treasureMax: func() uint32 { return 0 },
		report:      func(*Object) {},
		audio:       func(uint32, *Object, int32, uint32) {},
	}
}

func TestTreasureDrop4ED710NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantObjectUpdate := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantCount := uintptr(2152)
	wantMaximum := uintptr(2156)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantObjectUpdate = 872
		wantUpdateSize = 640
		wantUpdatePlayer = 320
		wantPlayerSize = 6160
		wantCount = 2156
		wantMaximum = 2160
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.Field2152", unsafe.Offsetof(Player{}.Field2152), wantCount},
		{"Player.Field2156", unsafe.Offsetof(Player{}.Field2156), wantMaximum},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestTreasureDropNative4ED710BindsFieldsAndServices(t *testing.T) {
	player := &Player{Field2152: 3, Field2156: 99}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	treasure := &Object{}
	point := &types.Pointf{X: 4.5, Y: -7.25}
	events := make([]string, 0, 5)
	deps := defaultTreasureDropNativeDeps4ED710()
	deps.defaultDrop = func(gotOwner, gotTreasure *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotTreasure != treasure || gotPoint != point {
			t.Fatalf("default args = %p/%p/%p", gotOwner, gotTreasure, gotPoint)
		}
		return -1
	}
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, "game")
		if flag != 64 {
			t.Fatalf("game flag = %d, want 64", flag)
		}
		return -1
	}
	deps.treasureMax = func() uint32 {
		events = append(events, "maximum")
		return 0xf1234567
	}
	deps.report = func(gotOwner *Object) {
		events = append(events, "report")
		if gotOwner != owner {
			t.Fatalf("report owner = %p, want %p", gotOwner, owner)
		}
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 308 || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio = %d/%p/%d/%d", id, gotOwner, kind, code)
		}
	}

	if got := treasureDropNative4ED710(owner, treasure, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if player.Field2152 != 2 || player.Field2156 != 0xf1234567 {
		t.Fatalf("count/max = %#x/%#x", player.Field2152, player.Field2156)
	}
	if !reflect.DeepEqual(events, []string{"default", "game", "maximum", "report", "audio"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestTreasureDropNative4ED710LoadsUpdateAfterGameAndReloadsCachedPlayer(t *testing.T) {
	oldPlayer := &Player{Field2152: 100}
	oldUpdate := &PlayerUpdateData{Player: oldPlayer}
	firstPlayer := &Player{Field2152: 9}
	cachedUpdate := &PlayerUpdateData{Player: firstPlayer}
	secondPlayer := &Player{Field2152: 77}
	replacementUpdate := &PlayerUpdateData{Player: &Player{Field2152: 55}}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(oldUpdate)}
	deps := defaultTreasureDropNativeDeps4ED710()
	deps.defaultDrop = func(*Object, *Object, *types.Pointf) int32 { return 1 }
	deps.gameFlag = func(uint32) int32 {
		owner.UpdateData = unsafe.Pointer(cachedUpdate)
		return 1
	}
	deps.treasureMax = func() uint32 {
		owner.UpdateData = unsafe.Pointer(replacementUpdate)
		cachedUpdate.Player = secondPlayer
		return 12
	}

	if got := treasureDropNative4ED710(owner, &Object{}, &types.Pointf{}, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if oldPlayer.Field2152 != 100 || firstPlayer.Field2152 != 8 {
		t.Fatalf("old/first count = %d/%d, want 100/8", oldPlayer.Field2152, firstPlayer.Field2152)
	}
	if secondPlayer.Field2152 != 77 || secondPlayer.Field2156 != 12 {
		t.Fatalf("second count/max = %d/%d, want 77/12", secondPlayer.Field2152, secondPlayer.Field2156)
	}
	if replacementUpdate.Player.Field2156 != 0 {
		t.Fatalf("replacement owner update was used: max = %d", replacementUpdate.Player.Field2156)
	}
}

func TestTreasureDropNative4ED710DefaultFailureAcceptsNilPointers(t *testing.T) {
	deps := defaultTreasureDropNativeDeps4ED710()
	deps.defaultDrop = func(owner, treasure *Object, point *types.Pointf) int32 {
		if owner != nil || treasure != nil || point != nil {
			t.Fatalf("default args = %p/%p/%p", owner, treasure, point)
		}
		return 0
	}
	if got := treasureDropNative4ED710(nil, nil, nil, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}
