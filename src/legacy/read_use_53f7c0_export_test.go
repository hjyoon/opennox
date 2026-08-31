package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type readUseLegacyServer53F7C0 struct {
	Server
	srv *server.Server
}

func (s *readUseLegacyServer53F7C0) S() *server.Server {
	return s.srv
}

func installReadUseLegacyServer53F7C0(t *testing.T, srv *server.Server) {
	t.Helper()
	oldGetServer := GetServer
	GetServer = func() Server {
		return &readUseLegacyServer53F7C0{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
}

func TestReadUseExport53F7C0PreservesNativePointers(t *testing.T) {
	installReadUseLegacyServer53F7C0(t, &server.Server{})
	owner := &server.Object{ObjClass: object.ClassImmobile}
	readable := &server.Object{}

	var pin runtime.Pinner
	pin.Pin(owner)
	pin.Pin(readable)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(owner)) <= math.MaxUint32 || uintptr(unsafe.Pointer(readable)) <= math.MaxUint32 {
			t.Fatalf("test pointers do not exercise native high addresses: owner=%p readable=%p", owner, readable)
		}
	}

	if got := readUseExportCall53F7C0(owner, readable); got != 1 {
		t.Fatalf("export result = %d, want 1", got)
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(readable)
}

func TestReadUseRegistration53F7C0DispatchesNativeImplementation(t *testing.T) {
	srv := &server.Server{}
	srv.SetTickRate(20)
	srv.SetFrame(61)
	installReadUseLegacyServer53F7C0(t, srv)

	owner := &server.Object{ObjClass: object.ClassPlayer}
	data := &server.ReadableUseData{TransientReadState: 1}
	readable := &server.Object{}
	readable.UseData.SetPtr(unsafe.Pointer(data))
	use := server.UseFuncPtr{Ptr: Get_nox_xxx_useRead_53F7C0()}.Get()
	if use == nil {
		t.Fatal("registered ReadUse callback is nil")
	}
	if !use(owner, readable) {
		t.Fatal("registered ReadUse result = false, want true")
	}
	if data.TransientReadState != 1 {
		t.Fatalf("cooldown state = %#x, want unchanged 1", data.TransientReadState)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"owner":    uintptr(unsafe.Pointer(owner)),
			"readable": uintptr(unsafe.Pointer(readable)),
			"data":     uintptr(unsafe.Pointer(data)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(readable)
	runtime.KeepAlive(data)
}
