package legacy

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type mapEntryStateSetLegacyServer4FC580 struct {
	Server
	srv *server.Server
}

func (s *mapEntryStateSetLegacyServer4FC580) S() *server.Server {
	return s.srv
}

func TestMapEntryStateSetExport4FC580PreservesSignedInt32(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &mapEntryStateSetLegacyServer4FC580{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})

	for _, value := range []int32{
		math.MinInt32,
		-1,
		0,
		1,
		math.MaxInt32,
		-1985229329, // 0x89abcdef
	} {
		if got := mapEntryStateSetExportCall4FC580(value); uint32(got) != uint32(value) {
			t.Fatalf("export(%#08x) = %#08x", uint32(value), uint32(got))
		}
		if got := srv.MapEntryState4FC580(); uint32(got) != uint32(value) {
			t.Fatalf("state after %#08x = %#08x", uint32(value), uint32(got))
		}
	}
}
