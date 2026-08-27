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

func defaultPickupFoodNativeDeps4F3350() pickupFoodNativeDeps4F3350 {
	return pickupFoodNativeDeps4F3350{
		playerState:   func(*Object) int32 { return 0 },
		defaultPickup: func(*Object, *Object, int32, int32) int32 { return 0 },
		audio:         func(uint32, *Object, int32, uint32) {},
	}
}

func TestPickupFood4F3350NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantSubClass := uintptr(12)
	wantFlags := uintptr(16)
	wantMaterial := uintptr(24)
	wantUse := uintptr(732)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantSubClass = 16
		wantFlags = 20
		wantMaterial = 28
		wantUse = 840
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Material", unsafe.Offsetof(Object{}.Material), wantMaterial},
		{"Object.Use", unsafe.Offsetof(Object{}.Use), wantUse},
		{"Object.ObjSubClass size", unsafe.Sizeof(Object{}.ObjSubClass), 4},
		{"Object.ObjFlags size", unsafe.Sizeof(Object{}.ObjFlags), 4},
		{"Object.Material size", unsafe.Sizeof(Object{}.Material), 2},
		{"UseFuncPtr size", unsafe.Sizeof(UseFuncPtr{}), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPickupFood4F3350NativeSoundTableMatchesOracleAndEnums(t *testing.T) {
	want := [...]pickupFoodSoundRule4F3350{
		{materialMask: uint16(object.MaterialFlesh), sound: uint16(sound.SoundMeatPickup)},
		{subClassMask: uint32(object.FoodApple), sound: uint16(sound.SoundApplePickup)},
		{subClassMask: uint32(object.FoodJug), sound: uint16(sound.SoundPotionPickup)},
		{subClassMask: uint32(object.FoodMushroom), sound: uint16(sound.SoundShroomPickup)},
		{},
	}
	if !reflect.DeepEqual(pickupFoodSoundRules4F3350, want) {
		t.Fatalf("sound rules = %#v, want %#v", pickupFoodSoundRules4F3350, want)
	}
}

func TestPickupFoodNative4F3350BindsUseAndLiveDestroyedFlag(t *testing.T) {
	var token byte
	usePointer := unsafe.Pointer(&token)
	owner := &Object{}
	item := &Object{Use: UseFuncPtr{Ptr: usePointer}}
	calls := 0
	objUse.Register(usePointer, func(gotOwner, gotItem *Object) bool {
		calls++
		if gotOwner != owner || gotItem != item {
			t.Fatalf("Use args = %p/%p, want %p/%p", gotOwner, gotItem, owner, item)
		}
		item.ObjFlags = object.FlagDestroyed | object.Flags(0x80000000)
		return false
	})
	deps := defaultPickupFoodNativeDeps4F3350()
	deps.defaultPickup = func(*Object, *Object, int32, int32) int32 {
		t.Fatal("destroyed Use path called DefaultPickup")
		return 0
	}
	deps.audio = func(uint32, *Object, int32, uint32) {
		t.Fatal("destroyed Use path played pickup audio")
	}
	if got := pickupFoodNative4F3350(owner, item, -1, -2, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if calls != 1 {
		t.Fatalf("Use calls = %d, want 1", calls)
	}
}

func TestPickupFoodNative4F3350BindsFourArgsLiveFieldsAndAudio(t *testing.T) {
	owner := &Object{}
	item := &Object{ObjSubClass: object.SubClass(object.FoodJug)}
	events := make([]string, 0, 2)
	deps := defaultPickupFoodNativeDeps4F3350()
	deps.playerState = func(gotOwner *Object) int32 {
		if gotOwner != owner {
			t.Fatalf("player-state owner = %p, want %p", gotOwner, owner)
		}
		return math.MinInt32
	}
	deps.defaultPickup = func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || arg3 != math.MinInt32 || arg4 != math.MaxInt32 {
			t.Fatalf("default args = %p/%p/%d/%d", gotOwner, gotItem, arg3, arg4)
		}
		item.ObjSubClass = object.SubClass(object.FoodMushroom)
		return math.MinInt32
	}
	deps.audio = func(id uint32, gotOwner *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != uint32(sound.SoundShroomPickup) || gotOwner != owner || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotOwner, kind, code)
		}
	}
	if got := pickupFoodNative4F3350(owner, item, math.MinInt32, math.MaxInt32, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if want := []string{"default", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPickupFoodNative4F3350NilGuardsSkipDependencies(t *testing.T) {
	owner := &Object{}
	item := &Object{}
	for _, tc := range []struct {
		name        string
		owner, item *Object
	}{
		{name: "owner", item: item},
		{name: "item", owner: owner},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deps := defaultPickupFoodNativeDeps4F3350()
			deps.playerState = func(*Object) int32 { t.Fatal("nil guard read player state"); return 0 }
			deps.defaultPickup = func(*Object, *Object, int32, int32) int32 { t.Fatal("nil guard called DefaultPickup"); return 0 }
			deps.audio = func(uint32, *Object, int32, uint32) { t.Fatal("nil guard played audio") }
			if got := pickupFoodNative4F3350(tc.owner, tc.item, 1, 2, deps); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
		})
	}
}

func TestPickupFoodNative4F3350NilUseFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil Use callback did not fault")
		}
	}()
	pickupFoodNative4F3350(&Object{}, &Object{}, 0, 0, defaultPickupFoodNativeDeps4F3350())
}

func TestPickupFood4F3350ServerBindingDefaultPickupAndQueuesAudio(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := &Server{}
	owner := &Object{CarryCapacity: 50}
	item := &Object{
		TypeInd:     17,
		ObjClass:    object.ClassFood,
		ObjSubClass: object.SubClass(object.FoodJug),
		ObjFlags:    object.FlagActive,
		Weight:      3,
	}
	events := make([]string, 0, 2)
	runtime := PickupFoodRuntime4F3350{
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
	if got := s.PickupFood4F3350(owner, item, math.MinInt32, math.MaxInt32, runtime); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := []string{"delete", "put"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != sound.SoundPotionPickup || audio.Obj != owner || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("queued audio = %#v", audio)
	}
}
