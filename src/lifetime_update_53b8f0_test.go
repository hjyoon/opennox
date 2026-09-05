package opennox

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectCallUpdateRoutesLifetimeUpdateWithNativePointer53B8F0(t *testing.T) {
	oldServer := noxServer
	t.Cleanup(func() { noxServer = oldServer })
	noxServer = &Server{Server: new(server.Server)}
	noxServer.SetFrame(31)

	data := &server.LifetimeUpdateData53B8F0{Duration: 30}
	source := &server.Object{
		Field32:    1,
		Update:     legacy.Get_nox_xxx_updateLifetime_53B8F0(),
		UpdateData: unsafe.Pointer(data),
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(source.CObj()) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}

	asObjectS(source).CallUpdate()
	if source.ObjFlags != 0 {
		t.Fatalf("exact-boundary flags = %#x, want unchanged", source.ObjFlags)
	}
	runtime.KeepAlive(data)
	runtime.KeepAlive(source)
}
