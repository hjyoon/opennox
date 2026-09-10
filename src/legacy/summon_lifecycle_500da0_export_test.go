package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestSummonLifecycleExportsPreserveNativeRecordPointerAndInt32(t *testing.T) {
	oldStart := summonStartCall500DA0
	oldFinish := summonFinishCall5010D0
	oldCancel := summonCancelCall5011C0
	t.Cleanup(func() {
		summonStartCall500DA0 = oldStart
		summonFinishCall5010D0 = oldFinish
		summonCancelCall5011C0 = oldCancel
	})

	record := new(server.DurSpell)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(record)) <= math.MaxUint32 {
		t.Fatalf("record pointer = %p, want native address above 4 GiB", record)
	}
	var pin runtime.Pinner
	pin.Pin(record)
	defer pin.Unpin()

	var startRecord, finishRecord, cancelRecord *server.DurSpell
	summonStartCall500DA0 = func(got *server.DurSpell) int32 {
		startRecord = got
		return math.MinInt32
	}
	summonFinishCall5010D0 = func(got *server.DurSpell) int32 {
		finishRecord = got
		return math.MaxInt32
	}
	summonCancelCall5011C0 = func(got *server.DurSpell) {
		cancelRecord = got
	}

	if got := summonStartExportCall500DA0(record); got != math.MinInt32 {
		t.Fatalf("start result = %d, want %d", got, int32(math.MinInt32))
	}
	if got := summonFinishExportCall5010D0(record); got != math.MaxInt32 {
		t.Fatalf("finish result = %d, want %d", got, int32(math.MaxInt32))
	}
	summonCancelExportCall5011C0(record)
	if startRecord != record || finishRecord != record || cancelRecord != record {
		t.Fatalf("records = %p/%p/%p, want %p", startRecord, finishRecord, cancelRecord, record)
	}

	if got := summonStartExportCall500DA0(nil); got != math.MinInt32 || startRecord != nil {
		t.Fatalf("nil start = %d/%p", got, startRecord)
	}
	if got := summonFinishExportCall5010D0(nil); got != math.MaxInt32 || finishRecord != nil {
		t.Fatalf("nil finish = %d/%p", got, finishRecord)
	}
	summonCancelExportCall5011C0(nil)
	if cancelRecord != nil {
		t.Fatalf("nil cancel record = %p", cancelRecord)
	}
	runtime.KeepAlive(record)
}
