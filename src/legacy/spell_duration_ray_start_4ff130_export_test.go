package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type durationRayStartLegacyServer4FF130 struct {
	Server
	srv *server.Server
}

func (s *durationRayStartLegacyServer4FF130) S() *server.Server {
	return s.srv
}

func TestDurationRayStartExport4FF130PreservesNativeRecordAndObjectPointers(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &durationRayStartLegacyServer4FF130{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	object := new(server.Object)
	record := &server.DurSpell{Spell: 35, Caster16: object, Target48: object}
	var pin runtime.Pinner
	pin.Pin(object)
	pin.Pin(record)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"record": uintptr(unsafe.Pointer(record)),
			"object": uintptr(unsafe.Pointer(object)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}

	// Equal full-width identities return before any server callback. A
	// truncated record pointer faults on the initial Spell read, while a
	// truncated nested pointer would fail the identity branch and reach the
	// intentionally uninitialised server dependencies.
	durationRayStartExportCall4FF130(unsafe.Pointer(record))
	runtime.KeepAlive(object)
	runtime.KeepAlive(record)
}
