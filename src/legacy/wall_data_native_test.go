package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestWallDataExportsPreserveNativePointers(t *testing.T) {
	wall, freeWall := alloc.New(server.Wall{})
	serverData, freeServerData := alloc.Malloc(1)
	clientData, freeClientData := alloc.Malloc(1)
	t.Cleanup(freeWall)
	t.Cleanup(freeServerData)
	t.Cleanup(freeClientData)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"wall":        unsafe.Pointer(wall),
			"server data": serverData,
			"client data": clientData,
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	wall.Data = serverData
	wall.ClientData = clientData
	if got := wallServerDataExportCall(unsafe.Pointer(wall)); got != serverData {
		t.Fatalf("server data CGo result = %p, want %p", got, serverData)
	}
	if got := wallClientDataExportCall(unsafe.Pointer(wall)); got != clientData {
		t.Fatalf("client data CGo result = %p, want %p", got, clientData)
	}
	if got := wallServerDataExportCall(nil); got != nil {
		t.Fatalf("nil-wall server data CGo result = %p, want nil", got)
	}
	if got := wallClientDataExportCall(nil); got != nil {
		t.Fatalf("nil-wall client data CGo result = %p, want nil", got)
	}
	runtime.KeepAlive(wall)
	runtime.KeepAlive(serverData)
	runtime.KeepAlive(clientData)
}
