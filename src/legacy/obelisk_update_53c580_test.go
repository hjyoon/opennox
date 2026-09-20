package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type obeliskUpdateLegacyServer53C580 struct {
	Server
	srv *server.Server
}

func (s *obeliskUpdateLegacyServer53C580) S() *server.Server {
	return s.srv
}

func TestObeliskUpdate53C580RegistrationUsesNativeRecord(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("ObeliskUpdate")
	if !ok || callback == nil || size != unsafe.Sizeof(server.ObeliskUpdateData{}) {
		t.Fatalf("ObeliskUpdate registration = %p/%d/%t", callback, size, ok)
	}

	oldGetServer := GetServer
	GetServer = func() Server {
		return &obeliskUpdateLegacyServer53C580{srv: new(server.Server)}
	}
	t.Cleanup(func() { GetServer = oldGetServer })
	// A missing update record is rejected by the Go callback. The PE32 C
	// implementation would dereference the object's obsolete +748 slot.
	server.CallObjectUpdate(callback, new(server.Object))
}
