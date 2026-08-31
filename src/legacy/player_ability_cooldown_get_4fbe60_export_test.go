package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

type playerAbilityCooldownGetLegacyServer4FBE60 struct {
	Server
	srv *server.Server
}

func (s *playerAbilityCooldownGetLegacyServer4FBE60) S() *server.Server {
	return s.srv
}

func TestPlayerAbilityCooldownGetExport4FBE60PreservesNativePointerAndWidths(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Abils.Init4FB990()

	lookupUnit := &server.Object{NetCode: 0xfedcba98}
	indexedUnit := &server.Object{NetCode: 0x12345678}
	wrongUnit := new(server.Object)
	lookupPlayer := srv.Players.ResetInd(5)
	lookupPlayer.NetCodeVal = lookupUnit.NetCode
	lookupPlayer.PlayerInd = 7
	lookupPlayer.PlayerUnit = wrongUnit
	indexedPlayer := srv.Players.ResetInd(7)
	indexedPlayer.NetCodeVal = indexedUnit.NetCode
	indexedPlayer.PlayerUnit = indexedUnit

	srv.Abils.GetFor(indexedUnit)
	want := int32(math.MinInt32 + 0x1234)
	srv.Abils.SetCooldownForUnit(indexedUnit, server.AbilityTreadLightly, int(want))

	oldGetServer := GetServer
	GetServer = func() Server {
		return &playerAbilityCooldownGetLegacyServer4FBE60{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	var pin runtime.Pinner
	pin.Pin(lookupUnit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(lookupUnit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", lookupUnit)
	}

	if got := playerAbilityCooldownGetExportCall4FBE60(lookupUnit, int32(server.AbilityTreadLightly)); got != want {
		t.Fatalf("export cooldown = %#08x, want %#08x", uint32(got), uint32(want))
	}
	if got := srv.Abils.GetCooldownForUnit(lookupUnit, server.AbilityTreadLightly); got != int(want) {
		t.Fatalf("Go cooldown = %#08x, want %#08x", uint32(got), uint32(want))
	}
	runtime.KeepAlive(lookupUnit)
}
