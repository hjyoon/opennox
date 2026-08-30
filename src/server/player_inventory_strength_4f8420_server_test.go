package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerInventoryStrength4F8420NativeLayout(t *testing.T) {
	wantFlags := uintptr(16)
	wantNext := uintptr(496)
	wantFirst := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlags = 20
		wantNext = 528
		wantFirst = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerInventoryStrengthNative4F8420PreservesPointersAndLiveLinks(t *testing.T) {
	player := &Object{}
	item := &Object{ObjFlags: object.FlagEquipped}
	stale := &Object{TypeInd: 1, ObjFlags: object.FlagEquipped}
	replacement := &Object{TypeInd: 2, ObjFlags: object.FlagEquipped}
	player.InvFirstItem = item
	item.InvNextItem = stale
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(player),
			unsafe.Pointer(item),
			unsafe.Pointer(stale),
			unsafe.Pointer(replacement),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	var events []string
	playerInventoryStrengthNative4F8420(player, playerInventoryStrengthNativeDeps4F8420{
		checkStrength: func(gotPlayer, gotItem *Object) int32 {
			if gotPlayer != player {
				t.Fatalf("strength player = %p, want %p", gotPlayer, player)
			}
			switch gotItem {
			case item:
				events = append(events, "strength:item")
				item.InvNextItem = replacement
				return 0
			case replacement:
				events = append(events, "strength:replacement")
				return 1
			case stale:
				t.Fatal("strength visited stale pre-callback link")
			default:
				t.Fatalf("strength item = %p", gotItem)
			}
			return 1
		},
		forceDrop: func(gotPlayer, gotItem *Object) int32 {
			events = append(events, "drop:item")
			if gotPlayer != player || gotItem != item {
				t.Fatalf("drop = %p/%p, want %p/%p", gotPlayer, gotItem, player, item)
			}
			return math.MinInt32
		},
	})
	if want := []string{"strength:item", "drop:item", "strength:replacement"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(item)
	runtime.KeepAlive(stale)
	runtime.KeepAlive(replacement)
}

func TestServerPlayerInventoryStrength4F8420UsesRestoredStrength(t *testing.T) {
	s := new(Server)
	weapon := &Modifier{TypeInd: 0x1234, ReqStrength60: 37}
	armor := &Modifier{TypeInd: 0x5678, ReqStrength60: 38}
	s.Modif.Dword_5d4594_251600 = weapon
	s.Modif.Dword_5d4594_251608 = armor

	playerData := &Player{}
	playerData.Info().SetField2239(37)
	player := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: playerData}),
	}
	weakArmor := &Object{
		ObjClass: object.ClassArmor,
		ObjFlags: object.FlagEquipped,
		TypeInd:  0x5678,
	}
	strongWeapon := &Object{
		ObjClass: object.ClassWeapon,
		ObjFlags: object.FlagEquipped,
		TypeInd:  0x1234,
	}
	unequipped := &Object{
		ObjClass: object.ClassArmor,
		TypeInd:  0x5678,
	}
	player.InvFirstItem = weakArmor
	weakArmor.InvNextItem = strongWeapon
	strongWeapon.InvNextItem = unequipped

	var dropped []*Object
	s.PlayerInventoryStrength4F8420(player, PlayerInventoryStrengthRuntime4F8420{
		ForceDrop: func(gotPlayer, item *Object) int32 {
			if gotPlayer != player {
				t.Fatalf("drop player = %p, want %p", gotPlayer, player)
			}
			dropped = append(dropped, item)
			return math.MinInt32
		},
	})
	if !reflect.DeepEqual(dropped, []*Object{weakArmor}) {
		t.Fatalf("dropped = %p, want [%p]", dropped, weakArmor)
	}
}
