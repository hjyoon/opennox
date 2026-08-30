package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPickupAmmo4F3B00NativeLayouts(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantClass := uintptr(8)
	wantInvNext := uintptr(496)
	wantInvFirst := uintptr(504)
	wantInitData := uintptr(692)
	wantUseData := uintptr(736)
	wantUpdateData := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantUpdatePlayer := uintptr(276)
	wantPlayerSize := uintptr(4828)
	wantPlayerInd := uintptr(2064)
	wantModifierInitSize := uintptr(20)
	if ptrSize == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantClass = 12
		wantInvNext = 528
		wantInvFirst = 544
		wantInitData = 760
		wantUseData = 848
		wantUpdateData = 872
		wantUpdateSize = 656
		wantUpdatePlayer = 336
		wantPlayerSize = 6160
		wantPlayerInd = 2068
		wantModifierInitSize = 40
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantInvNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantInvFirst},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"UseDataPtr size", unsafe.Sizeof(UseDataPtr{}), ptrSize},
		{"ModifierInitData size", unsafe.Sizeof(ModifierInitData{}), wantModifierInitSize},
		{"ModifierInitData.Modifiers[1]", unsafe.Offsetof(ModifierInitData{}.Modifiers) + ptrSize, ptrSize},
		{"ModifierInitData.Modifiers[3]", unsafe.Offsetof(ModifierInitData{}.Modifiers) + 3*ptrSize, 3 * ptrSize},
		{"AmmoUseData size", unsafe.Sizeof(AmmoUseData{}), 3},
		{"AmmoUseData.Charge0", unsafe.Offsetof(AmmoUseData{}.Charge0), 0},
		{"AmmoUseData.Charge1", unsafe.Offsetof(AmmoUseData{}.Charge1), 1},
		{"AmmoUseData.Field2", unsafe.Offsetof(AmmoUseData{}.Field2), 2},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerInd},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	if def := useFuncs["AmmoUse"]; def.DataSize != unsafe.Sizeof(AmmoUseData{}) || def.DataSize != 3 {
		t.Fatalf("AmmoUse registration size = %d, want 3", def.DataSize)
	}
}

func TestPickupAmmoNative4F3B00UsesNativePointersAndCachedData(t *testing.T) {
	mods := [4]*ModifierEff{&ModifierEff{}, nil, &ModifierEff{}, &ModifierEff{}}
	itemInit := &ModifierInitData{Modifiers: mods}
	candidateInit := &ModifierInitData{Modifiers: mods}
	itemUse := &AmmoUseData{Charge0: 50, Charge1: 250, Field2: 0xff}
	candidateUse := &AmmoUseData{Charge0: 200, Charge1: 10}
	player := &Player{PlayerInd: 0xfe}
	update := &PlayerUpdateData{Player: player}
	replacement := &PlayerUpdateData{Player: &Player{PlayerInd: 1}}
	item := &Object{
		TypeInd:  0x8001,
		InitData: unsafe.Pointer(itemInit),
		UseData:  UseDataPtr{Ptr: unsafe.Pointer(itemUse)},
	}
	candidate := &Object{
		TypeInd:  0x8001,
		ObjClass: object.ClassWeapon,
		InitData: unsafe.Pointer(candidateInit),
		UseData:  UseDataPtr{Ptr: unsafe.Pointer(candidateUse)},
	}
	owner := &Object{
		ObjClass:     object.ClassPlayer,
		InvFirstItem: candidate,
		UpdateData:   unsafe.Pointer(update),
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"owner": unsafe.Pointer(owner), "item": unsafe.Pointer(item),
			"candidate": unsafe.Pointer(candidate), "item init": unsafe.Pointer(itemInit),
			"candidate init": unsafe.Pointer(candidateInit), "item use": unsafe.Pointer(itemUse),
			"candidate use": unsafe.Pointer(candidateUse), "update": unsafe.Pointer(update),
			"player": unsafe.Pointer(player),
		} {
			if uintptr(pointer) <= uintptr(math.MaxUint32) {
				t.Fatalf("%s pointer %#x does not exercise the high 64-bit half", name, uintptr(pointer))
			}
		}
	}

	var events []string
	deps := pickupAmmoNativeDeps4F3B00{
		weaponEquipFlags: func(obj *Object) uint32 {
			switch obj {
			case item:
				events = append(events, "flags:item")
				return 0x82
			case candidate:
				events = append(events, "flags:candidate")
				owner.UpdateData = unsafe.Pointer(replacement)
				return 0x80
			default:
				t.Fatalf("unexpected flags object %p", obj)
				return 0
			}
		},
		weaponPickup: func(*Object, *Object, int32, int32) int32 {
			t.Fatal("compatible ammunition fell back to WeaponPickup")
			return 0
		},
		reportCharges: func(index uint8, got *Object, charge1, charge0 uint8) {
			events = append(events, "report")
			if index != 0xfe || got != candidate || charge1 != 4 || charge0 != 250 {
				t.Fatalf("report = (%02x,%p,%d,%d), want (fe,%p,4,250)", index, got, charge1, charge0, candidate)
			}
		},
		delayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != item {
				t.Fatalf("deleted = %p, want %p", got, item)
			}
		},
		pickupAudio: func(gotOwner, gotItem *Object) {
			events = append(events, "audio")
			if gotOwner != owner || gotItem != item {
				t.Fatalf("audio objects = (%p,%p), want (%p,%p)", gotOwner, gotItem, owner, item)
			}
		},
	}
	if got := pickupAmmoNative4F3B00(owner, item, math.MinInt32, math.MaxInt32, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if candidateUse.Charge0 != 250 || candidateUse.Charge1 != 4 || candidateUse.Field2 != 0 {
		t.Fatalf("candidate charges = %+v, want 250/4/0", *candidateUse)
	}
	if !reflect.DeepEqual(events, []string{"flags:item", "flags:candidate", "report", "delete", "audio"}) {
		t.Fatalf("events = %v", events)
	}
}

