package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPixieTeleportAllNative4FD090UsesNativePointersAndLiveSuccessor(t *testing.T) {
	const pixieType = uint16(0x1234)
	owner := &Object{}
	mismatch := &Object{TypeInd: pixieType + 1}
	dead := &Object{TypeInd: pixieType, ObjFlags: object.FlagDead}
	targetedUpdate := &PixieUpdateData{Target: &Object{}}
	targeted := &Object{TypeInd: pixieType, UpdateData: unsafe.Pointer(targetedUpdate)}
	eligibleUpdate := &PixieUpdateData{}
	eligible := &Object{
		TypeInd:    pixieType,
		ObjFlags:   ^object.FlagDead,
		UpdateData: unsafe.Pointer(eligibleUpdate),
	}
	originalEndUpdate := &PixieUpdateData{}
	originalEnd := &Object{TypeInd: pixieType, UpdateData: unsafe.Pointer(originalEndUpdate)}
	replacement := &Object{TypeInd: pixieType + 1}
	owner.Field129 = mismatch
	mismatch.Field128 = dead
	dead.Field128 = targeted
	targeted.Field128 = eligible
	eligible.Field128 = originalEnd

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"owner":  uintptr(unsafe.Pointer(owner)),
			"pixie":  uintptr(unsafe.Pointer(eligible)),
			"update": uintptr(unsafe.Pointer(eligibleUpdate)),
			"target": uintptr(unsafe.Pointer(targetedUpdate.Target)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	typeIDLoads := 0
	teleports := 0
	pixieTeleportAllNative4FD090(owner, pixieTeleportAllNativeDeps4FD090{
		loadPixieTypeID: func() uint32 {
			typeIDLoads++
			return uint32(pixieType)
		},
		teleport: func(gotPixie, gotOwner *Object) {
			teleports++
			if gotPixie != eligible || gotOwner != owner {
				t.Fatalf("teleport = (%p, %p), want (%p, %p)", gotPixie, gotOwner, eligible, owner)
			}
			eligible.Field128 = replacement
		},
	})
	if typeIDLoads != 5 {
		t.Fatalf("type ID loads = %d, want 5", typeIDLoads)
	}
	if teleports != 1 {
		t.Fatalf("teleports = %d, want 1", teleports)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(targetedUpdate)
	runtime.KeepAlive(eligibleUpdate)
	runtime.KeepAlive(originalEndUpdate)
}

func TestPixieTeleportAllNative4FD090NilOwnerFaultsBeforeTypeLoad(t *testing.T) {
	typeIDLoads := 0
	teleports := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner did not fault")
		}
		if typeIDLoads != 0 || teleports != 0 {
			t.Fatalf("nil owner side effects = type IDs %d, teleports %d", typeIDLoads, teleports)
		}
	}()
	pixieTeleportAllNative4FD090(nil, pixieTeleportAllNativeDeps4FD090{
		loadPixieTypeID: func() uint32 { typeIDLoads++; return 1 },
		teleport:        func(*Object, *Object) { teleports++ },
	})
}

func TestPixieTeleportAllNative4FD090NilUpdateDataFaultsBeforeTeleport(t *testing.T) {
	pixie := &Object{TypeInd: 1}
	owner := &Object{Field129: pixie}
	typeIDLoads := 0
	teleports := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil Pixie update data did not fault")
		}
		if typeIDLoads != 1 || teleports != 0 {
			t.Fatalf("nil update side effects = type IDs %d, teleports %d", typeIDLoads, teleports)
		}
	}()
	pixieTeleportAllNative4FD090(owner, pixieTeleportAllNativeDeps4FD090{
		loadPixieTypeID: func() uint32 { typeIDLoads++; return 1 },
		teleport:        func(*Object, *Object) { teleports++ },
	})
}

func TestPixieTeleportAll4FD090ServerMethodUsesNativeAdapter(t *testing.T) {
	update := &PixieUpdateData{}
	pixie := &Object{TypeInd: 7, UpdateData: unsafe.Pointer(update)}
	owner := &Object{Field129: pixie}
	called := false
	new(Server).PixieTeleportAll4FD090(owner, func() uint32 { return 7 }, func(gotPixie, gotOwner *Object) {
		called = true
		if gotPixie != pixie || gotOwner != owner {
			t.Fatalf("teleport = (%p, %p), want (%p, %p)", gotPixie, gotOwner, pixie, owner)
		}
	})
	if !called {
		t.Fatal("server method did not invoke teleport")
	}
}

func TestPixieTeleportAll4FD090NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantFlags := uintptr(16)
	wantNextOwned := uintptr(512)
	wantFirstOwned := uintptr(516)
	wantUpdateData := uintptr(748)
	wantPixieUpdateSize := uintptr(28)
	wantTarget := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantFlags = 20
		wantNextOwned = 560
		wantFirstOwned = 568
		wantUpdateData = 872
		wantPixieUpdateSize = 40
		wantTarget = 8
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Field128", unsafe.Offsetof(Object{}.Field128), wantNextOwned},
		{"Object.Field129", unsafe.Offsetof(Object{}.Field129), wantFirstOwned},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"PixieUpdateData size", unsafe.Sizeof(PixieUpdateData{}), wantPixieUpdateSize},
		{"PixieUpdateData.Target", unsafe.Offsetof(PixieUpdateData{}.Target), wantTarget},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
