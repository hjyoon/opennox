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

func defaultPickupTrapNativeDeps4F3510() pickupTrapNativeDeps4F3510 {
	return pickupTrapNativeDeps4F3510{
		hasOwner:      func(*Object, *Object) int32 { return 0 },
		defaultPickup: func(*Object, *Object, int32, int32) int32 { return 0 },
		audio:         func(uint32, *Object, int32, uint32) {},
	}
}

func TestPickupTrap4F3510NativeLayoutAndSoundIDs(t *testing.T) {
	wantSize := uintptr(780)
	wantClass := uintptr(8)
	wantNetCode := uintptr(36)
	wantOwner := uintptr(508)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantClass = 12
		wantNetCode = 40
		wantOwner = 552
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.ObjClass size", unsafe.Sizeof(Object{}.ObjClass), 4},
		{"Object.NetCode size", unsafe.Sizeof(Object{}.NetCode), 4},
		{"Object.ObjOwner size", unsafe.Sizeof(Object{}.ObjOwner), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	if got := uint32(sound.SoundTrapPickup); got != pickupTrapSuccessAudio4F3510 {
		t.Errorf("SoundTrapPickup = %d, want %d", got, pickupTrapSuccessAudio4F3510)
	}
	if got := uint32(sound.SoundNoCanDo); got != pickupTrapRejectAudio4F3510 {
		t.Errorf("SoundNoCanDo = %d, want %d", got, pickupTrapRejectAudio4F3510)
	}
}

func TestPickupTrapNative4F3510BindsOwnerChainFourArgsAndAudio(t *testing.T) {
	owner := &Object{}
	middle := &Object{ObjOwner: owner}
	item := &Object{ObjOwner: middle}
	events := make([]string, 0, 3)
	deps := defaultPickupTrapNativeDeps4F3510()
	deps.hasOwner = func(gotItem, gotOwner *Object) int32 {
		events = append(events, "has-owner")
		if gotItem != item || gotOwner != owner || !gotItem.HasOwner(gotOwner) {
			t.Fatalf("owner-chain args = %p/%p", gotItem, gotOwner)
		}
		return math.MinInt32
	}
	deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
			t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
		}
		return math.MinInt32
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != uint32(sound.SoundTrapPickup) || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	if got := pickupTrapNative4F3510(owner, item, math.MinInt32, math.MaxInt32, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if want := []string{"has-owner", "default", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPickupTrapNative4F3510RejectsWithLivePlayerFields(t *testing.T) {
	owner := &Object{ObjClass: object.ClassPlayer | object.Class(0x80000000), NetCode: 0xfedcba98}
	item := &Object{}
	deps := defaultPickupTrapNativeDeps4F3510()
	deps.hasOwner = func(gotItem, gotOwner *Object) int32 {
		if gotItem != item || gotOwner != owner {
			t.Fatalf("owner-chain args = %p/%p", gotItem, gotOwner)
		}
		return 0
	}
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
		t.Fatal("rejected trap called DefaultPickup")
		return 0
	}
	called := false
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		called = true
		if id != uint32(sound.SoundNoCanDo) || gotOwner != owner || kind != 2 || code != owner.NetCode {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	if got := pickupTrapNative4F3510(owner, item, 3, 4, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if !called {
		t.Fatal("rejected Player trap did not play NoCanDo")
	}
}

func TestPickupTrapNative4F3510NilItemRejectsButNilOwnerFaults(t *testing.T) {
	owner := &Object{ObjClass: object.ClassPlayer, NetCode: 7}
	deps := defaultPickupTrapNativeDeps4F3510()
	deps.hasOwner = func(item, gotOwner *Object) int32 {
		if item != nil || gotOwner != owner {
			t.Fatalf("nil-item owner-chain args = %p/%p", item, gotOwner)
		}
		return 0
	}
	called := false
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		called = true
		if id != uint32(sound.SoundNoCanDo) || gotOwner != owner || kind != 2 || code != 7 {
			t.Fatalf("audio args = %d/%p/%d/%d", id, gotOwner, kind, code)
		}
	}
	if got := pickupTrapNative4F3510(owner, nil, 3, 4, deps); got != 0 || !called {
		t.Fatalf("nil-item result/audio = %d/%v, want 0/true", got, called)
	}

	deps = pickupTrapServerDeps4F3510(&Server{}, PickupTrapRuntime4F3510{})
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner did not fault at live class load")
		}
	}()
	pickupTrapNative4F3510(nil, nil, 3, 4, deps)
}

func TestPickupTrap4F3510ServerBindingDefaultPickupAndQueuesAudio(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := &Server{}
	owner := &Object{CarryCapacity: 50}
	item := &Object{
		TypeInd:  17,
		ObjClass: object.ClassPickup,
		ObjFlags: object.FlagActive,
		Weight:   3,
		ObjOwner: owner,
		NetCode:  42,
	}
	events := make([]string, 0, 2)
	runtime := PickupTrapRuntime4F3510{
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
				if gotOwner != owner || gotItem != item || report != math.MinInt32 {
					t.Fatalf("put args = %p/%p/%d", gotOwner, gotItem, report)
				}
			},
		},
	}
	if got := s.PickupTrap4F3510(owner, item, math.MinInt32, math.MaxInt32, runtime); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := []string{"delete", "put"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != sound.SoundTrapPickup || audio.Obj != owner || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("queued audio = %#v", audio)
	}
}

func TestPickupTrap4F3510ServerRejectQueuesNoCanDo(t *testing.T) {
	s := &Server{}
	owner := &Object{ObjClass: object.ClassPlayer, NetCode: 0xfedcba98}
	item := &Object{}
	if got := s.PickupTrap4F3510(owner, item, 3, 4, PickupTrapRuntime4F3510{}); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != sound.SoundNoCanDo || audio.Obj != owner || audio.Kind != 2 || audio.Code != owner.NetCode {
		t.Fatalf("queued audio = %#v", audio)
	}
}
