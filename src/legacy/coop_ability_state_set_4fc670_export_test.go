package legacy

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

type coopAbilityStateSetLegacyServer4FC670 struct {
	Server
	srv *server.Server
}

func (s *coopAbilityStateSetLegacyServer4FC670) S() *server.Server {
	return s.srv
}

func TestCoopAbilityStateSetExport4FC670PreservesSignedInt32(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &coopAbilityStateSetLegacyServer4FC670{srv: srv}
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
		if got := coopAbilityStateSetExportCall4FC670(value); uint32(got) != uint32(value) {
			t.Fatalf("export(%#08x) = %#08x", uint32(value), uint32(got))
		}
		if got := srv.CoopAbilityState4FC670(); uint32(got) != uint32(value) {
			t.Fatalf("state after %#08x = %#08x", uint32(value), uint32(got))
		}
	}
}
