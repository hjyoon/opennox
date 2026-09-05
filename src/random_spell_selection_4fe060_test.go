package opennox

import (
	"math"
	"testing"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestRandomSpellLegacyRoute4FE060UsesLiveRootServer(t *testing.T) {
	oldServer := noxServer
	t.Cleanup(func() { noxServer = oldServer })
	noxServer = &Server{Server: new(server.Server)}

	if legacy.Nox_xxx_unused_4FE060 == nil {
		t.Fatal("legacy random-spell callback is not registered")
	}
	if got := legacy.Nox_xxx_unused_4FE060(math.MaxUint32, 0x80000000); got != 0 {
		t.Fatalf("empty live server result = %d, want 0", got)
	}
}
