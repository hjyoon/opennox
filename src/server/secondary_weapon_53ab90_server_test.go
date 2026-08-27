package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestSecondaryWeaponReport53AB90NativePointersAndSidecar(t *testing.T) {
	srv := new(Server)
	owner, freeOwner := alloc.New(Object{})
	defer freeOwner()
	item, freeItem := alloc.New(Object{})
	defer freeItem()
	update, freeUpdate := alloc.New(PlayerUpdateData{})
	defer freeUpdate()
	playerValue, freePlayer := alloc.New(Player{})
	defer freePlayer()
	owner.ObjClass = object.ClassPlayer
	update.Field27 = 0xA5A5A5A5
	playerValue.PlayerInd = 6
	playerValue.Info().SetPlayerClass(player.Warrior)
	update.Player = playerValue
	owner.UpdateData = unsafe.Pointer(update)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(item), unsafe.Pointer(update), unsafe.Pointer(playerValue)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}
	var calls []string
	srv.SecondaryWeaponReport53AB90(owner, item,
		func(got *Object, class player.Class) bool {
			if got != item || class != player.Warrior {
				t.Fatalf("class callback = (%p,%v), want (%p,%v)", got, class, item, player.Warrior)
			}
			calls = append(calls, "class")
			return true
		},
		func(gotOwner, gotItem *Object) bool {
			if gotOwner != owner || gotItem != item {
				t.Fatalf("strength callback = (%p,%p), want (%p,%p)", gotOwner, gotItem, owner, item)
			}
			calls = append(calls, "strength")
			return true
		},
		func(byte) { t.Fatal("valid item cleared client selection") },
	)
	if len(calls) != 2 || calls[0] != "class" || calls[1] != "strength" {
		t.Fatalf("callbacks = %#v, want class then strength", calls)
	}
	if got := srv.SecondaryWeapon53AB90(owner); got != item {
		t.Fatalf("native secondary weapon = %p, want %p", got, item)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && update.Field27 != 0xA5A5A5A5 {
		t.Fatalf("64-bit ABI32 slot = %#x, want unchanged sentinel", update.Field27)
	}

	srv.SecondaryWeaponReport53AB90(owner, nil, nil, nil, func(byte) { t.Fatal("nil item cleared client selection") })
	if got := srv.SecondaryWeapon53AB90(owner); got != nil {
		t.Fatalf("cleared secondary weapon = %p, want nil", got)
	}
	if unsafe.Sizeof(uintptr(0)) == 4 && update.Field27 != 0 {
		t.Fatalf("32-bit ABI slot after clear = %#x, want zero", update.Field27)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
	runtime.KeepAlive(update)
	runtime.KeepAlive(playerValue)
}

func TestSecondaryWeaponReport53AB90NativeInvalidItemClearsBeforeStore(t *testing.T) {
	srv := new(Server)
	owner := &Object{ObjClass: object.ClassPlayer}
	item := new(Object)
	update := new(PlayerUpdateData)
	playerValue := new(Player)
	playerValue.PlayerInd = 9
	playerValue.Info().SetPlayerClass(player.Wizard)
	update.Player = playerValue
	owner.UpdateData = unsafe.Pointer(update)

	cleared := false
	srv.SecondaryWeaponReport53AB90(owner, item,
		func(*Object, player.Class) bool { return false },
		func(*Object, *Object) bool {
			t.Fatal("class rejection did not short-circuit strength")
			return false
		},
		func(index byte) {
			if index != 9 || srv.SecondaryWeapon53AB90(owner) != nil {
				t.Fatalf("clear observed index/state = %d/%p, want 9/nil", index, srv.SecondaryWeapon53AB90(owner))
			}
			cleared = true
		},
	)
	if !cleared || srv.SecondaryWeapon53AB90(owner) != item {
		t.Fatalf("clear/store = %t/%p, want true/%p", cleared, srv.SecondaryWeapon53AB90(owner), item)
	}
}
