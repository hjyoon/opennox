package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestCharmLifecycleExportsPreserveNativeRecordPointerAndInt32(t *testing.T) {
	oldStart := charmStartCall5011F0
	oldFinish := charmFinishCall5013E0
	oldCancel := charmCancelCall501690
	t.Cleanup(func() {
		charmStartCall5011F0 = oldStart
		charmFinishCall5013E0 = oldFinish
		charmCancelCall501690 = oldCancel
	})

	record := new(server.DurSpell)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(record)) <= math.MaxUint32 {
		t.Fatalf("record pointer = %p, want native address above 4 GiB", record)
	}
	var pin runtime.Pinner
	pin.Pin(record)
	defer pin.Unpin()

	var startRecord, finishRecord, cancelRecord *server.DurSpell
	charmStartCall5011F0 = func(got *server.DurSpell) int32 {
		startRecord = got
		return math.MinInt32
	}
	charmFinishCall5013E0 = func(got *server.DurSpell) int32 {
		finishRecord = got
		return math.MaxInt32
	}
	charmCancelCall501690 = func(got *server.DurSpell) int32 {
		cancelRecord = got
		return -1
	}

	if got := charmStartExportCall5011F0(record); got != math.MinInt32 {
		t.Fatalf("start result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := charmFinishExportCall5013E0(record); got != math.MaxInt32 {
		t.Fatalf("finish result = %d, want %d", got, int32(math.MaxInt32))
	}
	if got := charmCancelExportCall501690(record); got != -1 {
		t.Fatalf("cancel result = %d, want -1", got)
	}
	if startRecord != record || finishRecord != record || cancelRecord != record {
		t.Fatalf("records = %p/%p/%p, want %p", startRecord, finishRecord, cancelRecord, record)
	}

	if got := charmStartExportCall5011F0(nil); got != math.MinInt32 || startRecord != nil {
		t.Fatalf("nil start = %d/%p", got, startRecord)
	}
	if got := charmFinishExportCall5013E0(nil); got != math.MaxInt32 || finishRecord != nil {
		t.Fatalf("nil finish = %d/%p", got, finishRecord)
	}
	if got := charmCancelExportCall501690(nil); got != -1 || cancelRecord != nil {
		t.Fatalf("nil cancel = %d/%p", got, cancelRecord)
	}
	runtime.KeepAlive(record)
}
