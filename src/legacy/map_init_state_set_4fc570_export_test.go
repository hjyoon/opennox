package legacy

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type mapInitStateSetLegacyServer4FC570 struct {
	Server
	srv *server.Server
}

func (s *mapInitStateSetLegacyServer4FC570) S() *server.Server {
	return s.srv
}

func TestMapInitStateSetExport4FC570PreservesSignedInt32(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &mapInitStateSetLegacyServer4FC570{srv: srv}
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
		if got := mapInitStateSetExportCall4FC570(value); uint32(got) != uint32(value) {
			t.Fatalf("export(%#08x) = %#08x", uint32(value), uint32(got))
		}
		if got := srv.MapInitState4FC570(); uint32(got) != uint32(value) {
			t.Fatalf("state after %#08x = %#08x", uint32(value), uint32(got))
		}
	}
}
