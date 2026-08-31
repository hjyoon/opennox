package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

type playerAbilityCooldownSetLegacyServer4FBEA0 struct {
	Server
	srv *server.Server
}

func (s *playerAbilityCooldownSetLegacyServer4FBEA0) S() *server.Server {
	return s.srv
}

func TestPlayerAbilityCooldownSetExport4FBEA0PreservesNativePointerAndWidths(t *testing.T) {
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

	oldGetServer := GetServer
	GetServer = func() Server {
		return &playerAbilityCooldownSetLegacyServer4FBEA0{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	var pin runtime.Pinner
	pin.Pin(lookupUnit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(lookupUnit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", lookupUnit)
	}

	want := int32(math.MinInt32 + 0x1234)
	if got := playerAbilityCooldownSetExportCall4FBEA0(lookupUnit, int32(server.AbilityTreadLightly), want); got != want {
		t.Fatalf("export return = %#08x, want %#08x", uint32(got), uint32(want))
	}
	if got := srv.Abils.GetCooldownForUnit(lookupUnit, server.AbilityTreadLightly); got != int(want) {
		t.Fatalf("Go cooldown = %#08x, want %#08x", uint32(got), uint32(want))
	}
	if got := srv.Abils.GetCooldownForUnit(indexedUnit, server.AbilityTreadLightly); got != int(want) {
		t.Fatalf("indexed cooldown = %#08x, want %#08x", uint32(got), uint32(want))
	}
	runtime.KeepAlive(lookupUnit)
}

func TestPlayerAbilityCooldownSetExport4FBEA0MissingPlayerReturnsZero(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Abils.Init4FB990()
	unit := &server.Object{NetCode: 0x12345678}

	oldGetServer := GetServer
	GetServer = func() Server {
		return &playerAbilityCooldownSetLegacyServer4FBEA0{srv: srv}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	if got := playerAbilityCooldownSetExportCall4FBEA0(unit, int32(server.AbilityBerserk), 77); got != 0 {
		t.Fatalf("export return = %d, want 0", got)
	}
	runtime.KeepAlive(unit)
}
