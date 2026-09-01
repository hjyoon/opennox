package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type spellRewardUseLegacyServer53F9E0 struct {
	Server
	useOwner  *server.Object
	useItem   *server.Object
	useResult int32
}

func (s *spellRewardUseLegacyServer53F9E0) UseSpellReward53F9E0(
	owner, item *server.Object,
) int32 {
	s.useOwner = owner
	s.useItem = item
	return s.useResult
}

func installSpellRewardUseLegacyServer53F9E0(
	t *testing.T,
	fake *spellRewardUseLegacyServer53F9E0,
) {
	t.Helper()
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })
}

func TestSpellRewardUseExport53F9E0PreservesNativePointersAndResult(t *testing.T) {
	fake := &spellRewardUseLegacyServer53F9E0{useResult: math.MinInt32 + 0x4321}
	installSpellRewardUseLegacyServer53F9E0(t, fake)

	owner := new(server.Object)
	item := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(owner)
	pin.Pin(item)
	defer pin.Unpin()
	assertSpellRewardUseHighPointers53F9E0(t, owner, item)

	if got := spellRewardUseExportCall53F9E0(owner, item); got != fake.useResult {
		t.Fatalf("use export result = %d, want %d", got, fake.useResult)
	}
	if fake.useOwner != owner || fake.useItem != item {
		t.Fatalf("use export call = %p/%p, want %p/%p", fake.useOwner, fake.useItem, owner, item)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}

func TestSpellRewardUseRegistration53F9E0DispatchesNativeImplementation(t *testing.T) {
	fake := &spellRewardUseLegacyServer53F9E0{useResult: 1}
	installSpellRewardUseLegacyServer53F9E0(t, fake)
	owner := new(server.Object)
	item := new(server.Object)
	assertSpellRewardUseHighPointers53F9E0(t, owner, item)

	use := server.UseFuncPtr{Ptr: Get_nox_xxx_useSpellReward_53F9E0()}.Get()
	if use == nil {
		t.Fatal("registered SpellRewardUse callback is nil")
	}
	if !use(owner, item) {
		t.Fatal("registered SpellRewardUse result = false, want true")
	}
	if fake.useOwner != owner || fake.useItem != item {
		t.Fatalf("registered call = %p/%p, want %p/%p", fake.useOwner, fake.useItem, owner, item)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}

func TestSpellRewardUseSignCollide53F9E0PreservesHighPointers(t *testing.T) {
	fake := &spellRewardUseLegacyServer53F9E0{useResult: 1}
	installSpellRewardUseLegacyServer53F9E0(t, fake)

	owner := &server.Object{ObjClass: object.ClassPlayer}
	item := &server.Object{Use: server.UseFuncPtr{Ptr: Get_nox_xxx_useSpellReward_53F9E0()}}
	assertSpellRewardUseHighPointers53F9E0(t, owner, item)

	new(server.Server).SignCollide4EAB40(item, owner, nil)
	if fake.useOwner != owner || fake.useItem != item {
		t.Fatalf("sign collision call = %p/%p, want %p/%p", fake.useOwner, fake.useItem, owner, item)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}

func assertSpellRewardUseHighPointers53F9E0(
	t *testing.T,
	owner, item *server.Object,
) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	if uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(item)) <= math.MaxUint32 {
		t.Fatalf("test pointers do not exercise native high addresses: owner=%p item=%p", owner, item)
	}
}
