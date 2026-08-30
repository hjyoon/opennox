package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type mapFindPlayerStartLegacyServer4F7AB0 struct {
	Server
	srv *server.Server
}

func (s *mapFindPlayerStartLegacyServer4F7AB0) S() *server.Server {
	return s.srv
}

func TestMapFindPlayerStartExport4F7AB0PreservesNativePointers(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &mapFindPlayerStartLegacyServer4F7AB0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	player := new(server.Object)
	output := &types.Pointf{X: 1, Y: 2}
	var pin runtime.Pinner
	pin.Pin(player)
	pin.Pin(output)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if pointer := uintptr(unsafe.Pointer(player)); pointer <= math.MaxUint32 {
			t.Fatalf("player pointer = %#x, want native address above 4 GiB", pointer)
		}
		if pointer := uintptr(unsafe.Pointer(output)); pointer <= math.MaxUint32 {
			t.Fatalf("output pointer = %#x, want native address above 4 GiB", pointer)
		}
	}

	mapFindPlayerStartExportCall4F7AB0(output, player)
	if *output != (types.Pointf{X: 2000, Y: 2000}) {
		t.Fatalf("export output = %+v, want (2000,2000)", *output)
	}

	*output = types.Pointf{X: 3, Y: 4}
	mapFindPlayerStartExportCall4F7AB0(output, nil)
	if *output != (types.Pointf{X: 3, Y: 4}) {
		t.Fatalf("nil-player output = %+v, want untouched (3,4)", *output)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(output)
}
