package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type chatLegacyServer528AC0 struct {
	Server
	srv *server.Server
}

func (s *chatLegacyServer528AC0) S() *server.Server { return s.srv }

func TestNetSendChat528AC0PreservesNativeObject(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	obj := &server.Object{NetCode: 0x1234}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= 0xffffffff {
		t.Skip("allocator did not provide a high-address object")
	}
	// No players are registered. The call still builds the complete packet and
	// reaches the native-width player iteration without entering PE32 C code.
	Nox_xxx_netSendChat_528AC0(obj, "scripted chat", 10)
}
