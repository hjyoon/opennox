package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestDropOwnedCrowns4ED050NativeLayout(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	wantObjectSize := uintptr(780)
	wantType := uintptr(4)
	wantPosition := uintptr(56)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	wantUpdate := uintptr(748)
	wantCrownUpdateSize := uintptr(12)
	wantPickupTarget := uintptr(4)
	if ptrSize == 8 {
		wantObjectSize = 928
		wantType = 8
		wantPosition = 60
		wantNextOwned = 560
		wantFirstOwned = 568
		wantUpdate = 872
		wantCrownUpdateSize = 24
		wantPickupTarget = 8
	}

	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantNextOwned},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantFirstOwned},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"CrownUpdateData size", unsafe.Sizeof(CrownUpdateData{}), wantCrownUpdateSize},
		{"CrownUpdateData.PickupTarget", unsafe.Offsetof(CrownUpdateData{}.PickupTarget), wantPickupTarget},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDropOwnedCrownsNative4ED050UsesCachedUpdateAndLiveNext(t *testing.T) {
	cache := uint32(0)
	target := &Object{}
	firstData := &CrownUpdateData{}
	replacement := &CrownUpdateData{}
	secondData := &CrownUpdateData{}
	second := &Object{TypeInd: 9, UpdateData: unsafe.Pointer(secondData)}
	stale := &Object{TypeInd: 7, UpdateData: unsafe.Pointer(&CrownUpdateData{})}
	first := &Object{TypeInd: 7, Field128: stale, UpdateData: unsafe.Pointer(firstData)}
	other := &Object{TypeInd: 6, Field128: first}
	owner := &Object{PosVec: types.Pointf{X: 12.5, Y: -7.25}, Field129: other}
	drops := 0

	dropOwnedCrownsNative4ED050(owner, target, dropOwnedCrownsNativeDeps4ED050{
		loadCrownTypeCache: func() uint32 {
			return cache
		},
		lookupCrownType: func() uint32 {
			return 7
		},
		storeCrownTypeCache: func(value uint32) {
			cache = value
		},
		dropCrown: func(gotOwner, crown *Object, position *types.Pointf) uint32 {
			drops++
			if gotOwner != owner || position != &owner.PosVec {
				t.Fatalf("drop owner/position = (%p,%p), want (%p,%p)", gotOwner, position, owner, &owner.PosVec)
			}
			if crown == first {
				first.UpdateData = unsafe.Pointer(replacement)
				first.Field128 = second
				cache = uint32(second.TypeInd)
			} else if crown != second {
				t.Fatalf("unexpected dropped object %p", crown)
			}
			return 0xf1234567
		},
	})

	if drops != 2 {
		t.Fatalf("drop calls = %d, want 2", drops)
	}
	if firstData.PickupTarget != target || secondData.PickupTarget != target {
		t.Fatalf("cached update targets = (%p,%p), want %p", firstData.PickupTarget, secondData.PickupTarget, target)
	}
	if replacement.PickupTarget != nil {
		t.Fatalf("replacement update target = %p, want nil", replacement.PickupTarget)
	}
}

func TestDropOwnedCrowns4ED050ServerForwardsNativeRuntime(t *testing.T) {
	s := &Server{}
	cache := uint32(0)
	target := &Object{}
	update := &CrownUpdateData{}
	crown := &Object{TypeInd: 0xffff, UpdateData: unsafe.Pointer(update)}
	owner := &Object{Field129: crown}
	lookupCalls := 0
	s.DropOwnedCrowns4ED050(owner, target, DropOwnedCrownsRuntime4ED050{
		LoadCrownTypeCache: func() uint32 {
			return cache
		},
		LookupCrownType: func() uint32 {
			lookupCalls++
			return 0xffff
		},
		StoreCrownTypeCache: func(value uint32) {
			cache = value
		},
		DropCrown: func(gotOwner, gotCrown *Object, position *types.Pointf) uint32 {
			if gotOwner != owner || gotCrown != crown || position != &owner.PosVec {
				t.Fatalf("drop args = (%p,%p,%p)", gotOwner, gotCrown, position)
			}
			return 0
		},
	})
	if lookupCalls != 1 || cache != 0xffff || update.PickupTarget != target {
		t.Fatalf("runtime result = (lookups %d, cache %#x, target %p)", lookupCalls, cache, update.PickupTarget)
	}
}
