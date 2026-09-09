package server

import (
	"math"
	"runtime"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func creatureMonitoredTestServer500CC0(t *testing.T) *Server {
	t.Helper()
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	return s
}

func TestCreatureIsMonitoredNative500CC0PreservesNativePointers(t *testing.T) {
	owner := new(Object)
	update := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned | object.MonsterStatus(0x7fffff00)}
	unit := &Object{
		ObjClass:   object.ClassMonster | object.Class(0x80000000),
		ObjFlags:   object.Flags(0x7fff0000),
		ObjOwner:   owner,
		UpdateData: unsafe.Pointer(update),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"owner":  uintptr(unsafe.Pointer(owner)),
			"unit":   uintptr(unsafe.Pointer(unit)),
			"update": uintptr(unsafe.Pointer(update)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}

	if !creatureIsMonitoredNative500CC0(owner, unit, func(got *Object) bool {
		t.Fatalf("alive Monster unexpectedly called zombie predicate with %p", got)
		return false
	}) {
		t.Fatal("alive summoned Monster with native-width owner was rejected")
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
}

func TestCreatureIsMonitoredNative500CC0AcceptsNonMonsterZombie(t *testing.T) {
	owner := new(Object)
	update := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagDead,
		ObjOwner:   owner,
		UpdateData: unsafe.Pointer(update),
	}
	calls := 0
	if !creatureIsMonitoredNative500CC0(owner, unit, func(got *Object) bool {
		calls++
		if got != unit {
			t.Fatalf("zombie unit = %p, want %p", got, unit)
		}
		return true
	}) {
		t.Fatal("summoned non-Monster zombie accepted by the original expression was rejected")
	}
	if calls != 1 {
		t.Fatalf("zombie calls = %d, want 1", calls)
	}
	runtime.KeepAlive(update)
}

func TestCreatureIsMonitoredNative500CC0ExactFaultAndDeadGates(t *testing.T) {
	t.Run("nil unit faults at class read", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit was hidden by an added preflight")
			}
		}()
		creatureIsMonitoredNative500CC0(new(Object), nil, func(*Object) bool { return false })
	})

	t.Run("eligible unit without update faults at status read", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("missing update data was hidden by an added preflight")
			}
		}()
		creatureIsMonitoredNative500CC0(new(Object), &Object{ObjClass: object.ClassMonster}, func(*Object) bool {
			t.Fatal("alive Monster unexpectedly called zombie predicate")
			return false
		})
	})

	t.Run("dead non-zombie stops before update", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassMonster, ObjFlags: object.FlagDead}
		calls := 0
		if creatureIsMonitoredNative500CC0(new(Object), unit, func(got *Object) bool {
			calls++
			return false
		}) {
			t.Fatal("dead non-zombie was accepted")
		}
		if calls != 1 {
			t.Fatalf("zombie calls = %d, want 1", calls)
		}
	})
}

func TestNoxCreatureIsMonitored500CC0UsesServerZombiePredicate(t *testing.T) {
	s := creatureMonitoredTestServer500CC0(t)
	s.Types.fast.zombie = 0x1234
	s.Types.fast.zombieVile = 0x1235
	owner := new(Object)
	update := &MonsterUpdateData{StatusFlags: object.MonStatusSummoned}
	unit := &Object{
		TypeInd:      0x1234,
		ObjClass:     object.ClassPlayer,
		ObjOwner:     owner,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	if !Nox_xxx_creatureIsMonitored_500CC0(owner, unit) {
		t.Fatal("server-recognized non-Monster zombie was rejected")
	}
	runtime.KeepAlive(update)
}

func TestCreatureIsMonitoredNative500CC0Layout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantOwner := uintptr(508)
	wantUpdate := uintptr(748)
	wantMonsterSize := uintptr(2200)
	wantStatus := uintptr(1440)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantFlags = 20
		wantOwner = 552
		wantUpdate = 872
		wantMonsterSize = 2960
		wantStatus = 2180
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjClass width", unsafe.Sizeof(Object{}.ObjClass), 4},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.ObjFlags width", unsafe.Sizeof(Object{}.ObjFlags), 4},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"Object.UpdateData width", unsafe.Sizeof(Object{}.UpdateData), unsafe.Sizeof(uintptr(0))},
		{"MonsterUpdateData size", unsafe.Sizeof(MonsterUpdateData{}), wantMonsterSize},
		{"MonsterUpdateData.StatusFlags", unsafe.Offsetof(MonsterUpdateData{}.StatusFlags), wantStatus},
		{"MonsterUpdateData.StatusFlags width", unsafe.Sizeof(MonsterUpdateData{}.StatusFlags), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}