func TestPickupAmmoNative4F3B00FallbackPreservesSignedResult(t *testing.T) {
	owner := &Object{ObjClass: object.ClassMonster}
	item := &Object{}
	var gotArgs [2]int32
	deps := pickupAmmoNativeDeps4F3B00{
		weaponEquipFlags: func(got *Object) uint32 {
			if got != item {
				t.Fatalf("flags object = %p, want %p", got, item)
			}
			return 0x82
		},
		weaponPickup: func(gotOwner, gotItem *Object, arg3, arg4 int32) int32 {
			if gotOwner != owner || gotItem != item {
				t.Fatalf("fallback objects = (%p,%p), want (%p,%p)", gotOwner, gotItem, owner, item)
			}
			gotArgs = [2]int32{arg3, arg4}
			return math.MinInt32
		},
	}
	if got := pickupAmmoNative4F3B00(owner, item, math.MinInt32, math.MaxInt32, deps); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if gotArgs != [2]int32{math.MinInt32, math.MaxInt32} {
		t.Fatalf("fallback args = %v", gotArgs)
	}
}

func TestPickupAmmoNative4F3B00PreservesNilDataFaultBoundaries(t *testing.T) {
	newObjects := func() (*Object, *Object, *Object, *AmmoUseData) {
		use := &AmmoUseData{Charge0: 1}
		candidate := &Object{
			TypeInd: 1, ObjClass: object.ClassWeapon,
			InitData: unsafe.Pointer(&ModifierInitData{}),
			UseData:  UseDataPtr{Ptr: unsafe.Pointer(&AmmoUseData{})},
		}
		item := &Object{TypeInd: 1, UseData: UseDataPtr{Ptr: unsafe.Pointer(use)}}
		owner := &Object{
			ObjClass: object.ClassPlayer, InvFirstItem: candidate,
			UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: &Player{}}),
		}
		return owner, item, candidate, use
	}
	newDeps := func(item, candidate *Object) pickupAmmoNativeDeps4F3B00 {
		return pickupAmmoNativeDeps4F3B00{
			weaponEquipFlags: func(obj *Object) uint32 {
				if obj == item {
					return 0x82
				}
				if obj == candidate {
					return 0x80
				}
				return 0
			},
			weaponPickup:  func(*Object, *Object, int32, int32) int32 { return 0 },
			reportCharges: func(uint8, *Object, uint8, uint8) {},
			delayedDelete: func(*Object) {},
			pickupAudio:   func(*Object, *Object) {},
		}
	}

	t.Run("nil item InitData faults after candidate data loads", func(t *testing.T) {
		owner, item, candidate, _ := newObjects()
		defer func() {
			if recover() == nil {
				t.Fatal("nil item InitData did not fault")
			}
		}()
		pickupAmmoNative4F3B00(owner, item, 0, 0, newDeps(item, candidate))
	})

	t.Run("nil cached update faults only after charge stores", func(t *testing.T) {
		owner, item, candidate, _ := newObjects()
		owner.UpdateData = nil
		item.InitData = candidate.InitData
		candidateUse := candidate.UseDataAmmo()
		defer func() {
			if recover() == nil {
				t.Fatal("nil cached update did not fault")
			}
			if candidateUse.Charge0 != 1 {
				t.Fatalf("charge before report fault = %d, want 1", candidateUse.Charge0)
			}
		}()
		pickupAmmoNative4F3B00(owner, item, 0, 0, newDeps(item, candidate))
	})
}
