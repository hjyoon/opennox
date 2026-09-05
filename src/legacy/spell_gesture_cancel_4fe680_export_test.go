package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type spellGestureCancelLegacyServer4FE680 struct {
	Server
	source *server.Object
	radius float32
	calls  int
}

func (s *spellGestureCancelLegacyServer4FE680) SpellGestureCancel4FE680(
	source *server.Object,
	radius float32,
) {
	s.source = source
	s.radius = radius
	s.calls++
}

func TestSpellGestureCancelExport4FE680PreservesNativePointerAndBinary32(t *testing.T) {
	fake := new(spellGestureCancelLegacyServer4FE680)
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	source := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(source)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want native address above 4 GiB", source)
	}

	wantRadius := math.Float32frombits(0xc3960001)
	spellGestureCancelExportCall4FE680(source, wantRadius)
	if fake.calls != 1 || fake.source != source {
		t.Fatalf("export calls/source = %d/%p, want 1/%p", fake.calls, fake.source, source)
	}
	if got, want := math.Float32bits(fake.radius), math.Float32bits(wantRadius); got != want {
		t.Fatalf("export radius bits = %#08x, want %#08x", got, want)
	}
	runtime.KeepAlive(source)
}
