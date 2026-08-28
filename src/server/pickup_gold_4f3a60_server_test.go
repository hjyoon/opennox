package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultPickupGoldNativeDeps4F3A60() pickupGoldNativeDeps4F3A60 {
	return pickupGoldNativeDeps4F3A60{
		protectGold:     func(uint32, int32) {},
		delayedDelete:   func(*Object) {},
		loadString:      func(string, string, int) string { return "" },
		sendLineMessage: func(*Object, string, uint32) {},
		audio:           func(uint32, *Object, int32, uint32) {},
		defaultPickup:   func(*Object, *Object, int32, int32) int32 { return 0 },
	}
}

func TestPickupGold4F3A60NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantInitData := uintptr(692)
	wantUpdateData := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantGold := uintptr(2164)
	wantToken := uintptr(4588)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantInitData = 760
		wantUpdateData = 872
		wantUpdateSize = 640
		wantUpdatePlayer = 320
		wantPlayerSize = 6160
		wantGold = 2168
		wantToken = 5892
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"GoldInitData size", unsafe.Sizeof(GoldInitData{}), 4},
		{"GoldInitData.Amount", unsafe.Offsetof(GoldInitData{}.Amount), 0},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.GoldVal", unsafe.Offsetof(Player{}.GoldVal), wantGold},
		{"Player.ProtPlayerGold", unsafe.Offsetof(Player{}.ProtPlayerGold), wantToken},
		{"callback result width", unsafe.Sizeof(int32(0)), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerAddGoldNative4FA590WrapsAndProtects(t *testing.T) {
	player := &Player{GoldVal: math.MaxUint32, ProtPlayerGold: 0xfedcba98}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{UpdateData: unsafe.Pointer(update)}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(update)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(player)) <= math.MaxUint32) {
		t.Fatalf("native pointer test did not allocate above 32 bits: owner=%p update=%p player=%p", owner, update, player)
	}

	calls := 0
	playerAddGoldNative4FA590(owner, 0x80000001, func(token uint32, delta int32) {
		calls++
		if token != 0xfedcba98 || uint32(delta) != 0x80000001 || player.GoldVal != 0x80000000 {
			t.Fatalf("protection state = token %#x delta %#x gold %#x", token, uint32(delta), player.GoldVal)
		}
	})
	if calls != 1 || player.GoldVal != 0x80000000 {
		t.Fatalf("calls/gold = %d/%#x, want 1/0x80000000", calls, player.GoldVal)
	}
}

func TestPickupGoldNative4F3A60UsesCachedDataAndNativeGold(t *testing.T) {
	player := &Player{GoldVal: 5, ProtPlayerGold: 0x12345678}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	cached := &GoldInitData{Amount: 7}
	replacement := &GoldInitData{Amount: 99}
	item := &Object{InitData: unsafe.Pointer(cached)}
	if unsafe.Sizeof(uintptr(0)) == 8 &&
		(uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(item)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(update)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(player)) <= math.MaxUint32 ||
			uintptr(unsafe.Pointer(cached)) <= math.MaxUint32) {
		t.Fatalf("native pointer test did not allocate above 32 bits: owner=%p item=%p update=%p player=%p data=%p", owner, item, update, player, cached)
	}

	deps := defaultPickupGoldNativeDeps4F3A60()
	var events []string
	deps.protectGold = func(token uint32, delta int32) {
		events = append(events, "protect")
		if token != 0x12345678 || delta != 7 || player.GoldVal != 12 {
			t.Fatalf("protect state = %#x/%d/%d", token, delta, player.GoldVal)
		}
	}
	deps.delayedDelete = func(got *Object) {
		events = append(events, "delete")
		if got != item {
			t.Fatalf("deleted item = %p, want %p", got, item)
		}
		item.InitData = unsafe.Pointer(replacement)
		cached.Amount = 11
	}
	deps.loadString = func(key, path string, line int) string {
		events = append(events, "string")
		if key != pickupGoldMessageKey4F3A60 || path != pickupGoldMessagePath4F3A60 || line != 709 {
			t.Fatalf("string provenance = %q/%q/%d", key, path, line)
		}
		return "gold %d"
	}
	deps.sendLineMessage = func(got *Object, message string, amount uint32) {
		events = append(events, "line")
		if got != owner || message != "gold %d" || amount != 11 {
			t.Fatalf("line args = %p/%q/%d", got, message, amount)
		}
	}
	deps.audio = func(id uint32, got *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 307 || got != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%#x", id, got, kind, code)
		}
	}
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
		t.Fatal("Player path called DefaultPickup")
		return 0
	}

	if got := pickupGoldNative4F3A60(owner, item, math.MinInt32, math.MaxInt32, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"protect", "delete", "string", "line", "audio"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if player.GoldVal != 12 || item.InitData != unsafe.Pointer(replacement) {
		t.Fatalf("final gold/data = %d/%p", player.GoldVal, item.InitData)
	}
}

func TestPickupGoldNative4F3A60NonPlayerForwardsExactResult(t *testing.T) {
	owner := &Object{ObjClass: object.ClassMonster}
	item := &Object{}
	for _, result := range []int32{0, 1, 2, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(result)), func(t *testing.T) {
			deps := defaultPickupGoldNativeDeps4F3A60()
			var events []string
			deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
				events = append(events, "default")
				if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
					t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
				}
				return result
			}
			deps.audio = func(id uint32, got *Object, kind int32, code uint32) {
				events = append(events, "audio")
				if id != 307 || got != owner || kind != 0 || code != 0 {
					t.Fatalf("audio args = %d/%p/%d/%#x", id, got, kind, code)
				}
			}
			if got := pickupGoldNative4F3A60(owner, item, math.MinInt32, math.MaxInt32, deps); got != result {
				t.Fatalf("result = %d, want %d", got, result)
			}
			want := []string{"default"}
			if result != 0 {
				want = append(want, "audio")
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestPickupGoldNative4F3A60PreservesNilFaults(t *testing.T) {
	t.Run("owner", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil owner did not preserve the class-load fault")
			}
		}()
		pickupGoldNative4F3A60(nil, nil, 0, 0, defaultPickupGoldNativeDeps4F3A60())
	})

	t.Run("gold data", func(t *testing.T) {
		owner := &Object{ObjClass: object.ClassPlayer}
		item := &Object{}
		calls := 0
		deps := defaultPickupGoldNativeDeps4F3A60()
		deps.protectGold = func(uint32, int32) { calls++ }
		defer func() {
			if recover() == nil || calls != 0 {
				t.Fatalf("nil GoldInitData recover/calls = %d, want panic/0", calls)
			}
		}()
		pickupGoldNative4F3A60(owner, item, 0, 0, deps)
	})
}
