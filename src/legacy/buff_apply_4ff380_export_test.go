package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestBuffApplyExport4FF380PreservesNativePointerAndFixedWidths(t *testing.T) {
	type call struct {
		unit     *server.Object
		buff     int32
		duration int16
		power    int8
	}
	var calls []call
	old := buffApplyExportImpl4FF380
	buffApplyExportImpl4FF380 = func(unit *server.Object, buff int32, duration int16, power int8) {
		calls = append(calls, call{unit: unit, buff: buff, duration: duration, power: power})
	}
	t.Cleanup(func() { buffApplyExportImpl4FF380 = old })

	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	buffApplyExportCall4FF380(unit, math.MinInt32, math.MinInt16, math.MinInt8)
	buffApplyExportCall4FF380(nil, math.MaxInt32, math.MaxInt16, math.MaxInt8)

	want := []call{
		{unit: unit, buff: math.MinInt32, duration: math.MinInt16, power: math.MinInt8},
		{unit: nil, buff: math.MaxInt32, duration: math.MaxInt16, power: math.MaxInt8},
	}
	if len(calls) != len(want) {
		t.Fatalf("export calls = %d, want %d", len(calls), len(want))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("export call %d = %+v, want %+v", i, calls[i], want[i])
		}
	}

	runtime.KeepAlive(unit)
}
