package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type playerPhonemeBroadcastLegacyServer4FC960 struct {
	Server
	source    *server.Object
	phoneme   int8
	result    int32
	callCount int
}

func (s *playerPhonemeBroadcastLegacyServer4FC960) PlayerPhonemeBroadcast4FC960(
	source *server.Object,
	phoneme int8,
) int32 {
	s.source = source
	s.phoneme = phoneme
	s.callCount++
	return s.result
}

func TestPlayerPhonemeBroadcastExport4FC960PreservesNativePointerAndScalars(t *testing.T) {
	fake := &playerPhonemeBroadcastLegacyServer4FC960{result: math.MinInt32 + 0x4321}
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

	if got := playerPhonemeBroadcastExportCall4FC960(source, -128); got != fake.result {
		t.Fatalf("export result = %d, want %d", got, fake.result)
	}
	if fake.source != source || fake.phoneme != -128 || fake.callCount != 1 {
		t.Fatalf("export call = %p/%d/%d, want %p/-128/1",
			fake.source, fake.phoneme, fake.callCount, source)
	}
	runtime.KeepAlive(source)
}
