package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestRewardContainerNativeLayout4F1F20(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantClass := uintptr(8)
	wantSubclass := uintptr(12)
	wantPosition := uintptr(56)
	wantNext := uintptr(444)
	wantInvNext := uintptr(496)
	wantInvFirst := uintptr(504)
	wantInit := uintptr(688)
	wantInitData := uintptr(692)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantClass = 12
		wantSubclass = 16
		wantPosition = 60
		wantNext = 448
		wantInvNext = 528
		wantInvFirst = 544
		wantInit = 752
		wantInitData = 760
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubclass},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.ObjNext", unsafe.Offsetof(Object{}.ObjNext), wantNext},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantInvNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantInvFirst},
		{"Object.Init", unsafe.Offsetof(Object{}.Init), wantInit},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"RewardMarkerInitData size", unsafe.Sizeof(RewardMarkerInitData{}), 220},
		{"RewardMarkerInitData.Field216", unsafe.Offsetof(RewardMarkerInitData{}.Field216), 216},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestRewardContainerNativeUsesLiveNativePositionPointer4F1F20(t *testing.T) {
	data := &RewardMarkerInitData{Field216: 0x12345680}
	marker := &Object{TypeInd: 7, InitData: unsafe.Pointer(data), PosVec: types.Pointf{X: 5, Y: 6}}
	reward := &Object{
		ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(object.WeaponBow),
		PosVec: types.Pointf{X: -1, Y: -2},
	}
	quiver := new(Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 {
		t.Fatalf("marker address %#x does not exercise a native 64-bit pointer", uintptr(unsafe.Pointer(marker)))
	}

	markerCache := uint32(7)
	plusCache := uint32(8)
	deleted := false
	randomCalled := false
	rewardContainerNative4F1F20(rewardContainerNativeDeps4F1F20{
		loadMarkerCache:  func() uint32 { return markerCache },
		storeMarkerCache: func(value uint32) { markerCache = value },
		loadPlusCache:    func() uint32 { return plusCache },
		storePlusCache:   func(value uint32) { plusCache = value },
		lookupType:       func(string) uint32 { t.Fatal("cached types were looked up"); return 0 },
		firstObject:      func() *Object { return marker },
		chestInit:        unsafe.Pointer(new(byte)),
		newObject: func(name string) *Object {
			if name != rewardContainerQuiverTypeName4F1F20 {
				t.Fatalf("new object name = %q", name)
			}
			return quiver
		},
		randomReachable: func(radius float32, center, output *types.Pointf) *types.Pointf {
			randomCalled = true
			if radius != 30 || center != &reward.PosVec || output == center ||
				*center != (types.Pointf{X: 5, Y: 6}) || *output != *center {
				t.Fatalf("random args = %g/%p/%p/%+v/%+v", radius, center, output, *center, *output)
			}
			*output = types.Pointf{X: 11, Y: 12}
			return output
		},
		runtime: RewardContainerRuntime4F1F20{
			QuestStage:        func() uint32 { return 41 },
			PreprocessMarkers: func() {},
			PreprocessRewards: func() {},
			ActivateMarker: func(got *Object, stage uint32) *Object {
				if got != marker || stage != 41 {
					t.Fatalf("activate args = %p/%d", got, stage)
				}
				return reward
			},
			CreateAt: func(object, owner *Object, point types.Pointf) {
				if owner != nil {
					t.Fatalf("CreateAt owner = %p, want nil", owner)
				}
				object.PosVec = point
			},
			DelayedDelete: func(got *Object) {
				if got != marker {
					t.Fatalf("deleted = %p, want marker", got)
				}
				deleted = true
			},
			DetachInventory: func(*Object, *Object) { t.Fatal("world marker detached") },
			InventoryPut:    func(*Object, *Object, uint32) { t.Fatal("world reward put in inventory") },
		},
	})
	if !randomCalled || !deleted || markerCache != 7 || plusCache != 8 {
		t.Fatalf("random/deleted/caches = %v/%v/%d/%d", randomCalled, deleted, markerCache, plusCache)
	}
	if reward.PosVec != (types.Pointf{X: 5, Y: 6}) || quiver.PosVec != (types.Pointf{X: 11, Y: 12}) {
		t.Fatalf("reward/quiver positions = %+v/%+v", reward.PosVec, quiver.PosVec)
	}
}
