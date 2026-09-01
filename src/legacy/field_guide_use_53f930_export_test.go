package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type fieldGuideUseLegacyServer53F930 struct {
	Server
	awardUnit   *server.Object
	awardArgs   [2]int32
	awardResult int32
	useOwner    *server.Object
	useItem     *server.Object
	useResult   int32
}

func (s *fieldGuideUseLegacyServer53F930) AwardBeastGuide4FAE80(
	unit *server.Object,
	guide, notify int32,
) int32 {
	s.awardUnit = unit
	s.awardArgs = [2]int32{guide, notify}
	return s.awardResult
}

func (s *fieldGuideUseLegacyServer53F930) UseFieldGuide53F930(
	owner, item *server.Object,
) int32 {
	s.useOwner = owner
	s.useItem = item
	return s.useResult
}

func installFieldGuideUseLegacyServer53F930(t *testing.T, fake *fieldGuideUseLegacyServer53F930) {
	t.Helper()
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })
}

func TestFieldGuideUseExportsPreserveNativePointersAndScalars(t *testing.T) {
	fake := &fieldGuideUseLegacyServer53F930{
		awardResult: math.MinInt32 + 0x1234,
		useResult:   -0x2345678,
	}
	installFieldGuideUseLegacyServer53F930(t, fake)

	owner := new(server.Object)
	item := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(owner)
	pin.Pin(item)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(item)) <= math.MaxUint32 {
			t.Fatalf("test pointers do not exercise native high addresses: owner=%p item=%p", owner, item)
		}
	}

	wantAwardArgs := [2]int32{math.MaxInt32, math.MinInt32 + 0x4321}
	if got := beastGuideAwardExportCall4FAE80(owner, wantAwardArgs[0], wantAwardArgs[1]); got != fake.awardResult {
		t.Fatalf("award export result = %d, want %d", got, fake.awardResult)
	}
	if fake.awardUnit != owner || fake.awardArgs != wantAwardArgs {
		t.Fatalf("award export call = %p/%v, want %p/%v", fake.awardUnit, fake.awardArgs, owner, wantAwardArgs)
	}

	fake.awardResult = 0x1234567
	wantAwardArgs = [2]int32{24, -7}
	if got := Nox_xxx_awardBeastGuide_4FAE80_magic_plyrgide(owner, wantAwardArgs[0], wantAwardArgs[1]); got != fake.awardResult {
		t.Fatalf("award Go wrapper result = %d, want %d", got, fake.awardResult)
	}
	if fake.awardUnit != owner || fake.awardArgs != wantAwardArgs {
		t.Fatalf("award Go wrapper call = %p/%v, want %p/%v", fake.awardUnit, fake.awardArgs, owner, wantAwardArgs)
	}

	if got := fieldGuideUseExportCall53F930(owner, item); got != fake.useResult {
		t.Fatalf("use export result = %d, want %d", got, fake.useResult)
	}
	if fake.useOwner != owner || fake.useItem != item {
		t.Fatalf("use export call = %p/%p, want %p/%p", fake.useOwner, fake.useItem, owner, item)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}

func TestFieldGuideUseRegistration53F930DispatchesNativeImplementation(t *testing.T) {
	fake := &fieldGuideUseLegacyServer53F930{useResult: 1}
	installFieldGuideUseLegacyServer53F930(t, fake)
	owner := new(server.Object)
	item := new(server.Object)

	use := server.UseFuncPtr{Ptr: Get_sub_53F930()}.Get()
	if use == nil {
		t.Fatal("registered FieldGuideUse callback is nil")
	}
	if !use(owner, item) {
		t.Fatal("registered FieldGuideUse result = false, want true")
	}
	if fake.useOwner != owner || fake.useItem != item {
		t.Fatalf("registered call = %p/%p, want %p/%p", fake.useOwner, fake.useItem, owner, item)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(item)) <= math.MaxUint32 {
			t.Fatalf("test pointers do not exercise native high addresses: owner=%p item=%p", owner, item)
		}
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(item)
}
