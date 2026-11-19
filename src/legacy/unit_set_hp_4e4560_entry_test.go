package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestUnitSetHPPublicCEntry4E4560KeepsNativeHolder(t *testing.T) {
	unit, releaseUnit := alloc.NewHigh(server.Object{})
	holder, releaseHolder := alloc.NewHigh(server.Object{})
	hp, releaseHP := alloc.NewHigh(server.HealthData{})
	t.Cleanup(releaseUnit)
	t.Cleanup(releaseHolder)
	t.Cleanup(releaseHP)
	for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(holder), hp.C()} {
		if cgoABIPointerSize == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native fixture pointer %p must be above 4 GiB", p)
		}
	}
	*holder = server.Object{ObjClass: object.ClassMonster}
	*unit = server.Object{InvHolder: holder, Field38: 0x12345678}
	if got := unitSetHPCEntry4E4560(unit, 213); got != 0 || unit.Field38 != 0x12345678 {
		t.Fatal("missing health must return before sync")
	}
	*hp = server.HealthData{Cur: 34, Max: 100, Field2: 0xa6a6, Field16: 0x34567890}
	unit.HealthData = hp
	beforeHolder := *holder
	if got := unitSetHPCEntry4E4560(unit, 213); got != uintptr(unsafe.Pointer(holder)) {
		t.Fatalf("holder result %#x, want %p", got, holder)
	}
	if *hp != (server.HealthData{Cur: 213, Max: 100, Field2: 0xa6a6, Field16: 0x34567890}) ||
		unit.Field38 != math.MaxUint32 || *holder != beforeHolder {
		t.Fatal("original write must change only current health and sync state")
	}
}
